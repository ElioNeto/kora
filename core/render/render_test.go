package render_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/render"
)

// --- Camera ---

func TestCameraWorldToScreen(t *testing.T) {
	cam := render.NewCamera(320, 240)
	// Camera at origin, zoom=1: world (0,0) should map to screen centre.
	sx, sy := cam.WorldToScreen(0, 0)
	if sx != 160 || sy != 120 {
		t.Errorf("expected (160,120), got (%v,%v)", sx, sy)
	}
}

func TestCameraScreenToWorld(t *testing.T) {
	cam := render.NewCamera(320, 240)
	wx, wy := cam.ScreenToWorld(160, 120)
	if wx != 0 || wy != 0 {
		t.Errorf("expected (0,0), got (%v,%v)", wx, wy)
	}
}

func TestCameraZoom(t *testing.T) {
	cam := render.NewCamera(320, 240)
	cam.Zoom = 2.0
	sx, sy := cam.WorldToScreen(10, 5)
	// (10*2 + 160, 5*2 + 120) = (180, 130)
	if sx != 180 || sy != 130 {
		t.Errorf("expected (180,130) with zoom=2, got (%v,%v)", sx, sy)
	}
}

func TestCameraFollow(t *testing.T) {
	cam := render.NewCamera(320, 240)
	cam.Follow(100, 50, 1000, 0.016) // very fast follow
	if cam.X == 0 && cam.Y == 0 {
		t.Error("camera should have moved toward target")
	}
}

// --- Animator ---
// (No Ebitengine image needed — we test frame logic directly via a nil-safe sheet stub.)

func TestAnimatorPlayAndUpdate(t *testing.T) {
	// Build a fake 4-frame sheet stub (no actual GPU image).
	// We only test the Animator state machine, not rendering.
	anim := &render.Animation{
		Name:   "run",
		Frames: []int{0, 1, 2, 3},
		FPS:    10,
		Loop:   true,
	}
	a := render.NewAnimator(nil) // nil sheet: CurrentSprite returns nil, won't panic
	a.Add(anim)
	a.Play("run")

	// After 0.1s at 10 FPS, we should advance 1 frame.
	a.Update(0.1)
	// Frame index 0 → 1; Done should be false (looping).
	if a.Done {
		t.Error("looping animation should not be Done")
	}
}

func TestAnimatorNonLoopDone(t *testing.T) {
	anim := &render.Animation{
		Name:   "die",
		Frames: []int{0, 1},
		FPS:    10,
		Loop:   false,
	}
	a := render.NewAnimator(nil)
	a.Add(anim)
	a.Play("die")
	a.Update(0.3) // 3 frames worth, but only 2 frames in clip
	if !a.Done {
		t.Error("non-looping animation should be Done after all frames play")
	}
}

// --- Tilemap ---

func TestTilemapSetGet(t *testing.T) {
	ts := &render.Tileset{TileW: 16, TileH: 16, Cols: 4} // stub, no atlas
	tm := render.NewTilemap(ts, 10, 8)
	tm.Set(3, 2, 5)
	if got := tm.Get(3, 2); got != 5 {
		t.Errorf("expected tile 5, got %d", got)
	}
}

func TestTilemapOutOfBounds(t *testing.T) {
	ts := &render.Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := render.NewTilemap(ts, 5, 5)
	tm.Set(99, 99, 1) // should not panic
	if got := tm.Get(99, 99); got != -1 {
		t.Errorf("out-of-bounds get should return -1, got %d", got)
	}
}

func TestTilemapTileAtWorld(t *testing.T) {
	ts := &render.Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := render.NewTilemap(ts, 10, 10)
	tm.Set(2, 3, 7)
	// World point inside tile (2,3): x=2*16+4=36, y=3*16+4=52
	if got := tm.TileAtWorld(36, 52); got != 7 {
		t.Errorf("expected tile 7 at world pos, got %d", got)
	}
}
