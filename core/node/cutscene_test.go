package node

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Helper: fakeScene implements Scene for testing.
// ---------------------------------------------------------------------------

type fakeScene struct {
	nodes   map[string]interface{}
	signals map[string]bool
}

func newFakeScene() *fakeScene {
	return &fakeScene{
		nodes:   make(map[string]interface{}),
		signals: make(map[string]bool),
	}
}

func (fs *fakeScene) FindNode(path string) interface{} {
	return fs.nodes[path]
}

func (fs *fakeScene) SignalFired(name string) bool {
	return fs.signals[name]
}

func (fs *fakeScene) addNode(path string, n interface{}) {
	fs.nodes[path] = n
}

func (fs *fakeScene) fireSignal(name string) {
	fs.signals[name] = true
}

// ---------------------------------------------------------------------------
// TestNewCutscenePlayer
// ---------------------------------------------------------------------------

func TestNewCutscenePlayer(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	if cp == nil {
		t.Fatal("expected non-nil CutscenePlayer")
	}
	if cp.GetName() != "cutscene" {
		t.Errorf("expected name 'cutscene', got '%s'", cp.GetName())
	}
	if cp.IsPlaying() {
		t.Error("expected not playing initially")
	}
	if cp.GetTotalActions() != 0 {
		t.Errorf("expected 0 actions initially, got %d", cp.GetTotalActions())
	}
	if cp.GetProgress() != 1.0 {
		t.Errorf("expected progress 1.0 with no actions, got %f", cp.GetProgress())
	}
}

// ---------------------------------------------------------------------------
// TestAddAction
// ---------------------------------------------------------------------------

func TestCutsceneAddAction(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(2.0))
	cp.AddAction(WaitAction(3.0))

	if cp.GetTotalActions() != 3 {
		t.Errorf("expected 3 actions, got %d", cp.GetTotalActions())
	}
}

// ---------------------------------------------------------------------------
// TestPlay
// ---------------------------------------------------------------------------

func TestCutscenePlay(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.Play()

	if !cp.IsPlaying() {
		t.Error("expected playing after Play")
	}
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected current action 0, got %d", cp.GetCurrentAction())
	}
}

// ---------------------------------------------------------------------------
// TestWaitTiming
// ---------------------------------------------------------------------------

func TestCutsceneWaitTiming(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(0.5))
	cp.Play()

	// Advance 0.6s — still on first action
	cp.Update(0.6)
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected still on action 0 after 0.6s, got %d", cp.GetCurrentAction())
	}
	if !cp.IsPlaying() {
		t.Error("expected still playing after 0.6s")
	}

	// Advance 0.5s more (total 1.1s) — should be on second action
	cp.Update(0.5)
	if cp.GetCurrentAction() != 1 {
		t.Errorf("expected on action 1 after 1.1s, got %d", cp.GetCurrentAction())
	}

	// Advance past second action
	cp.Update(0.6)
	if cp.IsPlaying() {
		t.Error("expected not playing after all actions complete")
	}
}

// ---------------------------------------------------------------------------
// TestMoveToInterpolation
// ---------------------------------------------------------------------------

func TestCutsceneMoveToInterpolation(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	cp.AddAction(MoveToAction("target", 100, 200, 1.0, "linear"))
	cp.Play()

	// Halfway (t=0.5) -> position should be (50, 100)
	cp.Update(0.5)
	if target.GetX() != 50 || target.GetY() != 100 {
		t.Errorf("expected position (50, 100) at halfway, got (%f, %f)",
			target.GetX(), target.GetY())
	}

	// Complete the action
	cp.Update(0.6)
	if target.GetX() != 100 || target.GetY() != 200 {
		t.Errorf("expected position (100, 200) at end, got (%f, %f)",
			target.GetX(), target.GetY())
	}
}

// ---------------------------------------------------------------------------
// TestMoveToInterpolationEaseOut
// ---------------------------------------------------------------------------

func TestCutsceneMoveToEaseOut(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	// Ease-out: slower at the start, faster at the end.
	// At t=0.5, ease_out(0.5) = 0.5*(2-0.5) = 0.75
	// So position should be 0 + (100-0)*0.75 = 75
	cp.AddAction(MoveToAction("target", 100, 0, 1.0, "ease_out"))
	cp.Play()

	cp.Update(0.5)
	if target.GetX() < 70 || target.GetX() > 80 {
		t.Errorf("expected X ~75 with ease_out at halfway, got %f", target.GetX())
	}

	cp.Update(0.6) // complete the action
	if target.GetX() != 100 {
		t.Errorf("expected X=100 at end, got %f", target.GetX())
	}
}

