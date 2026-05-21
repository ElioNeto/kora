// Tests for Skeleton2D and Bone2D
package node

import (
	stdMath "math"
	"testing"
)

// ---------------------------------------------------------------------------
// Bone2D tests
// ---------------------------------------------------------------------------

func TestNewBone2D(t *testing.T) {
	b := NewBone2D("test_bone")
	if b == nil {
		t.Fatal("expected non-nil Bone2D")
	}
	if b.GetName() != "test_bone" {
		t.Errorf("expected name 'test_bone', got '%s'", b.GetName())
	}
	if b.GetLength() != 32.0 {
		t.Errorf("expected default length 32.0, got %f", b.GetLength())
	}
	if b.DefaultLength != 32.0 {
		t.Errorf("expected DefaultLength 32.0, got %f", b.DefaultLength)
	}
	if b.IKMode {
		t.Error("expected IKMode to be false by default")
	}
	if b.IKSegments != 0 {
		t.Errorf("expected IKSegments 0 by default, got %d", b.IKSegments)
	}
	if b.BendDirection != 0 {
		t.Errorf("expected BendDirection 0 by default, got %d", b.BendDirection)
	}
	if b.SpriteName != "" {
		t.Errorf("expected empty SpriteName by default, got '%s'", b.SpriteName)
	}
	if b.ShowGizmo {
		t.Error("expected ShowGizmo to be false by default")
	}
	if b.GizmoColor.R != 1 || b.GizmoColor.G != 1 || b.GizmoColor.B != 1 || b.GizmoColor.A != 1 {
		t.Errorf("expected GizmoColor {1,1,1,1}, got %+v", b.GizmoColor)
	}
}

func TestBone2D_SetGetLength(t *testing.T) {
	b := NewBone2D("bone")
	b.SetLength(64.0)
	if b.GetLength() != 64.0 {
		t.Errorf("expected length 64.0, got %f", b.GetLength())
	}
	// Test zero length
	b.SetLength(0)
	if b.GetLength() != 0 {
		t.Errorf("expected length 0, got %f", b.GetLength())
	}
}

func TestBone2D_GetEndPosition(t *testing.T) {
	b := NewBone2D("bone")
	b.SetLength(100.0)
	// Default rotation is 0°, so the bone points right (positive X)
	endPos := b.GetEndPosition()
	if endPos.X != 100.0 || endPos.Y != 0 {
		t.Errorf("expected end position (100, 0) for 0° rotation, got (%f, %f)", endPos.X, endPos.Y)
	}

	// Rotate 90° -> bone points down (positive Y)
	b.SetRotation(90)
	endPos = b.GetEndPosition()
	if absFloat(endPos.X-0) > 0.001 || absFloat(endPos.Y-100) > 0.001 {
		t.Errorf("expected end position ~(0, 100) for 90° rotation, got (%f, %f)", endPos.X, endPos.Y)
	}

	// Rotate -90° (270°) -> bone points up (negative Y)
	b.SetRotation(-90)
	endPos = b.GetEndPosition()
	if absFloat(endPos.X-0) > 0.001 || absFloat(endPos.Y+100) > 0.001 {
		t.Errorf("expected end position ~(0, -100) for -90° rotation, got (%f, %f)", endPos.X, endPos.Y)
	}
}

func TestBone2D_GetEndPositionWithParent(t *testing.T) {
	parent := NewBone2D("parent")
	parent.SetLength(50.0)
	parent.SetPosition(100, 200)

	child := NewBone2D("child")
	child.SetLength(30.0)
	// Position child at the end of the parent bone
	child.SetPosition(50, 0)
	parent.AddChild(child)

	// Parent at (100,200) rotation 0, parent tip at (150,200)
	// Child world position = parent.TransformPoint((50,0)) = (100+50, 200+0) = (150, 200)
	// Child tip = (150 + 30, 200) = (180, 200)
	endPos := child.GetEndPosition()
	if absFloat(endPos.X-180) > 0.001 || absFloat(endPos.Y-200) > 0.001 {
		t.Errorf("expected child end position ~(180, 200), got (%f, %f)", endPos.X, endPos.Y)
	}
}

