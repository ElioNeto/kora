# core/physics

Lightweight 2D physics for the Kora game engine.

## Overview

| File | Responsibility |
|---|---|
| `body.go` | `Vec2`, `BodyType`, `ShapeType`, `RigidBody` — physical state of one entity |
| `collider.go` | Collision detection/resolution for Rect and Circle shapes |
| `world.go` | `PhysicsWorld` — 2D gravity, fixed 60 TPS step, layer/mask filtering |
| `raycast.go` | `Raycast` and `OverlapRect` with mask support |
| `world_test.go` | 19+ table-driven unit tests |

## Body types

| Type | Gravity | Moved by engine | Use case |
|---|---|---|---|
| `BodyDynamic` | ✅ | ✅ | Players, enemies, projectiles |
| `BodyKinematic` | ❌ | ✅ | Moving platforms, doors |
| `BodyStatic` | ❌ | ❌ | Ground, walls, ceilings |

## Shape types

| Type | Description |
|---|---|
| `ShapeRect` | Axis-aligned rectangle (uses HalfW, HalfH) |
| `ShapeCircle` | Circle (uses Radius) |

## Collision layers

Each body has a 16-bit `Layer` (which layer it belongs to) and `Mask` (which layers it collides with). Use `DefaultLayer` and `DefaultMask` for standard setups.

## Usage

```go
import "github.com/ElioNeto/kora/core/physics"

// Create world (pass a TileQuery if using a tilemap)
world := physics.NewWorld(tilemap.IsSolid)

// Set 2D gravity (x, y)
world.SetGravity(0, 980)

// Register rect body
player := physics.NewBody(playerID, 100, 50, 32, 48, physics.BodyDynamic)
// Register circle body
enemy := physics.NewCircleBody(enemyID, 200, 50, 16, physics.BodyDynamic)

ground := physics.NewBody(groundID, 180, 300, 360, 24, physics.BodyStatic)
world.Register(player)
world.Register(enemy)
world.Register(ground)

// Game loop: pass actual frame dt, physics runs at fixed 60 TPS
func Update(dt float32) {
    world.Step(dt)
}

// Apply forces/impulses
player.ApplyForce(physics.Vec2{0, -500})
player.ApplyImpulse(physics.Vec2{200, 0})

// Raycast
hit := world.Raycast(physics.Vec2{0,0}, physics.Vec2{0, 200}, 0xFFFF)
if hit.Hit {
    fmt.Printf("Hit body %d at %v\n", hit.Body.EntityID, hit.Point)
}

// Overlap rect
bodies := world.OverlapRect(40, 40, 60, 60, 0xFFFF)
```

## Tilemap integration

`TileQuery` is `func(px, py float32) bool` — return `true` if the world
coordinate (px, py) falls inside a solid tile. The physics world probes
four cardinal points around each body every step.

```go
func (t *Tilemap) IsSolid(px, py float32) bool {
    col := int(px) / t.TileSize
    row := int(py) / t.TileSize
    if col < 0 || row < 0 || col >= t.Cols || row >= t.Rows {
        return false
    }
    return t.Tiles[row][col].Solid
}
```

## Running tests

```bash
go test ./core/physics/...
# ok  github.com/ElioNeto/kora/core/physics
```

## Constants

| Constant | Value | Unit |
|---|---|---|
| `DefaultGravityX` | 0 | px/s² |
| `DefaultGravityY` | 980 | px/s² |
| `FixedPhysicsStep` | 1/60 | s (60 TPS) |
| `TerminalVelocity` | 1200 | px/s |
| Tile size (probe step) | 32 | px |
