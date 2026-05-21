// Tests for DebugConsole
package node

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// NewDebugConsole
// ---------------------------------------------------------------------------

func TestNewDebugConsole(t *testing.T) {
	dc := NewDebugConsole("Debug")
	if dc == nil {
		t.Fatal("expected non-nil DebugConsole")
	}
	if dc.GetName() != "Debug" {
		t.Errorf("expected name 'Debug', got '%s'", dc.GetName())
	}
}

func TestDebugConsole_DefaultState(t *testing.T) {
	dc := NewDebugConsole("debug")

	if dc.Visible {
		t.Error("expected Visible to be false by default")
	}

	if dc.ToggleKey != "F3" {
		t.Errorf("expected ToggleKey 'F3', got '%s'", dc.ToggleKey)
	}

	if !dc.ShowFPS {
		t.Error("expected ShowFPS to be true by default")
	}

	if !dc.ShowEntityCount {
		t.Error("expected ShowEntityCount to be true by default")
	}

	if dc.ShowPhysics {
		t.Error("expected ShowPhysics to be false by default")
	}

	if !dc.ShowTaskCount {
		t.Error("expected ShowTaskCount to be true by default")
	}

	if dc.ShowNodeTree {
		t.Error("expected ShowNodeTree to be false by default")
	}

	if dc.ShowCameraInfo {
		t.Error("expected ShowCameraInfo to be false by default")
	}

	if dc.ShowMemory {
		t.Error("expected ShowMemory to be false by default")
	}

	if dc.Scale != 1.0 {
		t.Errorf("expected Scale 1.0, got %f", dc.Scale)
	}

	if len(dc.trackedGroups) != 0 {
		t.Errorf("expected empty trackedGroups, got %d", len(dc.trackedGroups))
	}
}

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestDebugConsole_Toggle(t *testing.T) {
	dc := NewDebugConsole("debug")

	if dc.Visible {
		t.Error("expected Visible to be false initially")
	}

	dc.Toggle()
	if !dc.Visible {
		t.Error("expected Visible to be true after first Toggle")
	}

	dc.Toggle()
	if dc.Visible {
		t.Error("expected Visible to be false after second Toggle")
	}
}

// ---------------------------------------------------------------------------
// IsVisible
// ---------------------------------------------------------------------------

func TestDebugConsole_IsVisible(t *testing.T) {
	dc := NewDebugConsole("debug")

	if dc.IsVisible() {
		t.Error("expected IsVisible to be false initially")
	}

	dc.Toggle()
	if !dc.IsVisible() {
		t.Error("expected IsVisible to be true after Toggle")
	}
}

// ---------------------------------------------------------------------------
// SetShowFPS / SetShowEntityCount / etc.
// ---------------------------------------------------------------------------

func TestDebugConsole_SetShowFPS(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowFPS(false)
	if dc.ShowFPS {
		t.Error("expected ShowFPS to be false after SetShowFPS(false)")
	}
	dc.SetShowFPS(true)
	if !dc.ShowFPS {
		t.Error("expected ShowFPS to be true after SetShowFPS(true)")
	}
}

func TestDebugConsole_SetShowEntityCount(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowEntityCount(false)
	if dc.ShowEntityCount {
		t.Error("expected ShowEntityCount to be false")
	}
	dc.SetShowEntityCount(true)
	if !dc.ShowEntityCount {
		t.Error("expected ShowEntityCount to be true")
	}
}

func TestDebugConsole_SetShowPhysics(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowPhysics(true)
	if !dc.ShowPhysics {
		t.Error("expected ShowPhysics to be true")
	}
	dc.SetShowPhysics(false)
	if dc.ShowPhysics {
		t.Error("expected ShowPhysics to be false")
	}
}

func TestDebugConsole_SetShowTaskCount(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowTaskCount(false)
	if dc.ShowTaskCount {
		t.Error("expected ShowTaskCount to be false")
	}
	dc.SetShowTaskCount(true)
	if !dc.ShowTaskCount {
		t.Error("expected ShowTaskCount to be true")
	}
}

func TestDebugConsole_SetShowNodeTree(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowNodeTree(true)
	if !dc.ShowNodeTree {
		t.Error("expected ShowNodeTree to be true")
	}
	dc.SetShowNodeTree(false)
	if dc.ShowNodeTree {
		t.Error("expected ShowNodeTree to be false")
	}
}

func TestDebugConsole_SetShowCameraInfo(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowCameraInfo(true)
	if !dc.ShowCameraInfo {
		t.Error("expected ShowCameraInfo to be true")
	}
	dc.SetShowCameraInfo(false)
	if dc.ShowCameraInfo {
		t.Error("expected ShowCameraInfo to be false")
	}
}

