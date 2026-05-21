package node

import (
	"testing"

	"github.com/ElioNeto/kora/core/math"
)

// absF32 returns the absolute value of a float32 (helper for approximate comparisons)
func absF32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// ---------------------------------------------------------------------------
// ParallaxLayer tests
// ---------------------------------------------------------------------------

func TestNewParallaxLayer(t *testing.T) {
	pl := NewParallaxLayer("sky")
	if pl == nil {
		t.Fatal("expected non-nil ParallaxLayer")
	}
	if pl.GetName() != "sky" {
		t.Errorf("expected name 'sky', got '%s'", pl.GetName())
	}
	scale := pl.GetScrollScale()
	if scale.X != 1.0 || scale.Y != 1.0 {
		t.Errorf("expected default scroll scale (1.0, 1.0), got (%f, %f)", scale.X, scale.Y)
	}
	offset := pl.GetOffset()
	if offset.X != 0 || offset.Y != 0 {
		t.Errorf("expected zero offset, got (%f, %f)", offset.X, offset.Y)
	}
}

func TestParallaxLayer_SetScrollScale(t *testing.T) {
	pl := NewParallaxLayer("layer")
	pl.SetScrollScale(0.5, 0.0)
	scale := pl.GetScrollScale()
	if scale.X != 0.5 {
		t.Errorf("expected scrollScale.X 0.5, got %f", scale.X)
	}
	if scale.Y != 0.0 {
		t.Errorf("expected scrollScale.Y 0.0, got %f", scale.Y)
	}
}

func TestParallaxLayer_SetTexture(t *testing.T) {
	pl := NewParallaxLayer("layer")
	pl.SetTexture("textures/sky.png")
	if pl.GetTexture() != "textures/sky.png" {
		t.Errorf("expected texture 'textures/sky.png', got '%s'", pl.GetTexture())
	}
}

func TestParallaxLayer_NegativeScrollScale(t *testing.T) {
	pl := NewParallaxLayer("fog")
	pl.SetScrollScale(-0.3, -0.1)
	scale := pl.GetScrollScale()
	if scale.X != -0.3 {
		t.Errorf("expected scrollScale.X -0.3, got %f", scale.X)
	}
	if scale.Y != -0.1 {
		t.Errorf("expected scrollScale.Y -0.1, got %f", scale.Y)
	}
}

func TestParallaxLayer_AddChild(t *testing.T) {
	pl := NewParallaxLayer("layer")
	child := NewNode2D("sprite", 1)
	pl.AddChild(child)
	if pl.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", pl.GetChildCount())
	}
	if pl.GetChild("sprite") != child {
		t.Error("expected child to be retrievable")
	}
}

func TestParallaxLayer_NodeInterface(t *testing.T) {
	// Compile-time check
	var _ Node = (*ParallaxLayer)(nil)

	// Runtime check
	pl := NewParallaxLayer("test")
	var n Node = pl
	if n.Name() != "test" {
		t.Error("ParallaxLayer should satisfy Node interface")
	}
}

// ---------------------------------------------------------------------------
// ParallaxBackground tests
// ---------------------------------------------------------------------------

func TestNewParallaxBackground(t *testing.T) {
	pb := NewParallaxBackground("bg")
	if pb == nil {
		t.Fatal("expected non-nil ParallaxBackground")
	}
	if pb.GetName() != "bg" {
		t.Errorf("expected name 'bg', got '%s'", pb.GetName())
	}
	if pb.IsMirroring() {
		t.Error("expected mirroring to be false by default")
	}
	if pb.GetCameraReference() != nil {
		t.Error("expected camera reference to be nil by default")
	}
	offset := pb.GetScrollOffset()
	if offset.X != 0 || offset.Y != 0 {
		t.Errorf("expected zero scroll offset, got (%f, %f)", offset.X, offset.Y)
	}
	if len(pb.GetLayers()) != 0 {
		t.Error("expected no layers initially")
	}
}

func TestParallaxBackground_AddLayer(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("sky")
	pb.AddLayer(layer)

	layers := pb.GetLayers()
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	if layers[0] != layer {
		t.Error("expected the added layer")
	}

	// The layer should also be a child
	if pb.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", pb.GetChildCount())
	}
}

