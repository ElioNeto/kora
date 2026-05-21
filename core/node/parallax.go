package node

import (
	"github.com/ElioNeto/kora/core/math"
	"github.com/hajimehoshi/ebiten/v2"
)

// ParallaxLayer is a child node for parallax backgrounds.
// Each layer can have its own scroll speed and contain child nodes
// (sprites, etc.) that are drawn with parallax offset applied.
type ParallaxLayer struct {
	*Node2D

	// scrollScale controls how fast this layer scrolls relative to camera movement.
	//   1.0 = moves exactly with camera (foreground objects)
	//   0.5 = moves half as fast (midground)
	//   0.0 = stays fixed (sky/background)
	//   negative = moves opposite direction (fog/haze)
	scrollScale math.Vector2

	// texture is a path reference to the texture file for serialization.
	texture string

	// offset is the accumulated parallax offset applied when drawing children.
	offset math.Vector2
}

// NewParallaxLayer creates a new ParallaxLayer node.
func NewParallaxLayer(name string) *ParallaxLayer {
	return &ParallaxLayer{
		Node2D:      NewNode2D(name, 0),
		scrollScale: math.NewVector2(1.0, 1.0),
	}
}

// SetScrollScale sets the scroll multiplier for this layer.
// sx: X-axis scroll multiplier, sy: Y-axis scroll multiplier.
func (pl *ParallaxLayer) SetScrollScale(sx, sy float32) {
	pl.scrollScale.X = sx
	pl.scrollScale.Y = sy
}

// GetScrollScale returns the scroll multiplier for this layer.
func (pl *ParallaxLayer) GetScrollScale() math.Vector2 {
	return pl.scrollScale
}

// SetTexture sets the texture path reference for serialization.
func (pl *ParallaxLayer) SetTexture(path string) {
	pl.texture = path
}

// GetTexture returns the texture path reference.
func (pl *ParallaxLayer) GetTexture() string {
	return pl.texture
}

// GetOffset returns the accumulated parallax offset for this layer.
func (pl *ParallaxLayer) GetOffset() math.Vector2 {
	return pl.offset
}

// Draw renders the layer's children with the parallax offset applied.
// It temporarily adjusts the layer's position so children inherit the offset,
// then restores the original position after drawing.
func (pl *ParallaxLayer) Draw(screen *ebiten.Image) {
	if !pl.visible || !pl.alive {
		return
	}

	// Save original position and apply parallax offset
	origPos := pl.pos
	pl.pos.X = origPos.X + pl.offset.X
	pl.pos.Y = origPos.Y + pl.offset.Y

	// Draw children with the offset applied
	for _, child := range pl.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	// Restore original position
	pl.pos = origPos
}

// ---------------------------------------------------------------------------
// Compile-time interface check
// ---------------------------------------------------------------------------

var _ Node = (*ParallaxLayer)(nil)

// ParallaxBackground is a root node that manages parallax layers.
// It tracks camera movement and distributes scroll offsets to each layer
// based on their individual scroll scales.
type ParallaxBackground struct {
	*Node2D

	// scrollOffset is the total accumulated scroll offset from camera movement.
	scrollOffset math.Vector2

	// previousCameraPos stores the camera position from the previous frame
	// to calculate delta movement each tick.
	previousCameraPos math.Vector2

	// cameraReference is an optional camera node to track for parallax calculations.
	// When set, the background follows this camera's movement.
	cameraReference *Node2D

	// mirroring enables infinite scrolling / tiling support.
	// When true, the background seamlessly tiles by wrapping offsets
	// to prevent floating-point drift.
	mirroring bool

	// layers is a separate list of ParallaxLayer children, maintained alongside
	// the standard children list for type-safe access during Update and Draw.
	layers []*ParallaxLayer
}

// NewParallaxBackground creates a new ParallaxBackground node.
func NewParallaxBackground(name string) *ParallaxBackground {
	return &ParallaxBackground{
		Node2D:            NewNode2D(name, 0),
		scrollOffset:      math.Vector2{},
		previousCameraPos: math.Vector2{},
		mirroring:         false,
		layers:            make([]*ParallaxLayer, 0),
	}
}

// ---------------------------------------------------------------------------
// Camera reference
// ---------------------------------------------------------------------------

// SetCameraReference sets the camera to track for parallax calculations.
// Pass nil to stop tracking. The initial previous position is captured from
// the camera's current world position.
func (pb *ParallaxBackground) SetCameraReference(cam *Node2D) {
	pb.cameraReference = cam
	if cam != nil {
		pb.previousCameraPos = cam.GetWorldPosition()
	} else {
		pb.previousCameraPos = math.Vector2{}
	}
}

// GetCameraReference returns the tracked camera, or nil if none is set.
func (pb *ParallaxBackground) GetCameraReference() *Node2D {
	return pb.cameraReference
}

// ---------------------------------------------------------------------------
// Mirroring
// ---------------------------------------------------------------------------

// SetMirroring enables or disables infinite scrolling / tiling.
// When enabled, layer offsets are wrapped into a tileable range [0, tileSize)
// to prevent floating-point drift over long scrolling distances.
func (pb *ParallaxBackground) SetMirroring(enabled bool) {
	pb.mirroring = enabled
}

// IsMirroring returns whether mirroring is enabled.
func (pb *ParallaxBackground) IsMirroring() bool {
	return pb.mirroring
}

// ---------------------------------------------------------------------------
// Scroll offset
// ---------------------------------------------------------------------------

// GetScrollOffset returns the total accumulated scroll offset.
func (pb *ParallaxBackground) GetScrollOffset() math.Vector2 {
	return pb.scrollOffset
}