func TestBone2D_SetIKMode(t *testing.T) {
	b := NewBone2D("bone")
	b.SetIKMode(true)
	if !b.IKMode {
		t.Error("expected IKMode true after SetIKMode(true)")
	}
	b.SetIKMode(false)
	if b.IKMode {
		t.Error("expected IKMode false after SetIKMode(false)")
	}
}

func TestBone2D_SetBendDirection(t *testing.T) {
	b := NewBone2D("bone")
	b.SetBendDirection(1)
	if b.BendDirection != 1 {
		t.Errorf("expected BendDirection 1, got %d", b.BendDirection)
	}
	b.SetBendDirection(-1)
	if b.BendDirection != -1 {
		t.Errorf("expected BendDirection -1, got %d", b.BendDirection)
	}
	// Clamp to valid range
	b.SetBendDirection(5)
	if b.BendDirection != 1 {
		t.Errorf("expected BendDirection clamped to 1, got %d", b.BendDirection)
	}
	b.SetBendDirection(-5)
	if b.BendDirection != -1 {
		t.Errorf("expected BendDirection clamped to -1, got %d", b.BendDirection)
	}
	// Reset to 0
	b.SetBendDirection(0)
	if b.BendDirection != 0 {
		t.Errorf("expected BendDirection 0, got %d", b.BendDirection)
	}
}

func TestBone2D_NodeInterface(t *testing.T) {
	var _ Node = (*Bone2D)(nil)

	b := NewBone2D("node_test")
	var n Node = b
	if n.Name() != "node_test" {
		t.Error("Bone2D should satisfy Node interface")
	}
}

func TestBone2D_ParentChild(t *testing.T) {
	parent := NewBone2D("parent")
	child := NewBone2D("child")

	parent.AddChild(child)
	if child.GetParent() != parent.Node2D {
		t.Error("child's parent should be the parent Bone2D's Node2D")
	}
}

// ---------------------------------------------------------------------------
// Skeleton2D tests
// ---------------------------------------------------------------------------

func TestNewSkeleton2D(t *testing.T) {
	s := NewSkeleton2D("test_skeleton")
	if s == nil {
		t.Fatal("expected non-nil Skeleton2D")
	}
	if s.GetName() != "test_skeleton" {
		t.Errorf("expected name 'test_skeleton', got '%s'", s.GetName())
	}
	if len(s.GetAllBones()) != 0 {
		t.Errorf("expected no bones initially, got %d", len(s.GetAllBones()))
	}
	if s.animator != nil {
		t.Error("expected animator to be nil initially")
	}
	if s.ikActive {
		t.Error("expected ikActive to be false initially")
	}
}

func TestSkeleton2D_AddBone(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone1 := NewBone2D("bone1")
	bone2 := NewBone2D("bone2")

	s.AddBone("", bone1)
	s.AddBone("", bone2)

	// Should have 2 bones
	bones := s.GetAllBones()
	if len(bones) != 2 {
		t.Errorf("expected 2 bones, got %d", len(bones))
	}
}

func TestSkeleton2D_AddBoneWithParent(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	root := NewBone2D("root")
	child := NewBone2D("child")

	s.AddBone("", root)
	s.AddBone("root", child)

	// Verify child is under root
	rootChildren := root.GetChildren()
	if len(rootChildren) != 1 {
		t.Errorf("expected root to have 1 child, got %d", len(rootChildren))
	}
	if rootChildren[0] != child.Node2D {
		t.Error("root's child should be child's Node2D")
	}
}

func TestSkeleton2D_AddBoneNilParent(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("bone")

	// Adding with nonexistent parent should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding bone to nonexistent parent")
		}
	}()
	s.AddBone("nonexistent", bone)
}

func TestSkeleton2D_GetBone(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("my_bone")
	s.AddBone("", bone)

	got := s.GetBone("my_bone")
	if got != bone {
		t.Error("GetBone should return the same bone pointer")
	}

	got = s.GetBone("nonexistent")
	if got != nil {
		t.Error("GetBone should return nil for nonexistent bone")
	}
}

