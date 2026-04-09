# Kora Engine — Architecture

## Overview

Kora is divided into three independent layers:

1. **Core Runtime** (Go) — the engine itself.
2. **KScript Compiler** — transpiles user code to Go.
3. **Editor** — visual tool for building games.

All three converge at build time into a single native binary for Android.

---

## Core Runtime

Written in Go. Responsible for:

- **Render** — 2D sprites, tilemaps, camera, draw calls.
- **Input** — keyboard, mouse, touch, virtual gamepads.
- **Audio** — sound effects, background music, volume control.
- **Physics** — AABB collision, velocity, gravity (simple 2D).
- **Scene** — scene graph, entity lifecycle, transitions.
- **ECS** — optional entity-component system for advanced objects.
- **Async Scheduler** — runs compiled async tasks each frame.
- **Asset Loader** — sprites, sounds, tilemaps, fonts.

### Async Scheduler

The scheduler is a lightweight per-frame task runner. Each `async` function compiles to a state machine struct. The scheduler advances ready tasks every game tick, handling:

- Timer expiry
- Signal emission
- Tween completion
- Resource loading

No goroutines are exposed to user code.

---

## KScript Compiler

The compiler processes `.ks` files and emits `.go` source files, which are compiled alongside the core runtime.

### Phases

| Phase | Input | Output |
|---|---|---|
| Lexer | `.ks` source text | Token stream |
| Parser | Token stream | Raw AST |
| Type Checker | AST | Annotated AST + errors |
| Async Transformer | Annotated AST | State machine AST |
| Go Emitter | Final AST | `.go` source files |

### Async Transform

Each `async` function is rewritten as a struct implementing:
```go
type Task interface {
    Tick(dt float64) TaskStatus
}
```
`await` points become state transitions. Local variables become struct fields.

---

## Editor

Desktop application. Features:

- **Scene Editor** — drag and drop objects, set properties.
- **Sprite Editor** — import, crop, animate.
- **KScript Editor** — syntax highlighting, error markers, autocomplete.
- **Inspector** — edit object variables exposed via KScript.
- **Project Manager** — create, open, configure projects.
- **Export** — one-click build to APK or AAB.

---

## Export Pipeline (Android)

1. Compile KScript → Go.
2. `go build` with `GOOS=android GOARCH=arm64`.
3. Package assets into APK structure.
4. Sign with keystore.
5. Output `.apk` (debug) or `.aab` (release / Google Play).

---

## Project File Format

Projects are folders with a `project.kora` file (TOML-based):

```toml
[project]
name = "My Game"
package = "com.example.mygame"
version = "1.0.0"
target_sdk = 35
min_sdk = 26

[window]
width = 360
height = 640
orientation = "portrait"

[audio]
master_volume = 1.0

[export.android]
keystore = "./keys/release.keystore"
```
