// Package node implements the core node system for Kora Engine
package node

import (
	"image"
	"image/color"
	"math"

	kmath "github.com/ElioNeto/kora/core/math"
	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// Shared gradient textures (lazily initialised per type)
// ---------------------------------------------------------------------------

var (
	// pointLightGradient is a precomputed radial gradient texture used for
	// rendering point lights. It is a gradientSize×gradientSize image with
	// white at centre fading to transparent at the edges.
	pointLightGradient *ebiten.Image

	// dirLightFill is a 1×1 white image used as a uniform fill for
	// directional lights.
	dirLightFill *ebiten.Image

	// shadowOverlay is a 1×1 black image used for drawing shadow overlays
	// on the light map.
	shadowOverlay *ebiten.Image
)

// gradientSize is the resolution of the precomputed gradient textures.
// Higher values produce smoother falloffs at the cost of slightly more
// texture memory.
const gradientSize = 64

// ---------------------------------------------------------------------------
// Gradient initialisers
// ---------------------------------------------------------------------------

// ensurePointLightGradient creates the shared radial gradient texture on
// first use.
func ensurePointLightGradient() *ebiten.Image {
	if pointLightGradient != nil {
		return pointLightGradient
	}

	img := image.NewRGBA(image.Rect(0, 0, gradientSize, gradientSize))
	cx, cy := float64(gradientSize)/2, float64(gradientSize)/2
	maxDist := cx

	for y := 0; y < gradientSize; y++ {
		for x := 0; x < gradientSize; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			a := 1.0 - dist/maxDist
			if a < 0 {
				a = 0
			}
			alpha := uint8(a * 255)
			offset := y*img.Stride + x*4
			img.Pix[offset+0] = 255 // R
			img.Pix[offset+1] = 255 // G
			img.Pix[offset+2] = 255 // B
			img.Pix[offset+3] = alpha
		}
	}

	pointLightGradient = ebiten.NewImageFromImage(img)
	return pointLightGradient
}

// ensureDirLightFill returns a 1×1 white pixel image used for rendering
// directional lights as a uniform fill across the viewport.
func ensureDirLightFill() *ebiten.Image {
	if dirLightFill != nil {
		return dirLightFill
	}
	dirLightFill = ebiten.NewImage(1, 1)
	dirLightFill.Fill(color.White)
	return dirLightFill
}

// ensureShadowOverlay returns a 1×1 black pixel image for shadow rendering.
func ensureShadowOverlay() *ebiten.Image {
	if shadowOverlay != nil {
		return shadowOverlay
	}
	shadowOverlay = ebiten.NewImage(1, 1)
	shadowOverlay.Fill(color.Black)
	return shadowOverlay
}

// ---------------------------------------------------------------------------
// PointLight2D
// ---------------------------------------------------------------------------

// PointLight2D is a node that emits light in all directions from its
// world position with radial falloff.
type PointLight2D struct {
	*Node2D

	// Energy is the light intensity multiplier (0–1).
	Energy float64

	// Color is the light tint (RGBA components in range 0–1).
	Color struct{ R, G, B, A float32 }

	// Range is the maximum radius of the light in world units.
	Range float64

	// Attenuation controls the light falloff curve.
	//   1.0 = linear falloff
	//   2.0 = quadratic falloff (more realistic)
	Attenuation float64

	// ShadowsEnabled controls whether this light casts shadows.
	ShadowsEnabled bool

	// ShadowColor is the tint of the shadow cast by this light
	// (RGBA components in range 0–1).
	ShadowColor struct{ R, G, B, A float32 }

	// Enabled controls whether this light is active.
	Enabled bool
}

// NewPointLight2D creates a new PointLight2D node with default values.
func NewPointLight2D(name string) *PointLight2D {
	return &PointLight2D{
		Node2D:          NewNode2D(name, 0),
		Energy:          1.0,
		Color:           struct{ R, G, B, A float32 }{R: 1, G: 1, B: 1, A: 1},
		Range:           300.0,
		Attenuation:     1.0,
		ShadowsEnabled:  true,
		ShadowColor:     struct{ R, G, B, A float32 }{R: 0, G: 0, B: 0, A: 0.8},
		Enabled:         true,
	}
}

