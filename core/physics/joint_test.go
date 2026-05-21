package physics

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Factory tests

func TestNewDistanceJoint(t *testing.T) {
	a := dynBody(0, 0, 16, 16)
	b := dynBody(100, 0, 16, 16)
	anchorA := Vec2{8, 8}
	anchorB := Vec2{-8, -8}

	j := NewDistanceJoint(a, b, anchorA, anchorB, 50)

	if j.Type != JointDistance {
		t.Errorf("expected JointDistance, got %d", j.Type)
	}
	if j.BodyA != a {
		t.Error("BodyA mismatch")
	}
	if j.BodyB != b {
		t.Error("BodyB mismatch")
	}
	if j.AnchorA != anchorA {
		t.Errorf("AnchorA mismatch: got %v", j.AnchorA)
	}
	if j.AnchorB != anchorB {
		t.Errorf("AnchorB mismatch: got %v", j.AnchorB)
	}
	if j.Length != 50 {
		t.Errorf("expected Length 50, got %.2f", j.Length)
	}
	if j.Stiffness != DefaultDistanceStiffness {
		t.Errorf("expected default stiffness %.2f, got %.2f", DefaultDistanceStiffness, j.Stiffness)
	}
	if !j.Active {
		t.Error("expected joint to be active")
	}
}

func TestNewSpringJoint(t *testing.T) {
	a := dynBody(0, 0, 16, 16)
	b := dynBody(100, 0, 16, 16)

	j := NewSpringJoint(a, b, Vec2{}, Vec2{}, 50, 8.0, 0.3)

	if j.Type != JointSpring {
		t.Errorf("expected JointSpring, got %d", j.Type)
	}
	if j.Stiffness != 8.0 {
		t.Errorf("expected Stiffness 8.0, got %.2f", j.Stiffness)
	}
	if j.Damping != 0.3 {
		t.Errorf("expected Damping 0.3, got %.2f", j.Damping)
	}
	if j.Length != 50 {
		t.Errorf("expected Length 50, got %.2f", j.Length)
	}
}

func TestNewPinJoint(t *testing.T) {
	body := dynBody(100, 200, 16, 16)
	worldPos := Vec2{150, 250}

	j := NewPinJoint(body, worldPos)

	if j.Type != JointPin {
		t.Errorf("expected JointPin, got %d", j.Type)
	}
	if j.BodyA != body {
		t.Error("BodyA mismatch")
	}
	if j.BodyB != nil {
		t.Error("expected BodyB to be nil for pin joint")
	}

	// AnchorA should be the local offset from body centre to worldPos.
	expectedAnchor := Vec2{50, 50}
	if j.AnchorA != expectedAnchor {
		t.Errorf("expected AnchorA %v, got %v", expectedAnchor, j.AnchorA)
	}

	// AnchorB should store the target world position.
	if j.AnchorB != worldPos {
		t.Errorf("expected AnchorB (target) %v, got %v", worldPos, j.AnchorB)
	}
	if j.Stiffness != DefaultPinStiffness {
		t.Errorf("expected default pin stiffness %.2f, got %.2f", DefaultPinStiffness, j.Stiffness)
	}
}

// ---------------------------------------------------------------------------
// Joint break

func TestJointBreakForce(t *testing.T) {
	a := dynBody(0, 0, 16, 16)
	b := dynBody(200, 0, 16, 16)

	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 0)
	j.Breakable = true
	j.BreakForce = 50

	// correction = 200 - 0 = 200, forceMag = 200 * 1.0 = 200 > breakForce = 50
	j.solve(FixedPhysicsStep)

	if j.Active {
		t.Error("expected joint to break when force exceeds BreakForce")
	}
}

func TestJointNoBreakBelowThreshold(t *testing.T) {
	a := dynBody(0, 0, 16, 16)
	b := dynBody(10, 0, 16, 16)

	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 5)
	j.Breakable = true
	j.BreakForce = 100

	// correction = 10 - 5 = 5, forceMag = 5 * 1.0 = 5 < 100
	j.solve(FixedPhysicsStep)

	if !j.Active {
		t.Error("expected joint to stay active when force is below BreakForce")
	}
}

