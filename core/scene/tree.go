// Package scene provides the core scene management system for Kora Engine.
package scene

import (
	"github.com/ElioNeto/kora/core/async"
	"github.com/ElioNeto/kora/core/render"
)

// ProcessMode defines how an entity responds to pause state.
type ProcessMode int

const (
	// ProcessModeAlways runs even when the tree is paused.
	ProcessModeAlways ProcessMode = iota
	// ProcessModePausable runs only when not paused (default).
	ProcessModePausable
	// ProcessModeWhenPaused runs only when the tree is paused.
	ProcessModeWhenPaused
)

// EntityWithProcessMode extends Entity with pause behavior.
type EntityWithProcessMode interface {
	Entity
	GetProcessMode() ProcessMode
}

// DrawableEntity extends Entity with rendering.
type DrawableEntity interface {
	Entity
	Draw(r *render.Renderer)
}

// DrawingEntity is an alias for entities that use *ebiten.Image for drawing.
type DrawingEntity interface {
	Entity
	Draw(r interface{})
}

// PhysicsEntity extends Entity with physics updates.
type PhysicsEntity interface {
	Entity
	PhysicsUpdate(dt float64)
}

// SceneTree orchestrates the game loop by calling Update/PhysicsUpdate/Draw
// on all active entities. It supports pause state, scene transitions, and
// separate fixed-timestep physics from variable-timestep rendering.
type SceneTree struct {
	current      *Scene
	scenes       map[string]*Scene
	transitionTo *Scene // queued scene change
	paused       bool
	scheduler    *async.Scheduler
	physicsDt    float64 // fixed timestep for physics (1/60)
}

// NewSceneTree creates a new SceneTree with default physics timestep.
func NewSceneTree() *SceneTree {
	return &SceneTree{
		scenes:    make(map[string]*Scene),
		scheduler: async.NewScheduler(),
		physicsDt: 1.0 / 60.0,
	}
}

// GetCurrentScene returns the currently active scene.
func (st *SceneTree) GetCurrentScene() *Scene {
	if st.current == nil {
		st.current = New()
	}
	return st.current
}

// SetCurrentScene sets the current active scene.
func (st *SceneTree) SetCurrentScene(scene *Scene) {
	st.current = scene
}

// RegisterScene registers a scene by path for later switching.
func (st *SceneTree) RegisterScene(path string, scene *Scene) {
	st.scenes[path] = scene
}

// ChangeScene requests a scene change to be executed on the next Tick.
// This ensures we never change scenes mid-frame.
// Returns true if successful, false if scene not found.
func (st *SceneTree) ChangeScene(path string) bool {
	scene, ok := st.scenes[path]
	if !ok {
		return false
	}
	st.transitionTo = scene
	return true
}

// Pause sets the tree to pause state - Update/PhysicsUpdate are skipped
// but Draw continues (for pause menus, UI overlays, etc.).
func (st *SceneTree) Pause() {
	st.paused = true
}

// Resume undoes the effect of Pause.
func (st *SceneTree) Resume() {
	st.paused = false
}

// IsPaused returns whether the tree is currently paused.
func (st *SceneTree) IsPaused() bool {
	return st.paused
}

// Tick advances the entire scene tree by dt seconds.
// It properly separates physics (fixed timestep) from rendering (variable timestep).
func (st *SceneTree) Tick(dt float64) {
	// Handle pending scene transitions at start of tick.
	st.applySceneTransition()

	if st.current == nil {
		return
	}

	// Phase 1: Physics simulation (fixed timestep) - only when not paused.
	if !st.paused {
		st.current.PhysicsUpdate(st.physicsDt)
	}

	// Phase 2: Game logic (variable timestep).
	// When paused, only ProcessModeAlways and ProcessModeWhenPaused nodes run.
	st.current.UpdateProcessMode(dt, st.paused)

	// Phase 3: Scheduler tick (async coroutines).
	st.scheduler.Tick(dt)
}

// Draw renders the entire scene tree to the given renderer.
// This phase always runs regardless of pause state.
func (st *SceneTree) Draw(r *render.Renderer) {
	if st.current != nil {
		st.current.Draw(r)
	}
}

// Len returns the number of entities in the current scene.
func (st *SceneTree) Len() int {
	if st.current == nil {
		return 0
	}
	return st.current.Count()
}

// Find finds an entity in the current scene by group.
func (st *SceneTree) Find(group string) Entity {
	if st.current == nil {
		return nil
	}
	return st.current.Find(group)
}

// applySceneTransition applies a queued scene transition.
func (st *SceneTree) applySceneTransition() {
	if st.transitionTo == nil {
		return
	}

	// Clean up old scene if exists.
	if st.current != nil {
		st.current.DestroyAll()
	}

	// Switch to new scene.
	st.current = st.transitionTo
	st.transitionTo = nil
}