// SetColor sets the light tint colour (RGBA components in range 0–1).
func (pl *PointLight2D) SetColor(r, g, b, a float32) {
	pl.Color.R = clampLightColor(r)
	pl.Color.G = clampLightColor(g)
	pl.Color.B = clampLightColor(b)
	pl.Color.A = clampLightColor(a)
}

// SetEnergy sets the light intensity multiplier (clamped to 0–1).
func (pl *PointLight2D) SetEnergy(energy float64) {
	if energy < 0 {
		energy = 0
	}
	if energy > 1 {
		energy = 1
	}
	pl.Energy = energy
}

// SetRange sets the maximum radius of the light in world units.
// Negative values are clamped to zero.
func (pl *PointLight2D) SetRange(r float64) {
	if r < 0 {
		r = 0
	}
	pl.Range = r
}

// SetEnabled enables or disables this light.
func (pl *PointLight2D) SetEnabled(enabled bool) {
	pl.Enabled = enabled
}

// IsEnabled returns whether this light is currently enabled.
func (pl *PointLight2D) IsEnabled() bool {
	return pl.Enabled
}

// Draw satisfies the Node interface. PointLight2D does not draw
// itself directly; it is rendered by its parent LightWorld.
func (pl *PointLight2D) Draw(screen *ebiten.Image) {
	// Propagate to children (if any)
	for _, child := range pl.children {
		if child != nil {
			child.Draw(screen)
		}
	}
}

// GetNode2D returns the embedded *Node2D pointer, used by AddChild.
func (pl *PointLight2D) GetNode2D() *Node2D { return pl.Node2D }

// ---------------------------------------------------------------------------
// DirectionalLight2D
// ---------------------------------------------------------------------------

// DirectionalLight2D is a light that shines uniformly from a specific
// direction, like sunlight.
type DirectionalLight2D struct {
	*Node2D

	// Energy is the light intensity multiplier (0–1).
	Energy float64

	// Color is the light tint (RGBA components in range 0–1).
	Color struct{ R, G, B, A float32 }

	// Direction is the direction the light is shining (normalised).
	// The direction points toward the light source (i.e. shadows
	// extend in the opposite direction).
	Direction kmath.Vector2

	// ShadowsEnabled controls whether this light casts shadows.
	ShadowsEnabled bool

	// Enabled controls whether this light is active.
	Enabled bool
}

// NewDirectionalLight2D creates a new DirectionalLight2D node with default
// values. Default direction is upward (0, -1).
func NewDirectionalLight2D(name string) *DirectionalLight2D {
	return &DirectionalLight2D{
		Node2D:          NewNode2D(name, 0),
		Energy:          1.0,
		Color:           struct{ R, G, B, A float32 }{R: 1, G: 1, B: 1, A: 1},
		Direction:       kmath.Vector2{X: 0, Y: -1},
		ShadowsEnabled:  true,
		Enabled:         true,
	}
}

// SetColor sets the light colour (RGBA components in range 0–1).
func (dl *DirectionalLight2D) SetColor(r, g, b, a float32) {
	dl.Color.R = clampLightColor(r)
	dl.Color.G = clampLightColor(g)
	dl.Color.B = clampLightColor(b)
	dl.Color.A = clampLightColor(a)
}

// SetEnergy sets the light intensity multiplier (clamped to 0–1).
func (dl *DirectionalLight2D) SetEnergy(energy float64) {
	if energy < 0 {
		energy = 0
	}
	if energy > 1 {
		energy = 1
	}
	dl.Energy = energy
}

// SetDirection sets the light direction vector. The vector is
// normalised internally.
func (dl *DirectionalLight2D) SetDirection(x, y float32) {
	dl.Direction = kmath.Vector2{X: x, Y: y}.Normalize()
}

// SetEnabled enables or disables this light.
func (dl *DirectionalLight2D) SetEnabled(enabled bool) {
	dl.Enabled = enabled
}

// Draw satisfies the Node interface. DirectionalLight2D does not draw
// itself directly; it is rendered by its parent LightWorld.
func (dl *DirectionalLight2D) Draw(screen *ebiten.Image) {
	for _, child := range dl.children {
		if child != nil {
			child.Draw(screen)
		}
	}
}

// GetNode2D returns the embedded *Node2D pointer, used by AddChild.
func (dl *DirectionalLight2D) GetNode2D() *Node2D { return dl.Node2D }

