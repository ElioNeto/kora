package physics

import "math"

// SweepResult contains the result of a swept AABB test.
type SweepResult struct {
	Hit     bool
	Time    float64 // normalised time of first hit [0,1]
	NormalX float64 // collision normal X
	NormalY float64 // collision normal Y
	PointX  float64 // collision point X
	PointY  float64 // collision point Y
}

// SweptAABB tests if a moving AABB collides with a stationary AABB over [0,1] in time.
// It uses the slab method: for each axis, compute entry and exit times;
// the first entry time within [0,1] is the collision time, and the axis with
// the later entry time determines the collision normal.
func SweptAABB(minA, maxA Vec2, minB, maxB Vec2, velX, velY float64) SweepResult {
	// Check if already overlapping at t=0.
	if maxA.X > minB.X && minA.X < maxB.X && maxA.Y > minB.Y && minA.Y < maxB.Y {
		return SweepResult{
			Hit:    true,
			Time:   0,
			PointX: float64(minA.X+maxA.X) / 2,
			PointY: float64(minA.Y+maxA.Y) / 2,
		}
	}

	// No movement — cannot collide from motion.
	if math.Abs(velX) < 1e-10 && math.Abs(velY) < 1e-10 {
		return SweepResult{}
	}

	entryTime := -1.0 // start at -1 so entry at exactly 0 is still detected
	exitTime := 1.0
	normalX := 0.0
	normalY := 0.0

	// ── X axis ──────────────────────────────────────────────────────
	if math.Abs(velX) < 1e-10 {
		// Moving parallel to X slab — intervals must overlap along X.
		if maxA.X <= minB.X || minA.X >= maxB.X {
			return SweepResult{}
		}
	} else {
		invVx := 1.0 / velX
		var t1, t2 float64
		if velX > 0 {
			t1 = float64(minB.X-maxA.X) * invVx
			t2 = float64(maxB.X-minA.X) * invVx
		} else {
			t1 = float64(maxB.X-minA.X) * invVx
			t2 = float64(minB.X-maxA.X) * invVx
		}
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t1 >= entryTime {
			entryTime = t1
			if velX > 0 {
				normalX = -1 // moving right, hits left face of B
			} else {
				normalX = 1 // moving left, hits right face of B
			}
		}
		if t2 < exitTime {
			exitTime = t2
		}
		if entryTime > exitTime {
			return SweepResult{}
		}
	}

	// ── Y axis ──────────────────────────────────────────────────────
	if math.Abs(velY) < 1e-10 {
		// Moving parallel to Y slab — intervals must overlap along Y.
		if maxA.Y <= minB.Y || minA.Y >= maxB.Y {
			return SweepResult{}
		}
	} else {
		invVy := 1.0 / velY
		var t1, t2 float64
		if velY > 0 {
			t1 = float64(minB.Y-maxA.Y) * invVy
			t2 = float64(maxB.Y-minA.Y) * invVy
		} else {
			t1 = float64(maxB.Y-minA.Y) * invVy
			t2 = float64(minB.Y-maxA.Y) * invVy
		}
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t1 >= entryTime {
			entryTime = t1
			normalX = 0 // Y axis determines the normal
			if velY > 0 {
				normalY = -1 // moving down, hits top face of B
			} else {
				normalY = 1 // moving up, hits bottom face of B
			}
		}
		if t2 < exitTime {
			exitTime = t2
		}
		if entryTime > exitTime {
			return SweepResult{}
		}
	}

	// Collision must happen within [0, 1] and entry must occur before exit.
	if entryTime < 0 || entryTime > 1 || entryTime >= exitTime {
		return SweepResult{}
	}

	return SweepResult{
		Hit:     true,
		Time:    entryTime,
		NormalX: normalX,
		NormalY: normalY,
		PointX:  float64(minA.X+maxA.X)/2 + velX*entryTime,
		PointY:  float64(minA.Y+maxA.Y)/2 + velY*entryTime,
	}
}

// NeedsCCD checks if a body's velocity is high enough to warrant CCD.
// A body needs CCD when vel*dt > body's minimum dimension.
func NeedsCCD(body *RigidBody, dt float64) bool {
	speed := math.Sqrt(float64(body.Vel.X*body.Vel.X + body.Vel.Y*body.Vel.Y))
	dist := speed * dt
	var minDim float64
	switch body.Shape {
	case ShapeCircle:
		minDim = float64(body.Radius * 2)
	default: // ShapeRect
		w := float64(body.HalfW * 2)
		h := float64(body.HalfH * 2)
		if w < h {
			minDim = w
		} else {
			minDim = h
		}
	}
	return dist > minDim
}
