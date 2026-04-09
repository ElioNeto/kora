// Package async is the cooperative task runtime for the Kora engine.
//
// Tasks are lightweight coroutines driven by Tick(dt) calls from the
// game loop. Each Tick advances the task by one step and returns
// either Running (needs more ticks) or Done (finished).
//
// Design goals:
//   - Zero heap allocation for simple sequential tasks.
//   - Composable: Race, All, Seq wrap other tasks.
//   - Android-safe: no goroutines, no channels, no blocking.
package async

// Status is the result of a single Tick.
type Status int

const (
	Running Status = iota
	Done
)

// Task is anything that can be ticked frame-by-frame.
type Task interface {
	Tick(dt float64) Status
}

// ----------------------------------------------------------------------------
// Wait — pause for N seconds
// ----------------------------------------------------------------------------

// waitTask counts down a duration in seconds.
type waitTask struct {
	remaining float64
}

// Wait returns a Task that completes after `seconds` have elapsed.
func Wait(seconds float64) Task {
	return &waitTask{remaining: seconds}
}

func (t *waitTask) Tick(dt float64) Status {
	t.remaining -= dt
	if t.remaining <= 0 {
		return Done
	}
	return Running
}

// ----------------------------------------------------------------------------
// WaitFrames — pause for N frames
// ----------------------------------------------------------------------------

type waitFramesTask struct {
	remaining int
}

// WaitFrames returns a Task that completes after `n` Tick calls.
func WaitFrames(n int) Task {
	return &waitFramesTask{remaining: n}
}

func (t *waitFramesTask) Tick(_ float64) Status {
	t.remaining--
	if t.remaining <= 0 {
		return Done
	}
	return Running
}

// ----------------------------------------------------------------------------
// WaitSignal — pause until an Entity emits a named signal
// ----------------------------------------------------------------------------

// SignalEmitter is any object that can report whether it has emitted a signal.
type SignalEmitter interface {
	SignalFired(name string) bool
}

type waitSignalTask struct {
	emitter SignalEmitter
	signal  string
}

// WaitSignal returns a Task that completes when emitter fires signal.
func WaitSignal(emitter SignalEmitter, signal string) Task {
	return &waitSignalTask{emitter: emitter, signal: signal}
}

func (t *waitSignalTask) Tick(_ float64) Status {
	if t.emitter.SignalFired(t.signal) {
		return Done
	}
	return Running
}

// ----------------------------------------------------------------------------
// Func — wrap a callback as a one-shot task
// ----------------------------------------------------------------------------

type funcTask struct {
	fn   func(dt float64) Status
}

// Func wraps a function as a Task. Useful for inline lambdas.
func Func(fn func(dt float64) Status) Task {
	return &funcTask{fn: fn}
}

func (t *funcTask) Tick(dt float64) Status {
	return t.fn(dt)
}

// ----------------------------------------------------------------------------
// Done sentinel
// ----------------------------------------------------------------------------

type doneTask struct{}

// Immediate returns a Task that is already done on the first Tick.
func Immediate() Task { return &doneTask{} }

func (t *doneTask) Tick(_ float64) Status { return Done }