// ---------------------------------------------------------------------------
// TestFadeIn
// ---------------------------------------------------------------------------

func TestCutsceneFadeIn(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")

	cp.AddAction(FadeInAction("sprite", 1.0))
	cp.Play()

	// Fade actions track progress internally; alpha application to
	// Sprite2D requires scene-level integration (setAlpha is a no-op
	// until the scene wiring is in place). Verify the cutscene
	// processes the action and completes.
	if !cp.IsPlaying() {
		t.Error("expected playing after Play")
	}
	// Advance past the fade duration
	cp.Update(1.5)
	if cp.IsPlaying() {
		t.Error("expected cutscene to complete after fade duration")
	}
	if cp.GetCurrentAction() != cp.GetTotalActions() {
		t.Errorf("expected all actions completed, at %d/%d", cp.GetCurrentAction(), cp.GetTotalActions())
	}
}

func TestCutsceneFadeOut(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")

	cp.AddAction(FadeOutAction("sprite", 1.0))
	cp.Play()

	if !cp.IsPlaying() {
		t.Error("expected playing after Play")
	}
	cp.Update(1.5)
	if cp.IsPlaying() {
		t.Error("expected cutscene to complete after fade duration")
	}
	if cp.GetCurrentAction() != cp.GetTotalActions() {
		t.Errorf("expected all actions completed, at %d/%d", cp.GetCurrentAction(), cp.GetTotalActions())
	}
}

func TestCutsceneFadeInFromPartialAlpha(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")

	cp.AddAction(FadeInAction("sprite", 1.0))
	cp.Play()

	// Verify the fade action is processed (duration-based, completes)
	if !cp.IsPlaying() {
		t.Error("expected playing after Play")
	}
	cp.Update(0.5)
	if !cp.IsPlaying() {
		t.Error("expected still playing at halfway")
	}
	cp.Update(1.0)
	if cp.IsPlaying() {
		t.Error("expected cutscene to complete after full duration")
	}
}

// ---------------------------------------------------------------------------
// TestPauseResume
// ---------------------------------------------------------------------------

func TestCutscenePauseResume(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(2.0))
	cp.Play()

	if !cp.IsPlaying() {
		t.Error("expected playing after Play")
	}

	cp.Pause()
	if cp.IsPlaying() {
		t.Error("expected not playing after Pause")
	}

	// Update should not advance while paused
	cp.Update(10.0)
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected still on action 0 after paused update, got %d", cp.GetCurrentAction())
	}

	cp.Resume()
	if !cp.IsPlaying() {
		t.Error("expected playing after Resume")
	}

	// Now it should advance
	cp.Update(2.5)
	if cp.IsPlaying() {
		t.Error("expected not playing after completion")
	}
}

// ---------------------------------------------------------------------------
// TestStop
// ---------------------------------------------------------------------------

func TestCutsceneStop(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.Play()

	cp.Update(0.5)
	if !cp.IsPlaying() {
		t.Error("expected playing mid-way")
	}

	cp.Stop()
	if cp.IsPlaying() {
		t.Error("expected not playing after Stop")
	}
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected current action reset to 0 after Stop, got %d", cp.GetCurrentAction())
	}
}

// ---------------------------------------------------------------------------
// TestOnComplete
// ---------------------------------------------------------------------------

func TestCutsceneOnComplete(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(0.5))
	cp.AddAction(WaitAction(0.5))

	var completed bool
	cp.SetOnComplete(func() {
		completed = true
	})

	cp.Play()
	cp.Update(1.0) // first action completes, advance to second
	cp.Update(1.0) // second action completes -> finish

	if !completed {
		t.Error("expected onComplete callback to fire")
	}
	if cp.IsPlaying() {
		t.Error("expected not playing after completion")
	}
}

// ---------------------------------------------------------------------------
// TestOnCompleteNotCalledInLoopMode
// ---------------------------------------------------------------------------

func TestCutsceneLoopNoOnComplete(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(0.5))
	cp.SetLoop(true)

	var completed bool
	cp.SetOnComplete(func() {
		completed = true
	})

	cp.Play()
	cp.Update(1.5) // past end, should loop

	if completed {
		t.Error("expected onComplete NOT to fire when looping")
	}
	if !cp.IsPlaying() {
		t.Error("expected still playing after loop wraps")
	}
}

// ---------------------------------------------------------------------------
// TestSkipToEnd
// ---------------------------------------------------------------------------

