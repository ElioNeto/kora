package async

// entry wraps a running task with an optional cancel flag.
type entry struct {
	task      Task
	cancelled bool
}

// Scheduler advances all registered tasks once per game tick.
type Scheduler struct {
	active []*entry
}

// NewScheduler creates a Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Submit adds a task to the scheduler.
func (s *Scheduler) Submit(t Task) *entry {
	e := &entry{task: t}
	s.active = append(s.active, e)
	return e
}

// Cancel marks an entry for cancellation on the next tick.
func (s *Scheduler) Cancel(e *entry) {
	if e != nil {
		e.cancelled = true
	}
}

// Tick advances all active tasks and removes finished ones.
func (s *Scheduler) Tick(dt float64) {
	running := s.active[:0]
	for _, e := range s.active {
		if e.cancelled {
			continue
		}
		status := e.task.Tick(dt)
		if status == Running {
			running = append(running, e)
		}
	}
	s.active = running
}

// Len returns the number of currently active tasks.
func (s *Scheduler) Len() int { return len(s.active) }
