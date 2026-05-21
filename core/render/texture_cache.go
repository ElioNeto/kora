package render

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// TextureCache stores loaded *ebiten.Image assets keyed by an assetId string.
var (
	cache   = make(map[string]*ebiten.Image)
	cacheMu sync.RWMutex
)

// SetTexture stores an image under the given assetId.
func SetTexture(assetId string, img *ebiten.Image) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache[assetId] = img
}

// GetTexture retrieves an image by assetId. Returns nil if not present.
func GetTexture(assetId string) *ebiten.Image {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cache[assetId]
}

// LoadTexture loads an image file from the given path and caches it under
// that same path as the key. Supported formats: PNG, JPEG.
// Returns the loaded texture or an error.
func LoadTexture(path string) (*ebiten.Image, error) {
	// Check cache first
	if img := GetTexture(path); img != nil {
		return img, nil
	}

	// Open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode image
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	// Convert to ebiten image
	ebitenImg := ebiten.NewImageFromImage(img)

	// Cache it
	SetTexture(path, ebitenImg)

	return ebitenImg, nil
}

// LoadTextureFromImage loads an image.Image and caches it under the given key.
func LoadTextureFromImage(key string, img image.Image) *ebiten.Image {
	ebitenImg := ebiten.NewImageFromImage(img)
	SetTexture(key, ebitenImg)
	return ebitenImg
}

// ClearCache removes all cached textures.
func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	for k := range cache {
		delete(cache, k)
	}
}

// RemoveTexture removes a single texture from the cache.
func RemoveTexture(assetId string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	delete(cache, assetId)
}