func TestSkeleton2D_GetAllBones(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone1 := NewBone2D("bone1")
	bone2 := NewBone2D("bone2")
	bone3 := NewBone2D("bone3")

	s.AddBone("", bone1)
	s.AddBone("", bone2)
	s.AddBone("", bone3)

	bones := s.GetAllBones()
	if len(bones) != 3 {
		t.Errorf("expected 3 bones, got %d", len(bones))
	}

	// Collect names for comparison
	names := make(map[string]bool)
	for _, b := range bones {
		names[b.GetName()] = true
	}
	if !names["bone1"] || !names["bone2"] || !names["bone3"] {
		t.Error("GetAllBones should contain all added bones")
	}
}

func TestSkeleton2D_StoreRestPose(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("bone")
	s.AddBone("", bone)

	// Modify bone
	bone.SetPosition(10, 20)
	bone.SetRotation(45)
	bone.SetScaleX(2.0)
	bone.SetScaleY(0.5)

	// Store rest pose
	s.StoreRestPose()

	// Modify bone again
	bone.SetPosition(0, 0)
	bone.SetRotation(0)

	// Apply rest pose
	s.ApplyRestPose()

	// Verify bone is restored
	pos := bone.GetPosition()
	if pos.X != 10 || pos.Y != 20 {
		t.Errorf("expected rest position (10,20), got (%f,%f)", pos.X, pos.Y)
	}
	if bone.GetRotation() != 45 {
		t.Errorf("expected rest rotation 45, got %f", bone.GetRotation())
	}
	if bone.GetScaleX() != 2.0 || bone.GetScaleY() != 0.5 {
		t.Errorf("expected rest scale (2.0, 0.5), got (%f, %f)", bone.GetScaleX(), bone.GetScaleY())
	}
}

func TestSkeleton2D_ApplyRestPoseEmpty(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	// Should not panic when applying empty rest pose
	s.ApplyRestPose()
}

func TestSkeleton2D_StoreRestPoseMultiple(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	root := NewBone2D("root")
	child := NewBone2D("child")

	s.AddBone("", root)
	s.AddBone("root", child)

	root.SetPosition(100, 200)
	root.SetRotation(30)

	child.SetPosition(0, 0)
	child.SetRotation(45)

	s.StoreRestPose()

	// Change values
	root.SetRotation(0)
	child.SetRotation(0)

	s.ApplyRestPose()

	if root.GetRotation() != 30 {
		t.Errorf("expected root rotation 30, got %f", root.GetRotation())
	}
	if child.GetRotation() != 45 {
		t.Errorf("expected child rotation 45, got %f", child.GetRotation())
	}
}

func TestSkeleton2D_UpdateCollectsBones(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("dynamic_bone")
	s.AddBone("", bone)

	// Bones should be found immediately after AddBone
	if len(s.GetAllBones()) != 1 {
		t.Errorf("expected 1 bone after AddBone, got %d", len(s.GetAllBones()))
	}

	// Update should maintain the bone list
	s.Update(0.016)
	if len(s.GetAllBones()) != 1 {
		t.Errorf("expected 1 bone after Update, got %d", len(s.GetAllBones()))
	}
}

func TestSkeleton2D_GizmoDefaults(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("bone")
	s.AddBone("", bone)

	if bone.ShowGizmo {
		t.Error("expected ShowGizmo to be false by default")
	}
	if bone.GizmoColor.R != 1 || bone.GizmoColor.G != 1 || bone.GizmoColor.B != 1 || bone.GizmoColor.A != 1 {
		t.Error("expected default white gizmo color")
	}
}

func TestSkeleton2D_SetIKTarget(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	s.SetIKTarget(100, 200)
	if !s.ikActive {
		t.Error("expected ikActive true after SetIKTarget")
	}
	if s.ikTargetX != 100 || s.ikTargetY != 200 {
		t.Errorf("expected target (100, 200), got (%f, %f)", s.ikTargetX, s.ikTargetY)
	}
}

func TestSkeleton2D_ClearIKTarget(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	s.SetIKTarget(100, 200)
	s.ClearIKTarget()
	if s.ikActive {
		t.Error("expected ikActive false after ClearIKTarget")
	}
}