func TestDebugConsole_SetShowMemory(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetShowMemory(true)
	if !dc.ShowMemory {
		t.Error("expected ShowMemory to be true")
	}
	dc.SetShowMemory(false)
	if dc.ShowMemory {
		t.Error("expected ShowMemory to be false")
	}
}

// ---------------------------------------------------------------------------
// SetTextScale
// ---------------------------------------------------------------------------

func TestDebugConsole_SetTextScale(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetTextScale(2.0)
	if dc.Scale != 2.0 {
		t.Errorf("expected Scale 2.0, got %f", dc.Scale)
	}
	dc.SetTextScale(0.5)
	if dc.Scale != 0.5 {
		t.Errorf("expected Scale 0.5, got %f", dc.Scale)
	}
}

// ---------------------------------------------------------------------------
// AddTrackedGroup
// ---------------------------------------------------------------------------

func TestDebugConsole_AddTrackedGroup(t *testing.T) {
	dc := NewDebugConsole("debug")

	if len(dc.trackedGroups) != 0 {
		t.Errorf("expected no tracked groups initially, got %d", len(dc.trackedGroups))
	}

	dc.AddTrackedGroup("enemies")
	if len(dc.trackedGroups) != 1 {
		t.Errorf("expected 1 tracked group, got %d", len(dc.trackedGroups))
	}
	if dc.trackedGroups[0] != "enemies" {
		t.Errorf("expected 'enemies', got '%s'", dc.trackedGroups[0])
	}

	// Adding the same group again should be a no-op.
	dc.AddTrackedGroup("enemies")
	if len(dc.trackedGroups) != 1 {
		t.Errorf("expected 1 tracked group after duplicate add, got %d", len(dc.trackedGroups))
	}

	dc.AddTrackedGroup("players")
	if len(dc.trackedGroups) != 2 {
		t.Errorf("expected 2 tracked groups, got %d", len(dc.trackedGroups))
	}
}

// ---------------------------------------------------------------------------
// SetScene
// ---------------------------------------------------------------------------

// mockScene implements the interface required by SetScene.
type mockScene struct {
	count int
}

func (m *mockScene) Count() int                         { return m.count }
func (m *mockScene) Find(name string) interface{}        { return nil }

func TestDebugConsole_SetScene(t *testing.T) {
	dc := NewDebugConsole("debug")
	ms := &mockScene{count: 42}

	// Should not panic.
	dc.SetScene(ms)

	if dc.scene == nil {
		t.Error("expected scene to be set")
	}

	// Verify the scene is accessible through the interface.
	if sc, ok := dc.scene.(interface{ Count() int }); ok {
		if n := sc.Count(); n != 42 {
			t.Errorf("expected Count() == 42, got %d", n)
		}
	} else {
		t.Error("expected scene to satisfy Count() int")
	}
}

// ---------------------------------------------------------------------------
// SetSceneTree
// ---------------------------------------------------------------------------

// mockSceneTree implements the interface required by SetSceneTree.
type mockSceneTree struct {
	paused bool
	tps    float64
	sc     interface{ Count() int }
}

func (m *mockSceneTree) CurrentScene() interface{ Count() int } { return m.sc }
func (m *mockSceneTree) IsPaused() bool                         { return m.paused }
func (m *mockSceneTree) TPS() float64                           { return m.tps }

type mockSceneForTree struct {
	count int
}

func (m *mockSceneForTree) Count() int                  { return m.count }
func (m *mockSceneForTree) Find(name string) interface{} { return nil }

func TestDebugConsole_SetSceneTree(t *testing.T) {
	dc := NewDebugConsole("debug")
	ms := &mockSceneForTree{count: 10}
	st := &mockSceneTree{paused: true, tps: 60.0, sc: ms}

	// Should not panic.
	dc.SetSceneTree(st)

	if dc.sceneTree == nil {
		t.Fatal("expected sceneTree to be set")
	}

	// Verify the tree is accessible through its interface.
	if tree, ok := dc.sceneTree.(interface {
		CurrentScene() interface{ Count() int }
		IsPaused() bool
		TPS() float64
	}); ok {
		if tree.IsPaused() != true {
			t.Error("expected IsPaused() == true")
		}
		if tree.TPS() != 60.0 {
			t.Errorf("expected TPS() == 60.0, got %f", tree.TPS())
		}
		if cs := tree.CurrentScene(); cs != nil {
			if n := cs.Count(); n != 10 {
				t.Errorf("expected scene Count() == 10, got %d", n)
			}
		}
	} else {
		t.Error("expected sceneTree to satisfy the debug interface")
	}
}

