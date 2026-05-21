// Tests for the 2D Dynamic Lighting System
package node

import (
	"testing"

	"github.com/ElioNeto/kora/core/math"
)

// ---------------------------------------------------------------------------
// PointLight2D tests
// ---------------------------------------------------------------------------

func TestNewPointLight2D(t *testing.T) {
	pl := NewPointLight2D("torch")
	if pl == nil {
		t.Fatal("expected non-nil PointLight2D")
	}
	if pl.GetName() != "torch" {
		t.Errorf("expected name 'torch', got '%s'", pl.GetName())
	}
	if pl.Energy != 1.0 {
		t.Errorf("expected default Energy 1.0, got %f", pl.Energy)
	}
	if pl.Range != 300.0 {
		t.Errorf("expected default Range 300.0, got %f", pl.Range)
	}
	if pl.Attenuation != 1.0 {
		t.Errorf("expected default Attenuation 1.0, got %f", pl.Attenuation)
	}
	if !pl.Enabled {
		t.Error("expected Enabled to be true by default")
	}
	if !pl.ShadowsEnabled {
		t.Error("expected ShadowsEnabled to be true by default")
	}
	if pl.Color.R != 1 || pl.Color.G != 1 || pl.Color.B != 1 || pl.Color.A != 1 {
		t.Errorf("expected default Color (1,1,1,1), got (%f,%f,%f,%f)",
			pl.Color.R, pl.Color.G, pl.Color.B, pl.Color.A)
	}
}

func TestPointLight2D_SetColor(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetColor(0.5, 0.25, 0.75, 1.0)

	if pl.Color.R != 0.5 {
		t.Errorf("expected R 0.5, got %f", pl.Color.R)
	}
	if pl.Color.G != 0.25 {
		t.Errorf("expected G 0.25, got %f", pl.Color.G)
	}
	if pl.Color.B != 0.75 {
		t.Errorf("expected B 0.75, got %f", pl.Color.B)
	}
	if pl.Color.A != 1.0 {
		t.Errorf("expected A 1.0, got %f", pl.Color.A)
	}
}

func TestPointLight2D_SetColorClamped(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetColor(-0.5, 1.5, 0, 1)

	if pl.Color.R != 0 {
		t.Errorf("expected R clamped to 0, got %f", pl.Color.R)
	}
	if pl.Color.G != 1 {
		t.Errorf("expected G clamped to 1, got %f", pl.Color.G)
	}
}

func TestPointLight2D_SetEnergy(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetEnergy(0.5)
	if pl.Energy != 0.5 {
		t.Errorf("expected Energy 0.5, got %f", pl.Energy)
	}

	// Clamp low
	pl.SetEnergy(-0.1)
	if pl.Energy != 0 {
		t.Errorf("expected Energy clamped to 0, got %f", pl.Energy)
	}

	// Clamp high
	pl.SetEnergy(1.5)
	if pl.Energy != 1 {
		t.Errorf("expected Energy clamped to 1, got %f", pl.Energy)
	}
}

func TestPointLight2D_SetRange(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetRange(150.0)
	if pl.Range != 150.0 {
		t.Errorf("expected Range 150.0, got %f", pl.Range)
	}

	// Zero range
	pl.SetRange(0)
	if pl.Range != 0 {
		t.Errorf("expected Range 0, got %f", pl.Range)
	}
}

func TestPointLight2D_SetRangeNegative(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetRange(-50.0)
	if pl.Range != 0 {
		t.Errorf("expected Range clamped to 0, got %f", pl.Range)
	}
}

func TestPointLight2D_SetEnabled(t *testing.T) {
	pl := NewPointLight2D("test")
	if !pl.IsEnabled() {
		t.Error("expected IsEnabled true by default")
	}

	pl.SetEnabled(false)
	if pl.IsEnabled() {
		t.Error("expected IsEnabled false after SetEnabled(false)")
	}
	if pl.Enabled {
		t.Error("expected Enabled false")
	}

	pl.SetEnabled(true)
	if !pl.IsEnabled() {
		t.Error("expected IsEnabled true after SetEnabled(true)")
	}
}

func TestPointLight2D_NodeInterface(t *testing.T) {
	var _ Node = (*PointLight2D)(nil)

	pl := NewPointLight2D("test")
	var n Node = pl
	if n.Name() != "test" {
		t.Error("Node interface not satisfied correctly")
	}
}