// ---------------------------------------------------------------------------
// IK Solver tests
// ---------------------------------------------------------------------------

func TestIK_SolveSingleBone(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	bone := NewBone2D("bone")
	bone.SetLength(50.0)
	bone.SetIKMode(true)
	s.AddBone("", bone)

	// Target is directly ahead (angle 0), bone starts at 45 degrees
	bone.SetRotation(45)

	chain := []*Bone2D{bone}
	s.SolveIK(chain, 50, 0)

	// After solving, bone should rotate toward the target
	// Target is at (50, 0), bone is at (0, 0), so ideal angle is 0°
	if absFloat(bone.GetRotation()) > 5 {
		t.Errorf("expected bone rotation near 0° after solving, got %f°", bone.GetRotation())
	}
}

func TestIK_SolveTwoBoneChain(t *testing.T) {
	s := NewSkeleton2D("skeleton")

	root := NewBone2D("root")
	root.SetLength(50.0)
	root.SetIKMode(true)
	s.AddBone("", root)

	child := NewBone2D("child")
	child.SetLength(50.0)
	child.SetIKMode(true)
	// Position child at root's tip
	child.SetPosition(root.GetLength(), 0)
	s.AddBone("root", child)

	// Rotate both bones to make it interesting
	root.SetRotation(30)
	child.SetRotation(45)

	// Solve IK toward a reachable target
	chain := []*Bone2D{root, child}
	s.SolveIK(chain, 80, 20)

	// After solving, the end effector should be closer to the target
	endPos := child.GetEndPosition()
	targetDist := stdMath.Sqrt((float64(endPos.X)-80)*(float64(endPos.X)-80) + (float64(endPos.Y)-20)*(float64(endPos.Y)-20))
	if targetDist > 20 {
		t.Errorf("expected end effector within 20 units of target (80,20), got distance %f at (%f,%f)",
			targetDist, endPos.X, endPos.Y)
	}
}

func TestIK_SolveChainWithBendConstraint(t *testing.T) {
	s := NewSkeleton2D("skeleton")

	root := NewBone2D("root")
	root.SetLength(50.0)
	root.SetIKMode(true)
	root.SetBendDirection(1) // Only positive bend
	s.AddBone("", root)

	child := NewBone2D("child")
	child.SetLength(50.0)
	child.SetIKMode(true)
	child.SetBendDirection(1)
	// Position child at root's tip
	child.SetPosition(root.GetLength(), 0)
	s.AddBone("root", child)

	chain := []*Bone2D{root, child}
	s.SolveIK(chain, 80, 20)

	// With bend direction constraint, child rotation should be >= 0
	if child.GetRotation() < -0.1 {
		t.Errorf("expected child rotation >= 0 with BendDirection=1, got %f", child.GetRotation())
	}
}

func TestIK_SolveIKChain(t *testing.T) {
	s := NewSkeleton2D("skeleton")

	root := NewBone2D("root")
	root.SetLength(50.0)
	root.SetIKMode(true)
	s.AddBone("", root)

	middle := NewBone2D("middle")
	middle.SetLength(50.0)
	middle.SetIKMode(true)
	middle.SetPosition(root.GetLength(), 0)
	s.AddBone("root", middle)

	tip := NewBone2D("tip")
	tip.SetLength(50.0)
	tip.SetIKMode(true)
	tip.SetPosition(middle.GetLength(), 0)
	s.AddBone("middle", tip)

	root.SetRotation(-45)
	middle.SetRotation(30)
	tip.SetRotation(-20)

	// Solve IK chain from tip
	s.SolveIKChain("tip", 120, 10)

	// End effector should be closer to target
	endPos := tip.GetEndPosition()
	dist := stdMath.Sqrt((float64(endPos.X)-120)*(float64(endPos.X)-120) + (float64(endPos.Y)-10)*(float64(endPos.Y)-10))
	if dist > 30 {
		t.Errorf("expected end effector within 30 units of target (120,10), got distance %f at (%f,%f)",
			dist, endPos.X, endPos.Y)
	}
}

func TestIK_SolveIKChainNonexistentBone(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	// Should not panic
	s.SolveIKChain("nonexistent", 100, 100)
}

