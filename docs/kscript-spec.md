# KScript Language Specification — Draft v0.1

> KScript is the scripting language of the Kora Engine. It uses TypeScript-like syntax, is statically typed, and compiles ahead-of-time (AOT) to Go. No interpreter runs on the target device.

---

## Goals

- Familiar syntax for developers who know TypeScript or JavaScript.
- Simple enough that designers can learn it quickly.
- Statically typed and fully analyzable at compile time.
- Supports `async/await` via engine-managed coroutines.
- No access to system resources — sandboxed to the engine API.

---

## Types

| Type | Description |
|---|---|
| `int` | 32-bit integer |
| `float` | 64-bit float |
| `bool` | Boolean |
| `string` | Immutable UTF-8 string |
| `Vec2` | 2D vector `{ x: float, y: float }` |
| `Entity` | Reference to a game object |
| `Sprite` | Sprite resource |
| `Sound` | Audio resource |
| `Task` | Async operation handle |
| `Signal` | Named signal |
| `void` | No return value |

Arrays: `int[]`, `string[]`, `Entity[]`.

Dicts: `Dict<string, int>` (limited to primitive value types).

---

## Object Declaration

Every game entity is declared as an `object`. Objects map to an entity in the scene.

```ts
object Enemy {
  speed: float = 80
  hp: int = 3

  update(dt: float) { ... }
  draw() { ... }
}
```

---

## Lifecycle Events

| Event | Signature | When called |
|---|---|---|
| `create()` | `create()` or `async create()` | Object spawned |
| `update(dt)` | `update(dt: float)` | Every frame |
| `draw()` | `draw()` | Render pass |
| `destroy()` | `destroy()` | Object removed |
| `onCollision(other)` | `onCollision(other: Entity)` | Collision detected |
| `onSignal(name)` | `onSignal(name: string)` | Signal received |

---

## Async / Await

`async/await` is supported only for engine Task primitives. It is compiled to a state machine — no coroutine runtime is shipped with the game binary.

```ts
async create() {
  await wait(0.5)
  await tween(this, { alpha: 1.0 }, 0.3)
  emit("ready")
}
```

### Async Primitives

```ts
wait(seconds: float): Task
waitFrames(n: int): Task
waitSignal(obj: Entity, name: string): Task
tween(obj: Entity, props: TweenProps, duration: float): Task
race(...tasks: Task[]): Task
all(...tasks: Task[]): Task
cancel(task: Task): void
```

---

## Engine API (partial)

### Input
```ts
Input.pressed(key: string): bool
Input.held(key: string): bool
Input.released(key: string): bool
Input.axisX(): float   // -1.0 to 1.0
Input.axisY(): float
Input.touchPos(): Vec2
```

### Audio
```ts
Audio.play(sound: Sound): void
Audio.stop(sound: Sound): void
Audio.setVolume(v: float): void
```

### Scene
```ts
Scene.spawn(name: string): Entity
Scene.destroy(obj: Entity): void
Scene.find(name: string): Entity
Scene.load(name: string): void
```

### Entity (this)
```ts
this.x: float
this.y: float
this.velX: float
this.velY: float
this.alpha: float
this.visible: bool
this.tag: string
this.onGround(): bool
this.destroy(): void
this.emit(signal: string): void
```

---

## Compiler Pipeline

```
KScript source (.ks)
    │
    ▼
 Lexer  →  token stream
    │
    ▼
 Parser →  AST
    │
    ▼
 Type Checker + Semantic Analysis
    │
    ▼
 Async Transformer  →  state machine rewrite
    │
    ▼
 Go Emitter  →  generated .go files
    │
    ▼
 go build (Android target)
    │
    ▼
 APK / AAB
```

---

## Restrictions

The following are intentionally **not supported** in KScript:

- Dynamic typing or `any`
- `eval` or runtime code execution
- Reflection
- Free concurrency (no raw goroutines)
- Direct filesystem, network or OS access
- Prototype manipulation
- Unbounded recursion (compiler warning)
- Arbitrary generics (only engine-provided generics like `Dict<K,V>`)

---

## Reserved Keywords

```
object  func  async  await  return  if  else  for  while
break  continue  const  var  true  false  null
import  from  this  emit  signal
```
