// Package pool provides a generic object pool that reuses instances instead of
// allocating new ones. This reduces GC pressure for frequently spawned and
// destroyed objects such as bullets, particles, or enemies.
package pool

import "sync"

// Pool is a generic object pool that reuses instances instead of
// allocating new ones. This reduces GC pressure for frequently
// spawned/destroyed objects.
type Pool[T any] struct {
	pool    []*T
	factory func() *T
	reset   func(*T)
	initialSize int
	maxSize     int
	mu          sync.Mutex
}

// New creates a new Pool with the given factory function, optional reset
// function, and initial/max sizes.
//
//   - factory: called to create new objects when the pool is empty.
//     When nil, new(T) is used as default.
//   - reset: called on each object when it is returned via Put.
//     When nil, no reset is performed.
//   - initialSize: used by PreWarm to pre-populate the pool.
//   - maxSize: maximum number of objects the pool can hold.
//     Values <= 0 mean unlimited capacity.
func New[T any](factory func() *T, reset func(*T), initialSize, maxSize int) *Pool[T] {
	return &Pool[T]{
		pool:        make([]*T, 0, initialSize),
		factory:     factory,
		reset:       reset,
		initialSize: initialSize,
		maxSize:     maxSize,
	}
}

// Get returns an object from the pool, creating a new one if empty.
func (p *Pool[T]) Get() *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if n := len(p.pool); n > 0 {
		obj := p.pool[n-1]
		p.pool = p.pool[:n-1]
		return obj
	}

	if p.factory != nil {
		return p.factory()
	}
	return new(T)
}

// Put returns an object to the pool for reuse.
// If the pool is full (>= maxSize), the object is discarded (GC'd).
func (p *Pool[T]) Put(obj *T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.reset != nil {
		p.reset(obj)
	}

	if p.maxSize > 0 && len(p.pool) >= p.maxSize {
		return
	}

	p.pool = append(p.pool, obj)
}

// PreWarm creates initialSize objects and adds them to the pool.
func (p *Pool[T]) PreWarm() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.initialSize; i++ {
		if p.maxSize > 0 && len(p.pool) >= p.maxSize {
			return
		}

		var obj *T
		if p.factory != nil {
			obj = p.factory()
		} else {
			obj = new(T)
		}
		p.pool = append(p.pool, obj)
	}
}

// Len returns the current number of available objects in the pool.
func (p *Pool[T]) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pool)
}

// Cap returns the maximum pool size. Returns 0 when the pool has no limit.
func (p *Pool[T]) Cap() int {
	return p.maxSize
}