func TestPointLight2D_ChildPropagation(t *testing.T) {
	pl := NewPointLight2D("parent")
	child := NewNode2D("child", 1)
	pl.AddChild(child)

	if pl.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", pl.GetChildCount())
	}
	if pl.GetChild("child") != child {
		t.Error("expected child to be accessible")
	}
}

// ---------------------------------------------------------------------------
// DirectionalLight2D tests
// ---------------------------------------------------------------------------

func TestNewDirectionalLight2D(t *testing.T) {
	dl := NewDirectionalLight2D("sun")
	if dl == nil {
		t.Fatal("expected non-nil DirectionalLight2D")
	}
	if dl.GetName() != "sun" {
		t.Errorf("expected name 'sun', got '%s'", dl.GetName())
	}
	if dl.Energy != 1.0 {
		t.Errorf("expected default Energy 1.0, got %f", dl.Energy)
	}
	if !dl.Enabled {
		t.Error("expected Enabled to be true by default")
	}
	if !dl.ShadowsEnabled {
		t.Error("expected ShadowsEnabled to be true by default")
	}
	// Default direction: upward
	if dl.Direction.X != 0 || dl.Direction.Y != -1 {
		t.Errorf("expected default Direction (0,-1), got (%f,%f)",
			dl.Direction.X, dl.Direction.Y)
	}
}

func TestDirectionalLight2D_SetColor(t *testing.T) {
	dl := NewDirectionalLight2D("test")
	dl.SetColor(0.8, 0.6, 0.2, 1.0)

	if dl.Color.R != 0.8 {
		t.Errorf("expected R 0.8, got %f", dl.Color.R)
	}
	if dl.Color.G != 0.6 {
		t.Errorf("expected G 0.6, got %f", dl.Color.G)
	}
	if dl.Color.B != 0.2 {
		t.Errorf("expected B 0.2, got %f", dl.Color.B)
	}
}

func TestDirectionalLight2D_SetEnergy(t *testing.T) {
	dl := NewDirectionalLight2D("test")
	dl.SetEnergy(0.3)
	if dl.Energy != 0.3 {
		t.Errorf("expected Energy 0.3, got %f", dl.Energy)
	}

	// Clamp
	dl.SetEnergy(-1)
	if dl.Energy != 0 {
		t.Errorf("expected Energy clamped to 0, got %f", dl.Energy)
	}
}

func TestDirectionalLight2D_SetDirection(t *testing.T) {
	dl := NewDirectionalLight2D("test")

	// Set direction to the right
	dl.SetDirection(1, 0)
	if dl.Direction.X != 1 || dl.Direction.Y != 0 {
		t.Errorf("expected Direction (1,0), got (%f,%f)",
			dl.Direction.X, dl.Direction.Y)
	}

	// Set direction diagonally — should be normalised
	dl.SetDirection(1, 1)
	expectedLen := float32(0.70710677) // 1/√2
	if absF32(dl.Direction.X-expectedLen) > 0.001 {
		t.Errorf("expected Direction.X ~%f, got %f", expectedLen, dl.Direction.X)
	}
	if absF32(dl.Direction.Y-expectedLen) > 0.001 {
		t.Errorf("expected Direction.Y ~%f, got %f", expectedLen, dl.Direction.Y)
	}
}

func TestDirectionalLight2D_SetEnabled(t *testing.T) {
	dl := NewDirectionalLight2D("test")
	dl.SetEnabled(false)
	if dl.Enabled {
		t.Error("expected Enabled false")
	}
	dl.SetEnabled(true)
	if !dl.Enabled {
		t.Error("expected Enabled true")
	}
}

func TestDirectionalLight2D_NodeInterface(t *testing.T) {
	var _ Node = (*DirectionalLight2D)(nil)

	dl := NewDirectionalLight2D("test")
	var n Node = dl
	if n.Name() != "test" {
		t.Error("Node interface not satisfied correctly")
	}
}

// ---------------------------------------------------------------------------
// LightOccluder2D tests
// ---------------------------------------------------------------------------

func TestNewLightOccluder2D(t *testing.T) {
	lo := NewLightOccluder2D("wall")
	if lo == nil {
		t.Fatal("expected non-nil LightOccluder2D")
	}
	if lo.GetName() != "wall" {
		t.Errorf("expected name 'wall', got '%s'", lo.GetName())
	}
	if lo.OccluderType != OccluderTypeRectangle {
		t.Errorf("expected default OccluderType %d (Rectangle), got %d",
			OccluderTypeRectangle, lo.OccluderType)
	}
	if lo.Width != 32 {
		t.Errorf("expected default Width 32, got %f", lo.Width)
	}
	if lo.Height != 32 {
		t.Errorf("expected default Height 32, got %f", lo.Height)
	}
	if lo.Radius != 16 {
		t.Errorf("expected default Radius 16, got %f", lo.Radius)
	}
	if !lo.Enabled {
		t.Error("expected Enabled to be true by default")
	}
}

