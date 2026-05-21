package physics

import (
	"math"
	"testing"
)

// ------------------------------------------------------------------ SweptAABB

func TestSweptAABBNoCollisionMovingAway(t *testing.T) {
	// A at (0,0) half 5×5, B at (20,0) half 5×5, A moving left (away from B).
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{15, -5}
	maxB := Vec2{25, 5}

	res := SweptAABB(minA, maxA, minB, maxB, -10, 0)
	if res.Hit {
		t.Errorf("expected no hit when moving away, got %+v", res)
	}
}

func TestSweptAABBDirectHit(t *testing.T) {
	// A at (0,0) half 5×5, B at (20,0) half 5×5, A moving right at 100 px/s.
	// A's right edge (5) → B's left edge (15), distance 10, t = 10/100 = 0.1.
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{15, -5}
	maxB := Vec2{25, 5}

	res := SweptAABB(minA, maxA, minB, maxB, 100, 0)
	if !res.Hit {
		t.Fatal("expected hit")
	}
	if math.Abs(res.Time-0.1) > 1e-6 {
		t.Errorf("expected time 0.1, got %f", res.Time)
	}
	if res.NormalX != -1 || res.NormalY != 0 {
		t.Errorf("expected normal (-1,0), got (%f,%f)", res.NormalX, res.NormalY)
	}
	if math.Abs(res.PointX-10) > 1e-6 || math.Abs(res.PointY-0) > 1e-6 {
		t.Errorf("expected point (10,0), got (%f,%f)", res.PointX, res.PointY)
	}
}

func TestSweptAABBCornerHit(t *testing.T) {
	// A at (0,0) half 5×5, B at (20,20) half 5×5.
	// A moves right at 100 and down at 50.
	// X entry: (15-5)/100 = 0.1, Y entry: (15-5)/50 = 0.2.
	// Y has later entry → normal should be (0,-1), time = 0.2.
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{15, 15}
	maxB := Vec2{25, 25}

	res := SweptAABB(minA, maxA, minB, maxB, 100, 50)
	if !res.Hit {
		t.Fatal("expected hit")
	}
	if math.Abs(res.Time-0.2) > 1e-6 {
		t.Errorf("expected time 0.2, got %f", res.Time)
	}
	if res.NormalX != 0 || res.NormalY != -1 {
		t.Errorf("expected normal (0,-1), got (%f,%f)", res.NormalX, res.NormalY)
	}
	// At t=0.2: Point = (0 + 100*0.2, 0 + 50*0.2) = (20, 10)
	if math.Abs(res.PointX-20) > 1e-6 || math.Abs(res.PointY-10) > 1e-6 {
		t.Errorf("expected point (20,10), got (%f,%f)", res.PointX, res.PointY)
	}
}

func TestSweptAABBExactEdge(t *testing.T) {
	// A at (0,0) half 5×5, B at (10,0) half 5×5.
	// A's right edge (5) is exactly flush with B's left edge (5).
	// Moving right at 10 px/s → should hit at t=0.
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{5, -5}
	maxB := Vec2{15, 5}

	res := SweptAABB(minA, maxA, minB, maxB, 10, 0)
	if !res.Hit {
		t.Fatal("expected hit for exact edge touching (moving towards)")
	}
	if res.Time != 0 {
		t.Errorf("expected time 0 for exact edge, got %f", res.Time)
	}
	if res.NormalX != -1 || res.NormalY != 0 {
		t.Errorf("expected normal (-1,0), got (%f,%f)", res.NormalX, res.NormalY)
	}
}

func TestSweptAABBNegativeVelocity(t *testing.T) {
	// A at (0,0) half 5×5, B at (-20,0) half 5×5.
	// A moves left at 100 px/s.
	// X entry: (-15-5)/-100 → actually:
	//   velX < 0: t1 = (maxB.X - minA.X) / velX = (-15 - (-5)) / (-100) = -10/-100 = 0.1
	// Normal should be (1, 0) since hitting B's right face.
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{-25, -5}
	maxB := Vec2{-15, 5}

	res := SweptAABB(minA, maxA, minB, maxB, -100, 0)
	if !res.Hit {
		t.Fatal("expected hit with negative velocity")
	}
	if math.Abs(res.Time-0.1) > 1e-6 {
		t.Errorf("expected time 0.1, got %f", res.Time)
	}
	if res.NormalX != 1 || res.NormalY != 0 {
		t.Errorf("expected normal (1,0), got (%f,%f)", res.NormalX, res.NormalY)
	}
	// Point = (0 + (-100)*0.1, 0 + 0) = (-10, 0)
	if math.Abs(res.PointX+10) > 1e-6 || math.Abs(res.PointY-0) > 1e-6 {
		t.Errorf("expected point (-10,0), got (%f,%f)", res.PointX, res.PointY)
	}
}

