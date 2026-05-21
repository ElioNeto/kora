// Package asset provides a complete asset management system with async loading,
// reference counting, preloading, and lifecycle management.
//
// Architecture:
//
//	AssetManager — owns all loaded assets, provides Load/Unload/Preload
//	AssetRef     — reference-counted wrapper around a loaded asset
//	AssetType    — enum categorizing asset kinds (texture, audio, shader, …)
//
// Default loaders are registered for Texture, Audio, Shader, Font, and Prefab
// types. Custom loaders can be added via RegisterLoader.
package asset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/audio"
	"github.com/ElioNeto/kora/core/render"
	"github.com/ElioNeto/kora/core/scene"
)

// ---------------------------------------------------------------------------
// AssetType
// ---------------------------------------------------------------------------

// AssetType categorises the kind of asset.
type AssetType int

const (
	AssetTexture AssetType = iota
	AssetAudio
	AssetShader
	AssetFont
	AssetPrefab
	AssetScene
)

// ---------------------------------------------------------------------------
// AssetRef
// ---------------------------------------------------------------------------

// AssetRef is a reference-counted wrapper around a loaded asset.
// The Data field holds the decoded asset (e.g. *ebiten.Image, *audio.Sound).
// RefCount is updated atomically; the caller should not mutate it directly.
type AssetRef struct {
	Type     AssetType
	Name     string
	Data     interface{} // *ebiten.Image, *audio.Sound, *ebiten.Shader, *BitmapFont, *NodeEntity
	RefCount int32
}

// ---------------------------------------------------------------------------
// AssetStats
// ---------------------------------------------------------------------------

// AssetStats holds loading statistics for the AssetManager.
type AssetStats struct {
	Loaded  int
	Pending int
	Errors  int
	Memory  int64 // approximate bytes (sum of image dimensions * 4 for textures)
}

// ---------------------------------------------------------------------------
// AssetManager
// ---------------------------------------------------------------------------

// AssetManager manages the lifecycle of all game assets.
// It is safe for concurrent use via an internal RWMutex.
type AssetManager struct {
	assets     map[string]*AssetRef
	loaders    map[AssetType]func(path string) (interface{}, error)
	mu         sync.RWMutex
	baseDir    string
	loadErrors int64 // cumulative count of failed Load calls (atomic)
	pending    int64 // number of in-flight async loads (atomic)
}

// NewAssetManager creates an AssetManager rooted at baseDir.
// Base directory paths are prepended to relative asset paths when loading.
// Default loaders are registered for all standard asset types.
func NewAssetManager(baseDir string) *AssetManager {
	am := &AssetManager{
		assets:  make(map[string]*AssetRef),
		loaders: make(map[AssetType]func(path string) (interface{}, error)),
		baseDir: baseDir,
	}

	// Register default loaders
	am.loaders[AssetTexture] = am.defaultTextureLoader
	am.loaders[AssetAudio] = am.defaultAudioLoader
	am.loaders[AssetShader] = am.defaultShaderLoader
	am.loaders[AssetFont] = am.defaultFontLoader
	am.loaders[AssetPrefab] = am.defaultPrefabLoader

	return am
}

// RegisterLoader registers a custom loader function for the given asset type.
// The loader receives the asset path and must return the decoded asset or an error.
// Any previously registered loader for the same type is replaced.
func (am *AssetManager) RegisterLoader(assetType AssetType, loader func(path string) (interface{}, error)) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.loaders[assetType] = loader
}

// ---------------------------------------------------------------------------
// Path resolution
// ---------------------------------------------------------------------------

// resolvePath joins the asset path with baseDir if the path is relative.
func (am *AssetManager) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(am.baseDir, path)
}

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

