// Package async provides the coroutine-style task system for Kora.
//
// KScript async functions compile down to structs implementing Task.
// The Scheduler advances them each game tick — no goroutines are exposed
// to user code.
package async

// Status represents the result of a single Task tick.
type Status int

const (
	// Running means the task is still in progress.
	Running Status = iota
	// Done means the task completed successfully.
	Done
	// Cancelled means the task was cancelled before completion.
	Cancelled
)

// Task is the interface every compiled async operation must satisfy.
type Task interface {
	// Tick advances the task by dt seconds and returns its current status.
	Tick(dt float64) Status
}
