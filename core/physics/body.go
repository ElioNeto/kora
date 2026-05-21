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

// ShapeType defines the collision shape of the body.
type ShapeType int

const (
	// ShapeRect uses AABB collision (HalfW, HalfH).
	ShapeRect ShapeType = iota
	// ShapeCircle uses circle collision (Radius).
	ShapeCircle
)

// RigidBody holds the physical state of one entity.
type RigidBody struct {
	// Identity
	EntityID int

	// Transform (centre of the shape)
	Pos Vec2

	// Shape
	Shape  ShapeType
	HalfW, HalfH float32 // For ShapeRect
	Radius     float32    // For ShapeCircle

	// Dynamics
	Vel     Vec2
	Mass    float32 // kg; unused for Static bodies
	Gravity float32 // multiplier applied to world gravity (1 = normal, 0 = no gravity)

	Type BodyType

	// Collision filtering (16-bit layers)
	Layer uint16 // Which collision layer this body belongs to
	Mask  uint16 // Which layers this body collides with (bitmask)

	// State flags (set by PhysicsWorld each step)
	IsGrounded bool
	IsTouching [4]bool // Left, Right, Top, Bottom

	// Callbacks — set by KScript runtime binding
	OnCollision func(other *RigidBody, normal Vec2)

	// NodeRef stores a reference back to the game node (set by node package).
	// Used by Area2D overlap detection to map physics bodies → game nodes.
	NodeRef interface{}
}

// AABB returns the axis-aligned bounding box corners for b.
func (b *RigidBody) AABB() (minX, minY, maxX, maxY float32) {
	switch b.Shape {
	case ShapeCircle:
		return b.Pos.X - b.Radius,
			b.Pos.Y - b.Radius,
			b.Pos.X + b.Radius,
			b.Pos.Y + b.Radius
	default: // ShapeRect
		return b.Pos.X - b.HalfW,
			b.Pos.Y - b.HalfH,
			b.Pos.X + b.HalfW,
			b.Pos.Y + b.HalfH
	}
}

// Default collision layer and mask values
const (
	DefaultLayer uint16 = 1 << 0 // Layer 0
	DefaultMask  uint16 = 0xFFFF // Collides with all layers
)

// NewBody creates a RigidBody centred at (x, y) with the given dimensions (rect shape).
func NewBody(entityID int, x, y, w, h float32, bt BodyType) *RigidBody {
	return &RigidBody{
		EntityID: entityID,
		Pos:      Vec2{x, y},
		Shape:    ShapeRect,
		HalfW:    w / 2,
		HalfH:    h / 2,
		Mass:     1,
		Gravity:  1,
		Type:     bt,
		Layer:    DefaultLayer,
		Mask:     DefaultMask,
	}
}

// NewCircleBody creates a circle-shaped RigidBody.
func NewCircleBody(entityID int, x, y, radius float32, bt BodyType) *RigidBody {
	return &RigidBody{
		EntityID: entityID,
		Pos:      Vec2{x, y},
		Shape:    ShapeCircle,
		Radius:   radius,
		Mass:     1,
		Gravity:  1,
		Type:     bt,
		Layer:    DefaultLayer,
		Mask:     DefaultMask,
	}
}

// ApplyForce applies a continuous force (acceleration) to the body.
// Force is in px/s², applied over one physics step.
func (b *RigidBody) ApplyForce(force Vec2) {
	if b.Mass <= 0 {
		return
	}
	// Simplified: apply as acceleration (vel += force * dt, but dt is fixed 1/60)
	// For now, add directly to velocity (assumes force is impulse-like for simplicity)
	b.Vel.X += force.X / b.Mass
	b.Vel.Y += force.Y / b.Mass
}

// ApplyImpulse applies an instant velocity change to the body.
func (b *RigidBody) ApplyImpulse(impulse Vec2) {
	b.Vel.X += impulse.X
	b.Vel.Y += impulse.Y
}
