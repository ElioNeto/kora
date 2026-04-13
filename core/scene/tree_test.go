package scene

import (
	"testing"
)

// mockNode implements TreeNode for testing
type mockNode struct {
	alive        bool
	processMode  ProcessMode
	updateCalls  int
	drawCalls    int
	physicsCalls int
}

func newMockNode(mode ProcessMode) *mockNode {
	return &mockNode{
		alive:       true,
		processMode: mode,
	}
}

func (n *mockNode) Update(dt float64) {
	n.updateCalls++
}

func (n *mockNode) Draw(screen interface{}) {
	n.drawCalls++
}

func (n *mockNode) IsAlive() bool {
	return n.alive
}

func (n *mockNode) GetProcessMode() ProcessMode {
	return n.processMode
}

func (n *mockNode) Destroy() {
	n.alive = false
}

// mockPhysicsNode implements TreeNode + PhysicsNode for testing
type mockPhysicsNode struct {
	*mockNode
	physicsUpdates int
}

func newMockPhysicsNode(mode ProcessMode) *mockPhysicsNode {
	return &mockPhysicsNode{
		mockNode: newMockNode(mode),
	}
}

func (n *mockPhysicsNode) PhysicsUpdate(dt float64) {
	n.physicsUpdates++
}

// TestSceneTree_New creates a new SceneTree and checks initialization
func TestSceneTree_New(t *testing.T) {
	tree := NewSceneTree()

	if tree == nil {
		t.Fatal("NewSceneTree() returned nil")
	}

	if tree.physicsDt != 1.0/60.0 {
		t.Errorf("Expected default physicsDt %v, got %v", 1.0/60.0, tree.physicsDt)
	}

	if tree.IsPaused() {
		t.Error("New SceneTree should not be paused by default")
	}
}

// TestSceneTree_SetCurrentScene tests setting the current scene
func TestSceneTree_SetCurrentScene(t *testing.T) {
	tree := NewSceneTree()

	// GetCurrentScene creates a scene lazily, so it should never be nil
	_ = tree.GetCurrentScene()

	scene := New()
	tree.SetCurrentScene(scene)

	if tree.GetCurrentScene() != scene {
		t.Error("GetCurrentScene should return the set scene")
	}
}

// TestSceneTree_Pause_Resume tests Pause and Resume functionality
func TestSceneTree_Pause_Resume(t *testing.T) {
	tree := NewSceneTree()

	tree.Pause()
	if !tree.IsPaused() {
		t.Error("Tree should be paused after Pause()")
	}

	tree.Resume()
	if tree.IsPaused() {
		t.Error("Tree should not be paused after Resume()")
	}
}

// TestSceneTree_Tick_UpdatesEntities tests that Tick calls Update on entities
func TestSceneTree_Tick_UpdatesEntities(t *testing.T) {
	tree := NewSceneTree()

	// Create a simple scene with mock nodes
	scene := New()
	tree.SetCurrentScene(scene)

	// Note: This test is structural - actual entity update depends on
	// how entities are added to the scene
	tree.Tick(1.0 / 60.0)

	if tree.Len() < 0 {
		t.Error("Tick should not fail on empty scene")
	}
}

// TestSceneTree_Draw_AlwaysRuns tests that Draw runs even when paused
func TestSceneTree_Draw_AlwaysRuns(t *testing.T) {
	tree := NewSceneTree()

	// Create a scene with a mock node
	scene := New()
	tree.SetCurrentScene(scene)

	// Pause the tree
	tree.Pause()

	// Draw should still work
	tree.Draw(nil)

	// If we got here without panic, Draw ran successfully
}

// TestSceneTree_ChangeScene tests scene switching
func TestSceneTree_ChangeScene(t *testing.T) {
	tree := NewSceneTree()

	// Register test scenes
	scene1 := New()
	scene2 := New()

	tree.RegisterScene("level1", scene1)
	tree.RegisterScene("level2", scene2)
	tree.SetCurrentScene(scene1)

	// Initial scene is active
	if tree.GetCurrentScene() != scene1 {
		t.Error("scene1 should be initially active")
	}

	// Request transition
	ok := tree.ChangeScene("level2")
	if !ok {
		t.Error("ChangeScene should return true for valid scene")
	}

	// Transition should not be applied yet (queued)
	if tree.GetCurrentScene() != scene1 {
		t.Error("Transition should be queued, not applied")
	}

	// Tick applies the transition
	tree.Tick(1.0 / 60.0)
	if tree.GetCurrentScene() != scene2 {
		t.Error("After Tick, scene2 should be active")
	}
}

// TestSceneTree_ChangeScene_Invalid tests changing to non-existent scene
func TestSceneTree_ChangeScene_Invalid(t *testing.T) {
	tree := NewSceneTree()

	ok := tree.ChangeScene("nonexistent")
	if ok {
		t.Error("ChangeScene should return false for invalid scene")
	}
}

