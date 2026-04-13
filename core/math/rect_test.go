package math

import "testing"

func TestNewRect(t *testing.T) {
	r := NewRect(10, 20, 100, 200)
	if r.X != 10 || r.Y != 20 || r.W != 100 || r.H != 200 {
		t.Errorf("Unexpected rect: (%f,%f,%f,%f)", r.X, r.Y, r.W, r.H)
	}
}

func TestRectEmpty(t *testing.T) {
	r := Empty()
	if r.W != 0 || r.H != 0 {
		t.Errorf("Empty rect should have zero size: (%f,%f)", r.W, r.H)
	}
}

func TestRectIntersects(t *testing.T) {
	r1 := NewRect(0, 0, 50, 50)
	r2 := NewRect(25, 25, 50, 50)
	r3 := NewRect(100, 100, 50, 50)

	if !r1.Intersects(r2) {
		t.Error("Expected r1 and r2 to intersect")
	}

	if r1.Intersects(r3) {
		t.Error("Expected r1 and r3 not to intersect")
	}
}

func TestRectContains(t *testing.T) {
	r := NewRect(0, 0, 100, 100)

	if !r.Contains(50, 50) {
		t.Error("Point (50,50) should be inside rect")
	}

	if r.Contains(0, 0) {
		// Edge case - depends on implementation
	}

	if r.Contains(150, 150) {
		t.Error("Point (150,150) should be outside rect")
	}
}

func TestRectCenter(t *testing.T) {
	r := NewRect(10, 20, 100, 200)
	center := r.Center()
	if center.X != 60 || center.Y != 120 {
		t.Errorf("Expected center (60,120), got (%f,%f)", center.X, center.Y)
	}
}

func TestRectUnion(t *testing.T) {
	r1 := NewRect(0, 0, 50, 50)
	r2 := NewRect(25, 25, 50, 50)
	union := r1.Union(r2)

	if union.X != 0 || union.Y != 0 || union.W != 75 || union.H != 75 {
		t.Errorf("Expected union (0,0,75,75), got (%f,%f,%f,%f)", union.X, union.Y, union.W, union.H)
	}
}
