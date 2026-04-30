package physics

// PhysicsMaterial defines friction and bounce properties for collisions.
type PhysicsMaterial struct {
	Friction  float32 // 0 = no friction, 1 = full friction
	Bounciness float32 // 0 = no bounce, 1 = full bounce
}

// DefaultPhysicsMaterial returns a standard physics material.
func DefaultPhysicsMaterial() PhysicsMaterial {
	return PhysicsMaterial{
		Friction:  0.7,
		Bounciness: 0.3,
	}
}

// StaticBody2D is immovable; optimized for level geometry.
// Does not integrate velocity; collision-only.
type StaticBody2D struct {
	*RigidBody
	Material PhysicsMaterial
}

// NewStaticBody2D creates a new StaticBody2D.
func NewStaticBody2D(entityID int, x, y, w, h float32) *StaticBody2D {
	return &StaticBody2D{
		RigidBody: NewBody(entityID, x, y, w, h, BodyStatic),
		Material:  DefaultPhysicsMaterial(),
	}
}

// SetMaterial sets the physics material.
func (s *StaticBody2D) SetMaterial(material PhysicsMaterial) {
	s.Material = material
}

// GetMaterial returns the physics material.
func (s *StaticBody2D) GetMaterial() PhysicsMaterial {
	return s.Material
}

// GetFriction returns the friction coefficient.
func (s *StaticBody2D) GetFriction() float32 {
	return s.Material.Friction
}

// SetFriction sets the friction coefficient.
func (s *StaticBody2D) SetFriction(friction float32) {
	if friction < 0 {
		friction = 0
	}
	if friction > 1 {
		friction = 1
	}
	s.Material.Friction = friction
}

// GetBounciness returns the bounciness (restitution).
func (s *StaticBody2D) GetBounciness() float32 {
	return s.Material.Bounciness
}

// SetBounciness sets the bounciness (restitution).
func (s *StaticBody2D) SetBounciness(bounciness float32) {
	if bounciness < 0 {
		bounciness = 0
	}
	if bounciness > 1 {
		bounciness = 1
	}
	s.Material.Bounciness = bounciness
}

// KScript API helpers

// GetFrictionKS returns friction as float64 for KScript.
func (s *StaticBody2D) GetFrictionKS() float64 {
	return float64(s.Material.Friction)
}

// SetFrictionKS sets friction from KScript float64.
func (s *StaticBody2D) SetFrictionKS(f float64) {
	s.SetFriction(float32(f))
}

// GetBouncinessKS returns bounciness as float64 for KScript.
func (s *StaticBody2D) GetBouncinessKS() float64 {
	return float64(s.Material.Bounciness)
}

// SetBouncinessKS sets bounciness from KScript float64.
func (s *StaticBody2D) SetBouncinessKS(b float64) {
	s.SetBounciness(float32(b))
}
