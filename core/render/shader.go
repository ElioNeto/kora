package render

import (
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// ShaderManager manages compiled shaders with caching.
// Safe for concurrent use via sync.RWMutex.
type ShaderManager struct {
	mu    sync.RWMutex
	cache map[string]*ebiten.Shader
}

// NewShaderManager creates a new shader manager.
func NewShaderManager() *ShaderManager {
	return &ShaderManager{
		cache: make(map[string]*ebiten.Shader),
	}
}

// LoadShader compiles a Kage shader from source and caches it.
// Returns the compiled shader or an error.
func (sm *ShaderManager) LoadShader(name string, source []byte) (*ebiten.Shader, error) {
	shader, err := CompileShader(source)
	if err != nil {
		return nil, err
	}
	sm.mu.Lock()
	sm.cache[name] = shader
	sm.mu.Unlock()
	return shader, nil
}

// LoadShaderFile loads a shader from a file path.
func (sm *ShaderManager) LoadShaderFile(name, filePath string) (*ebiten.Shader, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return sm.LoadShader(name, source)
}

// GetShader returns a cached shader by name, or nil.
func (sm *ShaderManager) GetShader(name string) *ebiten.Shader {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.cache[name]
}

// UnloadShader removes a shader from the cache.
func (sm *ShaderManager) UnloadShader(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.cache, name)
}

// Clear removes all cached shaders.
func (sm *ShaderManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.cache = make(map[string]*ebiten.Shader)
}

// CompileShader is a helper to compile Kage shader source.
func CompileShader(source []byte) (*ebiten.Shader, error) {
	return ebiten.NewShader(source)
}

// DefaultUniforms returns a set of common shader uniforms:
//   - "Time" (float32) — elapsed time in seconds
//   - "Resolution" ([]float32{2}) — screen resolution in pixels
//   - "Mouse" ([]float32{2}) — mouse position in pixels
func DefaultUniforms(time float64, screenW, screenH int, mouseX, mouseY float64) map[string]interface{} {
	return map[string]interface{}{
		"Time":       float32(time),
		"Resolution": []float32{float32(screenW), float32(screenH)},
		"Mouse":      []float32{float32(mouseX), float32(mouseY)},
	}
}
