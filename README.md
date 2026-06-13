# Kora Engine

[![CI](https://github.com/ElioNeto/kora/actions/workflows/ci.yml/badge.svg)](https://github.com/ElioNeto/kora/actions/workflows/ci.yml)
[![Desktop Build](https://github.com/ElioNeto/kora/actions/workflows/desktop.yml/badge.svg)](https://github.com/ElioNeto/kora/actions/workflows/desktop.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ElioNeto/kora)](https://goreportcard.com/report/github.com/ElioNeto/kora)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> Motor de jogos 2D para desktop com runtime Go, linguagem KScript compilada e ferramentas CLI nativas.

---

## 🎮 O que é Kora?

**Kora Engine** é um motor de jogos 2D focado em desktop (Windows, macOS, Linux) que combina três pilares:

- **KScript** — Linguagem própria com sintaxe TypeScript, compilada diretamente para Go. Sem VM, sem runtime overhead, performance nativa.
- **Runtime Go** — Construído sobre Ebitengine, entregando renderização 2D acelerada, áudio multicanal, física AABB com CCD, pathfinding A\*, iluminação dinâmica, partículas, animação por keyframes, sistema de UI e muito mais.
- **Ferramentas CLI** — Compilador, executor e empacotador escritos em Go. Sem dependência de navegador, sem servidor web, sem JavaScript.

### Principais Características

- ✅ **KScript compilado para Go** — Performance nativa sem VM ou interpretador
- ✅ **Física 2D completa** — AABB, CCD, joints, raycast, CharacterBody2D, spatial hash
- ✅ **Renderização** — Sprites, spritesheets, tilemaps, frustum culling, câmera follow+zoom+shake
- ✅ **Áudio** — OGG/WAV/MP3, mixer multi-bus, som espacial 2D, pitch, pan
- ✅ **Animação** — Keyframe AnimationPlayer, cutscene sequencer, 28 funções de easing
- ✅ **Partículas** — CPU-based com emissão contínua/burst, gravidade, blend modes
- ✅ **Iluminação 2D** — PointLight2D, DirectionalLight2D, sombras dinâmicas
- ✅ **Pathfinding** — A\* com grid navigation, path smoothing, obstacles dinâmicos
- ✅ **Skeleton2D** — Hierarquia de ossos com CCD IK solver
- ✅ **Parallax** — Múltiplas camadas com scroll independente
- ✅ **UI** — Label, Button, Panel (bitmap font integrada)
- ✅ **Shader** — Kage (ESSL), ShaderManager, ShaderNode
- ✅ **Debugger** — Overlay com FPS, entidades, tasks, árvore de nós
- ✅ **Prefabs** — Templates reutilizáveis com deep copy
- ✅ **Object Pool** — Pool genérico thread-safe para redução de GC
- ✅ **Asset Manager** — Carregamento síncrono/assíncrono com ref counting
- ✅ **Bridge Editor↔Runtime** — `core/editor/` converte cenas do editor para Node2D tree
- ✅ **Animation Timeline** — Playback de keyframes no editor com easing visual
- ✅ **Hot-Reload** — Watcher de arquivos `.ks` com recompilação automática
- ✅ **Camera Gizmo** — Visualização de frustum, crosshair e handles no viewport
- ✅ **CLI tools** — `kora-run`, `kora-build`, `kora-android`
- ✅ **Exportação desktop** — Binário nativo para Windows, macOS e Linux
- ✅ **Exportação mobile** — APK/AAB Android via gomobile
- ✅ **Benchmarks** — 30+ benchmarks nos sistemas críticos

---

## 🏗️ Arquitetura

```
┌──────────────────────────────────────────┐
│     Ferramentas CLI (Go)                 │
│  kora-run · kora-build · kora-android    │
└────────────┬─────────────────────────────┘
             │ .ks → AST
             ▼
┌──────────────────────────────────────────┐
│     Compilador KScript (Go)              │
│  Lexer → Parser → Checker → Emitter      │
│  (Gera structs Go + interface Entity)    │
└────────────┬─────────────────────────────┘
             │ Código Go gerado
             ▼
┌──────────────────────────────────────────┐
│     Editor Kora (Go + Ebitengine)        │
│                                          │
│  ┌────────────┐  ┌──────────────────┐    │
│  │ SceneFile  │  │  Timeline Anim   │    │
│  │ .kora.json │  │  playhead/tracks │    │
│  └─────┬──────┘  └──────────────────┘    │
│        │                                 │
│  ┌─────▼──────────────────────────┐      │
│  │     Bridge (core/editor/)      │      │
│  │  SceneEntity ←→ Node2D tree    │      │
│  │  Instantiate / Preview / Play  │      │
│  └─────┬──────────────────────────┘      │
│        │                                 │
│        ▼                                 │
│  ┌──────────────────────────────────┐    │
│  │  Hot-Reload .ks                 │    │
│  │  Camera Gizmo / Gizmos          │    │
│  └──────────────────────────────────┘    │
└────────────┬─────────────────────────────┘
             │ Preview / Play
             ▼
┌──────────────────────────────────────────┐
│     Runtime Kora (Go + Ebitengine)       │
│                                          │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐   │
│  │ Render  │ │ Physics  │ │  Audio  │   │
│  │ Sprites │ │ AABB/CCD │ │ Mixer   │   │
│  │ Tilemap │ │ Joints   │ │ Spatial │   │
│  │ Camera  │ │ Raycast  │ │ OGG/WAV │   │
│  │ Shaders │ │ CharBody │ │ MP3     │   │
│  │ Culling │ │ Spatial  │ │         │   │
│  └─────────┘ └──────────┘ └─────────┘   │
│                                          │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐   │
│  │ Scene   │ │  Node2D  │ │  Async  │   │
│  │ Manager │ │ Sprite2D │ │  Tasks  │   │
│  │ Loader  │ │ Camera2D │ │  Tween  │   │
│  │ Prefab  │ │PhysicsBd │ │  Sched  │   │
│  │ AutoLoad│ │ AudioPl2 │ │  Pool   │   │
│  │ Tree    │ │Animation │ │  Easing │   │
│  │ Entity  │ │ UI       │ │         │   │
│  └─────────┘ └──────────┘ └─────────┘   │
│                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ Particles│ │ Light2D  │ │Skeleton2D│ │
│  │ Parallax │ │ Cutscene │ │ Nav A*   │ │
│  │ Shader   │ │ Debug    │ │ Pool     │ │
│  └──────────┘ └──────────┘ └──────────┘ │
└────────────┬─────────────────────────────┘
             │ go build / gomobile
             ▼
┌──────────────────────────────────────────┐
│   Binário Desktop (Win/Mac/Linux)        │
│   APK/AAB Android (via gomobile)         │
└──────────────────────────────────────────┘
```

---

## 🚀 Quickstart

### Pré-requisitos

- **Go 1.22+**
- **Ebitengine** — gerenciado automaticamente via `go.mod`

### Executar um jogo

```bash
# Compilar e executar um arquivo KScript
go run cmd/kora-run/main.go examples/hello/game.ks

# Ou usando o atalho do Makefile
make run
```

### Abrir o Editor Visual

```bash
# Editor Go nativo (Ebitengine) com viewport, hierarquia, inspetor
go run cmd/kora-editor/main.go

# Carregar uma cena existente
go run cmd/kora-editor/main.go scenes/mygame.kora.json
```

```bash
# Atalhos do editor:
# F1-F3: Tabs (Scene/Assets/Script)
# F4: Timeline de animação
# F5/F6: Painéis (Hierarquia/Inspetor)
# F7: Hot-Reload KScript
# F8: Play mode (preview)
# F9: Camera gizmo
# 1-3: Ferramentas (Select/Move/Scale)
```

### Compilar um binário desktop

```bash
# Gerar binário nativo para a plataforma atual
go build -o kora-game cmd/kora-run/main.go

# Para Windows (cross-compile)
GOOS=windows GOARCH=amd64 go build -o kora-game.exe cmd/kora-run/main.go
```

### Pipeline completa de um projeto

```bash
# 1. Crie seu jogo em KScript (.ks)
# 2. Compile e execute
kora-run game.ks

# 3. Exporte para desktop
kora-build --target desktop game.ks

# 4. Exporte para Android (opcional)
kora-android build game.ks
```

### Testes

```bash
# Executar todos os testes
make test

# Benchmarks nos sistemas críticos
go test -bench=. ./core/physics/... ./core/navigation/... ./core/scene/...
```

---

## 📖 Documentação

| Documentação | Descrição |
|---|---|
| [Guia KScript](docs/SCRIPT.md) | Linguagem completa com exemplos |
| [Referência de API](docs/API_REFERENCE.md) | Todas as APIs do runtime |
| [Guia de Assets](docs/ASSETS_GUIDE.md) | Importação e gerenciamento |
| [Arquitetura](docs/ARCHITECTURE.md) | Visão detalhada do sistema |
| [Guia Desktop](docs/DESKTOP_APP.md) | Exportação para desktop |
| [Contribuição](docs/CONTRIBUTING.md) | Como contribuir |

---

## 🎓 KScript — A Linguagem

KScript é uma linguagem com sintaxe similar a TypeScript, compilada estaticamente para Go. Zero overhead em runtime — seu código vira structs e funções Go compiladas nativamente.

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

### Diferenciais do KScript

- **Tipagem estática** com inferência
- **Async/await** nativo — sem goroutines expostas
- **Signals** — sistema de eventos desacoplado
- **Compilação direta para Go** — sem VM, sem runtime extra
- **Acesso completo à runtime** — física, áudio, input, nós

---

## 🛠️ Build & Development

```bash
# Compilar o compilador KScript
make compiler

# Executar um jogo de exemplo
make run

# Exportar para desktop
make build

# Exportar Android APK
make apk

# Executar testes
make test

# Benchmarks
go test -bench=. ./core/...
```

---

## 📊 Tecnologia

| Camada | Tecnologia |
|---|---|
| **Linguagem** | KScript (custom) → compilador Go |
| **Runtime** | Go 1.22+ + Ebitengine v2.7 |
| **CLI** | Go puro (cobra opcional para subcomandos) |
| **Render** | Sprites, tilemaps, bitmap font, câmera, shaders Kage, iluminação, parallax |
| **Física** | AABB, CCD, spatial hash, joints, raycast, CharacterBody2D |
| **Áudio** | OGG/WAV/MP3, mixer multi-bus, som espacial 2D |
| **Navegação** | A\* grid pathfinding, path smoothing, obstacles dinâmicos |
| **Animação** | Keyframe AnimationPlayer, cutscene sequencer, 28 easings |
| **Partículas** | CPU-based, burst/contínuo, gravidade, blend modes |
| **UI** | Label, Button, Panel (bitmap font interna) |
| **Skeleton** | Hierarquia de ossos, CCD IK solver, rest pose |
| **Asset Mgmt** | Ref-counted, sync/async, scene preload, 6 loaders |
| **Empacotamento** | `go build` para desktop, gomobile para Android |

---

## 📦 Estrutura do Projeto

```
kora/
├── cmd/
│   ├── kora-run/             # Compilador/executor CLI de KScript
│   └── kora-android/         # Entry point para Android (gomobile)
│
├── compiler/                 # Compilador KScript → Go
│   ├── lexer/parser/ast/checker/transform/emitter/
│   ├── compiler.go           # API pública CompileSource / CompileFile
│   └── kscript.go            # Registro de APIs da runtime
│
├── core/                     # Runtime do motor
│   ├── editor/               # Bridge SceneFile↔Node2D, anim timeline, hot-reload
│   ├── runner/               # Game loop, debug overlay, Config
│   ├── render/               # Renderer, Camera, Sprite, Tilemap,
│   │                         # TextureCache, ShaderManager, BitmapFont
│   ├── scene/                # Scene, Entity, SceneManager, SceneTree,
│   │                         # Loader, NodeEntity, PrefabManager
│   ├── node/                 # Node2D, Sprite2D, Camera2D, AnimationPlayer,
│   │                         # PhysicsBody2D, Area2D, AudioPlayer2D,
│   │                         # Particles2D, PointLight2D, Skeleton2D,
│   │                         # CutscenePlayer, ParallaxBackground,
│   │                         # DebugConsole, ShaderNode, UI (Label/Button/Panel)
│   ├── physics/              # PhysicsWorld, RigidBody, colliders, raycast,
│   │                         # SpatialHash, SweptAABB CCD, Joint
│   ├── input/                # InputManager, actions, virtual pad, gamepad
│   ├── audio/                # Manager, Mixer multi-bus, espacial
│   ├── async/                # Task, Scheduler, Tween (28 easings)
│   ├── math/                 # Vector2, Rect, funções de easing (28)
│   ├── navigation/           # Pathfinder A*, NavigationRegion2D, Agent2D
│   ├── autoload/             # Registro de singletons (thread-safe)
│   ├── asset/                # AssetManager, ref counting, loaders
│   └── pool/                 # Pool genérico thread-safe
│
├── editor/                   # Editor visual legado (HTML/JS)
│   └── ...                   # Será substituído por editor Go futuramente
│
├── android/                  # Pipeline de build Android
│   ├── build.sh, setup.sh, AndroidManifest.xml
│
├── examples/                 # Jogos e demos em KScript
│   └── hello/                # Exemplo mínimo
│
├── docs/                     # Documentação
│   ├── SCRIPT.md, API_REFERENCE.md
│   ├── ASSETS_GUIDE.md, DESKTOP_APP.md
│   └── ARCHITECTURE.md, CONTRIBUTING.md
│
├── Makefile                  # Comandos de build e desenvolvimento
└── main.go                   # Entry point da runtime (desktop)
```

---

## 🗺️ Roadmap

Visão geral do planejamento. Detalhes completos em [ROADMAP.md](ROADMAP.md).

| Versão | Foco | Status |
|---|---|---|
| **v1.0** | Runtime Desktop | 🚧 Em andamento |
| **v2.0** | Editor Go + Ecossistema | 🔲 Planejado |
| **v3.0** | Cloud & Multiplayer | 🔲 Futuro |

---

## 🧭 Direção do Projeto

Kora começou como um experimento de engine Android com editor web. A partir de 2025, o projeto **migrou para desktop-first** com as seguintes mudanças estratégicas:

- **Runtime Go como centro** — não mais Android-first, mas multiplataforma desktop com exportação opcional para Android
- **Editor nativo Go** — substituição gradual do editor HTML/JS por um editor desktop escrito em Go (Ebitengine + IMGUI)
- **Ferramentas CLI** — todo o fluxo de desenvolvimento centrado em terminal, sem dependência de navegador
- **KScript como linguagem primária** — compilador maduro com ecossistema próprio

> O editor web legado (`editor/`) permanece no repositório para referência, mas não receberá novas funcionalidades. O foco ativo é no runtime Go e nas ferramentas CLI.

---

## 🤝 Contribuindo

Veja [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) para guia de contribuição, padrões de código e processo de PR.

---

## 📄 Licença

MIT License — veja [LICENSE](LICENSE)

---

**Kora Engine** — Crie jogos 2D para desktop com Go e KScript.
Performance nativa. Zero overhead. CLI-first.

[Documentação KScript](docs/SCRIPT.md) · [Referência de API](docs/API_REFERENCE.md) · [Issues](https://github.com/ElioNeto/kora/issues)
