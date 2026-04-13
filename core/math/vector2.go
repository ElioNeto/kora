// Package math provides 2D math primitives for the engine
package math

import "math"

// Vector2 represents a 2D vector with X and Y coordinates
type Vector2 struct {
	X, Y float32
}

// NewVector2 creates a new Vector2
func NewVector2(x, y float32) Vector2 {
	return Vector2{X: x, Y: y}
}

// Add returns the sum of two vectors
func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference between two vectors
func (v Vector2) Sub(other Vector2) Vector2 {
	return Vector2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul returns the vector scaled by a float value
func (v Vector2) Mul(s float32) Vector2 {
	return Vector2{X: v.X * s, Y: v.Y * s}
}

// Length returns the length of the vector
func (v Vector2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X) + float64(v.Y*v.Y)))
}

// LengthSq returns the squared length of the vector (faster than Length)
func (v Vector2) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y
}

// Normalize returns a normalized (unit length) vector
func (v Vector2) Normalize() Vector2 {
	length := v.Length()
	if length == 0 {
		return Vector2{X: 1, Y: 0}
	}
	return Vector2{X: v.X / length, Y: v.Y / length}
}

// Dot returns the dot product of two vectors
func (v Vector2) Dot(other Vector2) float32 {
	return v.X*other.X + v.Y*other.Y
}

// Lerp returns a vector linearly interpolated between v1 and v2
func Lerp(v1, v2 Vector2, t float64) Vector2 {
	return Vector2{
		X: float32(lerp(float64(v1.X), float64(v2.X), t)),
		Y: float32(lerp(float64(v1.Y), float64(v2.Y), t)),
	}
}

// lerp performs linear interpolation between a and b
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
