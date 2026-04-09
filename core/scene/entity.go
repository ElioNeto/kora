// Package scene manages the lifecycle of all game objects (entities).
//
// Entities are spawned into a Scene, updated every frame, and destroyed
// when their alive flag is cleared. The Scene owns a Scheduler so that
// async tasks started by an entity are automatically cancelled when the
// entity is destroyed.
package scene

import "github.com/ElioNeto/kora/core/async"

// ----------------------------------------------------------------------------
// Entity interface
// ----------------------------------------------------------------------------

// Entity is the minimum contract every KScript object must satisfy.
// The compiler emits IsAlive / Destroy / EmitSignal on every generated struct.
type Entity interface {
	IsAlive() bool
	Destroy()
}

// Updater is an Entity that participates in the update loop.
type Updater interface {
	Entity
	Update(dt float64)
}

// Drawer is an Entity that knows how to draw itself.
// The render.Renderer parameter is kept as interface{} here to avoid
// an import cycle; the render package casts it back.
type Drawer interface {
	Entity
	Draw(r interface{})
}

// Initer is an Entity with an onCreate lifecycle hook.
type Initer interface {
	Entity
	Create()
}

// AsyncIniter is an Entity whose onCreate returns a runnable async Task.
type AsyncIniter interface {
	Entity
	CreateTask() async.Task
}

// SignalEmitter allows entities to fire named signals consumed by WaitSignal.
type SignalEmitter interface {
	EmitSignal(name string)
	SignalFired(name string) bool
	ClearSignals()
}

// ----------------------------------------------------------------------------
// BaseEntity — embed in generated structs for signal support
// ----------------------------------------------------------------------------

// BaseEntity provides signal bookkeeping. Generated structs embed this.
type BaseEntity struct {
	signals map[string]bool
}

func (b *BaseEntity) EmitSignal(name string) {
	if b.signals == nil {
		b.signals = make(map[string]bool)
	}
	b.signals[name] = true
}

func (b *BaseEntity) SignalFired(name string) bool {
	if b.signals == nil {
		return false
	}
	return b.signals[name]
}

func (b *BaseEntity) ClearSignals() {
	for k := range b.signals {
		delete(b.signals, k)
	}
}