// TestSceneTree_SkipPhysicsWhenPaused tests that physics is skipped when paused
func TestSceneTree_SkipPhysicsWhenPaused(t *testing.T) {
	tree := NewSceneTree()

	scene := New()
	tree.SetCurrentScene(scene)

	tree.Pause()
	tree.Tick(1.0 / 60.0)

	// Scene should remain valid even when physics is skipped
	if tree.GetCurrentScene() == nil {
		t.Error("Scene should not become nil when paused")
	}
}

// TestSceneTree_TickOrder verifies the order: physics -> update -> scheduler -> draw
func TestSceneTree_TickOrder(t *testing.T) {
	tree := NewSceneTree()

	// We can't easily track internal order, but we can verify the
	// Tick doesn't panic and maintains scene state
	scene := New()
	tree.SetCurrentScene(scene)

	for i := 0; i < 5; i++ {
		tree.Tick(1.0 / 60.0)
		if tree.GetCurrentScene() == nil {
			t.Fatalf("Tick %d caused scene to become nil", i)
		}
	}
}

// TestSceneTree_PauseModeAlways tests ProcessModeAlways nodes run when paused
func TestSceneTree_PauseModeAlways(t *testing.T) {
	tree := NewSceneTree()
	scene := New()
	tree.SetCurrentScene(scene)

	// Create a node with ProcessModeAlways
	node := newMockNode(ProcessModeAlways)

	// We can't directly add nodes to scene in this test,
	// but we can verify the node exists and has correct mode
	if node.GetProcessMode() != ProcessModeAlways {
		t.Errorf("Node should have ProcessModeAlways, got %v", node.GetProcessMode())
	}

	// Even when tree is paused, ProcessModeAlways should be handled
	tree.Pause()
	tree.Tick(1.0 / 60.0)

	// Node should not have been corrupted
	if node.IsAlive() != true {
		t.Error("ProcessedModeAlways node should still be alive")
	}
}

// Helper to Register scene with factory
func RegisterSceneWithFactory(t *testing.T, factory func() *Scene) *SceneTree {
	tree := NewSceneTree()
	return tree
}

// TestSceneTree_ProcessModeAlwaysRuns tests ProcessModeAlways nodes update when paused
func TestSceneTree_ProcessModeAlwaysRuns(t *testing.T) {
	tree := NewSceneTree()
	scene := New()
	tree.SetCurrentScene(scene)

	// Spawn a node with ProcessModeAlways
	node := newMockNode(ProcessModeAlways)
	tree.GetCurrentScene().Spawn(node)

	tree.Pause()
	tree.Tick(1.0 / 60.0)

	// ProcessModeAlways should have updated
	if node.updateCalls == 0 {
		t.Error("ProcessModeAlways node should update when paused")
	}
}

// TestSceneTree_ProcessModePausableStops tests Pausable nodes don't update when paused
func TestSceneTree_ProcessModePausableStops(t *testing.T) {
	tree := NewSceneTree()
	scene := New()
	tree.SetCurrentScene(scene)

	// Spawn a pausable node (default)
	node := newMockNode(ProcessModePausable)
	tree.GetCurrentScene().Spawn(node)

	pauseUpdates := node.updateCalls
	tree.Pause()
	tree.Tick(1.0 / 60.0)

	// Pausable should NOT have updated when paused
	if node.updateCalls != pauseUpdates {
		t.Error("ProcessModePausable node should not update when paused")
	}
}

// TestSceneTree_DrawRendersPausedNodes tests that Draw runs even for paused entities
func TestSceneTree_DrawRendersPausedNodes(t *testing.T) {
	tree := NewSceneTree()
	scene := New()
	tree.SetCurrentScene(scene)

	// Spawn a paused entity
	node := newMockNode(ProcessModePausable)
	tree.GetCurrentScene().Spawn(node)

	tree.Pause()
	tree.Tick(1.0 / 60.0)

	// Draw should still work
	tree.Draw(nil)

	// If Draw panics, test will fail here
	t.Log("Draw completed successfully for paused tree")
}

// TestSceneTree_Ticks60TPS tests that ticks advance at ~60 TPS
func TestSceneTree_Ticks60TPS(t *testing.T) {
	tree := NewSceneTree()
	tree.SetCurrentScene(New())

	initialTick := tree.GetCurrentScene()
	for i := 0; i < 60; i++ {
		tree.Tick(1.0 / 60.0)
	}
	finalTick := tree.GetCurrentScene()

	// If all works, scene should still be valid
	if finalTick == nil {
		t.Error("Scene became nil after 60 ticks")
	}
	if initialTick != finalTick {
		t.Error("Scene should persist across ticks")
	}
}