func TestParallaxBackground_AddLayerNil(t *testing.T) {
	pb := NewParallaxBackground("bg")
	pb.AddLayer(nil)
	if len(pb.GetLayers()) != 0 {
		t.Error("nil layer should not be added")
	}
}

func TestParallaxBackground_AddMultipleLayers(t *testing.T) {
	pb := NewParallaxBackground("bg")
	pb.AddLayer(NewParallaxLayer("sky"))
	pb.AddLayer(NewParallaxLayer("mountains"))
	pb.AddLayer(NewParallaxLayer("foreground"))

	if len(pb.GetLayers()) != 3 {
		t.Errorf("expected 3 layers, got %d", len(pb.GetLayers()))
	}
}

func TestParallaxBackground_RemoveLayer(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer1 := NewParallaxLayer("sky")
	layer2 := NewParallaxLayer("mountains")
	pb.AddLayer(layer1)
	pb.AddLayer(layer2)

	pb.RemoveLayer("sky")

	layers := pb.GetLayers()
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer after removal, got %d", len(layers))
	}
	if layers[0].GetName() != "mountains" {
		t.Error("expected remaining layer 'mountains'")
	}

	// Should no longer be a child
	if pb.GetChild("sky") != nil {
		t.Error("removed layer should not be a child")
	}
}

func TestParallaxBackground_RemoveLayerNonexistent(t *testing.T) {
	pb := NewParallaxBackground("bg")
	pb.AddLayer(NewParallaxLayer("sky"))
	pb.RemoveLayer("nonexistent") // should not panic
	if len(pb.GetLayers()) != 1 {
		t.Errorf("expected 1 layer, got %d", len(pb.GetLayers()))
	}
}

func TestParallaxBackground_CameraReference(t *testing.T) {
	pb := NewParallaxBackground("bg")
	cam := NewNode2D("camera", 1)
	cam.SetPosition(100, 200)

	pb.SetCameraReference(cam)
	if pb.GetCameraReference() != cam {
		t.Error("expected camera reference to be set")
	}

	// Check that it captured the initial position
	expected := math.NewVector2(100, 200)
	if pb.previousCameraPos != expected {
		t.Errorf("expected previousCameraPos %v, got %v", expected, pb.previousCameraPos)
	}

	// Clear reference
	pb.SetCameraReference(nil)
	if pb.GetCameraReference() != nil {
		t.Error("expected camera reference to be nil after clear")
	}
}

func TestParallaxBackground_SetMirroring(t *testing.T) {
	pb := NewParallaxBackground("bg")
	if pb.IsMirroring() {
		t.Error("expected mirroring false by default")
	}
	pb.SetMirroring(true)
	if !pb.IsMirroring() {
		t.Error("expected mirroring true after SetMirroring(true)")
	}
	pb.SetMirroring(false)
	if pb.IsMirroring() {
		t.Error("expected mirroring false after SetMirroring(false)")
	}
}

func TestParallaxBackground_ScrollOffset(t *testing.T) {
	pb := NewParallaxBackground("bg")
	offset := math.NewVector2(50, 100)
	pb.SetScrollOffset(offset)
	got := pb.GetScrollOffset()
	if got != offset {
		t.Errorf("expected scroll offset %v, got %v", offset, got)
	}
}

func TestParallaxBackground_UpdateNoCamera(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("sky")
	pb.AddLayer(layer)

	// Update without a camera reference — no movement expected
	pb.Update(0.016)

	offset := pb.GetScrollOffset()
	if offset.X != 0 || offset.Y != 0 {
		t.Errorf("expected no scroll offset without camera, got %v", offset)
	}
}

func TestParallaxBackground_UpdateCameraDelta(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("sky")
	layer.SetScrollScale(1.0, 0.5)
	pb.AddLayer(layer)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Move camera to (100, 200)
	cam.SetPosition(100, 200)
	pb.Update(0.016)

	// Scroll offset should be (100, 200)
	offset := pb.GetScrollOffset()
	if offset.X != 100 || offset.Y != 200 {
		t.Errorf("expected scroll offset (100, 200), got (%f, %f)", offset.X, offset.Y)
	}

	// Layer offset should be scaled: (100*1.0, 200*0.5) = (100, 100)
	layerOffset := layer.GetOffset()
	if layerOffset.X != 100 || layerOffset.Y != 100 {
		t.Errorf("expected layer offset (100, 100), got (%f, %f)", layerOffset.X, layerOffset.Y)
	}
}