// ---------------------------------------------------------------------------
// LightOccluder2D
// ---------------------------------------------------------------------------

const (
	// OccluderTypeRectangle is a rectangular occluder.
	OccluderTypeRectangle = 0
	// OccluderTypeCircle is a circular occluder.
	OccluderTypeCircle = 1
)

// LightOccluder2D is a node that blocks light, casting shadows. It can
// be either a rectangle or a circle.
type LightOccluder2D struct {
	*Node2D

	// OccluderType defines the shape of the occluder:
	//   0 = OccluderTypeRectangle
	//   1 = OccluderTypeCircle
	OccluderType int

	// Width and Height define the size of a rectangle occluder (in world
	// units).
	Width  float32
	Height float32

	// Radius defines the size of a circle occluder (in world units).
	Radius float32

	// Enabled controls whether this occluder actively casts shadows.
	Enabled bool
}

// NewLightOccluder2D creates a new LightOccluder2D node with default
// values (rectangle, 32×32 world units).
func NewLightOccluder2D(name string) *LightOccluder2D {
	return &LightOccluder2D{
		Node2D:       NewNode2D(name, 0),
		OccluderType: OccluderTypeRectangle,
		Width:        32,
		Height:       32,
		Radius:       16,
		Enabled:      true,
	}
}

// SetSize sets the dimensions of a rectangle occluder.
func (lo *LightOccluder2D) SetSize(w, h float32) {
	lo.Width = w
	lo.Height = h
}

// SetRadius sets the radius of a circle occluder.
func (lo *LightOccluder2D) SetRadius(r float32) {
	lo.Radius = r
}

// SetEnabled enables or disables this occluder.
func (lo *LightOccluder2D) SetEnabled(enabled bool) {
	lo.Enabled = enabled
}

// Draw satisfies the Node interface. LightOccluder2D does not draw
// itself directly.
func (lo *LightOccluder2D) Draw(screen *ebiten.Image) {
	for _, child := range lo.children {
		if child != nil {
			child.Draw(screen)
		}
	}
}

// GetNode2D returns the embedded *Node2D pointer, used by AddChild.
func (lo *LightOccluder2D) GetNode2D() *Node2D { return lo.Node2D }

// ---------------------------------------------------------------------------
// LightWorld
// ---------------------------------------------------------------------------

// LightWorld is a manager node that collects light sources and occluders
// from its children, renders a light map, and composites it onto the
// scene. It uses a deferred CPU-side approach.
//
// Usage:
//
//	lightWorld := NewLightWorld("LightWorld")
//	lightWorld.SetViewport(360, 640)
//	lightWorld.SetAmbient(0.05, 0.05, 0.1, 1.0, 0.3)
//	scene.AddChild(lightWorld)
//	lightWorld.AddChild(NewPointLight2D("torch"))
type LightWorld struct {
	*Node2D

	// AmbientColor is the minimum light level applied everywhere
	// (RGBA components in range 0–1).
	AmbientColor struct{ R, G, B, A float32 }

	// AmbientEnergy scales the ambient light intensity (0–1).
	AmbientEnergy float64

	// Internal state
	lightMap          *ebiten.Image
	viewportW         float64
	viewportH         float64
	pointLights       []*PointLight2D
	directionalLights []*DirectionalLight2D
	occluders         []*LightOccluder2D
	nodeLookup        map[*Node2D]Node // *Node2D → concrete type
}

// NewLightWorld creates a new LightWorld node with default ambient
// (very dim blue-white) and a viewport of 360×640.
func NewLightWorld(name string) *LightWorld {
	return &LightWorld{
		Node2D:            NewNode2D(name, 0),
		AmbientColor:      struct{ R, G, B, A float32 }{R: 0.05, G: 0.05, B: 0.1, A: 1.0},
		AmbientEnergy:     0.3,
		viewportW:         360,
		viewportH:         640,
		pointLights:       make([]*PointLight2D, 0),
		directionalLights: make([]*DirectionalLight2D, 0),
		occluders:         make([]*LightOccluder2D, 0),
		nodeLookup:        make(map[*Node2D]Node),
	}
}

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// SetViewport sets the viewport dimensions for the light map.
func (lw *LightWorld) SetViewport(w, h float64) {
	lw.viewportW = w
	lw.viewportH = h
}

