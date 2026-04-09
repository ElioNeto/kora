# Arquitetura do Kora Engine

## Visão Geral

Kora Engine é uma engine de jogos 2D para Android, composta por três partes principais:

1. **Kora Editor** - Editor visual desktop/web para criar cenas
2. **KScript Compiler** - Compilador que transforma KScript em Go
3. **Kora Runtime** - Runtime em Go que executa o jogo no Android

```
┌──────────────────────────────────────────────────────────────┐
│                    Kora Editor Desktop                        │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Scene Editor (Canvas 2D)                              │  │
│  │  ├─ Entity System                                      │  │
│  │  ├─ Asset Manager                                      │  │
│  │  └─ Inspector                                            │  │
│  └────────────────────────────────────────────────────────┘  │
│                         │                                      │
│                         │ exporta                              │
│                         ▼                                      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │        Scene Data (.kora.json)                          │  │
│  │  { entities: [...], meta: {...} }                       │  │
│  └────────────────────────────────────────────────────────┘  │
│                         │                                      │
│                         │ compila                              │
│                         ▼                                      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │     KScript Compiler (Go)                               │  │
│  │  ├─ Lexer (Tokenização)                                │  │
│  │  ├─ Parser (AST)                                       │  │
│  │  ├─ Type Checker                                       │  │
│  │  └─ Go Emitter                                         │  │
│  └────────────────────────────────────────────────────────┘  │
│                         │                                      │
│                         │ gera                                 │
│                         ▼                                      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │        Generated Go Code (.go)                          │  │
│  │  import "kora/runtime"                                  │  │
│  │  func (e *Entity) Update(dt float64) { ... }           │  │
│  └────────────────────────────────────────────────────────┘  │
│                         │                                      │
│                         │ compila via gomobile                 │
│                         ▼                                      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │     Kora Runtime (Go + gomobile)                        │  │
│  │  ├─ ECS (Entity Component System)                      │  │
│  │  ├─ 2D Renderer (sprites, tilemaps, camera)            │  │
│  │  ├─ Physics (AABB, collision detection)                │  │
│  │  ├─ Input System (keyboard, touch, gamepad)            │  │
│  │  ├─ Audio System (OGG, WAV)                            │  │
│  │  ├─ Asset Loader                                       │  │
│  │  └─ Async Scheduler (coroutines)                       │  │
│  └────────────────────────────────────────────────────────┘  │
│                         │                                      │
│                         │ empacota                             │
│                         ▼                                      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Android APK / AAB                          │  │
│  │        Native ARM64 Binary                              │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## Componentes

### 1. Editor

**Tecnologias**: Electron + HTML5 Canvas + JavaScript puro

**Localização**: `editor/` e `apps/desktop/`

**Arquitetura**:
```
editor/
├── editor.js          # Core do editor (state, render, entities)
├── assets-panel.js    # Importação e gerenciamento de assets
├── serializer.js      # Serialização JSON ↔ KScript
├── preview-panel.js   # Preview do jogo
├── idb.js             # IndexedDB wrapper
└── style.css          # Estilos

apps/desktop/
├── src/main/          # Electron main process
├── src/renderer/      # Electron renderer
└── package.json       # Dependencies Electron
```

**State Management**:
```javascript
const state = {
  entities: [],       // Lista de entidades
  selected: null,     // Entidade selecionada
  cam: { x, y, zoom }, // Câmera
  meta: { name, ... }, // Meta da cena
  dirty: false        // Flag de modificação
}
```

**Sistema de Entidades** (componente simples):
```javascript
{
  id: number,
  name: string,
  type: 'sprite' | 'tilemap' | 'camera' | 'audio' | 'custom',
  x, y: number,
  w, h: number,
  rotation: number,
  visible: boolean,
  locked: boolean,
  color: string,
  assetUrl?: string, // Para sprites
  script?: string    // KScript
}
```

### 2. Compiler KScript

**Tecnologias**: Go

**Localização**: `compiler/`

**Pipeline de Compilação**:
```
KScript Source (.ks)
    │
    ▼
[Lexer] → Tokens
    │
    ▼
[Parser] → AST
    │
    ▼
[Type Checker] → Validated AST
    │
    ▼
[Go Emitter] → Go Source (.go)
    │
    ▼
[Go Compiler] → Native Binary
```

**Exemplo de Transformação**:
```kscript
object Player {
  speed: float = 180
  update(dt: float) {
    this.x += this.speed * dt
  }
}
```

↓ Compilado para Go:

```go
type Player struct {
    Speed float64 = 180.0
    PosX  float64
}