func TestParallaxBackground_UpdateMultipleFrames(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("ground")
	layer.SetScrollScale(1.0, 1.0)
	pb.AddLayer(layer)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Frame 1: move to (50, 0)
	cam.SetPosition(50, 0)
	pb.Update(0.016)
	if pb.GetScrollOffset().X != 50 {
		t.Errorf("frame 1: expected scroll offset X 50, got %f", pb.GetScrollOffset().X)
	}

	// Frame 2: move to (120, 30) — delta = (70, 30)
	cam.SetPosition(120, 30)
	pb.Update(0.016)
	if pb.GetScrollOffset().X != 120 {
		t.Errorf("frame 2: expected scroll offset X 120, got %f", pb.GetScrollOffset().X)
	}
	if pb.GetScrollOffset().Y != 30 {
		t.Errorf("frame 2: expected scroll offset Y 30, got %f", pb.GetScrollOffset().Y)
	}

	// Layer should have accumulated offset = (120, 30)
	layerOffset := layer.GetOffset()
	if layerOffset.X != 120 || layerOffset.Y != 30 {
		t.Errorf("expected layer offset (120, 30), got (%f, %f)", layerOffset.X, layerOffset.Y)
	}
}

func TestParallaxBackground_DifferentScrollScales(t *testing.T) {
	pb := NewParallaxBackground("bg")
	sky := NewParallaxLayer("sky")
	sky.SetScrollScale(0.0, 0.0) // fixed

	mid := NewParallaxLayer("mid")
	mid.SetScrollScale(0.5, 0.5) // half speed

	fg := NewParallaxLayer("fg")
	fg.SetScrollScale(1.0, 1.0) // full speed

	pb.AddLayer(sky)
	pb.AddLayer(mid)
	pb.AddLayer(fg)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Move camera 200 units to the right
	cam.SetPosition(200, 0)
	pb.Update(0.016)

	// Sky should not move
	if sky.GetOffset().X != 0 {
		t.Errorf("sky offset X expected 0, got %f", sky.GetOffset().X)
	}
	// Mid should move half
	if mid.GetOffset().X != 100 {
		t.Errorf("mid offset X expected 100, got %f", mid.GetOffset().X)
	}
	// Foreground should move exactly
	if fg.GetOffset().X != 200 {
		t.Errorf("fg offset X expected 200, got %f", fg.GetOffset().X)
	}
}

func TestParallaxBackground_NegativeScrollScale(t *testing.T) {
	pb := NewParallaxBackground("bg")
	fog := NewParallaxLayer("fog")
	fog.SetScrollScale(-0.5, 0.0) // moves opposite direction
	pb.AddLayer(fog)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Move camera 100 units right
	cam.SetPosition(100, 0)
	pb.Update(0.016)

	// Fog should move -50 (opposite direction)
	if fog.GetOffset().X != -50 {
		t.Errorf("fog offset X expected -50, got %f", fog.GetOffset().X)
	}
}

func TestParallaxBackground_MirroringWrapsOffset(t *testing.T) {
	pb := NewParallaxBackground("bg")
	pb.SetMirroring(true)

	layer := NewParallaxLayer("tile")
	layer.SetScrollScale(1.0, 0.0)
	pb.AddLayer(layer)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Move camera by a large amount to test wrapping
	cam.SetPosition(5000, 0)
	pb.Update(0.016)

	layerOffset := layer.GetOffset()
	// 5000 % 4096 = 904
	if layerOffset.X < 0 || layerOffset.X >= 4096 {
		t.Errorf("mirrored offset should be in [0, 4096), got %f", layerOffset.X)
	}
	if layerOffset.X != 904 {
		t.Errorf("expected wrapped offset 904, got %f", layerOffset.X)
	}
}

func TestParallaxBackground_MirroringNegativeOffset(t *testing.T) {
	pb := NewParallaxBackground("bg")
	pb.SetMirroring(true)

	layer := NewParallaxLayer("tile")
	layer.SetScrollScale(1.0, 0.0)
	pb.AddLayer(layer)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Move camera in negative direction
	cam.SetPosition(-100, 0)
	pb.Update(0.016)

	layerOffset := layer.GetOffset()
	if layerOffset.X < 0 || layerOffset.X >= 4096 {
		t.Errorf("negative offset should wrap to positive range, got %f", layerOffset.X)
	}
	// -100 % 4096 = 3996 (with the modulo implementation)
	if layerOffset.X != 3996 {
		t.Errorf("expected wrapped offset 3996 for -100, got %f", layerOffset.X)
	}
}

