# core/physics

Lightweight 2D physics for the Kora game engine.

## Overview

| File | Responsibility |
|---|---|
| `body.go` | `Vec2`, `BodyType`, `RigidBody` — physical state of one entity |
| `collider.go` | `TestAABB` / `ResolveAABB` — overlap detection and resolution |
| `world.go` | `PhysicsWorld.Step(dt)` — gravity, integration, body-body + tilemap collision |
| `world_test.go` | 12 table-driven unit tests |

## Body types

| Type | Gravity | Moved by engine | Use case |
|---|---|---|---|
| `BodyDynamic` | ✅ | ✅ | Players, enemies, projectiles |
| `BodyKinematic` | ❌ | ✅ | Moving platforms, doors |
| `BodyStatic` | ❌ | ❌ | Ground, walls, ceilings |

## Usage

```go
import "github.com/ElioNeto/kora/core/physics"

// Create world (pass a TileQuery if using a tilemap)
world := physics.NewWorld(tilemap.IsSolid)

// Register bodies
player := physics.NewBody(playerID, 100, 50, 32, 48, physics.BodyDynamic)
ground := physics.NewBody(groundID, 180, 300, 360, 24, physics.BodyStatic)
world.Register(player)
world.Register(ground)

// Game loop
func Update(dt float32) {
    // Apply input before stepping
    if input.IsPressed(input.KeyRight) {
        player.Vel.X = 200
    }
    if input.IsJustPressed(input.KeySpace) && player.IsGrounded {
        player.Vel.Y = -500
    }

    world.Step(dt)

    // Sync render position from physics body
    renderEntity[playerID].X = player.Pos.X
    renderEntity[playerID].Y = player.Pos.Y
}
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
| `DefaultGravity` | 980 | px/s² |
| `TerminalVelocity` | 1200 | px/s |
| Tile size (probe step) | 32 | px |