// GetViewport returns the current viewport dimensions.
func (lw *LightWorld) GetViewport() (float64, float64) {
	return lw.viewportW, lw.viewportH
}

// SetAmbient sets the ambient light colour and energy.
func (lw *LightWorld) SetAmbient(r, g, b, a float32, energy float64) {
	lw.AmbientColor.R = clampLightColor(r)
	lw.AmbientColor.G = clampLightColor(g)
	lw.AmbientColor.B = clampLightColor(b)
	lw.AmbientColor.A = clampLightColor(a)
	if energy < 0 {
		energy = 0
	}
	if energy > 1 {
		energy = 1
	}
	lw.AmbientEnergy = energy
}

// GetLightMap returns the current light map image for advanced
// compositing. The returned image should not be modified directly.
func (lw *LightWorld) GetLightMap() *ebiten.Image {
	return lw.lightMap
}

// ---------------------------------------------------------------------------
// Child management overrides
// ---------------------------------------------------------------------------

// AddChild adds a child node to this LightWorld. If the child is a
// PointLight2D, DirectionalLight2D, or LightOccluder2D, it is also
// registered for internal collection.
func (lw *LightWorld) AddChild(child Node) {
	if child == nil {
		return
	}
	// Register the concrete type in the lookup map so that
	// refreshChildLists can identify it later.
	if node2d := extractNode2D(child); node2d != nil {
		lw.nodeLookup[node2d] = child
	}
	lw.Node2D.AddChild(child)
}

// RemoveChild removes a child node by name, including from the
// internal lookup table.
func (lw *LightWorld) RemoveChild(name string) {
	// Remove from lookup before calling parent (which clears children)
	for _, child := range lw.children {
		if child != nil && child.GetName() == name {
			delete(lw.nodeLookup, child)
		}
	}
	lw.Node2D.RemoveChild(name)
}

// RemoveAllChildren removes all children and clears internal state.
func (lw *LightWorld) RemoveAllChildren() {
	lw.Node2D.RemoveAllChildren()
	lw.pointLights = lw.pointLights[:0]
	lw.directionalLights = lw.directionalLights[:0]
	lw.occluders = lw.occluders[:0]
	lw.nodeLookup = make(map[*Node2D]Node)
}

// ---------------------------------------------------------------------------
// Accessors for the internal typed lists
// ---------------------------------------------------------------------------

// GetPointLights returns a copy of the internal point light list.
func (lw *LightWorld) GetPointLights() []*PointLight2D {
	result := make([]*PointLight2D, len(lw.pointLights))
	copy(result, lw.pointLights)
	return result
}

// GetDirectionalLights returns a copy of the internal directional light list.
func (lw *LightWorld) GetDirectionalLights() []*DirectionalLight2D {
	result := make([]*DirectionalLight2D, len(lw.directionalLights))
	copy(result, lw.directionalLights)
	return result
}

