package node

import (
	"testing"
)

func TestNewSprite2D(t *testing.T) {
	s := NewSprite2D("test")
	if s == nil {
		t.Fatal("expected non-nil sprite")
	}
	if s.GetName() != "test" {
		t.Errorf("expected name 'test', got '%s'", s.GetName())
	}
	if s.GetAlpha() != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", s.GetAlpha())
	}
	if s.IsFlipX() {
		t.Error("expected FlipX to be false by default")
	}
}

func TestSprite2D_SetGetAlpha(t *testing.T) {
	s := NewSprite2D("test")
	s.SetAlpha(0.5)
	if s.GetAlpha() != 0.5 {
		t.Errorf("expected alpha 0.5, got %f", s.GetAlpha())
	}
	// Clamp low
	s.SetAlpha(-0.1)
	if s.GetAlpha() != 0 {
		t.Errorf("expected alpha clamped to 0, got %f", s.GetAlpha())
	}
	// Clamp high
	s.SetAlpha(1.5)
	if s.GetAlpha() != 1 {
		t.Errorf("expected alpha clamped to 1, got %f", s.GetAlpha())
	}
}

func TestSprite2D_Flip(t *testing.T) {
	s := NewSprite2D("test")
	s.SetFlipX(true)
	if !s.IsFlipX() {
		t.Error("expected FlipX true after SetFlipX(true)")
	}
	s.SetFlipY(true)
	if !s.IsFlipY() {
		t.Error("expected FlipY true after SetFlipY(true)")
	}
}

func TestSprite2D_SetGetFrame(t *testing.T) {
	s := NewSprite2D("test")
	s.SetFrame(5)
	if s.GetFrame() != 5 {
		t.Errorf("expected frame 5, got %d", s.GetFrame())
	}
}

func TestSprite2D_SetSize(t *testing.T) {
	s := NewSprite2D("test")
	s.SetSize(32, 64)
	w, h := s.GetSize()
	if w != 32 || h != 64 {
		t.Errorf("expected size (32,64), got (%f,%f)", w, h)
	}
}

func TestSprite2D_FrameSize(t *testing.T) {
	s := NewSprite2D("test")
	s.SetFrameSize(16, 16)
	// FrameCount should be 0 since no grid set
	if s.FrameCount() != 0 {
		t.Errorf("expected FrameCount 0 with no grid, got %d", s.FrameCount())
	}
	s.SetGridSize(4, 4)
	if s.FrameCount() != 16 {
		t.Errorf("expected FrameCount 16 for 4x4 grid, got %d", s.FrameCount())
	}
}

func TestSprite2D_AddAnimation(t *testing.T) {
	s := NewSprite2D("test")
	s.AddAnimation("run", []int{0, 1, 2, 3}, 10, true)
	anim := s.Animator()
	if anim == nil {
		t.Fatal("expected non-nil animator after AddAnimation")
	}
	clip := anim.Current()
	if clip != nil {
		t.Error("expected current clip to be nil before Play")
	}
}

func TestSprite2D_PlayAnimation(t *testing.T) {
	s := NewSprite2D("test")
	s.AddAnimation("run", []int{0, 1, 2, 3}, 10, true)

	if !s.Play("run") {
		t.Fatal("expected Play('run') to return true")
	}
	if !s.IsPlayingAnimation() {
		t.Error("expected IsPlayingAnimation() true after Play")
	}

	// Non-looping animation
	s.AddAnimation("die", []int{0, 1}, 1, false)
	s.Play("die")
	if s.GetFrame() != 0 {
		t.Errorf("expected frame 0 after Play, got %d", s.GetFrame())
	}

	// Advance past last frame (dt=2s at 1fps = 2 frames)
	s.Update(2.0)
	if !s.Animator().Done {
		t.Error("expected non-looping animation to be Done after all frames")
	}
}

func TestSprite2D_PlayNonexistent(t *testing.T) {
	s := NewSprite2D("test")
	ok := s.Play("nonexistent")
	if ok {
		t.Error("expected Play('nonexistent') to return false")
	}
}

func TestSprite2D_StopAnimation(t *testing.T) {
	s := NewSprite2D("test")
	s.AddAnimation("run", []int{0, 1, 2}, 10, true)
	s.Play("run")
	s.StopAnimation()
	if s.IsPlayingAnimation() {
		t.Error("expected IsPlayingAnimation() false after Stop")
	}
}

func TestSprite2D_FrameCount(t *testing.T) {
	s := NewSprite2D("test")
	if s.FrameCount() != 0 {
		t.Errorf("expected FrameCount 0, got %d", s.FrameCount())
	}
	s.SetGridSize(3, 2)
	if s.FrameCount() != 6 {
		t.Errorf("expected FrameCount 6, got %d", s.FrameCount())
	}
}

func TestSprite2D_UpdateAnimation(t *testing.T) {
	s := NewSprite2D("test")
	s.AddAnimation("walk", []int{0, 1, 2, 3}, 10, true) // 10fps = 0.1s per frame
	s.Play("walk")

	// Advance 0.25s → should be on frame 2 (0.1 + 0.1 + 0.05)
	s.Update(0.25)
	frame1 := s.GetFrame()
	if frame1 != 2 {
		t.Errorf("expected frame ~2 after 0.25s, got %d", frame1)
	}

	// Advance 1s more → should loop around
	s.Update(1.0)
	frame2 := s.GetFrame()
	if frame2 < 0 || frame2 > 3 {
		t.Errorf("expected frame 0-3 after loop, got %d", frame2)
	}
}