func TestSweptAABBZeroVelocity(t *testing.T) {
	// A and B separated, zero velocity → no hit from movement.
	minA := Vec2{-5, -5}
	maxA := Vec2{5, 5}
	minB := Vec2{15, -5}
	maxB := Vec2{25, 5}

	res := SweptAABB(minA, maxA, minB, maxB, 0, 0)
	if res.Hit {
		t.Errorf("expected no hit with zero velocity, got %+v", res)
	}
}

// ------------------------------------------------------------------ NeedsCCD

func TestNeedsCCDFastBody(t *testing.T) {
	body := NewBody(0, 0, 0, 16, 16, BodyDynamic) // half 8×8, minDim = 16
	body.Vel = Vec2{1000, 0}
	const dt = 1.0 / 60.0

	if !NeedsCCD(body, dt) {
		t.Error("expected NeedsCCD true for fast body (1000 px/s)")
	}
}

func TestNeedsCCDSlowBody(t *testing.T) {
	body := NewBody(0, 0, 0, 16, 16, BodyDynamic) // half 8×8, minDim = 16
	body.Vel = Vec2{100, 0}
	const dt = 1.0 / 60.0

	if NeedsCCD(body, dt) {
		t.Error("expected NeedsCCD false for slow body (100 px/s)")
	}
}

func TestNeedsCCDZeroVelocity(t *testing.T) {
	body := NewBody(0, 0, 0, 16, 16, BodyDynamic)
	body.Vel = Vec2{0, 0}
	const dt = 1.0 / 60.0

	if NeedsCCD(body, dt) {
		t.Error("expected NeedsCCD false for zero velocity")
	}
}

// ------------------------------------------------------------------ Integration: CCD via PhysicsWorld

// TestCCDPreventsTunneling verifies that a fast-moving body that would tunnel
// through a thin wall gets stopped by CCD.
func TestCCDPreventsTunneling(t *testing.T) {
	w := NewWorld(nil)
	w.SetCCDThreshold(10) // enable CCD

	wall := staticBody(50, 0, 8, 200) // 8 px wide wall at x=50
	ball := dynBody(0, 0, 16, 16)     // starts at x=0, moving right at 3000 px/s
	ball.Vel.X = 3000
	w.Register(ball)
	w.Register(wall)

	const dt = 1.0 / 60.0
	for i := 0; i < 10; i++ {
		w.Step(dt)
	}

	// The ball should NOT have passed through the wall.
	// Wall's left edge is at 50-4 = 46, wall's right edge is at 50+4 = 54.
	// Ball should be somewhere near or before the wall's left edge.
	_, _, ballMaxX, _ := ball.AABB()
	wallMinX, _, _, _ := wall.AABB()
	if ballMaxX > wallMinX+5 {
		t.Errorf("ball tunneled through wall: ball right edge=%.2f, wall left edge=%.2f",
			ballMaxX, wallMinX)
	}
}

// TestCCDThresholdDisabled checks that CCD does nothing when threshold is 0.
func TestCCDThresholdDisabled(t *testing.T) {
	w := NewWorld(nil)
	w.SetCCDThreshold(0) // disabled

	wall := staticBody(50, 0, 8, 200)
	ball := dynBody(0, 0, 16, 16)
	ball.Vel.X = 3000
	w.Register(ball)
	w.Register(wall)

	const dt = 1.0 / 60.0
	for i := 0; i < 10; i++ {
		w.Step(dt)
	}

	// Without CCD the ball would pass through the thin wall.
	_, _, ballMaxX, _ := ball.AABB()
	wallMinX, _, _, _ := wall.AABB()
	if ballMaxX <= wallMinX {
		t.Errorf("ball should have passed through wall without CCD, ball right=%.2f wall left=%.2f",
			ballMaxX, wallMinX)
	}
}