func TestIK_SolveEmptyChain(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	// Should not panic
	s.SolveIK(nil, 100, 100)
	s.SolveIK([]*Bone2D{}, 100, 100)
}

// ---------------------------------------------------------------------------
// Animation tests
// ---------------------------------------------------------------------------

func TestSkeleton2D_PlayAnimationNoAnimator(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	ok := s.PlayAnimation("walk")
	if ok {
		t.Error("expected PlayAnimation to return false with no animator")
	}
}

func TestSkeleton2D_SetAnimator(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	ap := NewAnimationPlayer("animator")
	s.SetAnimator(ap)
	if s.Animator() != ap {
		t.Error("Animator() should return the same pointer")
	}
}

func TestSkeleton2D_StopAnimationNoAnimator(t *testing.T) {
	s := NewSkeleton2D("skeleton")
	// Should not panic
	s.StopAnimation()
}

// ---------------------------------------------------------------------------
// Integration test
// ---------------------------------------------------------------------------

func TestSkeleton2D_ComplexHierarchy(t *testing.T) {
	s := NewSkeleton2D("character")

	// Build a simple arm hierarchy
	shoulder := NewBone2D("shoulder")
	shoulder.SetLength(30.0)
	shoulder.SetIKMode(true)
	s.AddBone("", shoulder)

	upperArm := NewBone2D("upper_arm")
	upperArm.SetLength(40.0)
	upperArm.SetIKMode(true)
	upperArm.SetPosition(shoulder.GetLength(), 0)
	s.AddBone("shoulder", upperArm)

	forearm := NewBone2D("forearm")
	forearm.SetLength(35.0)
	forearm.SetIKMode(true)
	forearm.SetPosition(upperArm.GetLength(), 0)
	s.AddBone("upper_arm", forearm)

	hand := NewBone2D("hand")
	hand.SetLength(15.0)
	hand.SetIKMode(true)
	hand.SetPosition(forearm.GetLength(), 0)
	s.AddBone("forearm", hand)

	// Verify hierarchy
	if len(s.GetAllBones()) != 4 {
		t.Errorf("expected 4 bones, got %d", len(s.GetAllBones()))
	}

	// Verify parent-child relationships
	shoulderChildren := shoulder.GetChildren()
	if len(shoulderChildren) != 1 || shoulderChildren[0] != upperArm.Node2D {
		t.Error("shoulder should have upper_arm as child")
	}

	upperArmChildren := upperArm.GetChildren()
	if len(upperArmChildren) != 1 || upperArmChildren[0] != forearm.Node2D {
		t.Error("upper_arm should have forearm as child")
	}

	forearmChildren := forearm.GetChildren()
	if len(forearmChildren) != 1 || forearmChildren[0] != hand.Node2D {
		t.Error("forearm should have hand as child")
	}

	// Test bone lookups by name
	if s.GetBone("shoulder") != shoulder {
		t.Error("GetBone('shoulder') failed")
	}
	if s.GetBone("hand") != hand {
		t.Error("GetBone('hand') failed")
	}

	// Store rest pose
	s.StoreRestPose()

	// Rotate bones
	shoulder.SetRotation(45)
	upperArm.SetRotation(-30)
	forearm.SetRotation(60)
	hand.SetRotation(10)

	// Apply rest pose should reset all
	s.ApplyRestPose()

	if shoulder.GetRotation() != 0 {
		t.Errorf("expected shoulder rotation 0 after rest, got %f", shoulder.GetRotation())
	}
	if hand.GetRotation() != 0 {
		t.Errorf("expected hand rotation 0 after rest, got %f", hand.GetRotation())
	}

	// Set IK target
	s.SetIKTarget(100, 50)
	if !s.ikActive {
		t.Error("ikActive should be true after SetIKTarget")
	}

	// Solve IK chain
	s.SolveIKChain("hand", 100, 50)

	// Clear IK
	s.ClearIKTarget()
	if s.ikActive {
		t.Error("ikActive should be false after ClearIKTarget")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// absFloat returns the absolute value of a float32.
func absFloat(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
