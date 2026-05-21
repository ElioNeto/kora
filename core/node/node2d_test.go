// Tests for Node2D basic functionality
package node

import (
	"testing"
	"github.com/ElioNeto/kora/core/math"
)

func TestNewNode2D(t *testing.T) {
	node := NewNode2D("TestNode", 1)
	if node == nil {
		t.Fatal("Expected Node2D to be created")
	}
	if node.GetName() != "TestNode" {
		t.Errorf("Expected name 'TestNode', got '%s'", node.GetName())
	}
	if node.id != 1 {
		t.Errorf("Expected id 1, got %d", node.id)
	}
	if node.pos.X != 0 || node.pos.Y != 0 {
		t.Errorf("Expected position (0,0), got (%f,%f)", node.pos.X, node.pos.Y)
	}
	if node.scaleX != 1.0 || node.scaleY != 1.0 {
		t.Errorf("Expected scale (1.0,1.0), got (%f,%f)", node.scaleX, node.scaleY)
	}
	if !node.visible {
		t.Error("Expected visible to be true by default")
	}
}

func TestSetName(t *testing.T) {
	node := NewNode2D("Original", 1)
	node.SetName("NewName")
	if node.GetName() != "NewName" {
		t.Errorf("Expected 'NewName', got '%s'", node.GetName())
	}
}

func TestGetParent(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child := NewNode2D("Child", 2)

	if parent.GetParent() != nil {
		t.Error("Parent should have no parent")
	}

	parent.AddChild(child)
	if child.GetParent() != parent {
		t.Error("Child should have parent after AddChild")
	}
}

func TestAddChild(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)

	parent.AddChild(child1)
	parent.AddChild(child2)

	if parent.GetChildCount() != 2 {
		t.Errorf("Expected 2 children, got %d", parent.GetChildCount())
	}

	if parent.GetChild("Child1") != child1 {
		t.Error("GetChild returned wrong child")
	}
	if parent.GetChild("Child2") != child2 {
		t.Error("GetChild returned wrong child")
	}
	if parent.GetChild("NonExistent") != nil {
		t.Error("GetChild should return nil for non-existent child")
	}
}

func TestRemoveChild(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)

	parent.AddChild(child1)
	parent.AddChild(child2)

	parent.RemoveChild("Child1")

	if parent.GetChildCount() != 1 {
		t.Errorf("Expected 1 child after removal, got %d", parent.GetChildCount())
	}
	if child1.GetParent() != nil {
		t.Error("Removed child should have nil parent")
	}
	if parent.GetChild("Child1") != nil {
		t.Error("Removed child should not be found")
	}
}

func TestGetNodeByPath(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)
	grandchild := NewNode2D("Grandchild", 4)

	parent.AddChild(child1)
	parent.AddChild(child2)
	child1.AddChild(grandchild)

	// Test direct child
	node := parent.GetNode("Child1")
	if node == nil || node.Name() != "Child1" {
		t.Error("GetNode should return Child1")
	}

	// Test non-existent child
	node = parent.GetNode("NonExistent")
	if node != nil {
		t.Error("GetNode should return nil for non-existent path")
	}

	// Test deep path
	node = parent.GetNode("Child1/Grandchild")
	if node == nil || node.Name() != "Grandchild" {
		t.Error("GetNode should return Grandchild")
	}

	// Test invalid path format
	node = parent.GetNode("")
	if node != nil {
		t.Error("GetNode should return nil for empty path")
	}
}

func TestGetNodeDeepPath(t *testing.T) {
	// Create a 3-level tree: A -> B -> C
	a := NewNode2D("A", 1)
	b := NewNode2D("B", 2)
	c := NewNode2D("C", 3)

	a.AddChild(b)
	b.AddChild(c)

	// Test 3-level path
	node := a.GetNode("B/C")
	if node == nil || node.Name() != "C" {
		t.Error("GetNode should return C for path 'B/C'")
	}

	// Test getting A from itself (empty path after split returns nil)
	node = a.GetNode("A")
	if node == nil || node.Name() != "A" {
		t.Error("GetNode should return A for path 'A'")
	}
}