// SetScrollOffset sets the scroll offset directly.
// This can be used to initialise the offset to a specific value.
func (pb *ParallaxBackground) SetScrollOffset(offset math.Vector2) {
	pb.scrollOffset = offset
}

// ---------------------------------------------------------------------------
// Layer management
// ---------------------------------------------------------------------------

// AddLayer adds a ParallaxLayer as a child of this background.
// The layer is registered in both the standard children list and the
// internal layers list for type-safe parallax processing.
func (pb *ParallaxBackground) AddLayer(layer *ParallaxLayer) {
	if layer == nil {
		return
	}
	pb.AddChild(layer)
}

// RemoveLayer removes a ParallaxLayer child by name.
func (pb *ParallaxBackground) RemoveLayer(name string) {
	pb.RemoveChild(name)
	for i, l := range pb.layers {
		if l.GetName() == name {
			pb.layers = append(pb.layers[:i], pb.layers[i+1:]...)
			return
		}
	}
}

// GetLayers returns a copy of the internal layers slice.
func (pb *ParallaxBackground) GetLayers() []*ParallaxLayer {
	result := make([]*ParallaxLayer, len(pb.layers))
	copy(result, pb.layers)
	return result
}

// ---------------------------------------------------------------------------
// Update / Draw
// ---------------------------------------------------------------------------

// Update calculates camera movement since the last frame and distributes
// parallax offsets to each layer based on their scroll scales.
//
// For each frame:
//  1. Computes the camera delta = current position - previous position
//  2. Updates the accumulated scrollOffset
//  3. For each layer: layer.offset += cameraDelta * layer.scrollScale
//  4. If mirroring is enabled, wraps offsets to prevent drift
//  5. Propagates Update to all children (scripts, non-layer children, etc.)
func (pb *ParallaxBackground) Update(dt float64) {
	// 1. Calculate camera delta
	var cameraDelta math.Vector2

	if pb.cameraReference != nil {
		currentPos := pb.cameraReference.GetWorldPosition()
		cameraDelta = currentPos.Sub(pb.previousCameraPos)
		pb.previousCameraPos = currentPos
	}

	// 2. Update total scroll offset
	pb.scrollOffset = pb.scrollOffset.Add(cameraDelta)

	// 3-4. Distribute delta to each parallax layer
	for _, layer := range pb.layers {
		layer.offset.X += cameraDelta.X * layer.scrollScale.X
		layer.offset.Y += cameraDelta.Y * layer.scrollScale.Y

		if pb.mirroring {
			layer.offset.X = wrapOffset(layer.offset.X)
			layer.offset.Y = wrapOffset(layer.offset.Y)
		}
	}

	// 5. Propagate Update to all children (scripts, etc.)
	pb.Node2D.Update(dt)
}

// Draw renders all parallax layers (with their offsets applied) and any
// non-layer children using their default draw behaviour.
func (pb *ParallaxBackground) Draw(screen *ebiten.Image) {
	if !pb.visible || !pb.alive {
		return
	}

	// Draw parallax layers first (using their custom Draw that applies offsets)
	for _, layer := range pb.layers {
		layer.Draw(screen)
	}

	// Draw non-layer children
	for _, child := range pb.children {
		if child == nil {
			continue
		}
		// Skip children that are parallax layers (already drawn above)
		if pb.isLayerChild(child) {
			continue
		}
		child.Draw(screen)
	}
}

// isLayerChild checks whether the given *Node2D pointer corresponds to one of
// the registered parallax layers.
func (pb *ParallaxBackground) isLayerChild(child *Node2D) bool {
	for _, layer := range pb.layers {
		if layer.Node2D == child {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Child management overrides
// ---------------------------------------------------------------------------

// AddChild adds a child node to this background.
// If the child is a *ParallaxLayer, it is also registered in the internal
// layers list for parallax processing.
func (pb *ParallaxBackground) AddChild(child Node) {
	// Check if the child is a ParallaxLayer and register it
	if pl, ok := child.(*ParallaxLayer); ok {
		// Avoid duplicates
		for _, existing := range pb.layers {
			if existing == pl {
				pb.Node2D.AddChild(child)
				return
			}
		}
		pb.layers = append(pb.layers, pl)
	}
	pb.Node2D.AddChild(child)
}

// RemoveChild removes a child node by name, including from the internal layers
// list if the child is a ParallaxLayer.
func (pb *ParallaxBackground) RemoveChild(name string) {
	pb.Node2D.RemoveChild(name)
	for i, l := range pb.layers {
		if l.GetName() == name {
			pb.layers = append(pb.layers[:i], pb.layers[i+1:]...)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// tileSize is the assumed tile dimension used for mirroring wrap calculations.
// Offsets are wrapped into the range [0, tileSize) to prevent floating-point
// drift during infinite scrolling.
const tileSize float32 = 4096.0

// wrapOffset wraps a float32 value into a tileable range [0, tileSize).
func wrapOffset(v float32) float32 {
	// Use modulo to wrap into range
	v = float32(int(v) % int(tileSize))
	if v < 0 {
		v += tileSize
	}
	return v
}

// ---------------------------------------------------------------------------
// KScript API (documentation only — wiring is optional)
//
// The following methods would be exposed to KScript once the binding layer
// is implemented:
//
//   parallax.setScrollScale(layerName, sx, sy)
//     Sets the scroll scale for the named ParallaxLayer child.
//
//   parallax.setMirroring(enabled)
//     Enables or disables mirroring/tiling on the ParallaxBackground.
// ---------------------------------------------------------------------------

// Compile-time interface checks
var (
	_ Node = (*ParallaxBackground)(nil)
)