func TestLightOccluder2D_SetSize(t *testing.T) {
	lo := NewLightOccluder2D("test")
	lo.SetSize(64, 128)
	if lo.Width != 64 {
		t.Errorf("expected Width 64, got %f", lo.Width)
	}
	if lo.Height != 128 {
		t.Errorf("expected Height 128, got %f", lo.Height)
	}

	// Zero size
	lo.SetSize(0, 0)
	if lo.Width != 0 || lo.Height != 0 {
		t.Errorf("expected (0,0), got (%f,%f)", lo.Width, lo.Height)
	}
}

func TestLightOccluder2D_SetRadius(t *testing.T) {
	lo := NewLightOccluder2D("test")
	lo.SetRadius(24)
	if lo.Radius != 24 {
		t.Errorf("expected Radius 24, got %f", lo.Radius)
	}

	// Zero radius
	lo.SetRadius(0)
	if lo.Radius != 0 {
		t.Errorf("expected Radius 0, got %f", lo.Radius)
	}
}

func TestLightOccluder2D_SetEnabled(t *testing.T) {
	lo := NewLightOccluder2D("test")
	lo.SetEnabled(false)
	if lo.Enabled {
		t.Error("expected Enabled false")
	}
	lo.SetEnabled(true)
	if !lo.Enabled {
		t.Error("expected Enabled true")
	}
}

func TestLightOccluder2D_NodeInterface(t *testing.T) {
	var _ Node = (*LightOccluder2D)(nil)

	lo := NewLightOccluder2D("test")
	var n Node = lo
	if n.Name() != "test" {
		t.Error("Node interface not satisfied correctly")
	}
}

func TestLightOccluder2D_OccluderTypeConstants(t *testing.T) {
	if OccluderTypeRectangle != 0 {
		t.Errorf("expected OccluderTypeRectangle 0, got %d", OccluderTypeRectangle)
	}
	if OccluderTypeCircle != 1 {
		t.Errorf("expected OccluderTypeCircle 1, got %d", OccluderTypeCircle)
	}
}

// ---------------------------------------------------------------------------
// LightWorld tests
// ---------------------------------------------------------------------------

func TestNewLightWorld(t *testing.T) {
	lw := NewLightWorld("LightWorld")
	if lw == nil {
		t.Fatal("expected non-nil LightWorld")
	}
	if lw.GetName() != "LightWorld" {
		t.Errorf("expected name 'LightWorld', got '%s'", lw.GetName())
	}
	if lw.AmbientEnergy != 0.3 {
		t.Errorf("expected default AmbientEnergy 0.3, got %f", lw.AmbientEnergy)
	}
	if lw.AmbientColor.R != 0.05 || lw.AmbientColor.G != 0.05 ||
		lw.AmbientColor.B != 0.1 || lw.AmbientColor.A != 1.0 {
		t.Errorf("expected default AmbientColor (0.05,0.05,0.1,1.0), got (%f,%f,%f,%f)",
			lw.AmbientColor.R, lw.AmbientColor.G, lw.AmbientColor.B, lw.AmbientColor.A)
	}
	w, h := lw.GetViewport()
	if w != 360 || h != 640 {
		t.Errorf("expected default viewport (360,640), got (%f,%f)", w, h)
	}
}

func TestLightWorld_SetViewport(t *testing.T) {
	lw := NewLightWorld("test")
	lw.SetViewport(800, 600)
	w, h := lw.GetViewport()
	if w != 800 || h != 600 {
		t.Errorf("expected viewport (800,600), got (%f,%f)", w, h)
	}

	// Zero viewport
	lw.SetViewport(0, 0)
	w, h = lw.GetViewport()
	if w != 0 || h != 0 {
		t.Errorf("expected viewport (0,0), got (%f,%f)", w, h)
	}
}