// ---------------------------------------------------------------------------
// Node interface compliance
// ---------------------------------------------------------------------------

func TestDebugConsole_NodeInterface(t *testing.T) {
	// Compile-time check.
	var _ Node = (*DebugConsole)(nil)

	// Runtime check.
	dc := NewDebugConsole("debug-console")
	var n Node = dc
	if n.Name() != "debug-console" {
		t.Errorf("expected name 'debug-console', got '%s'", n.Name())
	}

	// Ensure embedded Node2D methods are accessible via the interface.
	if n.Parent() != nil {
		t.Error("expected nil parent")
	}
	if len(n.Children()) != 0 {
		t.Error("expected no children")
	}

	// AddChild / RemoveChild through the Node interface.
	child := NewNode2D("child", 1)
	n.AddChild(child)
	if len(n.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(n.Children()))
	}
	n.RemoveChild("child")
	if len(n.Children()) != 0 {
		t.Errorf("expected 0 children after removal, got %d", len(n.Children()))
	}
}

// ---------------------------------------------------------------------------
// Update — does not require ebiten to run, but won't process real input
// ---------------------------------------------------------------------------

func TestDebugConsole_Update(t *testing.T) {
	dc := NewDebugConsole("debug")

	// Update should not panic even without a scene reference.
	dc.Update(0.016)
}

func TestDebugConsole_UpdateSamplesFPS(t *testing.T) {
	dc := NewDebugConsole("debug")

	// fpsCount should be 0 initially.
	if dc.fpsCount != 0 {
		t.Errorf("expected fpsCount 0, got %d", dc.fpsCount)
	}

	// After Update, fpsCount should increase.
	dc.Update(0.016)
	if dc.fpsCount != 1 {
		t.Errorf("expected fpsCount 1 after one Update, got %d", dc.fpsCount)
	}

	dc.Update(0.016)
	if dc.fpsCount != 2 {
		t.Errorf("expected fpsCount 2 after two Updates, got %d", dc.fpsCount)
	}
}

// ---------------------------------------------------------------------------
// RefreshNodeTree — internal, but we can verify it populates nodeTree
// ---------------------------------------------------------------------------

func TestDebugConsole_RefreshNodeTreeNoScene(t *testing.T) {
	dc := NewDebugConsole("debug")

	// Without a scene, refreshNodeTree should produce "(no scene)".
	dc.refreshNodeTree()
	if len(dc.nodeTree) == 0 {
		t.Error("expected nodeTree to be populated")
	}
	if dc.nodeTree[0] != "(no scene)" {
		t.Errorf("expected '(no scene)', got '%s'", dc.nodeTree[0])
	}
}

func TestDebugConsole_RefreshNodeTreeWithScene(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.SetScene(&mockScene{count: 5})
	dc.AddTrackedGroup("enemies")

	dc.refreshNodeTree()
	if len(dc.nodeTree) == 0 {
		t.Fatal("expected nodeTree to be populated")
	}
	// Root entry should mention the scene.
	if dc.nodeTree[0] != "Scene (5 entities)" {
		t.Errorf("expected 'Scene (5 entities)', got '%s'", dc.nodeTree[0])
	}
}

// ---------------------------------------------------------------------------
// BuildDebugLines — internal, but can test logic without ebiten
// ---------------------------------------------------------------------------

func TestDebugConsole_BuildDebugLinesEmpty(t *testing.T) {
	dc := NewDebugConsole("debug")
	// All panels off.
	dc.ShowFPS = false
	dc.ShowEntityCount = false
	dc.ShowTaskCount = false
	dc.ShowNodeTree = false
	dc.ShowPhysics = false
	dc.ShowCameraInfo = false
	dc.ShowMemory = false

	lines := dc.buildDebugLines()
	// Only scene tree section appears (sceneTree is nil, so it's skipped).
	// Actually, with all panels off and no sceneTree, there should be no lines.
	if len(lines) != 0 {
		t.Errorf("expected empty lines, got %d lines: %v", len(lines), lines)
	}
}

func TestDebugConsole_BuildDebugLinesEntityCount(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.ShowFPS = false
	dc.ShowTaskCount = false
	dc.ShowNodeTree = false
	dc.ShowPhysics = false
	dc.ShowCameraInfo = false
	dc.ShowMemory = false
	dc.ShowEntityCount = true
	dc.SetScene(&mockScene{count: 10})

	lines := dc.buildDebugLines()
	// Should have "--- Entities ---" and "Total: 10"
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if lines[0] != "--- Entities ---" {
		t.Errorf("expected '--- Entities ---', got '%s'", lines[0])
	}
	if lines[1] != "Total: 10" {
		t.Errorf("expected 'Total: 10', got '%s'", lines[1])
	}
}