// ---------------------------------------------------------------------------
// SolveJoints correction

func TestSolveJointsCorrection(t *testing.T) {
	w := NewWorld(nil)

	// Two bodies positioned with a gap of 60 units.
	a := dynBody(0, 0, 16, 16)
	b := dynBody(60, 0, 16, 16)
	w.Register(a)
	w.Register(b)

	// Distance joint with target length of 30 (bodies are too far apart).
	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)
	j.Stiffness = 0.8
	w.AddJoint(j)

	// Record initial distance.
	initialDist := dist(a.Pos, b.Pos)

	// Step enough times for the joint to converge.
	const steps = 30
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	finalDist := dist(a.Pos, b.Pos)

	// Distance should have decreased.
	if finalDist >= initialDist {
		t.Errorf("distance should decrease after joint resolution: initial=%.2f, final=%.2f",
			initialDist, finalDist)
	}

	// After enough steps, distance should be close to target.
	if finalDist > 36 {
		t.Errorf("expected distance to approach target of 30, got %.2f after %d steps",
			finalDist, steps)
	}
}

func TestSolveSpringJointCorrection(t *testing.T) {
	w := NewWorld(nil)

	a := dynBody(0, 0, 16, 16)
	b := dynBody(60, 0, 16, 16)
	w.Register(a)
	w.Register(b)

	// Spring joint pulls bodies together with damping.
	j := NewSpringJoint(a, b, Vec2{}, Vec2{}, 30, 0.8, 0.3)
	w.AddJoint(j)

	initialDist := dist(a.Pos, b.Pos)

	const steps = 15
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	finalDist := dist(a.Pos, b.Pos)
	if finalDist >= initialDist {
		t.Errorf("spring joint should pull bodies together: initial=%.2f, final=%.2f",
			initialDist, finalDist)
	}
}

func TestPinJointCorrection(t *testing.T) {
	w := NewWorld(nil)

	// Body at (0,0), pin to world position (50,0).
	body := dynBody(0, 0, 16, 16)
	w.Register(body)

	j := NewPinJoint(body, Vec2{50, 0})
	j.Stiffness = 1.0
	w.AddJoint(j)

	// After stepping, the body should be pulled toward the pinned position.
	const steps = 30
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	// The body's position should be approaching the pin target.
	// With stiffness 1.0, the anchor (body pos + anchor offset = body pos + 50)
	// should be pulled back toward target (50,0).
	// body.Pos converges to target - anchorA = (50-50, 0-0) = (0,0)
	// So body.Pos should be near (0,0).
	if body.Pos.X < -10 || body.Pos.X > 0.5 {
		t.Errorf("pin joint should hold body near target: got Pos.X=%.2f", body.Pos.X)
	}
}

// ---------------------------------------------------------------------------
// World integration

func TestAddJoint(t *testing.T) {
	w := NewWorld(nil)
	a := dynBody(0, 0, 16, 16)
	b := dynBody(50, 0, 16, 16)
	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)

	w.AddJoint(j)

	if w.JointCount() != 1 {
		t.Errorf("expected 1 joint, got %d", w.JointCount())
	}
}

func TestRemoveJoint(t *testing.T) {
	w := NewWorld(nil)
	a := dynBody(0, 0, 16, 16)
	b := dynBody(50, 0, 16, 16)
	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)

	w.AddJoint(j)
	w.RemoveJoint(j)

	if w.JointCount() != 0 {
		t.Errorf("expected 0 joints after removal, got %d", w.JointCount())
	}
}