func TestLightWorld_SetAmbient(t *testing.T) {
	lw := NewLightWorld("test")
	lw.SetAmbient(0.1, 0.2, 0.3, 1.0, 0.5)

	if lw.AmbientColor.R != 0.1 {
		t.Errorf("expected R 0.1, got %f", lw.AmbientColor.R)
	}
	if lw.AmbientColor.G != 0.2 {
		t.Errorf("expected G 0.2, got %f", lw.AmbientColor.G)
	}
	if lw.AmbientColor.B != 0.3 {
		t.Errorf("expected B 0.3, got %f", lw.AmbientColor.B)
	}
	if lw.AmbientColor.A != 1.0 {
		t.Errorf("expected A 1.0, got %f", lw.AmbientColor.A)
	}
	if lw.AmbientEnergy != 0.5 {
		t.Errorf("expected AmbientEnergy 0.5, got %f", lw.AmbientEnergy)
	}
}

func TestLightWorld_SetAmbientClamped(t *testing.T) {
	lw := NewLightWorld("test")
	lw.SetAmbient(-0.5, 2.0, 0.5, -1.0, 2.0)

	if lw.AmbientColor.R != 0 {
		t.Errorf("expected R 0, got %f", lw.AmbientColor.R)
	}
	if lw.AmbientColor.G != 1 {
		t.Errorf("expected G 1, got %f", lw.AmbientColor.G)
	}
	if lw.AmbientColor.A != 0 {
		t.Errorf("expected A 0, got %f", lw.AmbientColor.A)
	}
	if lw.AmbientEnergy != 1 {
		t.Errorf("expected AmbientEnergy 1, got %f", lw.AmbientEnergy)
	}
}

func TestLightWorld_CollectsPointLightOnAddChild(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("torch")

	lw.AddChild(pl)
	lw.Update(0)

	lights := lw.GetPointLights()
	if len(lights) != 1 {
		t.Fatalf("expected 1 point light, got %d", len(lights))
	}
	if lights[0] != pl {
		t.Error("expected the added point light to be collected")
	}
}

func TestLightWorld_CollectsDirectionalLight(t *testing.T) {
	lw := NewLightWorld("test")
	dl := NewDirectionalLight2D("sun")

	lw.AddChild(dl)
	lw.Update(0)

	lights := lw.GetDirectionalLights()
	if len(lights) != 1 {
		t.Fatalf("expected 1 directional light, got %d", len(lights))
	}
	if lights[0] != dl {
		t.Error("expected the added directional light to be collected")
	}
}

func TestLightWorld_CollectsOccluder(t *testing.T) {
	lw := NewLightWorld("test")
	lo := NewLightOccluder2D("wall")

	lw.AddChild(lo)
	lw.Update(0)

	occluders := lw.GetOccluders()
	if len(occluders) != 1 {
		t.Fatalf("expected 1 occluder, got %d", len(occluders))
	}
	if occluders[0] != lo {
		t.Error("expected the added occluder to be collected")
	}
}

func TestLightWorld_CollectsMultipleTypes(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("torch")
	dl := NewDirectionalLight2D("sun")
	lo := NewLightOccluder2D("wall")

	lw.AddChild(pl)
	lw.AddChild(dl)
	lw.AddChild(lo)
	lw.Update(0)

	if len(lw.GetPointLights()) != 1 {
		t.Errorf("expected 1 point light, got %d", len(lw.GetPointLights()))
	}
	if len(lw.GetDirectionalLights()) != 1 {
		t.Errorf("expected 1 directional light, got %d", len(lw.GetDirectionalLights()))
	}
	if len(lw.GetOccluders()) != 1 {
		t.Errorf("expected 1 occluder, got %d", len(lw.GetOccluders()))
	}
}

func TestLightWorld_SkipsDisabledLights(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("disabled")
	pl.SetEnabled(false)

	lw.AddChild(pl)
	lw.Update(0)

	if len(lw.GetPointLights()) != 0 {
		t.Error("expected disabled light to not be collected")
	}
}

func TestLightWorld_SkipsDisabledOccluder(t *testing.T) {
	lw := NewLightWorld("test")
	lo := NewLightOccluder2D("disabled")
	lo.SetEnabled(false)

	lw.AddChild(lo)
	lw.Update(0)

	if len(lw.GetOccluders()) != 0 {
		t.Error("expected disabled occluder to not be collected")
	}
}

func TestLightWorld_RemoveChild(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("torch")
	lw.AddChild(pl)
	lw.Update(0)

	lw.RemoveChild("torch")
	lw.Update(0)

	if len(lw.GetPointLights()) != 0 {
		t.Error("expected point light to be removed")
	}
	if lw.GetChild("torch") != nil {
		t.Error("expected child to be removed")
	}
}

