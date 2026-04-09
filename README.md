# Kora Engine

> A 2D game engine for Android, built with a Go core and KScript — a statically compiled scripting language with async/await support.

---

## What is Kora?

Kora is an open 2D game engine targeting Android, inspired by GameMaker Studio. It combines a high-performance Go runtime with **KScript**, a TypeScript-like language designed exclusively for game logic. KScript is compiled ahead-of-time (AOT) — no VM runs on the device.

---

## Core Principles

- **Go Core** — rendering, input, audio, physics, scene graph and asset pipeline.
- **KScript** — familiar TypeScript-like syntax, compiled to Go, no script runtime on device.
- **AOT Compilation** — KScript transpiles to Go, then everything compiles to a native Android APK/AAB.
- **Async-first** — coroutine-style `async/await` backed by an engine scheduler, not a JS event loop.
- **2D only (v1)** — sprites, tilemaps, cameras, particles, UI.
- **Visual Editor** — scene editor, sprite editor, object inspector, project manager.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────┐
│                  Kora Editor                     │
│  (Scene Editor · Inspector · Asset Browser)      │
└───────────────────┬──────────────────────────────┘
                    │ project files (.kora, .ks, assets)
                    ▼
┌──────────────────────────────────────────────────┐
│              KScript Compiler                    │
│  Lexer → Parser → AST → Type Check → Go Emitter  │
└───────────────────┬──────────────────────────────┘
                    │ generated Go files
                    ▼
┌──────────────────────────────────────────────────┐
│               Kora Runtime (Go)                  │
│  Render · Input · Audio · Physics · Scene · ECS  │
│  Async Scheduler · Asset Loader · Export         │
└───────────────────┬──────────────────────────────┘
                    │ compiled binary
                    ▼
          Android APK / AAB
```

---

## KScript — Language Overview

KScript is a restricted, statically typed language with TypeScript-like syntax. It compiles to Go source code before the Android build — no script interpreter runs on the device.

### Key Constraints
- No dynamic types, reflection or eval.
- No access to OS, filesystem or network directly.
- Only engine-provided modules can be imported.
- `async/await` is supported but limited to engine Task primitives.

### Example

```ts
object Player {
  speed: float = 180
  hp: int = 5

  async create() {
    await wait(0.3)
    this.hp = 5
  }

  update(dt: float) {
    const ax = Input.axisX()
    if (ax != 0) {
      this.x += ax * this.speed * dt
    }

    if (Input.pressed("jump") && this.onGround()) {
      this.velY = -320
    }
  }

  async onHit(other: Entity) {
    this.hp -= 1
    await tween(this, { alpha: 0.0 }, 0.1)
    await tween(this, { alpha: 1.0 }, 0.1)
    if (this.hp <= 0) {
      await signal(this, "dead")
      this.destroy()
    }
  }
}
```

### Async Primitives

| Primitive | Description |
|---|---|
| `wait(seconds)` | Pause for N seconds |
| `waitFrames(n)` | Pause for N frames |
| `waitSignal(obj, name)` | Wait until a signal is emitted |
| `tween(obj, props, duration)` | Animate properties |
| `race(a, b)` | Continue on whichever task finishes first |
| `all(a, b, c)` | Wait for all tasks to complete |
| `cancel(task)` | Cancel a running task |

---

## Project Structure

```
kora/
├── core/              # Go runtime: render, input, audio, physics, ECS
│   ├── render/
│   ├── input/
│   ├── audio/
│   ├── physics/
│   ├── scene/
│   ├── async/         # Task scheduler and coroutine model
│   └── assets/
├── compiler/          # KScript compiler
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── checker/       # Type checker and semantic analysis
│   └── emitter/       # Go code emitter
├── editor/            # Visual editor (desktop app)
├── export/
│   └── android/       # APK/AAB build pipeline
├── stdlib/            # KScript standard library definitions
├── examples/          # Example games and scenes
└── docs/              # Documentation
```

---

## Roadmap

### v0.1 — Core Runtime
- [ ] 2D renderer (sprites, tilemaps, camera)
- [ ] Input system (keyboard, touch)
- [ ] Audio (sfx, music)
- [ ] Scene graph + entity system
- [ ] Basic AABB physics
- [ ] Asset loader

### v0.2 — KScript Compiler
- [ ] Lexer and parser
- [ ] AST and type checker
- [ ] Go code emitter
- [ ] Async/await → state machine compilation
- [ ] Engine API bindings

### v0.3 — Editor
- [ ] Scene editor (drag/drop objects)
- [ ] Sprite and animation editor
- [ ] Object inspector
- [ ] KScript editor with syntax highlighting
- [ ] Project manager

### v0.4 — Android Export
- [ ] APK build pipeline
- [ ] AAB build for Google Play
- [ ] Keystore and signing
- [ ] Android manifest configuration

### v1.0 — Public Release
- [ ] Stable KScript API
- [ ] Template projects
- [ ] Documentation site
- [ ] Example games

---

## Technology Stack

| Layer | Technology |
|---|---|
| Core runtime | Go |
| Scripting language | KScript (custom, TS-like) |
| Mobile export | gomobile + Android SDK |
| Editor | TBD (desktop, cross-platform) |
| 2D rendering | Ebitengine (Go) |

---

## License

MIT License — see [LICENSE](LICENSE).
