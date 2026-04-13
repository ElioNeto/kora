package scene_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/async"
	"github.com/ElioNeto/kora/core/scene"
)

// --- minimal test entity ---

type testEntity struct {
	scene.BaseEntity
	alive    bool
	updates  int
}

func newEntity() *testEntity { return &testEntity{alive: true} }
func (e *testEntity) IsAlive() bool  { return e.alive }
func (e *testEntity) Destroy()       { e.alive = false }
func (e *testEntity) Update(_ float64) { e.updates++ }

// --- async entity ---

type asyncEntity struct {
	scene.BaseEntity
	alive   bool
	started bool
}

func (e *asyncEntity) IsAlive() bool { return e.alive }
func (e *asyncEntity) Destroy()      { e.alive = false }
func (e *asyncEntity) CreateTask() async.Task {
	e.started = true
	return async.Wait(1.0)
}

// --- tests ---

func TestSpawnAndUpdate(t *testing.T) {
	s := scene.New()
	e := newEntity()
	s.Spawn(e)
	s.Update(0.016)
	if e.updates != 1 {
		t.Errorf("expected 1 update, got %d", e.updates)
	}
}

func TestDestroyPrunes(t *testing.T) {
	s := scene.New()
	e := newEntity()
	s.Spawn(e)
	s.Update(0.016)
	e.Destroy()
	s.Update(0.016)
	if s.Count() != 0 {
		t.Errorf("expected 0 entities after destroy, got %d", s.Count())
	}
}

func TestGroupFindAll(t *testing.T) {
	s := scene.New()
	e1 := newEntity()
	e2 := newEntity()
	s.SpawnInGroup("enemy", e1)
	s.SpawnInGroup("enemy", e2)
	s.Update(0.016)
	if len(s.FindAll("enemy")) != 2 {
		t.Errorf("expected 2 enemies")
	}
}

func TestGroupPrunesDeadEntities(t *testing.T) {
	s := scene.New()
	e := newEntity()
	s.SpawnInGroup("bullet", e)
	s.Update(0.016)
	e.Destroy()
	s.Update(0.016)
	if s.CountInGroup("bullet") != 0 {
		t.Errorf("expected group to be pruned")
	}
}

func TestAsyncIniterStartsTask(t *testing.T) {
	s := scene.New()
	e := &asyncEntity{alive: true}
	s.Spawn(e)
	if !e.started {
		t.Error("CreateTask should be called on Spawn")
	}
	// tick scheduler directly (Scene.Update doesn't tick it by default)
	s.Scheduler().Tick(0.5)
	if s.Scheduler().Len() == 0 {
		t.Error("expected async task to still be running")
	}
	// task completes after another 0.6s (total 1.1s > 1.0s duration)
	for i := 0; i < 10; i++ {
		s.Scheduler().Tick(0.1)
	}
	if s.Scheduler().Len() != 0 {
		t.Errorf("expected 0 running tasks, got %d", s.Scheduler().Len())
	}
}

func TestDestroyAll(t *testing.T) {
	s := scene.New()
	for i := 0; i < 5; i++ {
		s.Spawn(newEntity())
	}
	s.Update(0.016)
	s.DestroyAll()
	s.Update(0.016)
	if s.Count() != 0 {
		t.Errorf("expected 0 after DestroyAll")
	}
}

func TestSignalRoundtrip(t *testing.T) {
	e := newEntity()
	e.EmitSignal("hit")
	if !e.SignalFired("hit") {
		t.Error("signal should be fired")
	}
	e.ClearSignals()
	if e.SignalFired("hit") {
		t.Error("signal should be cleared")
	}
}

func TestFadeTransition(t *testing.T) {
	var f scene.FadeState
	done := false
	f.FadeOut(1.0, func() { done = true })
	f.Tick(0.5)
	if f.Alpha <= 0 || f.Alpha >= 1 {
		t.Errorf("mid-fade alpha out of range: %v", f.Alpha)
	}
	f.Tick(0.6)
	if !done {
		t.Error("onDone callback should have fired")
	}
	if f.Alpha != 1 {
		t.Errorf("expected alpha=1 at end of FadeOut, got %v", f.Alpha)
	}
}