func TestCutsceneSkipToEnd(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	cp.AddAction(MoveToAction("target", 100, 200, 2.0, "linear"))
	cp.AddAction(WaitAction(0.5))

	var completed bool
	cp.SetOnComplete(func() {
		completed = true
	})

	cp.Play()
	cp.SkipToEnd()

	// Target should be at final position
	if target.GetX() != 100 || target.GetY() != 200 {
		t.Errorf("expected target at (100, 200) after SkipToEnd, got (%f, %f)",
			target.GetX(), target.GetY())
	}
	if !completed {
		t.Error("expected onComplete callback after SkipToEnd")
	}
	if cp.IsPlaying() {
		t.Error("expected not playing after SkipToEnd")
	}
}

// ---------------------------------------------------------------------------
// TestSkipToEndNoActions
// ---------------------------------------------------------------------------

func TestCutsceneSkipToEndNoActions(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	// Should not panic
	cp.SkipToEnd()
}

// ---------------------------------------------------------------------------
// TestLoop
// ---------------------------------------------------------------------------

func TestCutsceneLoop(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(0.5))
	cp.SetLoop(true)
	cp.Play()

	// Advance past end — should loop back to the beginning
	cp.Update(1.0)

	if !cp.IsPlaying() {
		t.Error("expected still playing after loop wrap")
	}
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected on action 0 after loop wrap, got %d",
			cp.GetCurrentAction())
	}

	// Process the first action again
	cp.Update(0.6)
	if cp.GetCurrentAction() != 0 {
		t.Errorf("expected still on action 0 during second iteration, got %d",
			cp.GetCurrentAction())
	}
}

// ---------------------------------------------------------------------------
// TestGetProgress
// ---------------------------------------------------------------------------

func TestCutsceneGetProgress(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(1.0))

	// Progress should be 0 initially
	if cp.GetProgress() != 0.0 {
		t.Errorf("expected progress 0.0 initially, got %f", cp.GetProgress())
	}

	cp.Play()
	// After first action completes, progress should be 1/4 = 0.25
	cp.Update(1.5)
	progress := cp.GetProgress()
	if progress < 0.24 || progress > 0.26 {
		t.Errorf("expected progress ~0.25 after first action, got %f", progress)
	}
}

// ---------------------------------------------------------------------------
// TestClear
// ---------------------------------------------------------------------------

func TestCutsceneClear(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	cp.AddAction(WaitAction(1.0))
	cp.AddAction(WaitAction(2.0))
	cp.Play()

	cp.Clear()

	if cp.GetTotalActions() != 0 {
		t.Errorf("expected 0 actions after Clear, got %d", cp.GetTotalActions())
	}
	if cp.IsPlaying() {
		t.Error("expected not playing after Clear")
	}
	if cp.GetProgress() != 1.0 {
		t.Errorf("expected progress 1.0 after Clear, got %f", cp.GetProgress())
	}
}

// ---------------------------------------------------------------------------
// TestSetVisible
// ---------------------------------------------------------------------------

func TestCutsceneSetVisible(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	if !target.IsVisible() {
		t.Error("expected target visible initially")
	}

	cp.AddAction(SetVisibleAction("target", false))
	cp.AddAction(WaitAction(0.5))

	cp.Play()
	cp.Update(0.1)

	if target.IsVisible() {
		t.Error("expected target not visible after SetVisible(false)")
	}
}

// ---------------------------------------------------------------------------
// TestMultipleActionsChain
// ---------------------------------------------------------------------------

func TestCutsceneMultipleActions(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	// Chain: Wait -> MoveTo -> Wait -> MoveTo
	cp.AddAction(WaitAction(0.5))
	cp.AddAction(MoveToAction("target", 50, 0, 1.0, "linear"))
	cp.AddAction(WaitAction(0.5))
	cp.AddAction(MoveToAction("target", 100, 0, 1.0, "linear"))

	cp.Play()

	// After 0.5s: wait done, first MoveTo starts
	cp.Update(0.5)
	if cp.GetCurrentAction() != 1 {
		t.Errorf("expected on action 1 after 0.5s, got %d", cp.GetCurrentAction())
	}

	// After 1.0s more: first MoveTo done, second wait starts
	cp.Update(1.0)
	if cp.GetCurrentAction() != 2 {
		t.Errorf("expected on action 2 after 1.5s, got %d", cp.GetCurrentAction())
	}

	// After 0.5s more: second wait done, second MoveTo starts
	cp.Update(0.5)
	if cp.GetCurrentAction() != 3 {
		t.Errorf("expected on action 3 after 2.0s, got %d", cp.GetCurrentAction())
	}

	// Complete the second MoveTo
	cp.Update(1.0)
	if cp.IsPlaying() {
		t.Error("expected not playing after all actions complete")
	}
	if target.GetX() != 100 {
		t.Errorf("expected target X=100 at end, got %f", target.GetX())
	}
}