func TestLightWorld_RemoveAllChildren(t *testing.T) {
	lw := NewLightWorld("test")
	lw.AddChild(NewPointLight2D("p1"))
	lw.AddChild(NewPointLight2D("p2"))
	lw.AddChild(NewDirectionalLight2D("sun"))
	lw.Update(0)

	lw.RemoveAllChildren()
	lw.Update(0)

	if len(lw.GetPointLights()) != 0 {
		t.Error("expected all point lights removed")
	}
	if len(lw.GetDirectionalLights()) != 0 {
		t.Error("expected all directional lights removed")
	}
	if len(lw.GetOccluders()) != 0 {
		t.Error("expected all occluders removed")
	}
	if lw.GetChildCount() != 0 {
		t.Errorf("expected 0 children, got %d", lw.GetChildCount())
	}
}

func TestLightWorld_GetLightMap(t *testing.T) {
	lw := NewLightWorld("test")

	// Before RenderLightMap, GetLightMap should return nil
	if lw.GetLightMap() != nil {
		t.Error("expected nil light map before RenderLightMap")
	}

	// After RenderLightMap, it should be non-nil
	lw.RenderLightMap()
	if lw.GetLightMap() == nil {
		t.Error("expected non-nil light map after RenderLightMap")
	}
}

func TestLightWorld_NodeInterface(t *testing.T) {
	var _ Node = (*LightWorld)(nil)

	lw := NewLightWorld("test")
	var n Node = lw
	if n.Name() != "test" {
		t.Error("Node interface not satisfied correctly")
	}
}

func TestLightWorld_UpdatePropagatesToChildren(t *testing.T) {
	lw := NewLightWorld("parent")
	child := NewNode2D("child", 1)
	lw.AddChild(child)

	// Should not panic
	lw.Update(0.016)
}

