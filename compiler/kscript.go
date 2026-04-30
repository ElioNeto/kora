package compiler

import (
	"github.com/ElioNeto/kora/core/physics"
)

// RegisterPhysicsAPI registers the Physics namespace and Body methods with KScript.
func RegisterPhysicsAPI() {
	// Physics namespace methods
	registerGlobal("Physics", map[string]interface{}{
		"setGravity": func(x, y float64) {
			// Assumes a global default physics world; adjust as needed
			// For now, use a package-level world or pass from engine
			// TODO: Connect to the active PhysicsWorld instance
		},
		"raycast": func(fromX, fromY, toX, toY float64, mask int) map[string]interface{} {
			// TODO: Connect to active PhysicsWorld
			hit := physics.RaycastHit{}
			return map[string]interface{}{
				"hit":      hit.Hit,
				"pointX":   float64(hit.Point.X),
				"pointY":   float64(hit.Point.Y),
				"normalX":  float64(hit.Normal.X),
				"normalY":  float64(hit.Normal.Y),
				"fraction": float64(hit.Fraction),
			}
		},
		"overlapRect": func(x, y, w, h float64, mask int) []interface{} {
			// TODO: Connect to active PhysicsWorld
			return nil
		},
	})

	// Body instance methods
	registerType("RigidBody", map[string]interface{}{
		"applyForce": func(b *physics.RigidBody, x, y float64) {
			b.ApplyForce(physics.Vec2{float32(x), float32(y)})
		},
		"applyImpulse": func(b *physics.RigidBody, x, y float64) {
			b.ApplyImpulse(physics.Vec2{float32(x), float32(y)})
		},
		"getVelocity": func(b *physics.RigidBody) (float32, float32) {
			return b.Vel.X, b.Vel.Y
		},
		"setVelocity": func(b *physics.RigidBody, x, y float32) {
			b.Vel = physics.Vec2{x, y}
		},
		"getMass": func(b *physics.RigidBody) float32 {
			return b.Mass
		},
		"setMass": func(b *physics.RigidBody, m float32) {
			b.Mass = m
		},
	})
}

// TODO: Implement registerGlobal and registerType based on KScript runtime
// These are placeholders assuming the KScript compiler has a way to register native bindings.
func registerGlobal(name string, methods map[string]interface{}) {}
func registerType(name string, methods map[string]interface{}) {}