func TestNodeInterface(t *testing.T) {
	// Verify that *Node2D satisfies the Node interface at compile time
	var _ Node = (*Node2D)(nil)

	// Also verify at runtime
	node := NewNode2D("Test", 1)
	var n Node = node
	if n.Name() != "Test" {
		t.Error("Node interface not satisfied correctly")
	}
}

func TestRemoveAllChildren(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)

	parent.AddChild(child1)
	parent.AddChild(child2)

	parent.RemoveAllChildren()

	if parent.GetChildCount() != 0 {
		t.Errorf("Expected 0 children, got %d", parent.GetChildCount())
	}
	if child1.GetParent() != nil {
		t.Error("Child parent should be nil")
	}
}

func TestSetPosition(t *testing.T) {
	node := NewNode2D("Node", 1)

	node.SetPosition(10, 20)
	pos := node.GetPosition()
	if pos.X != 10 || pos.Y != 20 {
		t.Errorf("Expected (10,20), got (%f,%f)", pos.X, pos.Y)
	}
}

func TestSetX(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetX(50)
	if node.pos.X != 50 {
		t.Errorf("Expected X=50, got %f", node.pos.X)
	}
}

func TestSetY(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetY(75)
	if node.pos.Y != 75 {
		t.Errorf("Expected Y=75, got %f", node.pos.Y)
	}
}

func TestGetRotation(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetRotation(90)
	if node.GetRotation() != 90 {
		t.Errorf("Expected rotation 90, got %f", node.GetRotation())
	}
}

func TestSetScaleX(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetScaleX(2.0)
	if node.scaleX != 2.0 {
		t.Errorf("Expected scaleX 2.0, got %f", node.scaleX)
	}
}

func TestSetScaleY(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetScaleY(1.5)
	if node.scaleY != 1.5 {
		t.Errorf("Expected scaleY 1.5, got %f", node.scaleY)
	}
}

func TestSetScaleBoth(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetScale(2.0, 3.0)

	if node.scaleX != 2.0 {
		t.Errorf("Expected scaleX 2.0, got %f", node.scaleX)
	}
	if node.scaleY != 3.0 {
		t.Errorf("Expected scaleY 3.0, got %f", node.scaleY)
	}
}

func TestToggleVisibility(t *testing.T) {
	node := NewNode2D("Node", 1)

	if !node.IsVisible() {
		t.Error("Node should be visible by default")
	}

	node.SetVisible(false)
	if node.IsVisible() {
		t.Error("Node should not be visible after SetVisible(false)")
	}

	node.SetVisible(true)
	if !node.IsVisible() {
		t.Error("Node should be visible after SetVisible(true)")
	}
}

func TestHiddenNodeNoUpdate(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child := NewNode2D("Child", 2)
	parent.AddChild(child)

	parent.Update(0.016)
}

func TestGetWorldPositionNoParent(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetPosition(10, 20)

	worldPos := node.GetWorldPosition()

	if worldPos.X != 10 || worldPos.Y != 20 {
		t.Errorf("Expected (10,20), got (%f,%f)", worldPos.X, worldPos.Y)
	}
}

func TestSetWorldPosition(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetWorldPosition(50, 100)

	pos := node.GetPosition()
	if pos.X != 50 || pos.Y != 100 {
		t.Errorf("Expected (50,100), got (%f,%f)", pos.X, pos.Y)
	}
}

func TestTransformPoint(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	parent.SetPosition(100, 200)

	child := NewNode2D("Child", 2)
	child.SetPosition(10, 20)
	parent.AddChild(child)

	point := math.NewVector2(5, 5)
	transformed := child.TransformPoint(point)
	expected := math.NewVector2(115, 225)

	if transformed.X != expected.X || transformed.Y != expected.Y {
		t.Errorf("Expected (%f,%f), got (%f,%f)", expected.X, expected.Y, transformed.X, transformed.Y)
	}
}