func (e *Entity) Update(dt float64) {
    if p, ok := e.(*Player); ok {
        p.PosX += p.Speed * dt
    }
}
```

### 3. Runtime (Go)

**Tecnologias**: Go 1.22+, gomobile

**Localização**: `core/`, `cmd/`, `main.go`

**Módulos Core**:

#### ECS (Entity Component System)
```go
type Entity struct {
    ID       uuid.UUID
    Position Vector2
    Scale    Vector2
    Rotation float64
    Components map[string]Component
}

type Component interface {
    Type() string
    Update(dt float64)
}
```

#### 2D Renderer
```go
type Renderer interface {
    Clear()
    DrawSprite(sprite *Sprite, pos Vector2, rotation float64)
    DrawTilemap(tm *Tilemap, camera Camera)
    DrawRect(pos Vector2, size Vector2, color Color)
    Present()
}
```

#### Physics (AABB)
```go
type PhysicsSystem struct {
    gravity Vector2
    bodies  []*RigidBody
}

func (ps *PhysicsSystem) Update(dt float64) {
    for _, body := range ps.bodies {
        body.velocity += ps.gravity * dt
        body.position += body.velocity * dt
    }
    ps.checkCollisions()
}
```

#### Input System
```go
type InputState struct {
    Keys     map[string]bool
    AxisX    float64
    AxisY    float64
    Touches  []TouchEvent
}

func (in *InputState) Pressed(key string) bool
func (in *InputState) AxisX() float64
func (in *InputState) Touches() []TouchEvent
```

#### Async Scheduler
```go
type Task struct {
    State    TaskState
    ResumeAt float64
    Result   interface{}
}

type Scheduler struct {
    tasks     []*Task
    startTime time.Time
}

func (s *Scheduler) Spawn(fn func() interface{}) *Task
func (s *Scheduler) Wait(task *Task) interface{}
func (s *Scheduler) Cancel(task *Task)
```

### 4. Build Pipeline

#### APK Build Process:
```bash
1. KScript → Go Code
   kora-compiler game.ks > generated.go

2. Go Code + Runtime → Android
   gomobile bind -target=android -androidapi 21 -o kora.aar .

3. Bundle → APK
   ./createApk.sh kora.aar game.apk
```

**arquivo android/build.sh**:
```bash
#!/bin/bash
# Build pipeline para APK

# 1. Compilar KScript
go run cmd/build.go input.ks > generated.go

# 2. Bind como biblioteca Android
gomobile bind -target=android -androidapi 21 \
    -o kora.aar .

# 3. Criar APK com Android Studio/gradle
cd android/app
./gradlew assembleRelease

# 4. Resultado
# output: app-release.apk
```

## Fluxo de Dados

### Desenvolvimento
```
Editor → Exporta → Scene JSON → Editor valida
         ↓
    Preview (JS) → Teste rápido
         ↓
    Export KScript → Compila → Runtime Go
         ↓
    APK → Android
```

### Runtime
```
Android Start
    ↓
Initialize Runtime
    ↓
Load Scene JSON
    ↓
Create Entities
    ↓
Loop Principal:
  ├─ Update(dt)
  ├─ PhysicsStep()
  ├─ Render()
  └─ Present()
```

## Padrões de Código

### Editor (JavaScript)
- Vanilla JS (sem framework)
- Pattern: State → Render → Update
- Modular via módulos ES

### Runtime (Go)
- Interfaces para componentes
- ECS para entidades
- Async/await via corrotinas

### KScript
- Tipagem estática
- Classes/Objects para lógica de jogo
- Sinais para eventos

## Dependências

```
Kora Editor
├── Electron (Desktop wrapper)
├── HTML5 Canvas (Render)
└── IndexedDB (Storage)

KScript Compiler
└── Go stdlib

Kora Runtime
├── gomobile (Android bind)
├── Ebitengine (opcional render)
└── Android SDK
```

## Performance

### Objetivos
- **Editor**: 60 FPS preview
- **Runtime**: 60 FPS nativo
- **Build APK**: < 5 minutos
- **Memory**: < 50MB base

### Otimizações
- Batch rendering no editor
- Sprite atlas no runtime
- Async loading
- Object pooling
- Compiled KScript = nativo

## Segurança

### Sandboxing
- Editor roda com acesso limitado
- Runtime tem permissões mínimas
- KScript é compilado, não interpretado

### Validação
- Input validation no editor
- Type checking no compiler
- Bounds checking no runtime

## Roadmap de Arquitetura

### v0.3
- [ ] Multi-threaded physics
- [ ] Asset streaming
- [ ] Plugin system

### v0.4
- [ ] VR/AR preview
- [ ] Network multiplayer
- [ ] Shader editor

### v1.0
- [ ] Full IDE (debugger, profiler)
- [ ] Marketplace
- [ ] Cross-platform export (iOS, Desktop)

---

**Documento de Arquitetura - Kora Engine**
