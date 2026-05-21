// Package node implements the core node system for Kora Engine
package node

import (
	"github.com/ElioNeto/kora/core/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// LightCompositor
// ---------------------------------------------------------------------------

// LightCompositor combines a scene render with a light map using
// render targets. This replaces the current direct-to-screen approach
// by allowing offscreen compositing of scene and light layers.
//
// Usage:
//
//	lc := NewLightCompositor(360, 640)
//
//	// Each frame:
//	lc.RenderScene(func(screen *ebiten.Image) {
//	    // draw opaque scene objects
//	})
//	lc.RenderLights(func(lightMap *ebiten.Image) {
//	    // draw lights (point lights, directional lights, ambient fill)
//	})
//	lc.Compose(screen)
type LightCompositor struct {
	sceneTarget *render.RenderTarget
	lightTarget *render.RenderTarget
}

// NewLightCompositor creates a LightCompositor with the given viewport
// dimensions. If width or height is 0, the default render target size
// (360×640) is used.
func NewLightCompositor(width, height int) *LightCompositor {
	return &LightCompositor{
		sceneTarget: render.NewRenderTarget(width, height),
		lightTarget: render.NewRenderTarget(width, height),
	}
}

// RenderScene draws all opaque nodes to the scene target.
// The provided drawFn receives the scene target's image and should
// render the scene onto it.
func (lc *LightCompositor) RenderScene(drawFn func(screen *ebiten.Image)) {
	if drawFn == nil {
		return
	}
	lc.sceneTarget.Begin()
	drawFn(lc.sceneTarget.Image())
	lc.sceneTarget.End()
}

// RenderLights draws all lights to the light target.
// The provided drawFn receives the light target's image and should
// render the light map (ambient + lights) onto it.
func (lc *LightCompositor) RenderLights(drawFn func(screen *ebiten.Image)) {
	if drawFn == nil {
		return
	}
	lc.lightTarget.Begin()
	drawFn(lc.lightTarget.Image())
	lc.lightTarget.End()
}

// Compose combines the scene and light targets onto the final screen.
// The scene is drawn first (source-over), then the light map is
// composited with multiply blending so that dark areas of the light
// map darken the scene and bright areas preserve more of the scene.
func (lc *LightCompositor) Compose(screen *ebiten.Image) {
	if screen == nil {
		return
	}

	// Draw the scene onto the screen.
	screen.DrawImage(lc.sceneTarget.Image(), nil)

	// Composite the light map using multiply blending.
	opts := &ebiten.DrawImageOptions{
		CompositeMode: ebiten.CompositeModeMultiply,
	}
	screen.DrawImage(lc.lightTarget.Image(), opts)
}

// SceneTarget returns the underlying scene RenderTarget for advanced
// usage (e.g. post-processing effects on the scene alone).
func (lc *LightCompositor) SceneTarget() *render.RenderTarget {
	return lc.sceneTarget
}

// LightTarget returns the underlying light RenderTarget for advanced
// usage (e.g. applying additional effects to the light map).
func (lc *LightCompositor) LightTarget() *render.RenderTarget {
	return lc.lightTarget
}

// Resize resizes both internal render targets to the given dimensions.
func (lc *LightCompositor) Resize(w, h int) {
	lc.sceneTarget.Resize(w, h)
	lc.lightTarget.Resize(w, h)
}