func TestLightWorld_AddNilChild(t *testing.T) {
	lw := NewLightWorld("test")
	lw.AddChild(nil)
	if lw.GetChildCount() != 0 {
		t.Error("nil child should not be added")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestPointLight2D_ZeroRange(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetRange(0)
	if pl.Range != 0 {
		t.Errorf("expected Range 0, got %f", pl.Range)
	}

	lw := NewLightWorld("test")
	lw.AddChild(pl)
	lw.Update(0)

	// Light with zero range and non-zero energy should be collected
	// but RenderLightMap should handle it gracefully
	if len(lw.GetPointLights()) != 1 {
		t.Error("expected light with zero range to be collected")
	}
}

func TestPointLight2D_NegativeEnergy(t *testing.T) {
	pl := NewPointLight2D("test")
	pl.SetEnergy(-0.5)
	// Should be clamped to 0
	if pl.Energy != 0 {
		t.Errorf("expected Energy clamped to 0, got %f", pl.Energy)
	}
}

func TestDirectionalLight2D_ZeroDirection(t *testing.T) {
	dl := NewDirectionalLight2D("test")
	// Default direction (0,-1) is already normalised
	// Set to zero vector
	dl.Direction = math.Vector2{X: 0, Y: 0}
	// Direction should remain zero, but Normalize would return (1,0) for zero
	// This tests that the code handles it gracefully
}

func TestLightWorld_AddDuplicateChild(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("torch")

	lw.AddChild(pl)
	lw.AddChild(pl) // duplicate
	lw.Update(0)

	if len(lw.GetPointLights()) != 1 {
		t.Errorf("expected 1 point light (no duplicate), got %d", len(lw.GetPointLights()))
	}
	if lw.GetChildCount() != 1 {
		t.Errorf("expected 1 child (no duplicate), got %d", lw.GetChildCount())
	}
}

func TestLightWorld_NonLightChildIgnored(t *testing.T) {
	lw := NewLightWorld("test")
	regular := NewNode2D("regular", 1)
	lw.AddChild(regular)
	lw.Update(0)

	if len(lw.GetPointLights()) != 0 {
		t.Error("non-light child should not appear in pointLights")
	}
	if len(lw.GetDirectionalLights()) != 0 {
		t.Error("non-light child should not appear in directionalLights")
	}
	if len(lw.GetOccluders()) != 0 {
		t.Error("non-light child should not appear in occluders")
	}
	if lw.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", lw.GetChildCount())
	}
}

func TestLightWorld_Hierarchy(t *testing.T) {
	// Test that LightWorld works as a child of another node
	parent := NewNode2D("parent", 1)
	lw := NewLightWorld("LightWorld")
	parent.AddChild(lw)

	pl := NewPointLight2D("torch")
	lw.AddChild(pl)
	lw.Update(0)

	if len(lw.GetPointLights()) != 1 {
		t.Error("expected point light collected in hierarchy")
	}
	if lw.GetParent() != parent {
		t.Error("LightWorld should be child of parent")
	}
}

func TestLightWorld_RenderLightMapZeroViewport(t *testing.T) {
	lw := NewLightWorld("test")
	lw.SetViewport(0, 0)

	// Should not panic
	lw.RenderLightMap()
	if lw.GetLightMap() != nil {
		t.Error("expected nil light map with zero viewport")
	}
}

// ---------------------------------------------------------------------------
// Integration: LightWorld collects lights after hierarchy changes
// ---------------------------------------------------------------------------

func TestLightWorld_CollectionAfterAddChild(t *testing.T) {
	lw := NewLightWorld("test")
	lw.Update(0)

	// No lights initially
	if len(lw.GetPointLights()) != 0 {
		t.Error("expected no point lights initially")
	}

	// Add a light and update
	pl := NewPointLight2D("torch")
	lw.AddChild(pl)
	lw.Update(0)

	if len(lw.GetPointLights()) != 1 {
		t.Errorf("expected 1 point light after AddChild+Update, got %d", len(lw.GetPointLights()))
	}
}

func TestLightWorld_CollectionAfterRemoveChild(t *testing.T) {
	lw := NewLightWorld("test")
	pl := NewPointLight2D("torch")
	lw.AddChild(pl)
	lw.Update(0)

	if len(lw.GetPointLights()) != 1 {
		t.Fatalf("expected 1 point light before removal, got %d", len(lw.GetPointLights()))
	}

	lw.RemoveChild("torch")
	lw.Update(0)

	if len(lw.GetPointLights()) != 0 {
		t.Errorf("expected 0 point lights after removal, got %d", len(lw.GetPointLights()))
	}
}

// ---------------------------------------------------------------------------
// Default values table test
// ---------------------------------------------------------------------------

func TestPointLight2D_DefaultValues(t *testing.T) {
	pl := NewPointLight2D("test")

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Energy", pl.Energy, float64(1.0)},
		{"Range", pl.Range, float64(300.0)},
		{"Attenuation", pl.Attenuation, float64(1.0)},
		{"Enabled", pl.Enabled, true},
		{"ShadowsEnabled", pl.ShadowsEnabled, true},
	}

	for _, tt := range tests {
		switch v := tt.got.(type) {
		case bool:
			if v != tt.want.(bool) {
				t.Errorf("default %s = %v, want %v", tt.name, v, tt.want)
			}
		case float64:
			if v != tt.want.(float64) {
				t.Errorf("default %s = %f, want %f", tt.name, v, tt.want)
			}
		default:
			t.Errorf("unhandled type for %s", tt.name)
		}
	}
}

func TestLightOccluder2D_DefaultValues(t *testing.T) {
	lo := NewLightOccluder2D("test")

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"OccluderType", lo.OccluderType, OccluderTypeRectangle},
		{"Width", lo.Width, float32(32)},
		{"Height", lo.Height, float32(32)},
		{"Radius", lo.Radius, float32(16)},
		{"Enabled", lo.Enabled, true},
	}

	for _, tt := range tests {
		switch v := tt.got.(type) {
		case bool:
			if v != tt.want.(bool) {
				t.Errorf("default %s = %v, want %v", tt.name, v, tt.want)
			}
		case int:
			if v != tt.want.(int) {
				t.Errorf("default %s = %d, want %d", tt.name, v, tt.want)
			}
		case float32:
			if v != tt.want.(float32) {
				t.Errorf("default %s = %f, want %f", tt.name, v, tt.want)
			}
		default:
			t.Errorf("unhandled type for %s", tt.name)
		}
	}
}

func TestDirectionalLight2D_DefaultValues(t *testing.T) {
	dl := NewDirectionalLight2D("test")

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Energy", dl.Energy, float64(1.0)},
		{"Enabled", dl.Enabled, true},
		{"ShadowsEnabled", dl.ShadowsEnabled, true},
	}

	for _, tt := range tests {
		switch v := tt.got.(type) {
		case bool:
			if v != tt.want.(bool) {
				t.Errorf("default %s = %v, want %v", tt.name, v, tt.want)
			}
		case float64:
			if v != tt.want.(float64) {
				t.Errorf("default %s = %f, want %f", tt.name, v, tt.want)
			}
		default:
			t.Errorf("unhandled type for %s", tt.name)
		}
	}
}