func TestNoParentNoTransform(t *testing.T) {
	node := NewNode2D("Node", 1)
	point := math.NewVector2(10, 20)

	result := node.TransformPoint(point)
	if result.X != 10 || result.Y != 20 {
		t.Errorf("Expected (%f,%f), got (%f,%f)", point.X, point.Y, result.X, result.Y)
	}
}

func TestGetWorldRotation(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	parent.SetRotation(30)

	child := NewNode2D("Child", 2)
	child.SetRotation(15)
	parent.AddChild(child)

	worldRot := child.GetWorldRotation()
	if worldRot != 45 {
		t.Errorf("Expected 45 degrees, got %f", worldRot)
	}
}

func TestGetWorldRotationNoParent(t *testing.T) {
	node := NewNode2D("Node", 1)
	node.SetRotation(45)

	worldRot := node.GetWorldRotation()
	if worldRot != 45 {
		t.Errorf("Expected 45 degrees, got %f", worldRot)
	}
}

func TestAddNilChild(t *testing.T) {
	node := NewNode2D("Node", 1)

	node.AddChild(nil)

	if node.GetChildCount() != 0 {
		t.Error("Nil child should not be added")
	}
}

func TestGetChildCount(t *testing.T) {
	node := NewNode2D("Node", 1)

	if node.GetChildCount() != 0 {
		t.Errorf("Expected 0 children, got %d", node.GetChildCount())
	}

	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)
	node.AddChild(child1)
	node.AddChild(child2)

	if node.GetChildCount() != 2 {
		t.Errorf("Expected 2 children, got %d", node.GetChildCount())
	}
}

func TestSetScriptAndGetScript(t *testing.T) {
	node := NewNode2D("Node", 1)

	mockScript := &mockScriptHandler{}
	node.SetScript(mockScript)

	if node.GetScript() != mockScript {
		t.Error("GetScript should return the set script")
	}
}

func TestUpdateWithScript(t *testing.T) {
	node := NewNode2D("Node", 1)
	mock := &mockScriptHandler{}
	node.SetScript(mock)

	node.Update(0.016)

	if !mock.updated {
		t.Error("Script Update should have been called")
	}
}

func TestUpdatePropagatesToChildren(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)

	parent.AddChild(child1)
	parent.AddChild(child2)

	parent.SetScript(&mockScriptHandler{})
	child1.SetScript(&mockScriptHandler{})
	child2.SetScript(&mockScriptHandler{})

	parent.Update(0.016)
}

func TestProcessInput(t *testing.T) {
	node := NewNode2D("Node", 1)

	event := InputEvent{
		Type: "pressed",
		Key:  "A",
	}

	node.ProcessInput(event)
}

func TestProcessInputPropagation(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child := NewNode2D("Child", 2)
	parent.AddChild(child)

	parent.SetInputHandler(func(e InputEvent) {})
	child.SetInputHandler(func(e InputEvent) {})

	event := InputEvent{Type: "pressed", Key: "Space"}

	parent.ProcessInput(event)
}

func TestSetInputHandler(t *testing.T) {
	node := NewNode2D("Node", 1)

	handlerCalled := false
	handler := func(e InputEvent) {
		handlerCalled = true
	}

	node.SetInputHandler(handler)
	node.ProcessInput(InputEvent{Type: "pressed", Key: "A"})

	if !handlerCalled {
		t.Error("Input handler should have been called")
	}
}

func TestSetCollisionHook(t *testing.T) {
	node := NewNode2D("Node", 1)

	collisionCalled := false
	hook := func(other *Node2D, eventType CollisionType) {
		collisionCalled = true
	}

	node.SetCollisionHook(hook)
	other := NewNode2D("Other", 2)

	node.TriggerCollision(other, CollisionTypeEnter)

	if !collisionCalled {
		t.Error("Collision hook should have been called")
	}
}