func TestJointCount(t *testing.T) {
	w := NewWorld(nil)
	a := dynBody(0, 0, 16, 16)
	b := dynBody(50, 0, 16, 16)
	c := dynBody(100, 0, 16, 16)

	j1 := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)
	j2 := NewDistanceJoint(b, c, Vec2{}, Vec2{}, 40)
	j3 := NewDistanceJoint(a, c, Vec2{}, Vec2{}, 50)

	w.AddJoint(j1)
	w.AddJoint(j2)
	w.AddJoint(j3)

	if w.JointCount() != 3 {
		t.Errorf("expected 3 joints, got %d", w.JointCount())
	}

	// Deactivate one.
	j2.Active = false
	if w.JointCount() != 2 {
		t.Errorf("expected 2 active joints after deactivation, got %d", w.JointCount())
	}
}

func TestInactiveJoints(t *testing.T) {
	w := NewWorld(nil)
	a := dynBody(0, 0, 16, 16)
	b := dynBody(60, 0, 16, 16)
	w.Register(a)
	w.Register(b)

	j := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)
	j.Active = false
	w.AddJoint(j)

	initialDist := dist(a.Pos, b.Pos)

	const steps = 10
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	finalDist := dist(a.Pos, b.Pos)

	// Bodies should not move towards each other (or only due to gravity / other effects).
	// With zero velocity and no gravity (X) they should stay put.
	if finalDist < initialDist-eps {
		t.Errorf("inactive joint should not pull bodies together: initial=%.2f, final=%.2f",
			initialDist, finalDist)
	}
}

func TestJointSolveOnlyActive(t *testing.T) {
	w := NewWorld(nil)
	a := dynBody(0, 0, 16, 16)
	b := dynBody(60, 0, 16, 16)
	w.Register(a)
	w.Register(b)

	active := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)
	active.Stiffness = 0.8
	inactive := NewDistanceJoint(a, b, Vec2{}, Vec2{}, 30)
	inactive.Active = false

	w.AddJoint(inactive)
	w.AddJoint(active)

	const steps = 30
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	// The active joint should have pulled bodies together.
	d := dist(a.Pos, b.Pos)
	if d > 36 {
		t.Errorf("active joint should pull bodies together, got distance %.2f", d)
	}
}

func TestJointWithStaticBody(t *testing.T) {
	w := NewWorld(nil)

	// Dynamic body connected to a static body.
	staticBody := staticBody(0, 0, 16, 16)
	dynBody := dynBody(100, 0, 16, 16)
	w.Register(staticBody)
	w.Register(dynBody)

	j := NewDistanceJoint(dynBody, staticBody, Vec2{}, Vec2{}, 20)
	j.Stiffness = 1.0
	w.AddJoint(j)

	initialDist := dist(dynBody.Pos, staticBody.Pos)

	const steps = 20
	const dt = 1.0 / 60.0
	for i := 0; i < steps; i++ {
		w.Step(dt)
	}

	// Dynamic body should be pulled toward static body.
	finalDist := dist(dynBody.Pos, staticBody.Pos)
	if finalDist >= initialDist {
		t.Errorf("dynamic body should be pulled toward static body: initial=%.2f, final=%.2f",
			initialDist, finalDist)
	}

	// Static body should not move.
	if staticBody.Pos.X != 0 || staticBody.Pos.Y != 0 {
		t.Errorf("static body should not move: got %v", staticBody.Pos)
	}
}

func TestPinJointStaticBodyNoOp(t *testing.T) {
	w := NewWorld(nil)

	// Pin joint on a static body should be a no-op (static bodies don't move).
	s := staticBody(0, 0, 16, 16)
	w.Register(s)

	j := NewPinJoint(s, Vec2{100, 100})
	w.AddJoint(j)

	const dt = 1.0 / 60.0
	w.Step(dt)
	w.Step(dt)

	if s.Pos.X != 0 || s.Pos.Y != 0 {
		t.Errorf("static body should not move from pin joint: got %v", s.Pos)
	}
}

// ---------------------------------------------------------------------------
// Helpers

// dist returns the Euclidean distance between two positions.
func dist(a, b Vec2) float32 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}
