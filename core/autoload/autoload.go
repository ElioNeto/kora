// Package autoload provides a singleton registry for global game objects
// that persist across scene changes.
//
// AutoLoad objects are created before the first scene loads and remain
// alive for the entire game session. They are accessible by name from
// any scene, making them ideal for:
//   - Game state managers (score, lives, progress)
//   - Audio managers
//   - Save/load systems
//   - Input configuration
//   - Network services
//
// Usage:
//
//	// Define an AutoLoad object.
//	type ScoreManager struct {
//	    autoload.Base
//	    Score int
//	}
//	func (s *ScoreManager) Name() string { return "ScoreManager" }
//	func (s *ScoreManager) Update(dt float64) { /* per-frame logic */ }
//
//	// Register at startup.
//	autoload.Registry.Set("ScoreManager", &ScoreManager{})
//
//	// Access from anywhere.
//	sm := autoload.Registry.Get("ScoreManager").(*ScoreManager)
//	sm.Score += 100
package autoload

import (
	"fmt"
	"sync"
)

// AutoLoad is the interface that global singleton objects must satisfy.
//
// Name returns a globally unique identifier used to access the object
// from KScript and from other Go code.
//
// Update is called once per frame with the delta time in seconds.
type AutoLoad interface {
	Name() string
	Update(dt float64)
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

// Registry holds all active AutoLoad singletons.
// Use the global Registry variable to register and access singletons.
type Registry struct {
	mu        sync.RWMutex
	instances map[string]AutoLoad
	order     []string // insertion order for deterministic Update
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		instances: make(map[string]AutoLoad),
		order:     make([]string, 0),
	}
}

// Set registers an AutoLoad instance by its Name.
// If a singleton with the same name already exists, it is replaced.
// Set must be called before the game loop starts for predictable behavior;
// however, it is safe to call at any time.
func (r *Registry) Set(a AutoLoad) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := a.Name()
	if _, exists := r.instances[name]; !exists {
		r.order = append(r.order, name)
	}
	r.instances[name] = a
}

// Get returns the AutoLoad instance registered under name, or nil if
// no instance with that name exists.
func (r *Registry) Get(name string) AutoLoad {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.instances[name]
}

// Remove unregisters the AutoLoad instance with the given name.
// It is safe to call mid-frame; the instance will no longer receive
// Update calls.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.instances, name)
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// Update calls Update on every registered AutoLoad in insertion order.
// This is called once per frame by the runner.
func (r *Registry) Update(dt float64) {
	r.mu.RLock()
	// Copy the order slice under the lock so we can release before calling
	// user code (avoiding potential deadlocks if user code calls back into
	// the registry).
	order := make([]string, len(r.order))
	copy(order, r.order)
	r.mu.RUnlock()

	for _, name := range order {
		r.mu.RLock()
		inst := r.instances[name]
		r.mu.RUnlock()
		if inst != nil {
			inst.Update(dt)
		}
	}
}

// Len returns the number of registered AutoLoad instances.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.instances)
}

// Names returns all registered AutoLoad names in insertion order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// Clear removes all registered AutoLoad instances.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances = make(map[string]AutoLoad)
	r.order = make([]string, 0)
}

// ---------------------------------------------------------------------------
// Global registry
// ---------------------------------------------------------------------------

// Registry is the global AutoLoad registry used by the runner and KScript.
// It is initialised once at startup.
var Global = NewRegistry()

// ---------------------------------------------------------------------------
// Base — embeddable helper
// ---------------------------------------------------------------------------

// Base provides a no-op AutoLoad implementation that can be embedded.
// Override Name() and optionally Update() in your concrete type.
//
//	type ScoreManager struct {
//	    autoload.Base
//	    Score int
//	}
//	func (s *ScoreManager) Name() string { return "ScoreManager" }
type Base struct{}

// Name returns "AutoLoad" by default. Override in your concrete type.
func (Base) Name() string { return "AutoLoad" }

// Update is a no-op. Override if your singleton needs per-frame logic.
func (Base) Update(float64) {}

// String returns the type name for debugging.
func (Base) String() string { return "autoload.Base" }

// ---------------------------------------------------------------------------
// Convenience functions for the global Registry
// ---------------------------------------------------------------------------

// Set registers a on the global Registry.
func Set(a AutoLoad) { Global.Set(a) }

// Get returns an AutoLoad from the global Registry by name.
func Get(name string) AutoLoad { return Global.Get(name) }

// Remove removes an AutoLoad from the global Registry by name.
func Remove(name string) { Global.Remove(name) }

// MustGet returns the AutoLoad with the given name, panicking if not found.
// Useful in initialization code where a missing singleton is a programmer error.
func MustGet(name string) AutoLoad {
	a := Global.Get(name)
	if a == nil {
		panic(fmt.Sprintf("autoload: singleton %q not registered", name))
	}
	return a
}

// UpdateAll calls Update on every registered AutoLoad on the global Registry.
func UpdateAll(dt float64) { Global.Update(dt) }

// Len returns the number of registered AutoLoad instances on the global Registry.
func Len() int { return Global.Len() }

// Names returns all registered AutoLoad names in insertion order from the global Registry.
func Names() []string { return Global.Names() }
