package async

// WaitTask pauses execution for a given duration.
type WaitTask struct {
	remaining float64
}

// Wait creates a task that completes after `seconds`.
func Wait(seconds float64) *WaitTask {
	return &WaitTask{remaining: seconds}
}

func (t *WaitTask) Tick(dt float64) Status {
	t.remaining -= dt
	if t.remaining <= 0 {
		return Done
	}
	return Running
}

// FrameTask pauses execution for a number of frames.
type FrameTask struct {
	remaining int
}

// WaitFrames creates a task that completes after `n` frames.
func WaitFrames(n int) *FrameTask {
	return &FrameTask{remaining: n}
}

func (t *FrameTask) Tick(_ float64) Status {
	t.remaining--
	if t.remaining <= 0 {
		return Done
	}
	return Running
}

// RaceTask completes as soon as any of its children finishes.
type RaceTask struct {
	tasks []Task
}

// Race creates a task that finishes when the first child task finishes.
func Race(tasks ...Task) *RaceTask {
	return &RaceTask{tasks: tasks}
}

func (t *RaceTask) Tick(dt float64) Status {
	for _, child := range t.tasks {
		if child.Tick(dt) == Done {
			return Done
		}
	}
	return Running
}

// AllTask completes when every child task is done.
type AllTask struct {
	tasks []Task
}

// All creates a task that finishes when all children finish.
func All(tasks ...Task) *AllTask {
	return &AllTask{tasks: tasks}
}

func (t *AllTask) Tick(dt float64) Status {
	done := 0
	for _, child := range t.tasks {
		if child.Tick(dt) == Done {
			done++
		}
	}
	if done == len(t.tasks) {
		return Done
	}
	return Running
}

// SignalTask completes when a named signal is emitted on a target.
type SignalTask struct {
	targetID string
	name     string
	triggered bool
}

// WaitSignal creates a task waiting for `name` on the object with `targetID`.
func WaitSignal(targetID, name string) *SignalTask {
	return &SignalTask{targetID: targetID, name: name}
}

// Trigger is called by the signal bus when the signal fires.
func (t *SignalTask) Trigger() {
	t.triggered = true
}

func (t *SignalTask) Tick(_ float64) Status {
	if t.triggered {
		return Done
	}
	return Running
}
