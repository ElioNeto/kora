// Package render wraps Ebitengine to provide a 2D renderer for the Kora engine.
//
// Architecture:
//
//	Renderer  — owns the camera and draw state; passed to every Drawer each frame
//	Sprite    — a sub-region of a loaded image atlas
//	Camera    — 2D world-to-screen transform (pan + zoom)
//	Tilemap   — grid of tile indices drawn in one batched pass
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Vec2 is a 2D float64 vector used throughout the engine.
type Vec2 struct{ X, Y float64 }

// ----------------------------------------------------------------------------
// Renderer
// ----------------------------------------------------------------------------

// Renderer is the draw context passed to every Drawer each frame.
// It wraps an *ebiten.Image (the screen) and applies the Camera transform.
type Renderer struct {
	screen *ebiten.Image
	Camera Camera
}

// NewRenderer creates a Renderer. Screen is set via SetScreen.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// SetScreen sets the underlying screen image.
func (r *Renderer) SetScreen(screen *ebiten.Image) {
	r.screen = screen
}

// Screen returns the underlying Ebitengine screen image.
func (r *Renderer) Screen() *ebiten.Image { return r.screen }

// ----------------------------------------------------------------------------
// Helper functions (replacing deprecated ebitenutil)
// ----------------------------------------------------------------------------

// DrawRect draws a filled rectangle on the image.
func DrawRect(img *ebiten.Image, x, y, w, h float64, c color.Color) {
	if img == nil {
		return
	}
	// Create a 1x1 pixel image and stretch it
	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(c)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(w, h)
	opts.GeoM.Translate(x, y)

	img.DrawImage(pixel, opts)
}

// DebugTextAt draws debug text at screen coordinates.
func DebugTextAt(img *ebiten.Image, msg string, x, y int) {
	// Placeholder for debug text rendering
	// Full implementation requires a font atlas
	if img == nil {
		return
	}
	_ = msg
	_ = x
	_ = y
}

// Clear fills the screen with c.
func (r *Renderer) Clear(c color.Color) {
	if r.screen != nil {
		r.screen.Fill(c)
	}
}

// DrawSprite draws sp at world position (x, y) with the given options.
// The Camera transform is applied automatically.
func (r *Renderer) DrawSprite(sp *Sprite, x, y float64, opts *SpriteOpts) {
	if sp == nil || sp.image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}

	// Default pivot is top-left; centre if requested.
	px, py := 0.0, 0.0
	if opts != nil && opts.Centered {
		px = float64(sp.Bounds.Dx()) / 2
		py = float64(sp.Bounds.Dy()) / 2
	}
	op.GeoM.Translate(-px, -py)

	// Scale.
	if opts != nil {
		if opts.FlipX {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(sp.Bounds.Dx()), 0)
		}
		if opts.FlipY {
			op.GeoM.Scale(1, -1)
			op.GeoM.Translate(0, float64(sp.Bounds.Dy()))
		}
		if opts.ScaleX != 0 || opts.ScaleY != 0 {
			sx := opts.ScaleX
			sy := opts.ScaleY
			if sx == 0 { sx = 1 }
			if sy == 0 { sy = 1 }
			op.GeoM.Scale(sx, sy)
		}
		if opts.Rotation != 0 {
			op.GeoM.Rotate(opts.Rotation)
		}
		if opts.Alpha != 0 {
			op.ColorScale.ScaleAlpha(opts.Alpha)
		}
	}

	// World → screen via camera.
	screenX, screenY := r.Camera.WorldToScreen(x, y)
	op.GeoM.Translate(screenX, screenY)

	r.screen.DrawImage(sp.image.SubImage(sp.Bounds).(*ebiten.Image), op)
}

// DrawRect draws a filled rectangle in world coordinates.
func (r *Renderer) DrawRect(x, y, w, h float64, c color.Color) {
	sx, sy := r.Camera.WorldToScreen(x, y)
	DrawRect(r.screen, sx, sy, w*r.Camera.Zoom, h*r.Camera.Zoom, c)
}

// DrawDebugText draws a string at screen coordinates (ignores camera).
func (r *Renderer) DrawDebugText(x, y float64, msg string) {
	DebugTextAt(r.screen, msg, int(x), int(y))
}

// ----------------------------------------------------------------------------
// SpriteOpts
// ----------------------------------------------------------------------------

// SpriteOpts controls optional draw parameters.
type SpriteOpts struct {
	Centered bool
	FlipX    bool
	FlipY    bool
	ScaleX   float64
	ScaleY   float64
	Rotation float64 // radians
	Alpha    float32 // 0 = use default (1.0)
}
