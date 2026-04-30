package physics

import "math"

// RaycastHit stores information about a raycast hit.
type RaycastHit struct {
	Body     *RigidBody
	Point    Vec2
	Normal   Vec2
	Fraction float32 // 0..1, position along the ray
	Hit      bool
}

// Raycast fires a ray from 'from' to 'to' and returns the first body hit
// that matches the given collision mask.
func (w *PhysicsWorld) Raycast(from, to Vec2, mask uint16) RaycastHit {
	dir := to.Add(from.Scale(-1))
	bestHit := RaycastHit{Hit: false}
	bestFraction := float32(1.0)

	for _, b := range w.bodies {
		// Check layer/mask
		if (b.Layer & mask) == 0 {
			continue
		}

		// AABB-ray intersection using slab method
		minX, minY, maxX, maxY := b.AABB()
		tmin := float32(0)
		tmax := float32(1)
		normal := Vec2{}

		// X axis
		if math.Abs(float64(dir.X)) < 1e-6 {
			// Ray parallel to X slab
			if from.X < minX || from.X > maxX {
				continue
			}
		} else {
			invDirX := 1.0 / float64(dir.X)
			t1 := float32(float64(minX-from.X) * invDirX)
			t2 := float32(float64(maxX-from.X) * invDirX)
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			if t1 > tmin {
				tmin = t1
				normal = Vec2{-1, 0}
			}
			if t2 < tmax {
				tmax = t2
				normal = Vec2{1, 0}
			}
			if tmin > tmax {
				continue
			}
		}

		// Y axis
		if math.Abs(float64(dir.Y)) < 1e-6 {
			if from.Y < minY || from.Y > maxY {
				continue
			}
		} else {
			invDirY := 1.0 / float64(dir.Y)
			t1 := float32(float64(minY-from.Y) * invDirY)
			t2 := float32(float64(maxY-from.Y) * invDirY)
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			if t1 > tmin {
				tmin = t1
				normal = Vec2{0, -1}
			}
			if t2 < tmax {
				tmax = t2
				normal = Vec2{0, 1}
			}
			if tmin > tmax {
				continue
			}
		}

		if tmin < bestFraction && tmin >= 0 {
			bestFraction = tmin
			bestHit = RaycastHit{
				Body:     b,
				Point:    Vec2{from.X + dir.X*tmin, from.Y + dir.Y*tmin},
				Normal:   normal,
				Fraction: tmin,
				Hit:      true,
			}
		}
	}

	return bestHit
}
