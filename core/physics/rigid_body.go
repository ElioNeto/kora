package physics

import "github.com/ElioNeto/kora/core/math"

// RigidBody2D is a completely physics-simulated body (gravity, forces, impulses, torque).
// Wraps RigidBody with additional RigidBody2D-specific properties.
type RigidBody2D struct {
	*RigidBody

	// Additional properties
	LinearDamping  float32
	AngularDamping float32
	FreezeRotation bool
	AngularVelocity float32
}

// NewRigidBody2D creates a new RigidBody2D node.
func NewRigidBody2D(entityID int, x, y, w, h float32) *RigidBody2D {
	return &RigidBody2D{
		RigidBody:      NewBody(entityID, x, y, w, h, BodyDynamic),
		LinearDamping:  0.1,
		AngularDamping: 0.1,
		FreezeRotation: false,
	}
}

// ApplyForce applies a force vector to the body (Native Go API).
func (r *RigidBody2D) ApplyForce(force Vec2) {
	r.Vel.X += force.X / r.Mass
	r.Vel.Y += force.Y / r.Mass
}

// ApplyImpulse applies an instantaneous impulse to the body (Native Go API).
func (r *RigidBody2D) ApplyImpulse(impulse Vec2) {
	r.Vel.X += impulse.X
	r.Vel.Y += impulse.Y
}

// SetVelocity sets the linear velocity (Native Go API).
func (r *RigidBody2D) SetVelocity(vel Vec2) {
	r.Vel = vel
}

// GetVelocity returns the linear velocity (Native Go API).
func (r *RigidBody2D) GetVelocity() Vec2 {
	return r.Vel
}

// SetAngularVelocity sets the angular velocity (Native Go API).
func (r *RigidBody2D) SetAngularVelocity(omega float32) {
	r.AngularVelocity = omega
}

// GetAngularVelocity returns the angular velocity (Native Go API).
func (r *RigidBody2D) GetAngularVelocity() float32 {
	return r.AngularVelocity
}

// SetMass sets the body mass (Native Go API).
func (r *RigidBody2D) SetMass(mass float32) {
	if mass > 0 {
		r.Mass = mass
	}
}

// GetMass returns the body mass (Native Go API).
func (r *RigidBody2D) GetMass() float32 {
	return r.Mass
}

// SetGravityScale sets the gravity multiplier (Native Go API).
func (r *RigidBody2D) SetGravityScale(scale float32) {
	r.Gravity = scale
}

// GetGravityScale returns the gravity multiplier (Native Go API).
func (r *RigidBody2D) GetGravityScale() float32 {
	return r.Gravity
}

// KScript API helpers (used by compiler/kscript.go)
// These use float64 to match KScript number type

// ApplyForceKS applies force from KScript (float64 params).
func (r *RigidBody2D) ApplyForceKS(fx, fy float64) {
	r.ApplyForce(Vec2{float32(fx), float32(fy)})
}

// ApplyImpulseKS applies impulse from KScript (float64 params).
func (r *RigidBody2D) ApplyImpulseKS(ix, iy float64) {
	r.ApplyImpulse(Vec2{float32(ix), float32(iy)})
}

// SetVelocityKS sets velocity from KScript (float64 params).
func (r *RigidBody2D) SetVelocityKS(vx, vy float64) {
	r.SetVelocity(Vec2{float32(vx), float32(vy)})
}

// GetVelocityKS returns velocity as float64 for KScript.
func (r *RigidBody2D) GetVelocityKS() (float64, float64) {
	return float64(r.Vel.X), float64(r.Vel.Y)
}

// GetPositionVec returns position as math.Vector2 for node system.
func (r *RigidBody2D) GetPositionVec() math.Vector2 {
	return math.NewVector2(r.Pos.X, r.Pos.Y)
}

// SetPositionVec sets position from math.Vector2.
func (r *RigidBody2D) SetPositionVec(pos math.Vector2) {
	r.Pos.X = pos.X
	r.Pos.Y = pos.Y
}
