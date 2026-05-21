# Kora Engine

[![CI](https://github.com/ElioNeto/kora/actions/workflows/ci.yml/badge.svg)](https://github.com/ElioNeto/kora/actions/workflows/ci.yml)
[![Android Build](https://github.com/ElioNeto/kora/actions/workflows/android.yml/badge.svg)](https://github.com/ElioNeto/kora/actions/workflows/android.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ElioNeto/kora)](https://goreportcard.com/report/github.com/ElioNeto/kora)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> Engine de jogos 2D para Android com editor visual e linguagem KScript compilada para Go.

## 🎮 O que é Kora?

Kora Engine é uma engine de jogos 2D projetada para criar jogos nativos para Android. Componentes principais:

- **Editor Web** — Crie cenas visualmente com arrastar e soltar, hierarquia de entidades, inspetor de propriedades e editor de código KScript
- **KScript** — Linguagem própria compilada para Go (sem VM, zero overhead)
- **Runtime Go** — Performance nativa ARM64 via gomobile no Android

### Principais Características

- ✅ **Editor Visual** — Hierarquia, viewport, inspetor, grid snapping, multi-seleção
- ✅ **Assets Import** — PNG, JPG, OGG, WAV com persistência IndexedDB
- ✅ **KScript** — TypeScript-like, async/await, tipagem estática, signals
- ✅ **Física 2D** — AABB, gravidade, raycas, áreas, CharacterBody2D, CCD, joints, spatial hash
- ✅ **Renderização** — Sprites, spritesheets, tilemaps, frustum culling, câmera follow+zoom+shake
- ✅ **Áudio** — OGG/WAV/MP3, mixer multi-bus, som espacial, pitch, pan
- ✅ **Animação** — Keyframe AnimationPlayer, cutscene sequencer, 28 funções de easing
- ✅ **Sistema de Partículas** — CPU-based com emissão contínua/burst, gravidade, blend modes
- ✅ **Iluminação 2D** — PointLight2D, DirectionalLight2D, sombras dinâmicas
- ✅ **Pathfinding** — A* com grid navigation, smoothing, obstacles dinâmicos
- ✅ **Skeleton2D** — Hierarquia de ossos com CCD IK solver
- ✅ **Parallax** — Múltiplas camadas com scroll independente
- ✅ **UI** — Label, Button, Panel (usando bitmap font integrada)
- ✅ **Debugger** — Overlay com FPS, entidades, tasks, árvore de nós
- ✅ **Shaders** — Kage (ESSL), ShaderManager, ShaderNode
- ✅ **Prefabs** — Templates reutilizáveis com deep copy
- ✅ **Object Pool** — Pool genérico thread-safe para redução de GC
- ✅ **Asset Manager** — Carregamento síncrono/assíncrono com ref counting
- ✅ **Canvas/Inspector** — z-index, collapsible sections, tags, Ctrl+Enter apply
- ✅ **Git Panel** — Status, stage, unstage, commit, diff, histórico
- ✅ **APK Nativo** — Build para Android com runtime Go via gomobile
- ✅ **Benchmarks** — 30+ benchmarks nos sistemas críticos

## 🏗️ Arquitetura

```
┌──────────────────────────────────────────────────────────────┐
│               Editor Web (HTML5/JS)                          │
│  Cena · Preview · Assets · Script (CodeMirror 6) · Git      │
│  Hierarquia · Inspetor · Serializer · Grid Snap · Multi-sel │
└──────────────────────┬───────────────────────────────────────┘
                       │ .kora.json / .kora.prefab
                       ▼
┌──────────────────────────────────────────────────────────────┐
│          KScript Compiler (Go)                                │
│  Lexer → Parser → Type Checker → Async Transform → Emitter   │
│  (Gera structs Go com Entity interface + state machines)     │
└──────────────────────┬───────────────────────────────────────┘
                       │ Código Go gerado (+ runtime linkado)
                       ▼
┌──────────────────────────────────────────────────────────────┐
│               Kora Runtime (Go + Ebitengine v2.7)             │
│                                                              │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐ ┌─────────────────┐   │
│  │ Render  │ │ Physics  │ │  Audio  │ │     Input       │   │
│  │ Sprites │ │ AABB     │ │ OGG/WAV │ │  Keyboard/Touch │   │
│  │ Tilemap │ │ CCD      │ │ Mixer   │ │  Virtual Pad    │   │
│  │ Font    │ │ Joints   │ │ Spatial │ │  Gesture (P2)   │   │
│  │ Camera  │ │ Spatial  │ │         │ │  Gamepad (P2)   │   │
│  │ Shaders │ │ Raycast  │ │         │ │                 │   │
│  │ Culling │ │ Character│ │         │ │                 │   │
│  └─────────┘ └──────────┘ └─────────┘ └─────────────────┘   │
│                                                              │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐ ┌─────────────────┐   │
│  │ Scene   │ │  Node2D  │ │  Async  │ │    Extras       │   │
│  │ Entity  │ │ Camera2D │ │  Tasks  │ │ Particles2D     │   │
│  │ Manager │ │ Sprite2D │ │  Tweens │ │ Skeleton2D      │   │
│  │ Loader  │ │PhysicsBd │ │  Sched  │ │ Light2D         │   │
│  │ Prefab  │ │ AudioPl2 │ │  Pool   │ │ Parallax        │   │
│  │ AutoLoad│ │Animation │ │  Easing │ │ Cutscene        │   │
│  │ Tree    │ │ UI (Lbl) │ │         │ │ DebugConsole    │   │
│  └─────────┘ └──────────┘ └─────────┘ └─────────────────┘   │
│                                                              │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐                       │
│  │ Asset   │ │ Nav      │ │  Pool   │                       │
│  │Manager  │ │ A*       │ │  Generic│                       │
│  │ Loaders │ │Region2D  │ │ Thread  │                       │
│  │RefCount │ │Agent2D   │ │  Safe   │                       │
│  └─────────┘ └──────────┘ └─────────┘                       │
└──────────────────────┬───────────────────────────────────────┘
                       │ gomobile build
                       ▼
               Android APK/AAB (ARM64)
```

## 🚀 Quickstart

### Pré-requisitos

- **Go 1.22+** (compiler + runtime)
- **Android SDK** (apenas para build APK)

### Editor Web (Rápido)

```bash
cd editor
python3 -m http.server 8080
# Acesse: http://localhost:8080
```

### Workflow Básico

```bash
# 1. Inicie o editor web
cd editor && python3 -m http.server 8080

# 2. Crie cenas visualmente no editor
# 3. Adicione KScript nos objetos
# 4. Salve: Ctrl+S (.kora.json)

# 5. Compile o KScript
go run cmd/kora-run/main.go game.ks

# 6. Build APK
./build.sh debug

# 7. Instale no dispositivo
adb install build/android/kora-debug.apk
```

## 📖 Documentação

| Documentação | Descrição |
|--------------|-----------|
| [KScript Guide](docs/SCRIPT.md) | Linguagem completa com exemplos |
| [API Reference](docs/API_REFERENCE.md) | Referências de todas APIs |
| [Editor Guide](docs/EDITOR_GUIDE.md) | Guia do editor visual |
| [Assets Guide](docs/ASSETS_GUIDE.md) | Importação e gerenciamento |
| [Architecture](docs/ARCHITECTURE.md) | Arquitetura do sistema |
| [Contributing](docs/CONTRIBUTING.md) | Como contribuir |

## 🎓 KScript - Linguagem

KScript é uma linguagem tipo TypeScript, compilada para Go. Zero runtime overhead.

### Exemplo

```kscript
object Player {
  speed: float = 180
  hp: int = 5

  async create() {
    await wait(0.5)
    this.hp = 10
  }

  update(dt: float) {
    const move = Input.axisX()
    this.x += move * this.speed * dt
    if (Input.pressed("Space") && this.onGround()) {
      this.y -= 300
    }
  }

  onHit(damage: int) {
    this.hp -= damage
    if (this.hp <= 0) {
      emit(this, "dead")
      destroyAsync(this)
    }
  }
}
```

## 🛠️ Build & Development

```bash
# Run editor web
make dev

# Build compiler
make compiler

# Run example
make run

# Build APK (debug)
make apk

# Testes
make test

# Benchmarks
go test -bench=. ./core/physics/... ./core/navigation/... ./core/scene/...
```

## 📊 Tecnologia

| Camada | Tecnologia |
|--------|------------|
| **Editor** | HTML5 Canvas + Vanilla JS + CodeMirror 6 + IndexedDB |
| **KScript** | Custom language → Go compiler (lexer, parser, checker, emitter) |
| **Runtime** | Go 1.22+ + Ebitengine v2.7 + gomobile |
| **Render** | Sprites, tilemaps, bitmap font, camera, shaders Kage, lighting, parallax |
| **Physics** | AABB, CCD, spatial hash, joints, raycast, CharacterBody2D |
| **Audio** | OGG/WAV/MP3, multi-bus mixer, spatial panning |
| **Navigation** | A* grid pathfinding, path smoothing, dynamic obstacles |
| **Animation** | Keyframe AnimationPlayer, cutscene sequencer, 28 easings |
| **Particles** | CPU-based, burst/continuous, gravity, blend modes |
| **UI** | Label, Button, Panel (built-in bitmap font) |
| **Skeleton** | Bone hierarchy, CCD IK solver, rest pose |
| **Asset Mgmt** | Ref-counted, sync/async, scene preload, 6 loaders |
| **Output** | Native ARM64 APK/AAB via gomobile |

## 📦 Estrutura do Projeto

```
kora/
├── editor/                     # Editor web (HTML/JS)
│   ├── index.html, editor.js, style.css
│   ├── code-panel.js           # CodeMirror 6 KScript editor
│   ├── git-panel.js            # Git version control
│   ├── assets-panel.js         # Asset management
│   ├── serializer.js           # JSON ↔ KScript
│   ├── preview-panel.js        # Preview runtime
│   └── idb.js                  # IndexedDB wrapper
│
├── compiler/                   # KScript → Go compiler
│   ├── lexer/parser/ast/checker/transform/emitter/
│   ├── compiler.go             # Public CompileSource/CompileFile
│   └── kscript.go              # Physics API registration
│
├── core/
│   ├── engine/                 # (deprecated, delegates to runner)
│   ├── runner/                 # Game loop, debug overlay, Config
│   ├── render/                 # Renderer, Camera, Sprite, Tilemap,
│   │                           # TextureCache, ShaderManager, BitmapFont
│   ├── scene/                  # Scene, Entity, SceneManager, SceneTree,
│   │                           # Loader, NodeEntity, PrefabManager
│   ├── node/                   # Node2D, Sprite2D, Camera2D, AnimationPlayer,
│   │                           # PhysicsBody2D, Area2D, AudioPlayer2D,
│   │                           # Particles2D, PointLight2D, Skeleton2D,
│   │                           # CutscenePlayer, ParallaxBackground,
│   │                           # DebugConsole, ShaderNode, Label/Button/Panel
│   ├── physics/                # PhysicsWorld, RigidBody, colliders, raycast,
│   │                           # SpatialHash, SweptAABB CCD, Joint
│   ├── input/                  # InputManager, actions, virtual pad
│   ├── audio/                  # Manager, Mixer multi-bus, spatial
│   ├── async/                  # Task, Scheduler, Tween (28 easings)
│   ├── math/                   # Vector2, Rect, easing functions (28)
│   ├── navigation/             # Pathfinder A*, NavigationRegion2D, Agent2D
│   ├── autoload/               # Singleton registry (thread-safe)
│   ├── asset/                  # AssetManager, ref counting, loaders
│   └── pool/                   # Generic thread-safe object pool
│
├── cmd/
│   ├── kora-run/               # KScript compiler CLI
│   └── kora-android/           # Android entry point
│
├── android/                    # APK build pipeline
│   ├── build.sh, setup.sh, AndroidManifest.xml
│
├── examples/hello/             # Minimal game example
│
└── docs/                       # Documentation
    ├── SCRIPT.md, API_REFERENCE.md, EDITOR_GUIDE.md
    ├── ASSETS_GUIDE.md, DESKTOP_APP.md
    ├── ARCHITECTURE.md, CONTRIBUTING.md
```

## 📋 Roadmap

### v0.1 — Runtime + Editor Base ✅
- [x] [#33](https://github.com/ElioNeto/kora/issues/33) Sistema de Nós (Node2D) ✅
- [x] [#34](https://github.com/ElioNeto/kora/issues/34) Sistema de Cenas ✅
- [x] [#35](https://github.com/ElioNeto/kora/issues/35) SceneTree ✅
- [x] [#10](https://github.com/ElioNeto/kora/issues/10) Câmera 2D ✅
- [x] [#22](https://github.com/ElioNeto/kora/issues/22) Física 2D ✅
- [x] [#37](https://github.com/ElioNeto/kora/issues/37) Corpos físicos (4 tipos) ✅
- [x] [#16](https://github.com/ElioNeto/kora/issues/16) Sistema de Sprites ✅
- [x] [#11](https://github.com/ElioNeto/kora/issues/11) Animação de sprites ✅
- [x] [#5](https://github.com/ElioNeto/kora/issues/5) Editor KScript com syntax highlight ✅
- [x] [#4](https://github.com/ElioNeto/kora/issues/4) CI/CD GitHub Actions ✅
- [ ] [#6](https://github.com/ElioNeto/kora/issues/6) Loja de templates

### v0.2 — Editor Features ✅
- [x] Sistema de Objetos + Eventos KScript ✅
- [x] Editor de Salas (grid, snap, layers) ✅
- [x] Auto-tile + Tile System ✅
- [x] AnimationPlayer + CutscenePlayer ✅
- [x] Mixer de Áudio + som espacial ✅
- [x] Debugger (FPS, entities, tasks, node tree) ✅
- [x] AutoLoad (Singletons) ✅
- [x] Git Panel (status, stage, commit) ✅

### v0.3 — Runtime Features ✅
- [x] [#25](https://github.com/ElioNeto/kora/issues/25) Exportação Android (APK/AAB) ✅
- [x] [#40](https://github.com/ElioNeto/kora/issues/40) Sistema de Partículas ✅
- [x] [#41](https://github.com/ElioNeto/kora/issues/41) Iluminação 2D Dinâmica ✅
- [x] [#42](https://github.com/ElioNeto/kora/issues/42) Pathfinding A* ✅
- [x] [#43](https://github.com/ElioNeto/kora/issues/43) Parallax ✅
- [x] [#39](https://github.com/ElioNeto/kora/issues/39) Skeleton2D com IK ✅
- [x] [#24](https://github.com/ElioNeto/kora/issues/24) Shaders (Kage) ✅
- [ ] [#31](https://github.com/ElioNeto/kora/issues/31) Networking / Multiplayer 🔲

### v1.0 — Polimento (Em andamento)
- [ ] UI System (Label, Button, Panel) ✅ *(implementado)*
- [ ] Prefabs ✅
- [ ] Asset Manager ✅
- [ ] Spatial Hash + CCD + Joints ✅
- [ ] Benchmark tests ✅
- [ ] 28 easing functions ✅
- [ ] Timeline animation editor 🔲
- [ ] Gamepad input 🔲
- [ ] Sprite batching 🔲
- [ ] Completo documentation 🔲
- [ ] Example games 🔲
- [ ] Multi-platform export (Windows, Linux, iOS) 🔲

## 🤝 Contribuindo

Veja [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) para como contribuir.

## 📄 License

MIT License — see [LICENSE](LICENSE)

---

**Kora Engine** — Crie jogos 2D para Android com editor visual e performance nativa.

[Documentação KScript](docs/SCRIPT.md) · [API Reference](docs/API_REFERENCE.md) · [Issues](https://github.com/ElioNeto/kora/issues)
