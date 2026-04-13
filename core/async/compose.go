package async

// ----------------------------------------------------------------------------
// Seq — run tasks one after another
// ----------------------------------------------------------------------------

type seqTask struct {
	tasks []Task
	idx   int
}

// Seq returns a Task that runs each task in order, completing when the last one finishes.
func Seq(tasks ...Task) Task {
	return &seqTask{tasks: tasks}
}

func (t *seqTask) Tick(dt float64) Status {
	for t.idx < len(t.tasks) {
		if t.tasks[t.idx].Tick(dt) == Running {
			return Running
		}
		t.idx++
	}
	return Done
}

// ----------------------------------------------------------------------------
// Repeat — run a task N times (0 = forever)
// ----------------------------------------------------------------------------

type repeatTask struct {
	factory func() Task
	times   int // 0 = infinite
	count   int
	current Task
}

// Repeat runs the task produced by factory `times` times.
// Pass times=0 to repeat forever (cancel via Scheduler.Cancel).
func Repeat(times int, factory func() Task) Task {
	t := &repeatTask{factory: factory, times: times}
	t.current = factory()
	return t
}

func (t *repeatTask) Tick(dt float64) Status {
	if t.current == nil {
		return Done
	}
	if t.current.Tick(dt) == Done {
		t.count++
		if t.times > 0 && t.count >= t.times {
			return Done
		}
		t.current = t.factory()
	}
	return Running
}

// ----------------------------------------------------------------------------
// Delay — run a task after a delay
// ----------------------------------------------------------------------------

type delayTask struct {
	delay   float64
	task    Task
	started bool
}

// Delay waits `seconds` then runs task.
func Delay(seconds float64, task Task) Task {
	return &delayTask{delay: seconds, task: task}
}

func (t *delayTask) Tick(dt float64) Status {
	if !t.started {
		t.delay -= dt
		if t.delay > 0 {
			return Running
		}
		t.started = true
	}
	return t.task.Tick(dt)
}
