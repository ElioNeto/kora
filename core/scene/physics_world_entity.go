package scene

import (
	physics "github.com/ElioNeto/kora/core/physics"
)

// PhysicsWorldEntity wraps PhysicsWorld as a scene.Entity that implements PhysicsNode.
// This is the bridge between the physics system and the SceneTree.
type PhysicsWorldEntity struct {
	*physics.PhysicsWorld
	alive bool
}

// NewPhysicsWorldEntity creates a new PhysicsWorldEntity.
func NewPhysicsWorldEntity(tileQ physics.TileQuery) *PhysicsWorldEntity {
	return &PhysicsWorldEntity{
		PhysicsWorld: physics.NewWorld(tileQ),
		alive:       true,
	}
}

// IsAlive implements scene.Entity.
func (e *PhysicsWorldEntity) IsAlive() bool {
	return e.alive
}

// Destroy implements scene.Entity.
func (e *PhysicsWorldEntity) Destroy() {
	e.alive = false
}

// PhysicsUpdate implements scene.PhysicsNode.
// This is called by SceneTree at fixed timestep (60 TPS).
func (e *PhysicsWorldEntity) PhysicsUpdate(dt float64) {
	e.PhysicsWorld.Step(float32(dt))
}

// Ensure compile-time interface check.
var _ PhysicsNode = (*PhysicsWorldEntity)(nil)
