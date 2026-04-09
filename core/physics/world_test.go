package physics

import (
	"testing"
)

// ------------------------------------------------------------------ helpers

func dynBody(x, y, w, h float32) *RigidBody {
	return NewBody(0, x, y, w, h, BodyDynamic)
}
func staticBody(x, y, w, h float32) *RigidBody {
	return NewBody(1, x, y, w, h, BodyStatic)
}

const eps = float32(0.5) // pixel tolerance for float comparisons

func near(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

// ------------------------------------------------------------------ Vec2

func TestVec2Add(t *testing.T) {
	v := Vec2{1, 2}.Add(Vec2{3, 4})
	if v.X != 4 || v.Y != 6 {
		t.Errorf("Add: got %v", v)
	}
}

func TestVec2Scale(t *testing.T) {
	v := Vec2{2, 5}.Scale(3)
	if v.X != 6 || v.Y != 15 {
		t.Errorf("Scale: got %v", v)
	}
}

// ------------------------------------------------------------------ AABB

func TestAABBNoOverlap(t *testing.T) {
	a := dynBody(0, 0, 32, 32)
	b := staticBody(100, 0, 32, 32)
	ov := TestAABB(a, b)
	if ov.Hit {
		t.Error("expected no hit")
	}
}

func TestAABBOverlap(t *testing.T) {
	a := dynBody(0, 0, 32, 32)
	b := staticBody(20, 0, 32, 32)
	ov := TestAABB(a, b)
	if !ov.Hit {
		t.Error("expected hit")
	}
}

// ------------------------------------------------------------------ Gravity / free fall

func TestFreefall(t *testing.T) {
	w := NewWorld(nil)
	b := dynBody(0, 0, 16, 16)
	w.Register(b)

	const dt = 1.0 / 60.0
	for i := 0; i < 60; i++ {
		w.Step(dt)
	}
	// After 1 second of free fall, Y should have increased (down = +Y)
	if b.Pos.Y <= 0 {
		t.Errorf("expected downward movement, got Y=%.2f", b.Pos.Y)
	}
}

func TestTerminalVelocity(t *testing.T) {
	w := NewWorld(nil)
	b := dynBody(0, 0, 16, 16)
	w.Register(b)

	// Run for 10 simulated seconds
	const dt = 1.0 / 60.0
	for i := 0; i < 600; i++ {
		w.Step(dt)
	}
	if b.Vel.Y > TerminalVelocity+eps {
		t.Errorf("velocity exceeds terminal: %.2f", b.Vel.Y)
	}
}

// ------------------------------------------------------------------ Static body collision

func TestLandOnStaticBody(t *testing.T) {
	w := NewWorld(nil)
	ground := staticBody(0, 100, 200, 20) // top edge at Y=90
	player := dynBody(0, 50, 16, 16)      // starts above ground
	player.Vel.Y = 300
	w.Register(player)
	w.Register(ground)

	const dt = 1.0 / 60.0
	for i := 0; i < 120; i++ {
		w.Step(dt)
		if player.IsGrounded {
			break
		}
	}
	if !player.IsGrounded {
		t.Error("player should be grounded after landing on static body")
	}
	if player.Vel.Y > eps {
		t.Errorf("downward velocity should be zeroed after landing, got %.2f", player.Vel.Y)
	}
}

func TestNoPassThroughFloor(t *testing.T) {
	w := NewWorld(nil)
	floor := staticBody(0, 200, 400, 10)
	ball := dynBody(0, 100, 16, 16)
	ball.Vel.Y = 1000 // fast drop
	w.Register(ball)
	w.Register(floor)

	const dt = 1.0 / 60.0
	for i := 0; i < 60; i++ {
		w.Step(dt)
	}
	_, _, _, maxY := ball.AABB()
	_, floorMinY, _, _ := floor.AABB()
	if maxY > floorMinY+eps {
		t.Errorf("ball passed through floor: ball bottom=%.2f, floor top=%.2f", maxY, floorMinY)
	}
}

func TestWallCollision(t *testing.T) {
	w := NewWorld(nil)
	wall := staticBody(80, 0, 10, 200) // left face at X=75
	player := dynBody(0, 0, 16, 16)
	player.Vel.X = 500
	w.Register(player)
	w.Register(wall)

	const dt = 1.0 / 60.0
	for i := 0; i < 60; i++ {
		w.Step(dt)
	}
	if player.Vel.X > eps {
		t.Errorf("horizontal velocity should be zeroed against wall, got %.2f", player.Vel.X)
	}
}

// ------------------------------------------------------------------ Tilemap collision

func TestTilemapFloor(t *testing.T) {
	const tileSize = 32
	// Solid row at Y=192..224 (tile row 6)
	solidY := float32(6 * tileSize) // 192
	tileQ := func(px, py float32) bool {
		return py >= solidY && py < solidY+tileSize
	}
	w := NewWorld(tileQ)
	b := dynBody(16, 0, 16, 16)
	b.Vel.Y = 400
	w.Register(b)

	const dt = 1.0 / 60.0
	for i := 0; i < 120; i++ {
		w.Step(dt)
		if b.IsGrounded {
			break
		}
	}
	if !b.IsGrounded {
		t.Error("body should be grounded on solid tile")
	}
}

// ------------------------------------------------------------------ Kinematic body (no gravity)

func TestKinematicNoGravity(t *testing.T) {
	w := NewWorld(nil)
	b := NewBody(0, 0, 0, 16, 16, BodyKinematic)
	b.Vel.Y = 0
	w.Register(b)

	const dt = 1.0 / 60.0
	for i := 0; i < 60; i++ {
		w.Step(dt)
	}
	if b.Pos.Y != 0 {
		t.Errorf("kinematic body should not fall, got Y=%.2f", b.Pos.Y)
	}
}

// ------------------------------------------------------------------ Static body immovable

func TestStaticImmovable(t *testing.T) {
	w := NewWorld(nil)
	s := staticBody(0, 0, 32, 32)
	w.Register(s)
	const dt = 1.0 / 60.0
	for i := 0; i < 60; i++ {
		w.Step(dt)
	}
	if s.Pos.X != 0 || s.Pos.Y != 0 {
		t.Errorf("static body moved: %v", s.Pos)
	}
}

// ------------------------------------------------------------------ Callback

func TestOnCollisionCallback(t *testing.T) {
	w := NewWorld(nil)
	called := false
	player := dynBody(0, 50, 16, 16)
	player.Vel.Y = 300
	player.OnCollision = func(_ *RigidBody) { called = true }
	ground := staticBody(0, 100, 200, 20)
	w.Register(player)
	w.Register(ground)

	const dt = 1.0 / 60.0
	for i := 0; i < 120; i++ {
		w.Step(dt)
		if called {
			break
		}
	}
	if !called {
		t.Error("OnCollision callback was never fired")
	}
}
