package async_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/async"
)

// tick helper: advance task n times by dt each.
func tickN(t *testing.T, task async.Task, n int, dt float64) async.Status {
	t.Helper()
	var s async.Status
	for i := 0; i < n; i++ {
		s = task.Tick(dt)
	}
	return s
}

// --- Wait ---

func TestWaitRunning(t *testing.T) {
	task := async.Wait(1.0)
	if task.Tick(0.5) != async.Running {
		t.Error("expected Running after 0.5s of a 1.0s wait")
	}
}

func TestWaitDone(t *testing.T) {
	task := async.Wait(1.0)
	if tickN(t, task, 2, 0.5) != async.Done {
		t.Error("expected Done after 1.0s total")
	}
}

// --- WaitFrames ---

func TestWaitFrames(t *testing.T) {
	task := async.WaitFrames(3)
	if task.Tick(0) != async.Running { t.Error("frame 1 should be Running") }
	if task.Tick(0) != async.Running { t.Error("frame 2 should be Running") }
	if task.Tick(0) != async.Done   { t.Error("frame 3 should be Done") }
}

// --- Seq ---

func TestSeq(t *testing.T) {
	calls := 0
	task := async.Seq(
		async.Wait(0.5),
		async.Func(func(_ float64) async.Status {
			calls++
			return async.Done
		}),
	)
	task.Tick(0.6) // completes wait(0.5) and runs Func
	if calls != 1 {
		t.Errorf("expected Func to be called once, got %d", calls)
	}
}

// --- Race ---

func TestRace(t *testing.T) {
	task := async.Race(
		async.Wait(2.0),
		async.Wait(0.3),
	)
	if task.Tick(0.4) != async.Done {
		t.Error("Race should complete when the fastest child (0.3s) finishes")
	}
}

// --- All ---

func TestAll(t *testing.T) {
	task := async.All(
		async.Wait(0.5),
		async.Wait(1.0),
	)
	if task.Tick(0.6) != async.Running {
		t.Error("All should still be Running after 0.6s (second task needs 1.0s)")
	}
	if task.Tick(0.5) != async.Done {
		t.Error("All should be Done after 1.1s total")
	}
}

// --- Repeat ---

func TestRepeat(t *testing.T) {
	count := 0
	task := async.Repeat(3, func() async.Task {
		return async.Func(func(_ float64) async.Status {
			count++
			return async.Done
		})
	})
	for i := 0; i < 10; i++ {
		if task.Tick(0) == async.Done {
			break
		}
	}
	if count != 3 {
		t.Errorf("expected 3 repetitions, got %d", count)
	}
}

// --- Delay ---

func TestDelay(t *testing.T) {
	called := false
	task := async.Delay(0.5, async.Func(func(_ float64) async.Status {
		called = true
		return async.Done
	}))
	task.Tick(0.3)
	if called {
		t.Error("task should not have run yet")
	}
	task.Tick(0.3)
	if !called {
		t.Error("task should have run after delay")
	}
}

// --- Tween ---

func TestTweenFloat(t *testing.T) {
	v := 0.0
	task := async.TweenFloat(&v, 100.0, 1.0, nil)
	task.Tick(0.5)
	if v <= 0 || v >= 100 {
		t.Errorf("mid-tween value out of range: %v", v)
	}
	task.Tick(0.6)
	if v != 100.0 {
		t.Errorf("expected final value 100.0, got %v", v)
	}
}

// --- Scheduler ---

func TestScheduler(t *testing.T) {
	sched := async.NewScheduler()
	sched.Run(async.Wait(0.5))
	sched.Run(async.Wait(1.0))

	sched.Tick(0.6)
	if sched.Len() != 1 {
		t.Errorf("expected 1 remaining task, got %d", sched.Len())
	}
	sched.Tick(0.5)
	if sched.Len() != 0 {
		t.Errorf("expected 0 remaining tasks, got %d", sched.Len())
	}
}

func TestSchedulerCancel(t *testing.T) {
	sched := async.NewScheduler()
	h := sched.Run(async.Wait(10.0))
	sched.Cancel(h)
	sched.Tick(0.016)
	if sched.Len() != 0 {
		t.Errorf("cancelled task should be removed")
	}
}

// --- WaitSignal ---

type fakeEmitter struct{ fired bool }
func (f *fakeEmitter) SignalFired(_ string) bool { return f.fired }

func TestWaitSignal(t *testing.T) {
	em := &fakeEmitter{}
	task := async.WaitSignal(em, "ready")
	if task.Tick(0) != async.Running {
		t.Error("should be Running before signal fires")
	}
	em.fired = true
	if task.Tick(0) != async.Done {
		t.Error("should be Done after signal fires")
	}
}
