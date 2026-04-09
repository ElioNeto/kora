# Arquitetura do Kora Engine - v0.3

## Visão Geral

Kora Engine é uma engine de jogos 2D para Android, composta por:

1. **Kora Editor Desktop** - Aplicativo Electron para criação de cenas
2. **Editor Web** - Versão web para desenvolvimento rápido (fallback)
3. **KScript Compiler** - Compilador que transforma KScript em Go
4. **Kora Runtime** - Runtime em Go que executa o jogo no Android

```
┌──────────────────────────────────────────────────────────────────┐
│              Kora Editor Desktop (Electron)                       │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Editor Visual                                              │  │
│  │  ├─ Scene Editor (Canvas 2D)                               │  │
│  │  ├─ Asset Management (IndexedDB)                           │  │
│  │  ├─ Entity Inspector                                        │  │
│  │  ├─ Hierarchy Tree                                          │  │
│  │  └─ Console/Debug                                           │  │
│  └────────────────────────────────────────────────────────────┘  │
│                         │                                           │
│                         │ exporta cena                              │
│                         ▼                                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │        Scene Data (.kora.json)                              │  │
│  │  {                                                        │  │
│  │    entities: [...],                                        │  │
│  │    meta: { name, version, logicalW, logicalH },           │  │
│  │    assets: [...]                                           │  │
│  │  }                                                          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                         │                                           │
│                         │ compila                                   │
│                         ▼                                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │           KScript Compiler (Go)                             │  │
│  │  ┌──────────────────────────────────────────────────────┐  │  │
│  │  │  Lexer → Tokens                                       │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Parser → AST                                         │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Type Checker                                         │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Go Emitter                                           │  │  │
│  │  └──────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────┘  │
│                         │                                           │
│                         │ gera código Go                            │
│                         ▼                                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │        Generated Go Code (.go)                              │  │
│  │  import "kora/runtime"                                      │  │
│  │  func (e *Entity) Update(dt float64) { ... }                │  │
│  └────────────────────────────────────────────────────────────┘  │
│                         │                                           │
│                         │ compila via gomobile                      │
│                         ▼                                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │         Kora Runtime (Go + gomobile)                        │  │
│  │  ┌──────────────────────────────────────────────────────┐  │  │
│  │  │  ECS (Entity Component System)                        │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  2D Renderer (sprites, tilemaps, camera, particles)   │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Physics (AABB, forces, collision)                    │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Input (keyboard, touch, gamepad)                     │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Audio (OGG, WAV, MP3)                                │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Asset Loader                                         │  │  │
│  │  ├──────────────────────────────────────────────────────┤  │  │
│  │  │  Async Scheduler (coroutines, await/await)            │  │  │
│  │  └──────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────┘  │
│                         │                                           │
│                         │ empacota para Android                     │
│                         ▼                                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │           Android APK / AAB (Native ARM64)                  │  │
│  │        Zero VM Overhead                                     │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

## Componentes Principais

### 1. Kora Editor Desktop

**Tecnologias**: Electron 28 + Vite + HTML5 Canvas + JavaScript puro

**Localização**: `apps/desktop/` e `editor/`

#### Arquitetura Electron

```
apps/desktop/
├── src/
│   ├── main/                    # Main Process (Node.js)
│   │   ├── index.js             # Window, menu, IPC handlers
│   │   └── preload.js           # Secure API bridge (contextBridge)
│   │
│   └── renderer/                # Renderer Process (Chrome Blink)
│       ├── index.html           # Editor UI container
│       ├── assets/
│       │   └── style.css        # Editor styles
│       └── vite.config.js       # Vite build config
│
├── package.json                 # Dependencies & build scripts
└── vite.config.js               # Vite configuration
```

#### Flow de Processos

```javascript
// Main Process (index.js)
createWindow() {
  window = new BrowserWindow({
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  })
  loadFile('renderer/index.html')
}

createMenu() {
  Menu.buildFromTemplate([
    { label: 'File', submenu: [...] },
    { label: 'Edit', submenu: [...] },
    { label: 'View', submenu: [...] },
    { label: 'Help', submenu: [...] }
  ])
}