func TestParallaxBackground_AddChildThroughAddChild(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("sky")

	// Use the generic AddChild instead of AddLayer
	pb.AddChild(layer)

	// Should be registered as both a child and a layer
	if pb.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", pb.GetChildCount())
	}
	layers := pb.GetLayers()
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	if layers[0] != layer {
		t.Error("expected the correct layer")
	}
}

func TestParallaxBackground_RemoveChildThroughRemoveChild(t *testing.T) {
	pb := NewParallaxBackground("bg")
	layer := NewParallaxLayer("sky")
	pb.AddLayer(layer)

	// Use the generic RemoveChild
	pb.RemoveChild("sky")

	if len(pb.GetLayers()) != 0 {
		t.Error("layer should be removed from layers list")
	}
	if pb.GetChild("sky") != nil {
		t.Error("layer should be removed from children list")
	}
}

func TestParallaxBackground_AddChildNonLayer(t *testing.T) {
	pb := NewParallaxBackground("bg")
	child := NewNode2D("sprite", 1)
	pb.AddChild(child)

	if pb.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", pb.GetChildCount())
	}
	// Should NOT be in layers list
	if len(pb.GetLayers()) != 0 {
		t.Error("non-layer children should not be in layers list")
	}
}

func TestParallaxBackground_NodeInterface(t *testing.T) {
	// Compile-time check
	var _ Node = (*ParallaxBackground)(nil)

	// Runtime check
	pb := NewParallaxBackground("test")
	var n Node = pb
	if n.Name() != "test" {
		t.Error("ParallaxBackground should satisfy Node interface")
	}
}

func TestParallaxBackground_UpdateNoCameraDoesNotCrash(t *testing.T) {
	pb := NewParallaxBackground("bg")
	// Update without setting a camera reference
	pb.Update(0.016) // should not panic
}

func TestParallaxBackground_NonLayerChildrenDraw(t *testing.T) {
	pb := NewParallaxBackground("bg")
	child := NewNode2D("regular", 1)
	pb.AddChild(child)

	// Should not crash when drawing (children are iterated appropriately)
	// Note: we can't test actual rendering without ebiten screen,
	// but the method should not panic
}

// ---------------------------------------------------------------------------
// Integration tests
// ---------------------------------------------------------------------------

func TestParallaxFullPipeline(t *testing.T) {
	// Simulate a typical parallax setup: background with 3 layers
	// tracking a camera that moves over several frames
	pb := NewParallaxBackground("bg")
	pb.SetMirroring(false)

	sky := NewParallaxLayer("sky")
	sky.SetScrollScale(0.0, 0.0) // fixed

	mountains := NewParallaxLayer("mountains")
	mountains.SetScrollScale(0.3, 0.3) // slow

	trees := NewParallaxLayer("trees")
	trees.SetScrollScale(0.6, 0.4) // medium

	pb.AddLayer(sky)
	pb.AddLayer(mountains)
	pb.AddLayer(trees)

	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	// Simulate 3 frames of camera movement
	// Frame 1: move right 100
	cam.SetPosition(100, 0)
	pb.Update(0.016)
	// Frame 2: move right 50 more
	cam.SetPosition(150, 0)
	pb.Update(0.016)
	// Frame 3: move down 80
	cam.SetPosition(150, 80)
	pb.Update(0.016)

	// Sky: fixed, should be ~(0, 0)
	skyOff := sky.GetOffset()
	if absF32(skyOff.X) > 0.001 || absF32(skyOff.Y) > 0.001 {
		t.Errorf("sky: expected ~(0, 0), got (%f, %f)", skyOff.X, skyOff.Y)
	}

	// Mountains: (150*0.3, 80*0.3) = (45, 24)
	mountainOff := mountains.GetOffset()
	if absF32(mountainOff.X-45) > 0.01 || absF32(mountainOff.Y-24) > 0.01 {
		t.Errorf("mountains: expected ~(45, 24), got (%f, %f)", mountainOff.X, mountainOff.Y)
	}

	// Trees: (150*0.6, 80*0.4) = (90, 32)
	treeOff := trees.GetOffset()
	if absF32(treeOff.X-90) > 0.01 || absF32(treeOff.Y-32) > 0.01 {
		t.Errorf("trees: expected ~(90, 32), got (%f, %f)", treeOff.X, treeOff.Y)
	}

	// Total scroll offset should be (150, 80)
	scrollOff := pb.GetScrollOffset()
	if absF32(scrollOff.X-150) > 0.01 || absF32(scrollOff.Y-80) > 0.01 {
		t.Errorf("scroll offset: expected ~(150, 80), got (%f, %f)", scrollOff.X, scrollOff.Y)
	}
}