// Load loads an asset identified by its type and path.
// If the asset is already loaded, the existing reference is returned and its
// reference count is incremented. Otherwise the registered loader for the
// asset type is called, the decoded asset is cached, and a new AssetRef is
// returned.
func (am *AssetManager) Load(assetType AssetType, path string) (*AssetRef, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Return cached asset if already loaded.
	if ref, ok := am.assets[path]; ok {
		atomic.AddInt32(&ref.RefCount, 1)
		return ref, nil
	}

	// Look up the loader.
	loader, ok := am.loaders[assetType]
	if !ok {
		atomic.AddInt64(&am.loadErrors, 1)
		return nil, fmt.Errorf("asset: no loader registered for type %d", assetType)
	}

	// Load the asset.
	data, err := loader(path)
	if err != nil {
		atomic.AddInt64(&am.loadErrors, 1)
		return nil, fmt.Errorf("asset: load %s: %w", path, err)
	}

	// Create and cache the reference.
	ref := &AssetRef{
		Type:     assetType,
		Name:     path,
		Data:     data,
		RefCount: 1,
	}
	am.assets[path] = ref
	return ref, nil
}

// LoadAsync loads an asset asynchronously in a background goroutine.
// The callback is invoked on completion with the resulting reference or error.
// The callback is called exactly once and may be called from a different
// goroutine.
func (am *AssetManager) LoadAsync(assetType AssetType, path string, callback func(*AssetRef, error)) {
	atomic.AddInt64(&am.pending, 1)
	go func() {
		ref, err := am.Load(assetType, path)
		atomic.AddInt64(&am.pending, -1)
		callback(ref, err)
	}()
}

// ---------------------------------------------------------------------------
// Unload
// ---------------------------------------------------------------------------

