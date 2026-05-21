package physics

import "math"

// JointType identifies the type of constraint.
type JointType int

const (
	// JointDistance maintains a fixed distance between two anchor points.
	JointDistance JointType = iota
	// JointSpring acts like a spring with stiffness and damping.
	JointSpring
	// JointPin pins a body to a fixed world position.
	JointPin
)

// Default stiffness values used by factory functions.
const (
	DefaultDistanceStiffness float32 = 1.0
	DefaultSpringStiffness   float32 = 5.0
	DefaultSpringDamping     float32 = 0.5
	DefaultPinStiffness      float32 = 10.0
)

// Joint connects two bodies with a constraint.
// For PinJoint, BodyB is nil and AnchorB stores the target world position.
type Joint struct {
	Type JointType
	BodyA *RigidBody
	BodyB *RigidBody

	// AnchorA is the local offset from BodyA's centre to the anchor point.
	AnchorA Vec2
	// AnchorB is the local offset from BodyB's centre to the anchor point.
	// For PinJoint, AnchorB stores the fixed target world position instead.
	AnchorB Vec2

	Length    float32 // target distance (for Distance / Spring)
	Stiffness float32 // spring constant (for Spring and Distance)
	Damping   float32 // damping factor (for Spring)

	Breakable  bool
	BreakForce float32 // max force before joint breaks
	Active     bool
}

// ---------------------------------------------------------------------------
// Factory functions

// NewDistanceJoint creates a joint that maintains a fixed distance between
// two anchor points on two bodies.
func NewDistanceJoint(a, b *RigidBody, anchorA, anchorB Vec2, length float32) *Joint {
	return &Joint{
		Type:      JointDistance,
		BodyA:     a,
		BodyB:     b,
		AnchorA:   anchorA,
		AnchorB:   anchorB,
		Length:    length,
		Stiffness: DefaultDistanceStiffness,
		Active:    true,
	}
}

// NewSpringJoint creates a spring joint with configurable stiffness and
// damping.  Higher stiffness = stiffer spring; damping reduces oscillation.
func NewSpringJoint(a, b *RigidBody, anchorA, anchorB Vec2, length, stiffness, damping float32) *Joint {
	return &Joint{
		Type:      JointSpring,
		BodyA:     a,
		BodyB:     b,
		AnchorA:   anchorA,
		AnchorB:   anchorB,
		Length:    length,
		Stiffness: stiffness,
		Damping:   damping,
		Active:    true,
	}
}

// NewPinJoint pins a body so that a specific point on the body stays at a
// fixed world position.  BodyB is nil for pin joints.
func NewPinJoint(body *RigidBody, worldPos Vec2) *Joint {
	// Local offset from body centre to the pinned point.
	anchorA := Vec2{worldPos.X - body.Pos.X, worldPos.Y - body.Pos.Y}
	return &Joint{
		Type:      JointPin,
		BodyA:     body,
		AnchorA:   anchorA,
		AnchorB:   worldPos, // stores the fixed target world position
		Stiffness: DefaultPinStiffness,
		Active:    true,
	}
}

// ---------------------------------------------------------------------------
// Joint solving (called by PhysicsWorld.SolveJoints)

// solve resolves this joint for one physics step.
// It computes the constraint impulse and applies it to the connected bodies.
func (j *Joint) solve(dt float32) {
	if !j.Active {
		return
	}

	switch j.Type {
	case JointDistance:
		j.solveDistance()
	case JointSpring:
		j.solveSpring()
	case JointPin:
		j.solvePin()
	}
}

// solveDistance applies a correction impulse proportional to the distance
// error (current – target) times stiffness.
func (j *Joint) solveDistance() {
	worldA := j.BodyA.Pos.Add(j.AnchorA)
	worldB := j.BodyB.Pos.Add(j.AnchorB)

	dx := worldB.X - worldA.X
	dy := worldB.Y - worldA.Y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if dist < 0.0001 {
		return // anchors coincident – nothing to do
	}

	// Unit vector from A to B.
	nx := dx / dist
	ny := dy / dist

	correction := dist - j.Length // positive when stretched
	forceMag := correction * j.Stiffness

	j.applyImpulse(forceMag, nx, ny)
}

// solveSpring applies a spring force (Hooke's law) with velocity damping.
func (j *Joint) solveSpring() {
	worldA := j.BodyA.Pos.Add(j.AnchorA)
	worldB := j.BodyB.Pos.Add(j.AnchorB)

	dx := worldB.X - worldA.X
	dy := worldB.Y - worldA.Y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if dist < 0.0001 {
		return
	}

	nx := dx / dist
	ny := dy / dist

	// Relative velocity along the anchor-to-anchor axis (B relative to A).
	relVel := Vec2{j.BodyB.Vel.X - j.BodyA.Vel.X, j.BodyB.Vel.Y - j.BodyA.Vel.Y}
	relVelAlong := relVel.X*nx + relVel.Y*ny

	// Hooke's law: F = stiffness * displacement + damping * relative velocity.
	correction := dist - j.Length
	forceMag := correction*j.Stiffness + relVelAlong*j.Damping

	j.applyImpulse(forceMag, nx, ny)
}

// solvePin pulls BodyA's anchor point back to the fixed world position
// stored in AnchorB.
func (j *Joint) solvePin() {
	if j.BodyA.Type == BodyStatic || j.BodyA.Mass <= 0 {
		return
	}

	// Current world position of the anchor point on the body.
	current := j.BodyA.Pos.Add(j.AnchorA)
	target := j.AnchorB

	// Displacement from target.
	dx := target.X - current.X
	dy := target.Y - current.Y

	// Apply a spring-like force back to the target.
	j.BodyA.ApplyForce(Vec2{dx * j.Stiffness, dy * j.Stiffness})
}

// applyImpulse distributes an impulse of magnitude `mag` along the normal
// (nx, ny) across the two bodies, respecting their masses and types.
func (j *Joint) applyImpulse(mag, nx, ny float32) {
	fx := nx * mag
	fy := ny * mag

	if j.BodyA != nil && j.BodyA.Type != BodyStatic && j.BodyA.Mass > 0 {
		j.BodyA.ApplyForce(Vec2{fx, fy})
	}
	if j.BodyB != nil && j.BodyB.Type != BodyStatic && j.BodyB.Mass > 0 {
		j.BodyB.ApplyForce(Vec2{-fx, -fy})
	}

	// Check break threshold.
	if j.Breakable && mag < 0 {
		mag = -mag
	}
	if j.Breakable && mag > j.BreakForce {
		j.Active = false
	}
}
