// Package render handles all 2D drawing operations.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer wraps Ebitengine drawing primitives.
type Renderer struct {
	width  int
	height int
	screen *ebiten.Image
}

// New creates a Renderer with the given logical size.
func New(width, height int) *Renderer {
	return &Renderer{width: width, height: height}
}

// Begin starts a new frame, storing the target screen.
func (r *Renderer) Begin(screen *ebiten.Image) {
	r.screen = screen
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})
}

// End finalises the current frame. Reserved for future post-processing.
func (r *Renderer) End() {}

// DrawSprite draws an *ebiten.Image at position (x, y) with the given scale and rotation (radians).
func (r *Renderer) DrawSprite(img *ebiten.Image, x, y, scaleX, scaleY, rotation float64, alpha float32) {
	if img == nil || r.screen == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	// Pivot at sprite centre.
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleAlpha(alpha)
	r.screen.DrawImage(img, op)
}

// Width returns the logical render width.
func (r *Renderer) Width() int { return r.width }

// Height returns the logical render height.
func (r *Renderer) Height() int { return r.height }
