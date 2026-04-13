package async

// ----------------------------------------------------------------------------
// Vector2 — position/velocity helper
// ----------------------------------------------------------------------------

// Vec2 is a 2D vector with convenience methods for common game operations.
type Vec2 struct {
	X, Y float64
}

// Add returns a new vector with component-wise addition.
func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{v.X + o.X, v.Y + o.Y}
}

// Sub returns a new vector with component-wise subtraction.
func (v Vec2) Sub(o Vec2) Vec2 {
	return Vec2{v.X - o.X, v.Y - o.Y}
}

// Mul returns a new vector scaled by a scalar.
func (v Vec2) Mul(s float64) Vec2 {
	return Vec2{v.X * s, v.Y * s}
}

// Len returns the Euclidean length of the vector.
func (v Vec2) Len() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Normalized returns a unit vector in the same direction.
func (v Vec2) Normalized() Vec2 {
	l := v.Len()
	if l == 0 {
		return Vec2{}
	}
	return Vec2{v.X / l, v.Y / l}
}

// Dot returns the dot product of v and o.
func (v Vec2) Dot(o Vec2) float64 {
	return v.X*o.X + v.Y*o.Y
}

// Cross returns the 2D cross product (scalar z-component).
func (v Vec2) Cross(o Vec2) float64 {
	return v.X*o.Y - v.Y*o.X
}