ipcMain.handle('select-file', async () => {
  const result = await dialog.showOpenDialog(window, options)
  return result.canceled ? null : result.filePaths[0]
})
```

#### Preload Script (`preload.js`)

```javascript
const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('electronAPI', {
  // File system
  selectFile: (opts) => ipcRenderer.invoke('select-file', opts),
  readFile: (path) => ipcRenderer.invoke('read-file', path),
  writeFile: (path, data) => ipcRenderer.invoke('write-file', path, data),

  // Special handlers
  saveScene: (data, name) => ipcRenderer.invoke('save-scene', { data, name }),
  buildAPK: (config) => ipcRenderer.invoke('build-apk', config),

  // Window controls
  minimize: () => ipcRenderer.invoke('minimize'),
  maximize: () => ipcRenderer.invoke('maximize'),
  unmaximize: () => ipcRenderer.invoke('unmaximize'),

  // Events
  on: (channel, fn) => ipcRenderer.on(channel, (e, ...args) => fn(...args))
})
```

#### Editor Core (Web)

O editor web (`editor/`) funciona como uma camada de UI:

```
editor/
├── editor.js          # Core logic: state, entities, render
├── assets-panel.js    # Asset management + IndexedDB
├── serializer.js      # JSON ↔ KScript serialization
├── idb.js             # IndexedDB wrapper for assets
├── preview-panel.js   # Preview with physics simulation
├── style.css          # All editor styles
└── index.html         # Standalone editor fallback
```

**State Management**:
```javascript
const state = {
  entities: [],       // Array<Entity>
  selected: null,     // Entity id | null
  tool: 'select',     // 'select' | 'move' | 'scale'
  cam: {              // Camera
    x: 0,
    y: 0,
    zoom: 1.0
  },
  drag: null,         // Drag state
  idSeq: 1,           // Next entity id
  meta: {             // Scene metadata
    name: 'Untitled',
    version: 1,
    logicalW: 360,
    logicalH: 640
  },
  dirty: false        // Unsaved changes flag
}
```

**Entity Structure**:
```javascript
{
  id: number,           // Unique ID
  name: string,         // Display name
  type: string,         // 'sprite' | 'tilemap' | 'camera' | 'audio' | 'custom'
  x: number,            // World position X
  y: number,            // World position Y
  w: number,            // Width
  h: number,            // Height
  rotation: number,     // Rotation in degrees
  visible: boolean,     // Visibility flag
  locked: boolean,      // Locked position
  color: string,        // Tint color (hex)
  script: string,       // KScript code
  assetId?: string,     // Reference to asset
  assetUrl?: string     // Direct URL (for sprites)
}
```

### 2. Editor Web

Versão fallback para desenvolvimento rápido. Funciona em qualquer navegador moderno sem Electron.

**Limitações**:
- Não tem acesso direto ao sistema de arquivos (usa File API)
- Sem menus nativos
- Sem build APK integrado

**Vantagens**:
- Desenvolvimento mais rápido
- Não requer Node.js instalado
- Pode abrir diretamente no navegador

### 3. Editor de Assets

**Sistema**: IndexedDB wrapper

```javascript
// idb.js
class AssetDB {
  static async init() { ... }

  static async getAll() { ... }  // Read all assets

  static async add(asset) { ... } // Add single asset

  static async addAll(assets) { ... } // Bulk add

  static async delete(id) { ... }  // Remove by ID

  static async clear() { ... }     // Clear all

  static async getSize() { ... }   // Total storage used
}
```

**Asset Structure**:
```javascript
{
  id: string,           // Unique ID (asset_timestamp_rand)
  name: string,         // Original filename
  type: string,         // 'image' | 'audio' | 'tilemap' | 'script'
  url: string,          // Blob URL (recreated on load)
  size: number,         // File size in bytes
  ext: string,          // Extension (png, jpg, etc)
  blob: Blob,           // Binary data for storage
  contentType: string   // MIME type (image/png, audio/ogg)
}
```

**Operations**:
```javascript
// Import
async _importFiles(fileList) {
  for (const file of fileList) {
    const data = await file.arrayBuffer()
    const blob = new Blob([data], { type: file.type })
    const url = URL.createObjectURL(blob)
    const asset = { id, name, type, url, size, blob, ... }
    assets.set(id, asset)
    await AssetDB.add(asset)
  }
}

