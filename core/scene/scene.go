// Package scene manages the entity graph and lifecycle.
package scene

import "github.com/ElioNeto/kora/core/render"

// Entity is the base game object interface that KScript objects implement.
type Entity interface {
	Update(dt float64)
	Draw(r *render.Renderer)
	Destroy()
	IsAlive() bool
}

// Scene holds and drives a collection of entities.
type Scene struct {
	entities []Entity
	pending  []Entity
}

// New creates an empty Scene.
func New() *Scene {
	return &Scene{}
}

// Spawn adds an entity to be included from the next tick.
func (s *Scene) Spawn(e Entity) {
	s.pending = append(s.pending, e)
}

// Update advances all alive entities and flushes pending ones.
func (s *Scene) Update(dt float64) {
	// Flush pending entities.
	if len(s.pending) > 0 {
		s.entities = append(s.entities, s.pending...)
		s.pending = s.pending[:0]
	}

	// Update alive entities; collect dead ones.
	alive := s.entities[:0]
	for _, e := range s.entities {
		if e.IsAlive() {
			e.Update(dt)
			alive = append(alive, e)
		} else {
			e.Destroy()
		}
	}
	s.entities = alive
}

// Draw renders all alive entities.
func (s *Scene) Draw(r *render.Renderer) {
	for _, e := range s.entities {
		if e.IsAlive() {
			e.Draw(r)
		}
	}
}

// Clear removes all entities immediately.
func (s *Scene) Clear() {
	for _, e := range s.entities {
		e.Destroy()
	}
	s.entities = nil
	s.pending = nil
}