// mockSceneWithGroups implements both scene and sceneWithGroups interfaces.
type mockSceneWithGroups struct {
	total int
	groups map[string]int
}

func (m *mockSceneWithGroups) Count() int                    { return m.total }
func (m *mockSceneWithGroups) Find(name string) interface{}   { return nil }
func (m *mockSceneWithGroups) CountInGroup(group string) int { return m.groups[group] }

func TestDebugConsole_BuildDebugLinesEntityCountWithGroups(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.ShowFPS = false
	dc.ShowTaskCount = false
	dc.ShowNodeTree = false
	dc.ShowPhysics = false
	dc.ShowCameraInfo = false
	dc.ShowMemory = false
	dc.ShowEntityCount = true

	s := &mockSceneWithGroups{
		total:  4,
		groups: map[string]int{"enemies": 3, "players": 1},
	}
	dc.SetScene(s)
	dc.AddTrackedGroup("enemies")

	lines := dc.buildDebugLines()
	foundTotal := false
	foundGroup := false
	for _, line := range lines {
		if line == "Total: 4" {
			foundTotal = true
		}
		if line == "  enemies: 3" {
			foundGroup = true
		}
	}
	if !foundTotal {
		t.Error("expected 'Total: 4' in debug lines")
	}
	if !foundGroup {
		t.Error("expected '  enemies: 3' in debug lines")
	}
}

// ---------------------------------------------------------------------------
// FPS stats helper
// ---------------------------------------------------------------------------

func TestDebugConsole_ComputeFPSStatsEmpty(t *testing.T) {
	dc := NewDebugConsole("debug")
	min, max, avg := dc.computeFPSStats()
	if min != 0 || max != 0 || avg != 0 {
		t.Errorf("expected (0,0,0) for empty buffer, got (%f,%f,%f)", min, max, avg)
	}
}

// ---------------------------------------------------------------------------
// parseKeyName helper
// ---------------------------------------------------------------------------

func TestParseKeyName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"F3 default", "F3", true},
		{"lowercase", "f3", true},
		{"F1", "F1", true},
		{"F12", "F12", true},
		{"unknown key", "Space", false},
		{"empty string", "", false},
		{"garbage", "??", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := parseKeyName(tt.input)
			if tt.valid && key == ebiten.Key(0) {
				t.Errorf("expected valid key for '%s', got 0", tt.input)
			}
			if !tt.valid && key != ebiten.Key(0) {
				t.Errorf("expected invalid key for '%s', got non-zero", tt.input)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildDebugLines — scene tree panel
// ---------------------------------------------------------------------------

func TestDebugConsole_BuildDebugLinesWithSceneTree(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.ShowFPS = false
	dc.ShowEntityCount = false
	dc.ShowTaskCount = false
	dc.ShowNodeTree = false
	dc.ShowPhysics = false
	dc.ShowCameraInfo = false
	dc.ShowMemory = false

	ms := &mockSceneForTree{count: 7}
	st := &mockSceneTree{paused: false, tps: 60.0, sc: ms}
	dc.SetSceneTree(st)

	lines := dc.buildDebugLines()
	// Should include the scene tree status section.
	found := false
	for _, line := range lines {
		if line == "--- Scene Tree ---" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected '--- Scene Tree ---' in debug lines when sceneTree is set")
	}

	// Should show running state and entity count.
	foundState := false
	for _, line := range lines {
		if line == "State: running  Entities: 7" {
			foundState = true
			break
		}
	}
	if !foundState {
		t.Error("expected 'State: running  Entities: 7' in debug lines")
	}
}

func TestDebugConsole_BuildDebugLinesWithSceneTreePaused(t *testing.T) {
	dc := NewDebugConsole("debug")
	dc.ShowFPS = false
	dc.ShowEntityCount = false
	dc.ShowTaskCount = false
	dc.ShowNodeTree = false
	dc.ShowPhysics = false
	dc.ShowCameraInfo = false
	dc.ShowMemory = false

	ms := &mockSceneForTree{count: 3}
	st := &mockSceneTree{paused: true, tps: 60.0, sc: ms}
	dc.SetSceneTree(st)

	lines := dc.buildDebugLines()
	found := false
	for _, line := range lines {
		if line == "State: paused  Entities: 3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'State: paused  Entities: 3' in debug lines when scene tree is paused")
	}
}