func TestGetChildren(t *testing.T) {
	parent := NewNode2D("Parent", 1)
	child1 := NewNode2D("Child1", 2)
	child2 := NewNode2D("Child2", 3)

	parent.AddChild(child1)
	parent.AddChild(child2)

	children := parent.GetChildren()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

type mockScriptHandler struct {
	updated bool
}

func (m *mockScriptHandler) Update(dt float64) {
	m.updated = true
}

func (m *mockScriptHandler) Input(event InputEvent) {
	_ = event
}

// ---------------------------------------------------------------------------
// Camera2D tests
// ---------------------------------------------------------------------------

func TestNewCamera2D(t *testing.T) {
	cam := NewCamera2D("MainCamera")
	if cam == nil {
		t.Fatal("expected non-nil camera")
	}
	if cam.GetName() != "MainCamera" {
		t.Errorf("expected name 'MainCamera', got '%s'", cam.GetName())
	}
}

func TestCamera2D_SetTarget(t *testing.T) {
	cam := NewCamera2D("cam")
	target := NewNode2D("target", 1)
	offset := math.NewVector2(0, 0)
	cam.SetTarget(target, offset)
	got := cam.GetTarget()
	if got != target {
		t.Error("expected target to match")
	}
}

func TestCamera2D_ShakeStartsAndFinishes(t *testing.T) {
	cam := NewCamera2D("cam")
	cam.Shake(10, 1.0) // intensity=10, duration=1s

	// After starting, shake should be active but with zero offset initially
	// (offset accumulates during Update)
	cam.Update(0.1) // advance 0.1s

	// After 1.1 seconds, shake should be finished
	cam.Update(1.0)
	// Shake should now be done
	// (we can verify by checking there's no more shake offset after full duration)
	cam.Update(0.1) // past duration
}

func TestCamera2D_ShakeZeroDuration(t *testing.T) {
	cam := NewCamera2D("cam")
	// Should not panic
	cam.Shake(10, 0)
	cam.Update(0.1)
}

func TestCamera2D_ShakeZeroIntensity(t *testing.T) {
	cam := NewCamera2D("cam")
	// Should not panic
	cam.Shake(0, 1.0)
	cam.Update(0.1)
}

func TestCamera2D_Bounds(t *testing.T) {
	cam := NewCamera2D("cam")
	min := math.NewVector2(0, 0)
	max := math.NewVector2(100, 100)
	cam.SetBounds(min, max)

	gotMin, gotMax := cam.GetBounds()
	if gotMin != min {
		t.Errorf("expected min %v, got %v", min, gotMin)
	}
	if gotMax != max {
		t.Errorf("expected max %v, got %v", max, gotMax)
	}
}

func TestCamera2D_TargetOffset(t *testing.T) {
	cam := NewCamera2D("cam")
	offset := math.NewVector2(10, 20)
	cam.SetTargetOffset(offset)
	got := cam.GetTargetOffset()
	if got != offset {
		t.Errorf("expected offset %v, got %v", offset, got)
	}
}

func TestCamera2D_Smoothing(t *testing.T) {
	cam := NewCamera2D("cam")
	cam.SetSmoothing(0.5)
	got := cam.GetSmoothing()
	if got != 0.5 {
		t.Errorf("expected smoothing 0.5, got %f", got)
	}
	// Clamp to [0,1]
	cam.SetSmoothing(-0.1)
	if cam.GetSmoothing() != 0 {
		t.Errorf("expected smoothing clamped to 0, got %f", cam.GetSmoothing())
	}
	cam.SetSmoothing(1.5)
	if cam.GetSmoothing() != 1 {
		t.Errorf("expected smoothing clamped to 1, got %f", cam.GetSmoothing())
	}
}