// Load from DB
async _loadFromDB() {
  const dbAssets = await AssetDB.getAll()
  for (const a of dbAssets) {
    a.url = URL.createObjectURL(a.blob)  // Recreate Blob URL
    assets.set(a.id, a)
  }
}
```

### 4. KScript Compiler

**Tecnologias**: Go 1.22+

**Localização**: `compiler/`

#### Pipeline de Compilação

```
┌─────────────┐
│  .ks File   │  KScript source code
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Lexer     │  Tokenize → [Token, Token, ...]
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Parser    │  Token stream → AST (Abstract Syntax Tree)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Type Checker│  Validate types, check semantics
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Go Emitter  │  AST → go source code (.go)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  go build   │  Compile to native binary (with gomobile)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  APK Output │  Native Android application
└─────────────┘
```

#### AST Node Types

```go
type ASTExpr interface{ exprNode() }

// Expressions
type Identifier struct { Name string }
type NumberLit struct { Value float64 }
type StringLit struct { Value string }
type BinaryExpr struct { Left, Right Expr; Op string }
type CallExpr struct { Func string; Args []Expr }
type AwaitExpr struct { Expr Expr }

// Statements
type BlockStmt struct { Stmt []Statement }
type ExprStmt struct { Expr Expr }
type IfStmt struct { Cond Expr; Then, Else *BlockStmt }
type ForStmt struct { Init Stmt; Cond Expr; Post Stmt; Body *BlockStmt }
type ReturnStmt struct { Expr Expr }
type AsyncFuncDecl struct { Name string; Params []Param; Body *BlockStmt }

// Declarations
type VariableDecl struct { Name string; Type Type; Value Expr }
type TypeDecl struct { Name string; Fields []Field }
```

#### Exemplo de Transformação

**KScript Input**:
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
  }
}
```

**Generated Go Output**:
```go
type Player struct {
    Entity
    Speed float64 `json:"speed"`
    HP    int     `json:"hp"`
}

func (p *Player) Create() async.Task {
    defer func() { p.HP = 10 }()
    await async.Wait(TimeSpan.FromSeconds(0.5))
    return async.Success()
}

func (p *Player) Update(dt float64) {
    move := Input.AxisX()
    p.X += move * p.Speed * dt
}
```

### 5. Kora Runtime

**Tecnologias**: Go 1.22+, gomobile

**Localização**: `core/`, `cmd/`, `main.go`

#### Módulos Core

##### ECS (Entity Component System)

```go
type Entity struct {
    ID       uuid.UUID
    Name     string
    Position Vector2
    Scale    Vector2
    Rotation float64
    Components map[string]Component
    Visible    bool
    Tags       []string
}

type Component interface {
    Type() string
    Update(dt float64)
}

type EntityManager struct {
    entities map[uuid.UUID]*Entity
    pool     *ObjectPool
}
```

##### 2D Renderer

```go
type Renderer interface {
    Clear(color Color)
    BeginScene(camera Camera)
    EndScene()
    Present()

    // Sprites
    DrawSprite(sprite *Sprite, pos Vector2, rotation float64, color Color, alpha float32)
    DrawSpriteQuad(sprite *Sprite, quad SpriteQuad, color Color, alpha float32)

    // Tilemaps
    DrawTilemap(tm *Tilemap, camera Camera, layer string)

    // Primitives
    DrawLine(p1, p2 Vector2, color Color, width float32)
    DrawRect(pos, size Vector2, color Color, rotation float64)
    DrawCircle(pos Vector2, radius float32, color Color)

    // Text
    DrawText(font *Font, text string, pos Vector2, color Color, size float32)

    // Batch optimization
    BatchSprites()
    BatchTilemaps()
}
```

##### Physics (AABB)

