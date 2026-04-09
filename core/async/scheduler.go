package async

// Handle is an opaque reference to a running task in the Scheduler.
type Handle uint32

const InvalidHandle Handle = 0

// entry is an active task slot in the scheduler.
type entry struct {
	task   Task
	handle Handle
	alive  bool
}

// Scheduler owns and drives a set of concurrent tasks.
// Call Tick(dt) once per frame from the game loop.
//
// Thread-safety: NOT thread-safe. Call only from the main goroutine.
type Scheduler struct {
	entries []*entry
	nextID  Handle
}

// NewScheduler creates an empty Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{nextID: 1}
}

// Run registers task and returns its Handle.
func (s *Scheduler) Run(task Task) Handle {
	h := s.nextID
	s.nextID++
	s.entries = append(s.entries, &entry{task: task, handle: h, alive: true})
	return h
}

// Cancel stops a task before it finishes.
func (s *Scheduler) Cancel(h Handle) {
	for _, e := range s.entries {
		if e.handle == h {
			e.alive = false
			return
		}
	}
}

// Tick advances all running tasks by dt seconds.
// Completed tasks are removed.
func (s *Scheduler) Tick(dt float64) {
	live := s.entries[:0]
	for _, e := range s.entries {
		if !e.alive {
			continue
		}
		if e.task.Tick(dt) == Running {
			live = append(live, e)
		}
		// Done tasks are dropped.
	}
	s.entries = live
}

// Len returns the number of currently running tasks.
func (s *Scheduler) Len() int {
	return len(s.entries)
}

// Clear cancels all running tasks.
func (s *Scheduler) Clear() {
	s.entries = s.entries[:0]
}
