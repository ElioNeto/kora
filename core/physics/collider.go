package physics

import "math"

// Overlap holds the penetration depth and normal direction of a collision.
type Overlap struct {
	DepthX, DepthY float32
	// Normal points from B towards A (the direction A should be pushed).
	NormalX, NormalY float32
	Hit              bool
}

// TestAABB checks whether two AABB bodies overlap and returns the minimum
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

// TestCircleCircle checks collision between two circle bodies.
func TestCircleCircle(a, b *RigidBody) Overlap {
	dx := a.Pos.X - b.Pos.X
	dy := a.Pos.Y - b.Pos.Y
	distSq := dx*dx + dy*dy
	radiiSum := a.Radius + b.Radius
	if distSq > radiiSum*radiiSum {
		return Overlap{}
	}
	dist := float32(math.Sqrt(float64(distSq)))
	if dist == 0 {
		// Concentric circles, push up
		return Overlap{DepthY: a.Radius + b.Radius, NormalY: -1, Hit: true}
	}
	depth := radiiSum - dist
	nx := dx / dist
	ny := dy / dist
	return Overlap{DepthX: depth * nx, DepthY: depth * ny, NormalX: nx, NormalY: ny, Hit: true}
}

// TestCircleRect checks collision between a circle body (a) and a rect body (b).
func TestCircleRect(circle, rect *RigidBody) Overlap {
	// Closest point on rect to circle center
	closestX := clamp(circle.Pos.X, rect.Pos.X-rect.HalfW, rect.Pos.X+rect.HalfW)
	closestY := clamp(circle.Pos.Y, rect.Pos.Y-rect.HalfH, rect.Pos.Y+rect.HalfH)

	dx := circle.Pos.X - closestX
	dy := circle.Pos.Y - closestY
	distSq := dx*dx + dy*dy
	if distSq > circle.Radius*circle.Radius {
		return Overlap{}
	}
	if distSq == 0 {
		// Circle center inside rect, push out along shortest rect axis
		// Simplified: push up
		return Overlap{DepthY: circle.Radius, NormalY: -1, Hit: true}
	}
	dist := float32(math.Sqrt(float64(distSq)))
	depth := circle.Radius - dist
	nx := dx / dist
	ny := dy / dist
	return Overlap{DepthX: depth * nx, DepthY: depth * ny, NormalX: nx, NormalY: ny, Hit: true}
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ResolveCollision pushes body a out of body b using the minimum translation
// vector and zeroes the colliding velocity component. Handles all shape combinations.
func ResolveCollision(a, b *RigidBody) {
	var ov Overlap

	// Dispatch based on shape types
	switch {
	case a.Shape == ShapeRect && b.Shape == ShapeRect:
		ov = TestAABB(a, b)
	case a.Shape == ShapeCircle && b.Shape == ShapeCircle:
		ov = TestCircleCircle(a, b)
	case a.Shape == ShapeCircle && b.Shape == ShapeRect:
		ov = TestCircleRect(a, b)
	case a.Shape == ShapeRect && b.Shape == ShapeCircle:
		// Swap to use TestCircleRect with circle first
		ov = TestCircleRect(b, a)
		// Normal needs to be inverted since we swapped a and b
		ov.NormalX = -ov.NormalX
		ov.NormalY = -ov.NormalY
	default:
		return
	}

	if !ov.Hit {
		return
	}

	// Push a along the MTV
	a.Pos.X += ov.NormalX * ov.DepthX
	a.Pos.Y += ov.NormalY * ov.DepthY

	// Cancel velocity on the colliding axis
	if ov.DepthX != 0 {
		a.Vel.X = 0
		if ov.NormalX < 0 {
			a.IsTouching[0] = true // Left
		} else {
			a.IsTouching[1] = true // Right
		}
	}
	if ov.DepthY != 0 {
		a.Vel.Y = 0
		if ov.NormalY < 0 {
			a.IsTouching[2] = true // Top (player is above b, so grounded)
			a.IsGrounded = true
		} else {
			a.IsTouching[3] = true // Bottom (player is below b, ceiling)
			a.IsGrounded = false
		}
	}

	// Fire callback (if set) with collision normal
	if a.OnCollision != nil {
		normal := Vec2{ov.NormalX, ov.NormalY}
		a.OnCollision(b, normal)
	}
}
