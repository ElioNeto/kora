package math

// Rect represents a 2D rectangle
type Rect struct {
	X, Y, W, H float32
}

// NewRect creates a new Rect with position and size
func NewRect(x, y, w, h float32) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

// Empty creates a zero-size Rect at origin
func Empty() Rect {
	return Rect{}
}

// Intersects returns true if the rectangles overlap
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}

// Contains returns true if the point is inside the rect
func (r Rect) Contains(px, py float32) bool {
	return px >= r.X && px <= r.X+r.W &&
		py >= r.Y && py <= r.Y+r.H
}

// Center returns the center point of the rect
func (r Rect) Center() Vector2 {
	return Vector2{
		X: r.X + r.W/2,
		Y: r.Y + r.H/2,
	}
}

// Union returns a rect that contains both this rect and other
func (r Rect) Union(other Rect) Rect {
	minX := min(r.X, other.X)
	minY := min(r.Y, other.Y)
	maxX := max(r.X+r.W, other.X+other.W)
	maxY := max(r.Y+r.H, other.Y+other.H)
	return Rect{
		X: minX,
		Y: minY,
		W: maxX - minX,
		H: maxY - minY,
	}
}

// min returns the smaller of two float32 values
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two float32 values
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
