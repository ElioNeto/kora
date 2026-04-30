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
		"getGravityX": func() float64 {
			return float64(world.Gravity.X)
		},
		"getGravityY": func() float64 {
			return float64(world.Gravity.Y)
		},
		"raycast": func(fromX, fromY, toX, toY float64, mask int) map[string]interface{} {
			from := physics.Vec2{float32(fromX), float32(fromY)}
			to := physics.Vec2{float32(toX), float32(toY)}
			hit := world.Raycast(from, to, uint16(mask))
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
			bodies := world.OverlapRect(float32(x), float32(y), float32(w), float32(h), uint16(mask))
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
