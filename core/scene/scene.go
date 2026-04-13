package scene

import (
	"github.com/ElioNeto/kora/core/async"
)

// ----------------------------------------------------------------------------
// Scene
// ----------------------------------------------------------------------------

// Scene is the root container for all live entities.
// It owns a Scheduler and drives the update loop.
type Scene struct {
	entities  []Entity
	pending   []Entity          // spawned mid-frame, added after update
	scheduler *async.Scheduler
	groups    map[string][]Entity
}

// New creates an empty Scene.
func New() *Scene {
	return &Scene{
		scheduler: async.NewScheduler(),
		groups:    make(map[string][]Entity),
	}
}

// Scheduler returns the scene-wide async scheduler.
func (s *Scene) Scheduler() *async.Scheduler { return s.scheduler }

// ----------------------------------------------------------------------------
// Spawn / Destroy
// ----------------------------------------------------------------------------

// Spawn adds entity to the scene and calls Create / CreateTask if implemented.
// The entity is queued and added at the start of the next frame to avoid
// mutating the entity list mid-update.
func (s *Scene) Spawn(e Entity) Entity {
	if initer, ok := e.(Initer); ok {
		initer.Create()
	}
	if ai, ok := e.(AsyncIniter); ok {
		s.scheduler.Run(ai.CreateTask())
	}
	s.pending = append(s.pending, e)
	return e
}

// SpawnInGroup adds the entity to a named group in addition to the main list.
func (s *Scene) SpawnInGroup(group string, e Entity) Entity {
	s.Spawn(e)
	s.groups[group] = append(s.groups[group], e)
	return e
}

// DestroyAll destroys every entity in the scene and clears all tasks.
func (s *Scene) DestroyAll() {
	for _, e := range s.entities {
		e.Destroy()
	}
	s.entities = s.entities[:0]
	s.pending = s.pending[:0]
	s.scheduler.Clear()
	for k := range s.groups {
		delete(s.groups, k)
	}
}

// ----------------------------------------------------------------------------
// Update — called once per frame by the game loop
// ----------------------------------------------------------------------------

// Update advances the scene by dt seconds:
//  1. Flush pending spawns.
//  2. Tick the async scheduler.
//  3. Update every living Updater.
//  4. Clear per-frame signals.
//  5. Prune dead entities.
func (s *Scene) Update(dt float64) {
	// 1. Flush pending.
	s.entities = append(s.entities, s.pending...)
	s.pending = s.pending[:0]

	// 2. Async scheduler.
	s.scheduler.Tick(dt)

	// 3. Update entities.
	for _, e := range s.entities {
		if !e.IsAlive() {
			continue
		}
		if u, ok := e.(Updater); ok {
			u.Update(dt)
		}
	}

	// 4. Clear signals.
	for _, e := range s.entities {
		if se, ok := e.(SignalEmitter); ok {
			se.ClearSignals()
		}
	}

	// 5. Prune dead.
	s.prune()
}

func (s *Scene) prune() {
	live := s.entities[:0]
	for _, e := range s.entities {
		if e.IsAlive() {
			live = append(live, e)
		}
	}
	s.entities = live

	for g, group := range s.groups {
		live := group[:0]
		for _, e := range group {
			if e.IsAlive() {
				live = append(live, e)
			}
		}
		s.groups[g] = live
	}
}

// ----------------------------------------------------------------------------
// Draw — called once per frame after Update
// ----------------------------------------------------------------------------

// Draw calls Draw(r) on every living Drawer in insertion order.
func (s *Scene) Draw(r interface{}) {
	for _, e := range s.entities {
		if !e.IsAlive() {
			continue
		}
		if d, ok := e.(Drawer); ok {
			d.Draw(r)
		}
	}
}

// PhysicsUpdate calls PhysicsUpdate(dt) on every living PhysicsNode.
// This is called separately from Update to allow for fixed timestep physics.
func (s *Scene) PhysicsUpdate(dt float64) {
	for _, e := range s.entities {
		if !e.IsAlive() {
			continue
		}
		if p, ok := e.(PhysicsNode); ok {
			p.PhysicsUpdate(dt)
		}
	}
}

// ----------------------------------------------------------------------------
// Query
// ----------------------------------------------------------------------------

// Find returns the first living entity in group, or nil.
func (s *Scene) Find(group string) Entity {
	for _, e := range s.groups[group] {
		if e.IsAlive() {
			return e
		}
	}
	return nil
}

// FindAll returns all living entities in group.
func (s *Scene) FindAll(group string) []Entity {
	var out []Entity
	for _, e := range s.groups[group] {
		if e.IsAlive() {
			out = append(out, e)
		}
	}
	return out
}

// Count returns the number of living entities.
func (s *Scene) Count() int {
	n := 0
	for _, e := range s.entities {
		if e.IsAlive() {
			n++
		}
	}
	return n
}

// CountInGroup returns the number of living entities in a named group.
func (s *Scene) CountInGroup(group string) int {
	n := 0
	for _, e := range s.groups[group] {
		if e.IsAlive() {
			n++
		}
	}
	return n
}