// Unload decrements the reference count of the given asset ref.
// When the reference count reaches zero the underlying resource is released
// and the asset is removed from the cache.
func (am *AssetManager) Unload(ref *AssetRef) {
	if ref == nil {
		return
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	count := atomic.AddInt32(&ref.RefCount, -1)
	if count > 0 {
		return
	}

	// Remove from cache.
	delete(am.assets, ref.Name)

	// Release underlying resources.
	switch ref.Type {
	case AssetTexture:
		if _, ok := ref.Data.(*ebiten.Image); ok {
			render.RemoveTexture(ref.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

// Get returns a previously loaded asset by path, or nil if the asset has not
// been loaded (or has been freed). This does NOT increment the reference count.
func (am *AssetManager) Get(path string) *AssetRef {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.assets[path]
}

// ---------------------------------------------------------------------------
// Preload
// ---------------------------------------------------------------------------

// Preload loads multiple assets in bulk, grouped by asset type.
// It calls Load for each path and collects any errors that occur.
func (am *AssetManager) Preload(assets map[AssetType][]string) []error {
	var errs []error
	for assetType, paths := range assets {
		for _, path := range paths {
			if _, err := am.Load(assetType, path); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// ---------------------------------------------------------------------------
// SceneAssets
// ---------------------------------------------------------------------------

// SceneAssets preloads all assets referenced by a scene file.
// It parses the scene JSON for entity assetId fields and loads each unique
// assetId as a texture.
func (am *AssetManager) SceneAssets(scenePath string) error {
	data, err := os.ReadFile(am.resolvePath(scenePath))
	if err != nil {
		return fmt.Errorf("asset: read scene %s: %w", scenePath, err)
	}

	var scene struct {
		ParentScene string `json:"parentScene,omitempty"`
		Entities    []struct {
			AssetID string `json:"assetId"`
		} `json:"entities"`
	}
	if err := json.Unmarshal(data, &scene); err != nil {
		return fmt.Errorf("asset: parse scene %s: %w", scenePath, err)
	}

	// Load parent scene first if present.
	if scene.ParentScene != "" {
		parentPath := filepath.Join(filepath.Dir(scenePath), scene.ParentScene)
		if err := am.SceneAssets(parentPath); err != nil {
			return err
		}
	}

	// Load each unique assetId as a texture.
	seen := make(map[string]bool)
	for _, e := range scene.Entities {
		if e.AssetID == "" || seen[e.AssetID] {
			continue
		}
		seen[e.AssetID] = true
		if _, err := am.Load(AssetTexture, e.AssetID); err != nil {
			return fmt.Errorf("asset: scene asset %s: %w", e.AssetID, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// ReleaseUnused
// ---------------------------------------------------------------------------

// ReleaseUnused unloads all assets whose reference count has reached zero.
// Returns the number of assets that were freed.
func (am *AssetManager) ReleaseUnused() int {
	am.mu.Lock()
	defer am.mu.Unlock()

	var freed int
	for key, ref := range am.assets {
		if atomic.LoadInt32(&ref.RefCount) <= 0 {
			// Release underlying resources.
			switch ref.Type {
			case AssetTexture:
				if _, ok := ref.Data.(*ebiten.Image); ok {
					render.RemoveTexture(ref.Name)
				}
			}
			delete(am.assets, key)
			freed++
		}
	}
	return freed
}

// ---------------------------------------------------------------------------
// Count / Stats
// ---------------------------------------------------------------------------

// Count returns the total number of currently loaded assets.
func (am *AssetManager) Count() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.assets)
}

// Stats returns loading statistics including the number of loaded assets,
// pending async loads, cumulative errors, and approximate memory usage.
func (am *AssetManager) Stats() AssetStats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var mem int64
	for _, ref := range am.assets {
		mem += estimateMemory(ref)
	}

	return AssetStats{
		Loaded:  len(am.assets),
		Pending: int(atomic.LoadInt64(&am.pending)),
		Errors:  int(atomic.LoadInt64(&am.loadErrors)),
		Memory:  mem,
	}
}

// estimateMemory returns an approximate byte count for the asset.
func estimateMemory(ref *AssetRef) int64 {
	switch ref.Type {
	case AssetTexture:
		if img, ok := ref.Data.(*ebiten.Image); ok {
			bounds := img.Bounds()
			return int64(bounds.Dx()) * int64(bounds.Dy()) * 4
		}
	case AssetAudio:
		// For audio we cannot easily get the byte size without importing audio.
		// Return a conservative estimate.
		return 0
	}
	return 0
}

// ---------------------------------------------------------------------------
// Default loaders
// ---------------------------------------------------------------------------

// defaultTextureLoader loads a texture via the render package.
func (am *AssetManager) defaultTextureLoader(path string) (interface{}, error) {
	resolved := am.resolvePath(path)
	return render.LoadTexture(resolved)
}

// defaultAudioLoader loads an audio file by detecting the format from the
// file extension (.ogg, .wav, .mp3).
func (am *AssetManager) defaultAudioLoader(path string) (interface{}, error) {
	resolved := am.resolvePath(path)
	lower := strings.ToLower(resolved)
	switch {
	case strings.HasSuffix(lower, ".ogg"):
		return audio.LoadOGG(resolved)
	case strings.HasSuffix(lower, ".wav"):
		return audio.LoadWAV(resolved)
	case strings.HasSuffix(lower, ".mp3"):
		return audio.LoadMP3(resolved)
	default:
		return nil, fmt.Errorf("asset: unsupported audio format: %s", path)
	}
}

// defaultShaderLoader compiles a Kage shader from a file.
func (am *AssetManager) defaultShaderLoader(path string) (interface{}, error) {
	resolved := am.resolvePath(path)
	source, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	return ebiten.NewShader(source)
}

// fontConfig is the JSON structure for loading a BitmapFont.
type fontConfig struct {
	AtlasPath string `json:"atlasPath"`
	FirstChar rune   `json:"firstChar"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
	CharW     int    `json:"charW"`
	CharH     int    `json:"charH"`
}

// defaultFontLoader loads a bitmap font from a JSON config file.
// The atlas image referenced in the config is loaded separately via the
// render texture cache.
func (am *AssetManager) defaultFontLoader(path string) (interface{}, error) {
	resolved := am.resolvePath(path)
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	var cfg fontConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("asset: parse font config %s: %w", path, err)
	}

	// Load the atlas image if not already cached.
	atlas := render.GetTexture(cfg.AtlasPath)
	if atlas == nil {
		atlasPath := cfg.AtlasPath
		if !filepath.IsAbs(atlasPath) {
			atlasPath = filepath.Join(am.baseDir, atlasPath)
		}
		atlas, err = render.LoadTexture(atlasPath)
		if err != nil {
			return nil, fmt.Errorf("asset: load font atlas %s: %w", cfg.AtlasPath, err)
		}
	}

	return render.NewBitmapFont(atlas, cfg.FirstChar, cfg.Cols, cfg.Rows, cfg.CharW, cfg.CharH), nil
}

// defaultPrefabLoader loads a scene file as a prefab (NodeEntity).
func (am *AssetManager) defaultPrefabLoader(path string) (interface{}, error) {
	resolved := am.resolvePath(path)
	return scene.LoadSceneEntity(resolved)
}
