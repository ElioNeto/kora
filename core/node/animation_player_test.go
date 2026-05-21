package node

import (
	"testing"
)

func TestNewAnimationPlayer(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	if ap == nil {
		t.Fatal("expected non-nil AnimationPlayer")
	}
	if ap.GetName() != "anim" {
		t.Errorf("expected name 'anim', got '%s'", ap.GetName())
	}
	if ap.IsPlaying() {
		t.Error("expected not playing initially")
	}
}

func TestAddClip(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	clip := &AnimationClip{
		Name:      "test",
		Target:    "Node",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
	}
	ap.AddClip(clip)

	got := ap.GetClip("test")
	if got != clip {
		t.Error("expected clip to match")
	}
}

func TestPlayNonExistentClip(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ok := ap.Play("nonexistent")
	if ok {
		t.Error("expected false for nonexistent clip")
	}
	if ap.IsPlaying() {
		t.Error("expected not playing after failed play")
	}
}

func TestPlayClip(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "move",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
	})

	ok := ap.Play("move")
	if !ok {
		t.Fatal("expected true for existing clip")
	}
	if !ap.IsPlaying() {
		t.Error("expected playing after Play")
	}
}

func TestStop(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
	})
	ap.Play("test")
	ap.Stop()

	if ap.IsPlaying() {
		t.Error("expected not playing after Stop")
	}
	if ap.CurrentTime() != 0 {
		t.Errorf("expected elapsed reset to 0, got %f", ap.CurrentTime())
	}
}

func TestPauseResume(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
	})
	ap.Play("test")

	if ap.IsPaused() {
		t.Error("expected not paused initially")
	}

	ap.Pause()
	if !ap.IsPaused() {
		t.Error("expected paused after Pause")
	}
	if ap.IsPlaying() {
		t.Error("expected not playing while paused")
	}

	ap.Resume()
	if ap.IsPaused() {
		t.Error("expected not paused after Resume")
	}
	if !ap.IsPlaying() {
		t.Error("expected playing after Resume")
	}
}

func TestNonLoopingAnimationCompletes(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
		Loop:      false,
	})
	ap.Play("test")

	// Advance past duration
	ap.Update(1.5)

	if ap.IsPlaying() {
		t.Error("expected not playing after animation completes")
	}
	if !ap.IsDone() {
		t.Error("expected IsDone after completion")
	}
}

func TestLoopingAnimationRepeats(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
		Loop:      true,
	})
	ap.Play("test")

	// Advance past duration - should loop
	ap.Update(2.5)

	if !ap.IsPlaying() {
		t.Error("expected still playing after looping animation cycles")
	}
	if ap.IsDone() {
		t.Error("expected not done for looping animation")
	}
}

func TestOnFinishCallback(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
		Loop:      false,
	})

	var finishedName string
	ap.OnFinish(func(name string) {
		finishedName = name
	})

	ap.Play("test")
	ap.Update(1.5)

	if finishedName != "test" {
		t.Errorf("expected onFinish to fire with 'test', got '%s'", finishedName)
	}
}

func TestProgress(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	// No clip
	if ap.Progress() != 0 {
		t.Error("expected progress 0 with no clip")
	}

	ap.AddClip(&AnimationClip{
		Name:      "test",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}},
		Duration:  2,
	})
	ap.Play("test")
	ap.Update(1.0)

	progress := ap.Progress()
	if progress < 0.49 || progress > 0.51 {
		t.Errorf("expected progress ~0.5, got %f", progress)
	}
}

func TestClipNames(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{Name: "a"})
	ap.AddClip(&AnimationClip{Name: "b"})
	ap.AddClip(&AnimationClip{Name: "c"})

	names := ap.ClipNames()
	if len(names) != 3 {
		t.Errorf("expected 3 clip names, got %d", len(names))
	}
}