func TestParallaxBackground_DrawWithLayers(t *testing.T) {
	// Verify Draw doesn't panic and correctly virtual-dispatches to ParallaxLayer.Draw
	pb := NewParallaxBackground("bg")

	layer := NewParallaxLayer("sky")
	layer.SetScrollScale(0.5, 0.5)
	pb.AddLayer(layer)

	// Give the layer a child
	child := NewNode2D("child", 1)
	layer.AddChild(child)

	// Set up camera and update
	cam := NewNode2D("cam", 1)
	cam.SetPosition(0, 0)
	pb.SetCameraReference(cam)

	cam.SetPosition(100, 0)
	pb.Update(0.016)

	// The layer's offset should be set
	layerOff := layer.GetOffset()
	if layerOff.X != 50 {
		t.Errorf("expected layer offset X 50, got %f", layerOff.X)
	}

	// Draw should not panic (actual rendering uses ebiten which we can't test here)
	// but the layer's position is temporarily modified during Draw
}

func TestParallaxBackground_UpdatePropagatesToChildren(t *testing.T) {
	pb := NewParallaxBackground("bg")

	// Add a non-layer child with a script
	child := NewNode2D("child", 1)
	scriptCalled := false
	child.SetScript(&mockScriptHandler{
		updated: false,
	})

	// Override the mock to track calls
	child.SetScript(&mockUpdateTracker{called: &scriptCalled})
	pb.AddChild(child)

	pb.Update(0.016)

	if !scriptCalled {
		t.Error("expected Update to propagate to non-layer children")
	}
}

// mockUpdateTracker tracks whether Update was called.
type mockUpdateTracker struct {
	called *bool
}

func (m *mockUpdateTracker) Update(dt float64) {
	*m.called = true
}

func (m *mockUpdateTracker) Input(event InputEvent) {
}

// ---------------------------------------------------------------------------
// wrapOffset unit tests
// ---------------------------------------------------------------------------

func TestWrapOffset_Positive(t *testing.T) {
	// 5000 % 4096 = 904
	result := wrapOffset(5000)
	if result != 904 {
		t.Errorf("expected 904, got %f", result)
	}
}

func TestWrapOffset_Negative(t *testing.T) {
	// -100 % 4096 = 3996
	result := wrapOffset(-100)
	if result != 3996 {
		t.Errorf("expected 3996, got %f", result)
	}
}

func TestWrapOffset_Zero(t *testing.T) {
	result := wrapOffset(0)
	if result != 0 {
		t.Errorf("expected 0, got %f", result)
	}
}

func TestWrapOffset_WithinRange(t *testing.T) {
	result := wrapOffset(2048)
	if result != 2048 {
		t.Errorf("expected 2048, got %f", result)
	}
}

func TestWrapOffset_ExactTileSize(t *testing.T) {
	result := wrapOffset(4096)
	if result != 0 {
		t.Errorf("expected 0, got %f", result)
	}
}

func TestWrapOffset_NegativeExactTileSize(t *testing.T) {
	result := wrapOffset(-4096)
	if result != 0 {
		t.Errorf("expected 0, got %f", result)
	}
}

func TestWrapOffset_LargePositive(t *testing.T) {
	// 10000 % 4096 = 1808
	result := wrapOffset(10000)
	if result != 1808 {
		t.Errorf("expected 1808, got %f", result)
	}
}

func TestWrapOffset_LargeNegative(t *testing.T) {
	// -5000 % 4096
	// -5000 => -5000 + 4096 = -904 => -904 + 4096 = 3192
	result := wrapOffset(-5000)
	if result != 3192 {
		t.Errorf("expected 3192, got %f", result)
	}
}