```go
type RigidBody struct {
    Position   Vector2
    Velocity   Vector2
    Acceleration Vector2
    Mass       float64
    Friction   float64
    Restitution float64
    CollisionMask uint32
}

type PhysicsSystem struct {
    gravity    Vector2
    bodies     []*RigidBody
    collisions []CollisionEvent
}

func (ps *PhysicsSystem) Update(dt float64) {
    // Apply forces
    for _, body := range ps.bodies {
        body.Velocity += (body.Acceleration + ps.gravity) * dt
        body.Position += body.Velocity * dt
    }

    // Collision detection
    ps.checkCollisions()
}

func (ps *PhysicsSystem) checkCollisions() {
    // Broad phase: spatial partitioning
    // Narrow phase: AABB overlap test
    // Resolve: impulse-based resolution
}
```

##### Input System

```go
type InputSystem struct {
    keyboard   map[Key]bool
    mouse      MouseState
    gamepad    []GamepadState
    touch      []TouchState
    axes       map[string]float64 // Dual stick mapping
}

type MouseState struct {
    Position Vector2
    Delta    Vector2
    Buttons  [MouseButtonCount]bool
}

func (in *InputSystem) Pressed(key Key) bool
func (in *InputSystem) Down(key Key) bool
func (in *InputSystem) Released(key Key) bool
func (in *InputSystem) Axis(key string) float64
func (in *InputSystem) MousePos() Vector2
func (in *InputSystem) MousePressed(btn MouseButton) bool
```

##### Async Scheduler

```go
type TaskState int

const (
    TaskPending TaskState = iota
    TaskRunning
    TaskWaiting
    TaskComplete
    TaskCancelled
)

type Task struct {
    ID       uuid.UUID
    State    TaskState
    Generator func() async.Generator
    Result   interface{}
    Error    error
    ResumeAt float64 // For wait() operations
}

type Scheduler struct {
    tasks     map[uuid.UUID]*Task
    time      time.Time
    deltaTime time.Duration
}

func (s *Scheduler) Spawn(fn func() async.Generator) *Task
func (s *Scheduler) Wait(task *Task, duration float64) TaskState
func (s *Scheduler) Cancel(task *Task) bool
func (s *Scheduler) Update(dt float64)
```

##### Asset Loader

```go
type AssetType int

const (
    AssetImage AssetType = iota
    AssetAudio
    AssetTilemap
    AssetFont
    AssetScript
)

type AssetKey struct {
    Type AssetType
    Name string
}

type AssetManager struct {
    cache    map[AssetKey]Asset
    loading  map[AssetKey]*AssetLoadTask
    config   AssetConfig
}

func (am *AssetManager) Load(key AssetKey, path string) error
func (am *AssetManager) Get(key AssetKey) (Asset, bool)
func (am *AssetManager) IsLoaded(key AssetKey) bool
func (am *AssetManager) Preload(assets []AssetKey) error
```

##### Audio System

```go
type AudioSystem struct {
    channels   map[int]*AudioChannel
    masterGain float32
    bgmPool    *ObjectPool
}

type AudioChannel struct {
    ID        int
    Buffer    *SoundBuffer
    Position  float32
    Looping   bool
    Volume    float32
    Pitch     float32
    Playing   bool
}

func (aud *AudioSystem) Play(name string, channel int) AudioInstance
func (aud *AudioSystem) Stop(name string)
func (aud *AudioSystem) FadeOut(name string, duration float32)
func (aud *AudioSystem) PlayLoop(bgID, name string) error
func (aud *AudioSystem) SetMasterVolume(v float32)
```

## Build Pipeline

### APK Build

```bash
# 1. Compilar cena para KScript
kora-compiler scene.ks > generated.ks

# 2. Compilar KScript para Go
go run cmd/genscript.go generated.ks > main_kscript.go

# 3. Bind para Android AAR
gomobile bind -target=android \
    -androidapi 21 \
    -o kora.aar \
    ./cmd

# 4. Criar APK com Android Gradle
cd android/app
./gradlew assembleRelease

# Output: app/build/outputs/apk/release/app-release.apk
```

### android/build.sh Script

