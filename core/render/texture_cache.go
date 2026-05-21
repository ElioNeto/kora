package render

import (
    "sync"
    "github.com/hajimehoshi/ebiten/v2"
)

// TextureCache stores loaded *ebiten.Image assets keyed by an assetId string.
// It is a very small helper used by Sprite2D to avoid re‑loading the same image.
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

// ClearCache removes all cached textures.
func ClearCache() {
    cacheMu.Lock()
    defer cacheMu.Unlock()
    for k := range cache {
        delete(cache, k)
    }
}
