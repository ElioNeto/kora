package compiler

import (
	"github.com/ElioNeto/kora/core/physics"
)

// RegisterPhysicsAPI registers the Physics namespace and Body methods with KScript.
// world: active PhysicsWorld instance to bind to Physics namespace functions.
func RegisterPhysicsAPI(world *physics.PhysicsWorld) {
	// Physics namespace methods
	registerGlobal("Physics", map[string]interface{}{
		"setGravity": func(x, y float64) {
			world.SetGravity(float32(x), float32(y))
		},
		"raycast": func(fromX, fromY, toX, toY float64, mask int) map[string]interface{} {
			from := physics.Vec2{X: float32(fromX), Y: float32(fromY)}
			to := physics.Vec2{X: float32(toX), Y: float32(toY)}
			hit := world.Raycast(from, to, uint16(mask))
			return map[string]interface{}{
				"hit":     hit.Hit,
				"x":       float64(hit.Point.X),
				"y":       float64(hit.Point.Y),
				"normalX": float64(hit.Normal.X),
				"normalY": float64(hit.Normal.Y),
			}
		},
		"overlapRect": func(minX, minY, maxX, maxY float64, mask int) []interface{} {
			bodies := world.OverlapRect(float32(minX), float32(minY), float32(maxX), float32(maxY), uint16(mask))
			result := make([]interface{}, len(bodies))
			for i, b := range bodies {
				result[i] = b
			}
			return result
		},
	})

	// RigidBody2D instance methods
	registerType("RigidBody2D", physics.RegisterRigidBody2DAPI())

	// CharacterBody2D instance methods
	registerType("CharacterBody2D", physics.RegisterCharacterBody2DAPI())

	// StaticBody2D instance methods
	registerType("StaticBody2D", physics.RegisterStaticBody2DAPI())

	// Area2D instance methods
	registerType("Area2D", physics.RegisterArea2DAPI())
}

// TODO: Implement registerGlobal and registerType based on KScript runtime
// These are placeholders assuming the KScript compiler has a way to register native bindings.
func registerGlobal(name string, methods map[string]interface{}) {}
func registerType(name string, methods map[string]interface{}) {}
