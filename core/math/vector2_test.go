package math

import "testing"

func TestNewVector2(t *testing.T) {
	v := NewVector2(3, 4)
	if v.X != 3 || v.Y != 4 {
		t.Errorf("Expected (3,4), got (%f, %f)", v.X, v.Y)
	}
}

func TestVectorAdd(t *testing.T) {
	a := Vector2{X: 1, Y: 2}
	b := Vector2{X: 3, Y: 4}
	sum := a.Add(b)
	if sum.X != 4 || sum.Y != 6 {
		t.Errorf("Expected (4,6), got (%f,%f)", sum.X, sum.Y)
	}
}

func TestVectorSub(t *testing.T) {
	a := Vector2{X: 5, Y: 8}
	b := Vector2{X: 2, Y: 3}
	diff := a.Sub(b)
	if diff.X != 3 || diff.Y != 5 {
		t.Errorf("Expected (3,5), got (%f,%f)", diff.X, diff.Y)
	}
}

func TestVectorMul(t *testing.T) {
	v := Vector2{X: 2, Y: 3}
	scaled := v.Mul(2.0)
	if scaled.X != 4 || scaled.Y != 6 {
		t.Errorf("Expected (4,6), got (%f,%f)", scaled.X, scaled.Y)
	}
}

func TestVectorLength(t *testing.T) {
	v := Vector2{X: 3, Y: 4}
	len := v.Length()
	expected := float32(5.0)
	if len != expected {
		t.Errorf("Expected length %f, got %f", expected, len)
	}
}

func TestVectorNormalize(t *testing.T) {
	v := Vector2{X: 3, Y: 4}
	normalized := v.Normalize()
	expectedLen := float32(1.0)
	if abs(normalized.X-0.6) > 0.001 || abs(normalized.Y-0.8) > 0.001 {
		t.Errorf("Normalized vector incorrect: (%f,%f)", normalized.X, normalized.Y)
	}
	if abs(normalized.Length()-expectedLen) > 0.001 {
		t.Errorf("Normalized vector should have length 1, got %f", normalized.Length())
	}
}

func TestVectorDot(t *testing.T) {
	a := Vector2{X: 1, Y: 2}
	b := Vector2{X: 3, Y: 4}
	dot := a.Dot(b)
	if dot != 11 {
		t.Errorf("Expected dot product 11, got %f", dot)
	}
}

func TestLerp(t *testing.T) {
	v1 := Vector2{X: 0, Y: 0}
	v2 := Vector2{X: 10, Y: 20}
	lerped := Lerp(v1, v2, 0.5)
	if lerped.X != 5 || lerped.Y != 10 {
		t.Errorf("Expected (5,10), got (%f,%f)", lerped.X, lerped.Y)
	}
}

func TestLerpEndPoints(t *testing.T) {
	v1 := Vector2{X: 100, Y: 200}
	v2 := Vector2{X: 300, Y: 400}

	atStart := Lerp(v1, v2, 0.0)
	if atStart.X != 100 || atStart.Y != 200 {
		t.Errorf("t=0 should return v1: (%f,%f)", atStart.X, atStart.Y)
	}

	atEnd := Lerp(v1, v2, 1.0)
	if atEnd.X != 300 || atEnd.Y != 400 {
		t.Errorf("t=1 should return v2: (%f,%f)", atEnd.X, atEnd.Y)
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