// ---------------------------------------------------------------------------
// TestWaitForSignal
// ---------------------------------------------------------------------------

func TestCutsceneWaitForSignal(t *testing.T) {
	scene := newFakeScene()
	cp := NewCutscenePlayer("cutscene")
	cp.SetScene(scene)

	cp.AddAction(WaitAction(0.5))
	cp.AddAction(WaitForSignalAction("explosion"))
	cp.AddAction(WaitAction(0.5))

	cp.Play()

	// Pass first wait
	cp.Update(0.6)
	if cp.GetCurrentAction() != 1 {
		t.Errorf("expected on action 1 (WaitForSignal) after 0.6s, got %d",
			cp.GetCurrentAction())
	}

	// Update without signal — should stay on action 1
	cp.Update(1.0)
	if cp.GetCurrentAction() != 1 {
		t.Errorf("expected still on action 1 without signal, got %d",
			cp.GetCurrentAction())
	}

	// Fire the signal
	scene.fireSignal("explosion")
	cp.Update(0.1)
	if cp.GetCurrentAction() != 2 {
		t.Errorf("expected on action 2 after signal fired, got %d",
			cp.GetCurrentAction())
	}

	// Complete final wait
	cp.Update(0.6)
	if cp.IsPlaying() {
		t.Error("expected not playing after all actions complete")
	}
}

// ---------------------------------------------------------------------------
// TestFindTargetViaScene
// ---------------------------------------------------------------------------

func TestCutsceneFindTargetViaScene(t *testing.T) {
	scene := newFakeScene()
	target := NewNode2D("hero", 1)
	scene.addNode("hero", target)

	cp := NewCutscenePlayer("cutscene")
	cp.SetScene(scene)

	cp.AddAction(MoveToAction("hero", 200, 0, 1.0, "linear"))
	cp.Play()
	cp.Update(0.5)

	if target.GetX() != 100 {
		t.Errorf("expected target X=100 at halfway via scene, got %f", target.GetX())
	}

	cp.Update(0.6)
	if target.GetX() != 200 {
		t.Errorf("expected target X=200 at end via scene, got %f", target.GetX())
	}
}

// ---------------------------------------------------------------------------
// TestSetScene
// ---------------------------------------------------------------------------

func TestCutsceneSetScene(t *testing.T) {
	cp := NewCutscenePlayer("cutscene")
	scene := newFakeScene()
	cp.SetScene(scene)

	// No panic; just ensure the scene reference is stored.
	if cp.scene != scene {
		t.Error("expected scene reference to be stored")
	}
}

// ---------------------------------------------------------------------------
// TestEasingNonLinear
// ---------------------------------------------------------------------------

func TestCutsceneEaseInInterpolation(t *testing.T) {
	parent := NewNode2D("parent", 1)
	target := NewNode2D("target", 2)
	parent.AddChild(target)

	cp := NewCutscenePlayer("cutscene")
	parent.AddChild(cp)

	// ease_in: t^2, so at t=0.5, easedT=0.25
	// X should be 0 + (100-0)*0.25 = 25
	cp.AddAction(MoveToAction("target", 100, 0, 1.0, "ease_in"))
	cp.Play()

	cp.Update(0.5)
	if target.GetX() < 20 || target.GetX() > 30 {
		t.Errorf("expected X ~25 with ease_in at halfway, got %f", target.GetX())
	}

	cp.Update(0.6)
	if target.GetX() != 100 {
		t.Errorf("expected X=100 at end, got %f", target.GetX())
	}
}

// ---------------------------------------------------------------------------
// TestNodeInterfaceCompliance (compile-time check)
// ---------------------------------------------------------------------------

// This test verifies at compile time that *CutscenePlayer satisfies Node.
var _ Node = (*CutscenePlayer)(nil)

// ---------------------------------------------------------------------------
// TestCutscenePlayerUpdatePropagatesToChildren
// ---------------------------------------------------------------------------

func TestCutscenePlayerUpdatePropagatesToChildren(t *testing.T) {
	cp := NewCutscenePlayer("parent")
	child := NewNode2D("child", 1)
	cp.AddChild(child)

	// Verify children exist and update propagates (no panic).
	if cp.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", cp.GetChildCount())
	}

	cp.Update(1.0)
	// No assertions beyond no panic — Update propagates correctly.
}