func TestRemoveClip(t *testing.T) {
	ap := NewAnimationPlayer("anim")
	ap.AddClip(&AnimationClip{Name: "test"})
	ap.RemoveClip("test")

	if ap.GetClip("test") != nil {
		t.Error("expected nil after RemoveClip")
	}
}

func TestAnimPropertyFromString(t *testing.T) {
	tests := []struct {
		input string
		prop  AnimProperty
		ok    bool
	}{
		{"x", AnimPropPosX, true},
		{"y", AnimPropPosY, true},
		{"rotation", AnimPropRotation, true},
		{"scale_x", AnimPropScaleX, true},
		{"scale_y", AnimPropScaleY, true},
		{"alpha", AnimPropAlpha, true},
		{"invalid", AnimPropPosX, false},
	}

	for _, tt := range tests {
		prop, ok := AnimPropertyFromString(tt.input)
		if ok != tt.ok || prop != tt.prop {
			t.Errorf("AnimPropertyFromString(%q) = (%d, %v), want (%d, %v)",
				tt.input, prop, ok, tt.prop, tt.ok)
		}
	}
}

func TestEasingByName(t *testing.T) {
	tests := []struct {
		name string
		want EasingFunc
	}{
		{"linear", EaseLinear},
		{"ease_in", EaseIn},
		{"ease_out", EaseOut},
		{"ease_in_out", EaseInOut},
		{"unknown", EaseLinear}, // defaults to linear
	}

	for _, tt := range tests {
		fn := EasingByName(tt.name)
		if fn == nil {
			t.Errorf("EasingByName(%q) returned nil", tt.name)
		}
		// Test that fn(0)=0 and fn(1)=1 for valid easing functions
		if fn(0) != 0 || fn(1) != 1 {
			t.Errorf("EasingByName(%q) doesn't satisfy fn(0)=0, fn(1)=1", tt.name)
		}
	}
}

func TestAnimationPlayerApplyToNode(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	anim := NewAnimationPlayer("anim")
	parent.AddChild(anim)

	// Add a clip that animates the target's X position
	anim.AddClip(&AnimationClip{
		Name:      "slide",
		Target:    "target",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 100}},
		Duration:  1,
		Loop:      false,
	})

	// Verify initial position
	if target.GetX() != 0 {
		t.Errorf("expected initial X=0, got %f", target.GetX())
	}

	// Play and advance halfway
	anim.Play("slide")
	anim.Update(0.5)

	if target.GetX() < 40 || target.GetX() > 60 {
		t.Errorf("expected X ~50 at halfway, got %f", target.GetX())
	}

	// Advance to completion
	anim.Update(0.6)
	if target.GetX() != 100 {
		t.Errorf("expected X=100 at end, got %f", target.GetX())
	}
}

func TestAnimationPlayerApplyRotation(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	anim := NewAnimationPlayer("anim")
	parent.AddChild(anim)

	anim.AddClip(&AnimationClip{
		Name:      "spin",
		Target:    "target",
		Property:  AnimPropRotation,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 360}},
		Duration:  1,
	})

	anim.Play("spin")
	// Advance to completion
	anim.Update(1.5)
	if target.GetRotation() != 360 {
		t.Errorf("expected rotation=360 at end, got %f", target.GetRotation())
	}
}

func TestAnimationPlayerDefaultTarget(t *testing.T) {
	parent := NewNode2D("parent", 1)
	anim := NewAnimationPlayer("anim")
	parent.AddChild(anim)

	// Empty target means animate the parent
	anim.AddClip(&AnimationClip{
		Name:      "move",
		Target:    "",
		Property:  AnimPropPosX,
		Keyframes: []Keyframe{{Time: 0, Value: 0}, {Time: 1, Value: 50}},
		Duration:  1,
	})

	anim.Play("move")
	anim.Update(1.5)

	// The AnimationPlayer's parent should have been animated
	if parent.GetX() != 50 {
		t.Errorf("expected parent X=50 at end, got %f", parent.GetX())
	}
}

// Compile-time interface check
var _ Node = (*AnimationPlayer)(nil)
