package physics

import "math"

// Overlap holds the penetration depth and normal direction of a collision.
type Overlap struct {
	DepthX, DepthY float32
	// Normal points from B towards A (the direction A should be pushed).
	NormalX, NormalY float32
	Hit              bool
}

// TestAABB checks whether two bodies overlap and returns the minimum
// translation vector needed to separate them.
func TestAABB(a, b *RigidBody) Overlap {
	aMinX, aMinY, aMaxX, aMaxY := a.AABB()
	bMinX, bMinY, bMaxX, bMaxY := b.AABB()

	// Early-out: no overlap at all
	if aMaxX <= bMinX || aMinX >= bMaxX || aMaxY <= bMinY || aMinY >= bMaxY {
		return Overlap{}
	}

	// Penetration on each axis
	overlapX := float32(math.Min(float64(aMaxX-bMinX), float64(bMaxX-aMinX)))
	overlapY := float32(math.Min(float64(aMaxY-bMinY), float64(bMaxY-aMinY)))

	// Resolve along the axis of least penetration
	if overlapX < overlapY {
		nx := float32(1)
		if a.Pos.X < b.Pos.X {
			nx = -1
		}
		return Overlap{DepthX: overlapX, NormalX: nx, Hit: true}
	}
	ny := float32(1)
	if a.Pos.Y < b.Pos.Y {
		ny = -1
	}
	return Overlap{DepthY: overlapY, NormalY: ny, Hit: true}
}

// ResolveAABB pushes body a out of body b using the minimum translation
// vector and zeroes the colliding velocity component.
func ResolveAABB(a, b *RigidBody) {
	ov := TestAABB(a, b)
	if !ov.Hit {
		return
	}

	// Push a along the MTV
	a.Pos.X += ov.NormalX * ov.DepthX
	a.Pos.Y += ov.NormalY * ov.DepthY

	// Cancel velocity on the colliding axis
	if ov.DepthX > 0 {
		a.Vel.X = 0
		if ov.NormalX < 0 {
			a.IsTouching[0] = true // Left
		} else {
			a.IsTouching[1] = true // Right
		}
	}
	if ov.DepthY > 0 {
		a.Vel.Y = 0
		if ov.NormalY < 0 {
			a.IsTouching[2] = true // Top
			a.IsGrounded = false
		} else {
			a.IsTouching[3] = true // Bottom (landed)
			a.IsGrounded = true
		}
	}

	// Fire callback (if set)
	if a.OnCollision != nil {
		a.OnCollision(b)
	}
}