```bash
#!/bin/bash
set -e

# Configurações
APP_NAME="kora"
API_LEVEL=21
BUILD_TYPE="${1:-release}"

# Compilar KScript se necessário
if [ -f "game.ks" ]; then
    echo "Compiling KScript..."
    go run cmd/compile.go game.ks > src/main/generated.go
fi

# Build do projeto Android
cd android/app
./gradlew assemble${BUILD_TYPE^}

# Copiar APK
if [ "$BUILD_TYPE" = "release" ]; then
    cp build/outputs/apk/release/app-release.apk ../kora-release.apk
else
    cp build/outputs/apk/debug/app-debug.apk ../kora-debug.apk
fi

echo "Build complete: $APP_NAME-$BUILD_TYPE.apk"
```

## Fluxo de Dados

### Desenvolvimento

```
1. Editor Desktop/Web
   ↓ (criar cena)
2. Save Scene (file dialog)
   ↓ (serializar JSON)
3. scene.kora.json
   ↓ (validar estrutura)
4. Editor Preview (modo JS)
   ↓ (testar lógica)
5. Export KScript (.ks)
   ↓ (compiler)
6. main.go
   ↓ (gomobile bind)
7. kora.aar
   ↓ (gradle build)
8. app-release.apk
   ↓ (adb install)
9. Device Android
```

### Runtime Execution

```
Android App Start
    ↓
Application.onCreate()
    ↓
Initialize Runtime
    ├─ AssetManager
    ├─ InputSystem
    ├─ PhysicsEngine
    ├─ Renderer
    └─ Scheduler
    ↓
Load Scene (from assets)
    ↓
Create Entities
    ├─ Parse JSON scene
    ├─ Create Entity objects
    └─ Attach components
    ↓
Main Loop (60 FPS)
    ├─ InputInput.Update()
    ├─ Physics.Update(dt)
    ├─ Entity.Update(dt)
    ├─ Asset.LoadAsync()
    └─ Renderer.Update()
    ↓
    ├─ Physics.Step(dt)
    └─ Renderer.Render()
```

## Padrões de Código

### JavaScript (Editor)

- **Pattern Observer**: state → render
- **Dependency Injection**: panels recebem events via callbacks
- **Module Federation**: `window.AssetsPanel`, `window.AssetDB`
- **Event Delegation**: eventos no canvas → event handlers

```javascript
// Observer pattern
function render() {
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  drawGrid()
  state.entities.forEach(e => e.visible && drawEntity(e))
  if (state.selected) drawSelection(state.entities[state.selected])
}

// Event delegation
canvas.addEventListener('mousedown', (e) => {
  const [wx, wy] = screenToWorld(e.clientX, e.clientY)
  const hit = findHitEntity(wx, wy)
  if (hit) selectEntity(hit.id)
})
```

### Go (Runtime)

- **Interfaces**: componentes plugáveis
- **ECS**: entidades com componentes
- **Coroutine**: async/await via generator functions
- **Object Pooling**: reuso de objetos para performance

```go
// Interface pattern
type Component interface {
    Type() string
    Update(dt float64)
}

// Object pooling
type ObjectPool struct {
    pool   chan *T
    create func() *T
}

func (p *ObjectPool) Get() *T { ... }
func (p *ObjectPool) Put(item *T) { ... }
```

### KScript

- **Static Typing**: tipos declarados
- **No dynamic eval**: tudo compilado
- **Signals**: emit() → signal() para eventos

```kscript
// Declaration
var hp: int = 100
const MAX_SPEED: float = 200.0

// Object
object Player {
    speed: float = 180.0
    hp: int = 100

    async create() {
        await wait(1.0)
        emit(this, "ready")
    }

    onHit(amount: int) {
        this.hp -= amount
        if (this.hp <= 0) {
            emit(this, "dead")
            destroyAsync(this)
        }
    }
}
```

## Dependências

```
Kora Editor Desktop
├── Electron 28.x        # Desktop wrapper
├── Vite 5.x             # Build tool
├── Node.js 18+          # Runtime
└── HTML5                # Canvas, IndexedDB

KScript Compiler
└── Go 1.22+             # Compiler runtime

Kora Runtime
├── gomobile             # Android bind
├── Android SDK          # platform-tools, build-tools
└── gomobile bind target
```

