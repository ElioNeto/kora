// Package physics implements a lightweight 2D physics system for Kora.
// It provides RigidBody, AABB collision detection/resolution and
// a PhysicsWorld that drives the simulation each game tick.
package physics

// Vec2 is a 2-component float vector.
type Vec2 struct {
	X, Y float32
}

// Add returns the component-wise sum.
func (v Vec2) Add(o Vec2) Vec2 { return Vec2{v.X + o.X, v.Y + o.Y} }

// Scale multiplies both components by s.
func (v Vec2) Scale(s float32) Vec2 { return Vec2{v.X * s, v.Y * s} }

// BodyType controls how the physics world treats a body.
type BodyType int

const (
	// BodyDynamic is affected by gravity and resolves collisions.
	BodyDynamic BodyType = iota
	// BodyKinematic moves via velocity only (no gravity).
	BodyKinematic
	// BodyStatic never moves; used for terrain and walls.
	BodyStatic
)

// RigidBody holds the physical state of one entity.
type RigidBody struct {
	// Identity
	EntityID int

	// Transform (centre of the bounding box)
	Pos Vec2

	// Half-extents of the AABB
	HalfW, HalfH float32

	// Dynamics
	Vel     Vec2
	Mass    float32 // kg; unused for Static bodies
	Gravity float32 // multiplier applied to world gravity (1 = normal, 0 = no gravity)

	Type BodyType

	// State flags (set by PhysicsWorld each step)
	IsGrounded bool
	IsTouching [4]bool // Left, Right, Top, Bottom

	// Callbacks — set by KScript runtime binding
	OnCollision func(other *RigidBody)
}

// AABB returns the axis-aligned bounding box corners for b.
func (b *RigidBody) AABB() (minX, minY, maxX, maxY float32) {
	return b.Pos.X - b.HalfW,
		b.Pos.Y - b.HalfH,
		b.Pos.X + b.HalfW,
		b.Pos.Y + b.HalfH
}

// NewBody creates a RigidBody centred at (x, y) with the given dimensions.
func NewBody(entityID int, x, y, w, h float32, bt BodyType) *RigidBody {
	return &RigidBody{
		EntityID: entityID,
		Pos:      Vec2{x, y},
		HalfW:    w / 2,
		HalfH:    h / 2,
		Mass:     1,
		Gravity:  1,
		Type:     bt,
	}
}
