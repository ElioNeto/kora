// Package render wraps Ebitengine to provide a 2D renderer for the Kora engine.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// DefaultRenderTargetWidth is the default width used when 0 is passed to
	// NewRenderTarget.
	DefaultRenderTargetWidth = 360

	// DefaultRenderTargetHeight is the default height used when 0 is passed to
	// NewRenderTarget.
	DefaultRenderTargetHeight = 640
)

// ----------------------------------------------------------------------------
// RenderTarget
// ----------------------------------------------------------------------------

// RenderTarget is an offscreen buffer that can be drawn to and then
// composited onto the screen. Useful for post-processing, camera
// stacking, and scene compositing.
type RenderTarget struct {
	width, height int
	image         *ebiten.Image
	clearColor    color.RGBA
	dirty         bool
}

// NewRenderTarget creates an offscreen buffer of the given size.
// If width or height is <= 0, the default screen size (DefaultRenderTargetWidth
// × DefaultRenderTargetHeight) is used.
func NewRenderTarget(width, height int) *RenderTarget {
	if width <= 0 {
		width = DefaultRenderTargetWidth
	}
	if height <= 0 {
		height = DefaultRenderTargetHeight
	}
	return &RenderTarget{
		width:  width,
		height: height,
		image:  ebiten.NewImage(width, height),
	}
}

// Begin clears the target and prepares it for drawing.
// If SetClearColor has been called, the target is filled with that color;
// otherwise it is filled with transparent.
func (rt *RenderTarget) Begin() {
	if rt.image == nil {
		return
	}
	rt.image.Fill(rt.clearColor)
	rt.dirty = true
}

// End returns the rendered image for compositing.
func (rt *RenderTarget) End() *ebiten.Image {
	rt.dirty = false
	return rt.image
}

// Image returns the underlying ebiten image.
func (rt *RenderTarget) Image() *ebiten.Image {
	return rt.image
}

// SetClearColor sets the clear color for Begin().
// By default (before calling SetClearColor) Begin() fills with transparent.
func (rt *RenderTarget) SetClearColor(c color.RGBA) {
	rt.clearColor = c
}

// Resize resizes the render target. Creates a new image if the size has
// changed. If width or height is <= 0, the default is used.
func (rt *RenderTarget) Resize(w, h int) {
	if w <= 0 {
		w = DefaultRenderTargetWidth
	}
	if h <= 0 {
		h = DefaultRenderTargetHeight
	}
	if rt.width == w && rt.height == h && rt.image != nil {
		return
	}
	rt.width = w
	rt.height = h
	rt.image = ebiten.NewImage(w, h)
}

// DrawCompose draws the render target onto the screen with the given
// options (for post-processing, blending, etc.). If opts is nil, the
// default DrawImageOptions (source-over) will be used.
func (rt *RenderTarget) DrawCompose(screen *ebiten.Image, opts *ebiten.DrawImageOptions) {
	if rt.image == nil || screen == nil {
		return
	}
	screen.DrawImage(rt.image, opts)
}

// Dispose frees the underlying image.
func (rt *RenderTarget) Dispose() {
	rt.image = nil
	rt.dirty = false
}