## Performance

### Objetivos de Performance

| Componente | Target | Atual |
|------------|--------|-------|
| Editor UI | 60 FPS | ✅ 60 FPS |
| Canvas rendering | 60 FPS | ✅ 60 FPS |
| Asset loading | < 100ms | ✅ < 50ms |
| APK build time | < 5 min | ⚠️ 3-7 min |
| Runtime memory | < 50 MB | ✅ 30-40 MB |
| Runtime FPS | 60 FPS | ✅ 55-60 FPS |

### Otimizações Implementadas

1. **Batch rendering**: sprites agrupados por textura
2. **Sprite atlas**: múltiplas imagens em uma textura
3. **IndexedDB caching**: assets persistentes
4. **Object pooling**: reuso de entidades
5. **Async loading**: carregar assets em background
6. **Compiled code**: KScript → Go = nativo

## Segurança

### Sandboxing

| Camada | Sandbox |
|--------|---------|
| Editor Desktop | Web context (sem acesso direto ao FS) |
| Main Process | Node.js (com permissões controladas) |
| Runtime | Android app sandbox (permissões mínimas) |

### Validação

```javascript
// Editor input validation
function validateEntity(e) {
    if (e.name.length > 64) throw Error('Nome muito longo')
    if (e.w < 0 || e.h < 0) throw Error('Dimensões inválidas')
    return true
}

// Compiler type checking
typeCheck: switch (node.type) {
    case 'NumberLit':
        // OK - sempre número
        break
    case 'BinaryExpr':
        if (!isNumber(left) || !isNumber(right)) {
            throw Error('Operandos devem ser números')
        }
        break
}

// Runtime bounds checking
func (p *Player) Update(dt float64) {
    p.X = math.Max(0, math.Min(ScreenWidth, p.X))
    p.Y = math.Max(0, math.Min(ScreenHeight, p.Y))
}
```

## Roadmap de Arquitetura

### v0.3 (Current)
- [x] Desktop Electron
- [x] Editor web fallback
- [x] IndexedDB assets
- [x] KScript compiler basic
- [x] Go runtime
- [x] Physics AABB
- [x] Input system
- [x] Asset loader
- [ ] KScript editor (in progress)

### v0.4
- [x] Plugin system
- [x] Asset optimization (sprite sheets)
- [x] Version control (Git integration)
- [ ] Tilemap editor
- [ ] Particle system
- [ ] UI system
- [ ] Shader editor

### v0.5
- [ ] Multi-threaded physics
- [ ] Asset streaming
- [ ] Cloud sync (Google Drive, OneDrive)
- [ ] Collaboration mode
- [ ] Network multiplayer
- [ ] Performance profiler

### v1.0
- [ ] Full IDE (debugger, profiler)
- [ ] Marketplace de assets/templates
- [ ] Tutorial integrado
- [ ] Auto-update mechanism
- [ ] iOS export
- [ ] Desktop export (Windows, macOS, Linux)

## Troubleshooting

### Desktop App não inicia

```bash
# Verificar dependências
node -v  # Mínimo 18
npm -v

# Reinstalar
rm -rf node_modules package-lock.json
npm install

# Logs do Electron
electron --enable-logging 2>&1 | grep -i error

# Reset config
rm -rf ~/.config/com.koraengine.editor/
```

### Build APK falha

```bash
# ANDROID_HOME
export ANDROID_HOME=$HOME/Android/Sdk
echo $ANDROID_HOME

# gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Android SDK
sdkmanager "platform-tools" "platforms;android-33"

# Build manual debug
cd android
chmod +x build.sh
./build.sh debug
```

### Editor assets não carregam

```javascript
// Console logs
console.log('IndexedDB databases:', await indexedDB.databases())

// Verificar quota
await navigator.storage.estimate()

// Limpar IndexedDB
indexedDB.deleteDatabase('kora-editor')
location.reload()
```

---

**Documento de Arquitetura - Kora Engine v0.3**