// GetOccluders returns a copy of the internal occluder list.
func (lw *LightWorld) GetOccluders() []*LightOccluder2D {
	result := make([]*LightOccluder2D, len(lw.occluders))
	copy(result, lw.occluders)
	return result
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// Update scans children for lighting types and propagates to children.
// This refreshes the internal typed lists each frame by walking the
// children slice and looking up the concrete type in the nodeLookup map.
func (lw *LightWorld) Update(dt float64) {
	// Refresh internal lists by scanning children
	lw.refreshChildLists()

	// Propagate to children (scripts, etc.)
	lw.Node2D.Update(dt)
}

// refreshChildLists clears and rebuilds the typed child lists by
// walking the children slice and resolving each child's concrete type
// through the nodeLookup map.
func (lw *LightWorld) refreshChildLists() {
	lw.pointLights = lw.pointLights[:0]
	lw.directionalLights = lw.directionalLights[:0]
	lw.occluders = lw.occluders[:0]

	for _, child := range lw.children {
		if child == nil {
			continue
		}
		if original, ok := lw.nodeLookup[child]; ok {
			switch t := original.(type) {
			case *PointLight2D:
				if t.Enabled {
					lw.pointLights = append(lw.pointLights, t)
				}
			case *DirectionalLight2D:
				if t.Enabled {
					lw.directionalLights = append(lw.directionalLights, t)
				}
			case *LightOccluder2D:
				if t.Enabled {
					lw.occluders = append(lw.occluders, t)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// RenderLightMap
// ---------------------------------------------------------------------------

// RenderLightMap generates the light map for the current frame. It:
//  1. Creates or recreates the light map image sized to the viewport
//  2. Fills with ambient light colour (scaled by ambient energy)
//  3. Draws each enabled PointLight2D as a radial gradient
//  4. Draws each enabled DirectionalLight2D as a directional gradient
//  5. Applies shadow overlays for occluders
//
// Call this method once per frame before drawing the scene or after
// drawing opaque objects, depending on your compositing strategy.
func (lw *LightWorld) RenderLightMap() {
	if lw.viewportW <= 0 || lw.viewportH <= 0 {
		return
	}

	// 1. Create or recreate light map at correct size
	w := int(lw.viewportW)
	h := int(lw.viewportH)
	if lw.lightMap == nil || lw.lightMap.Bounds().Dx() != w || lw.lightMap.Bounds().Dy() != h {
		lw.lightMap = ebiten.NewImage(w, h)
	}

	// 2. Fill with ambient light
	ambientR := float64(lw.AmbientColor.R) * lw.AmbientEnergy
	ambientG := float64(lw.AmbientColor.G) * lw.AmbientEnergy
	ambientB := float64(lw.AmbientColor.B) * lw.AmbientEnergy
	ambientA := float64(lw.AmbientColor.A)

	ambientColor := color.RGBA{
		R: uint8(clampFloatToByte(ambientR * 255)),
		G: uint8(clampFloatToByte(ambientG * 255)),
		B: uint8(clampFloatToByte(ambientB * 255)),
		A: uint8(clampFloatToByte(ambientA * 255)),
	}
	lw.lightMap.Fill(ambientColor)

	// 3. Draw point lights (additive blending)
	for _, pl := range lw.pointLights {
		if !pl.Enabled || pl.Energy <= 0 || pl.Range <= 0 {
			continue
		}
		lw.drawPointLight(lw.lightMap, pl)
	}

	// 4. Draw directional lights
	for _, dl := range lw.directionalLights {
		if !dl.Enabled || dl.Energy <= 0 {
			continue
		}
		lw.drawDirectionalLight(lw.lightMap, dl)
	}

	// 5. Apply shadows for point lights
	for _, pl := range lw.pointLights {
		if !pl.Enabled || !pl.ShadowsEnabled {
			continue
		}
		for _, occluder := range lw.occluders {
			if !occluder.Enabled {
				continue
			}
			lw.renderPointLightShadow(lw.lightMap, pl, occluder)
		}
	}

	// 6. Apply shadows for directional lights
	for _, dl := range lw.directionalLights {
		if !dl.Enabled || !dl.ShadowsEnabled {
			continue
		}
		for _, occluder := range lw.occluders {
			if !occluder.Enabled {
				continue
			}
			lw.renderDirectionalShadow(lw.lightMap, dl, occluder)
		}
	}
}

// ---------------------------------------------------------------------------
// Light drawing helpers
// ---------------------------------------------------------------------------

// drawPointLight renders a single point light onto the target image.
func (lw *LightWorld) drawPointLight(target *ebiten.Image, pl *PointLight2D) {
	gradient := ensurePointLightGradient()

	// Convert light world position to light map (screen) position.
	lwPos := lw.GetWorldPosition()
	lightPos := pl.GetWorldPosition()

	// The light map covers the viewport. The LightWorld's position is
	// treated as the centre of the viewport.
	screenX := lw.viewportW/2 + float64(lightPos.X-lwPos.X)
	screenY := lw.viewportH/2 + float64(lightPos.Y-lwPos.Y)

	// Calculate the scale factor to map the gradient texture to the
	// light's range in world units.
	diameter := pl.Range * 2
	scale := diameter / gradientSize

	// Build the transformation: centre the gradient at the light
	// position, scaled to the light's range.
	var geo ebiten.GeoM
	geo.Translate(-gradientSize/2, -gradientSize/2) // centre origin
	geo.Scale(scale, scale)
	geo.Translate(screenX, screenY)

	// Apply energy and colour.
	colorScale := ebiten.ColorScale{}
	energyFactor := float32(pl.Energy)
	colorScale.SetR(pl.Color.R * energyFactor)
	colorScale.SetG(pl.Color.G * energyFactor)
	colorScale.SetB(pl.Color.B * energyFactor)
	colorScale.SetA(pl.Color.A * energyFactor)

	opts := &ebiten.DrawImageOptions{
		GeoM:          geo,
		ColorScale:    colorScale,
		CompositeMode: ebiten.CompositeModeLighter,
	}

	target.DrawImage(gradient, opts)
}

// drawDirectionalLight renders a single directional light onto the
// target image. It fills the viewport with a uniform colour scaled
// by the light's energy and tint. The direction is used primarily
// for shadow calculations.
func (lw *LightWorld) drawDirectionalLight(target *ebiten.Image, dl *DirectionalLight2D) {
	fill := ensureDirLightFill()

	// Scale the 1×1 fill to cover the entire viewport.
	var geo ebiten.GeoM
	geo.Scale(lw.viewportW, lw.viewportH)

	colorScale := ebiten.ColorScale{}
	energyFactor := float32(dl.Energy)
	colorScale.SetR(dl.Color.R * energyFactor)
	colorScale.SetG(dl.Color.G * energyFactor)
	colorScale.SetB(dl.Color.B * energyFactor)
	colorScale.SetA(dl.Color.A * energyFactor)

	opts := &ebiten.DrawImageOptions{
		GeoM:          geo,
		ColorScale:    colorScale,
		CompositeMode: ebiten.CompositeModeLighter,
	}

	target.DrawImage(fill, opts)
}

// ---------------------------------------------------------------------------
// Shadow rendering helpers
// ---------------------------------------------------------------------------

// renderPointLightShadow darkens the light map in areas occluded from a
// point light by an occluder.
func (lw *LightWorld) renderPointLightShadow(target *ebiten.Image, light *PointLight2D, occluder *LightOccluder2D) {
	overlay := ensureShadowOverlay()

	lwPos := lw.GetWorldPosition()
	lightPos := light.GetWorldPosition()
	occluderPos := occluder.GetWorldPosition()

	// Direction from light to occluder
	dx := float64(occluderPos.X - lightPos.X)
	dy := float64(occluderPos.Y - lightPos.Y)
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 1 {
		return // light is inside or too close to the occluder
	}

	// Check if occluder is within light range
	if dist > light.Range {
		return // occluder is outside the light's influence
	}

	// Normalised direction from light to occluder
	nx := dx / dist
	ny := dy / dist

	// Occluder size in world units (approximate for shadows)
	var oW, oH float64
	if occluder.OccluderType == OccluderTypeRectangle {
		oW = float64(occluder.Width)
		oH = float64(occluder.Height)
	} else {
		// Circle: use radius as half-dimensions
		oW = float64(occluder.Radius) * 2
		oH = float64(occluder.Radius) * 2
	}

	// Convert occluder centre to light map coordinates
	occluderScreenX := lw.viewportW/2 + float64(occluderPos.X-lwPos.X)
	occluderScreenY := lw.viewportH/2 + float64(occluderPos.Y-lwPos.Y)

	// Shadow length: how far the shadow extends from the occluder
	// away from the light. The shadow fades as it gets further from
	// the occluder.
	shadowLen := light.Range - dist

	// Draw multiple shadow quads extending behind the occluder.
	// We approximate the shadow as a series of dark rectangles with
	// decreasing opacity.
	segments := 4
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments)
		tNext := float64(i+1) / float64(segments)

		// The shadow expands as it goes further from the occluder
		// (penumbra effect)
		expandFactor := 1.0 + t*0.5

		segStart := t * shadowLen
		segEnd := tNext * shadowLen

		// Quad centre
		cx := occluderScreenX + nx*segStart + nx*(segEnd-segStart)/2
		cy := occluderScreenY + ny*segStart + ny*(segEnd-segStart)/2

		// Quad size: starts at occluder size and expands
		segW := oW * expandFactor
		segH := oH * expandFactor

		// Opacity decreases with distance from occluder
		opacity := 1.0 - tNext
		if opacity < 0 {
			opacity = 0
		}

		// Apply shadow colour
		shadowAlpha := float64(light.ShadowColor.A) * opacity
		if shadowAlpha <= 0 {
			continue
		}

		var geo ebiten.GeoM
		geo.Scale(segW, segH)
		geo.Translate(cx-segW/2, cy-segH/2)

		cs := ebiten.ColorScale{}
		cs.SetR(float32(light.ShadowColor.R))
		cs.SetG(float32(light.ShadowColor.G))
		cs.SetB(float32(light.ShadowColor.B))
		cs.SetA(float32(shadowAlpha))

		opts := &ebiten.DrawImageOptions{
			GeoM:          geo,
			ColorScale:    cs,
			CompositeMode: ebiten.CompositeModeSourceOver,
		}
		target.DrawImage(overlay, opts)
	}
}

// renderDirectionalShadow renders shadow cast by a directional light
// behind an occluder.
func (lw *LightWorld) renderDirectionalShadow(target *ebiten.Image, light *DirectionalLight2D, occluder *LightOccluder2D) {
	overlay := ensureShadowOverlay()

	lwPos := lw.GetWorldPosition()
	occluderPos := occluder.GetWorldPosition()

	// Directional light: shadows extend in the opposite direction of
	// the light (i.e., away from the light source).
	dir := light.Direction
	if dir.X == 0 && dir.Y == 0 {
		dir = kmath.Vector2{X: 0, Y: -1}
	}
	dir = dir.Normalize()

	// Shadows extend opposite to the light direction
	shadowDir := kmath.Vector2{X: -dir.X, Y: -dir.Y}

	// Occluder size
	var oW, oH float64
	if occluder.OccluderType == OccluderTypeRectangle {
		oW = float64(occluder.Width)
		oH = float64(occluder.Height)
	} else {
		oW = float64(occluder.Radius) * 2
		oH = float64(occluder.Radius) * 2
	}

	// Convert occluder centre to light map coordinates
	occluderScreenX := lw.viewportW/2 + float64(occluderPos.X-lwPos.X)
	occluderScreenY := lw.viewportH/2 + float64(occluderPos.Y-lwPos.Y)

	// Shadow extends across the viewport behind the occluder
	diag := math.Sqrt(lw.viewportW*lw.viewportW + lw.viewportH*lw.viewportH)
	shadowLen := diag * 2 // long enough to extend beyond the screen

	// Draw shadow as a stretched dark quad behind the occluder
	nx := float64(shadowDir.X)
	ny := float64(shadowDir.Y)

	cx := occluderScreenX + nx*shadowLen/2
	cy := occluderScreenY + ny*shadowLen/2

	segW := oW
	segH := oH

	shadowAlpha := float64(light.Color.A) * 0.3 // subtle shadow for directional
	if shadowAlpha <= 0 {
		return
	}

	var geo ebiten.GeoM
	geo.Scale(segW, segH)
	geo.Translate(cx-segW/2, cy-segH/2)

	cs := ebiten.ColorScale{}
	cs.SetR(0)
	cs.SetG(0)
	cs.SetB(0)
	cs.SetA(float32(shadowAlpha))

	opts := &ebiten.DrawImageOptions{
		GeoM:          geo,
		ColorScale:    cs,
		CompositeMode: ebiten.CompositeModeSourceOver,
	}
	target.DrawImage(overlay, opts)
}

// ---------------------------------------------------------------------------
// Draw
// ---------------------------------------------------------------------------

// Draw renders the light map onto the screen using additive blending.
// This should be called after the scene has been rendered so that the
// light map adds light to the existing pixels.
//
// The method skips rendering if the light map is nil or the node is
// not visible/alive.
func (lw *LightWorld) Draw(screen *ebiten.Image) {
	if !lw.visible || !lw.alive {
		return
	}
	if lw.lightMap == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{
		CompositeMode: ebiten.CompositeModeLighter,
	}
	screen.DrawImage(lw.lightMap, opts)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// clampLightColor clamps a colour component to [0, 1].
func clampLightColor(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// clampFloatToByte clamps a float64 to [0, 255] for conversion to uint8.
func clampFloatToByte(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var (
	_ Node = (*PointLight2D)(nil)
	_ Node = (*DirectionalLight2D)(nil)
	_ Node = (*LightOccluder2D)(nil)
	_ Node = (*LightWorld)(nil)
)
