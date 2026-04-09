# Kora Engine

> Engine de jogos 2D para Android com editor visual desktop e linguagem KScript compilada para Go.

## 🎮 O que é Kora?

Kora Engine é uma engine de jogos 2D inspirada no GameMaker Studio, projetada especificamente para criar jogos nativos para Android. Diferencia-se por:

- **Editor Desktop Nativo** - Aplicativo Electron com integração completa ao sistema
- **KScript** - Linguagem própria compilada para Go (sem VM)
- **Runtime Go** - Performance nativa ARM64 via gomobile
- **Zero Overhead** - Código compilado, executável direto no Android

### Principais Características

- ✅ **Editor Visual Desktop** - Crie cenas arrastando e soltando entidades
- ✅ **Assets Import** - PNG, JPG, WebP, OGG, WAV com persistência IndexedDB
- ✅ **Preview em Tempo Real** - Física 2D e teste imediato
- ✅ **KScript Compilado** - TypeScript-like, async/await, tipagem estática
- ✅ **APK Nativo** - Compile para Android com runtime 100% Go
- ✅ **Offline First** - Funciona completamente offline

## 🏗️ Arquitetura

```
┌────────────────────────────────────────────────────────────┐
│              Kora Editor Desktop (Electron)                 │
│  Scene Editor  ·  Asset Manager  ·  Inspector              │
│  File System  ·  Native Menus  ·  Build Pipeline           │
└─────────────────────┬──────────────────────────────────────┘
                      │ exporta cena
                      ▼
┌────────────────────────────────────────────────────────────┐
│              Scene Data (.kora.json)                        │
│  { entities: [...], meta: {...}, assets: [...] }           │
└─────────────────────┬──────────────────────────────────────┘
                      │ compila
                      ▼
┌────────────────────────────────────────────────────────────┐
│           KScript Compiler (Go)                             │
│  Lexer → Parser → Type Check → Go Emitter                  │
└─────────────────────┬──────────────────────────────────────┘
                      │ gera código Go
                      ▼
┌────────────────────────────────────────────────────────────┐
│           Kora Runtime (Go + gomobile)                      │
│  Render · Physics · Input · Audio · Scene · ECS            │
│  Async Scheduler · Asset Loader · Collision                │
└─────────────────────┬──────────────────────────────────────┘
                      │ compila ARM64
                      ▼
              Android APK / AAB (Native)
```

## 🚀 Quickstart

### Pré-requisitos

- **Node.js 18+** (editor desktop)
- **Go 1.22+** (compiler + runtime)
- **Android SDK** (apenas para build APK)

### Opção 1: Desktop App (Recomendado)

```bash
# Clone e instale
git clone https://github.com/koraengine/kora.git
cd kora/apps/desktop
npm install

# Inicie o editor
npm run dev
```

### Opção 2: Editor Web (Rápido)

```bash
cd editor
python -m http.server 8080
# Acesse: http://localhost:8080
```

### Workflow Básico

```bash
# 1. Crie cenas no editor
# 2. Importe assets (PNG, JPG, OGG, WAV)
# 3. Arraste para a cena → cria entidades
# 4. Adicione KScript → on Update(dt) { ... }
# 5. Salve: Ctrl+S (.kora.json)
# 6. Exporte KScript: botão ".ks"

# 7. Compile
go run cmd/build.go jogo.ks > main.go

# 8. Build APK
cd android && ./build.sh release
# APK em: android/app/build/outputs/apk/release/
```

## 📖 Documentação

| Documentação | Descrição |
|--------------|-----------|
| [KScript Guide](docs/SCRIPT.md) | Linguagem completa com exemplos |
| [API Reference](docs/API_REFERENCE.md) | Referências de todas APIs |
| [Editor Guide](docs/EDITOR_GUIDE.md) | Guia do editor visual |
| [Assets Guide](docs/ASSETS_GUIDE.md) | Importação e gerenciamento |
| [Desktop App](docs/DESKTOP_APP.md) | App Electron, APIs, build |
| [Architecture](docs/ARCHITECTURE.md) | Arquitetura do sistema |
| [Contributing](docs/CONTRIBUTING.md) | Como contribuir |

## 🎓 KScript - Linguagem

KScript é TypeScript-like, compilada para Go. Sem VM, zero runtime overhead.

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

### Recursos

- **Tipagem estática** - Erros em tempo de compilação
- **Async/await** - Corrotinas para lógica não-blocking
- **Signals** - Sistema reativo de eventos
- **Tweens** - Animação automática de propriedades
- **Collections** - Array, Map com generics

## 🛠️ Build & Development

### Scripts Makefile

```bash
# Setup inicial
make setup

# Run editor web
make dev

# Run desktop app
make desktop

# Build compiler
make compiler

# Run example
make run

# Build APK
make apk

# All builds
make build

# Tests
make test

# Clean artifacts
make clean
```

### Build Desktop App

```bash
cd apps/desktop

# Dev mode
npm run dev

# Build all
npm run build

# Platform-specific
npm run build:win    # Windows (nsis, portable)
npm run build:mac    # macOS (dmg, zip)
npm run build:linux  # Linux (AppImage, deb)
```

### Build APK

```bash
# Configurar ANDROID_HOME
export ANDROID_HOME=$HOME/Android/Sdk

# Debug (rápido, sem assinatura)
./android/build.sh debug

# Release (com keystore)
./android/build.sh release
```

## 📊 Tecnologia

| Camada | Tecnologia |
|--------|------------|
| **Editor Desktop** | Electron 28 + Vite + HTML5 Canvas |
| **Editor Web** | Vanilla JS + IndexedDB |
| **KScript** | Custom language → Go compiler |
| **Runtime** | Go 1.22+ + gomobile |
| **Output** | Native ARM64 APK/AAB |
| **Physics** | AABB collision + forces |
| **Assets** | Image, Audio, Tilemap loaders |

## 📦 Estrutura do Projeto

```
kora/
├── editor/                     # Editor web (HTML/JS)
│   ├── index.html
│   ├── editor.js              # Core editor logic
│   ├── assets-panel.js        # Asset management
│   ├── serializer.js          # JSON ↔ KScript
│   ├── idb.js                 # IndexedDB wrapper
│   ├── preview-panel.js       # Preview runtime
│   ├── style.css              # Styles
│   └── preview.html           # Preview page
│
├── apps/desktop/               # Desktop Electron app
│   ├── src/main/              # Electron main process
│   │   ├── index.js           # Window, menu, IPC
│   │   └── preload.js         # API bridge
│   ├── src/renderer/          # Renderer UI
│   │   ├── index.html
│   │   └── assets/
│   │       └── style.css
│   ├── package.json           # Dependencies
│   └── vite.config.js         # Build config
│
├── compiler/                   # KScript → Go compiler
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── checker/
│   └── emitter/
│
├── core/                       # Runtime Go
│   ├── render/                # 2D renderer
│   ├── physics/               # AABB physics
│   ├── input/                 # Input system
│   ├── async/                 # Task scheduler
│   ├── assets/                # Asset loading
│   └── scene/                 # Scene graph
│
├── cmd/                        # CLI commands
│   ├── build.go               # Build entry
│   └── main.go                # Runtime entry
│
├── android/                    # APK build pipeline
│   ├── build.sh               # Build script
│   └── app/                   # Gradle config
│
├── examples/                   # Example scenes/games
│
└── docs/                       # Documentation
    ├── SCRIPT.md              # Language reference
    ├── API_REFERENCE.md       # API docs
    ├── EDITOR_GUIDE.md        # Editor guide
    ├── ASSETS_GUIDE.md        # Assets guide
    ├── DESKTOP_APP.md         # Desktop app
    ├── ARCHITECTURE.md        # Architecture
    └── CONTRIBUTING.md        # How to contribute
```

## 📋 Roadmap

### v0.3 — Editor Desktop ✅
- [x] Electron wrapper
- [x] Native menus
- [x] File dialogs
- [x] Asset management
- [x] IndexedDB persistence
- [x] KScript export
- [ ] KScript editor (in progress)
- [ ] Plugin system

### v0.4 — Compiler
- [x] Basic types (int, float, bool, string)
- [x] Objects and methods
- [x] Async/await
- [x] Signals/emits
- [x] Type checking
- [ ] Collections (Array, Map)
- [ ] Enums
- [ ] Modules/imports

### v0.5 — Runtime
- [x] Entity system
- [x] 2D renderer
- [x] AABB physics
- [x] Input system
- [x] Asset loader
- [x] Sound (OGG, WAV)
- [ ] Tilemaps (Tiled)
- [ ] Particle system
- [ ] UI system

### v1.0 — Release
- [ ] Stable KScript API
- [ ] Full documentation
- [ ] Example games
- [ ] Template projects
- [ ] Auto-updater
- [ ] Multi-platform export (iOS, Windows, Linux)

## 🤝 Contribuindo

Veja [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) para como contribuir.

### Código de Conduta

[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## 💡 Exemplo Completo

```bash
# 1. Inicie o desktop app
cd apps/desktop
npm install
npm run dev

# 2. No editor:
#    - Importe player.png
#    - Arraste para a cena
#    - Adicione KScript:

on Update(dt) {
  this.x = this.x + 100 * dt
  if (this.x > 360) {
    this.x = 0
  }
}

# 3. Exporte cena: Ctrl+S → cena.kora.json
# 4. Exporte KScript: botão ".ks" → game.ks
# 5. Compile: go run cmd/build.go game.ks > main.go
# 6. Build APK: ./android/build.sh release
# 7. Instale: adb install app-release.apk
```

## 📄 License

MIT License — see [LICENSE](LICENSE)

---

**Kora Engine** - Crie jogos 2D para Android com editor visual e performance nativa.

[**Baixar Desktop App**](apps/desktop/README.md)  
[**Documentação KScript**](docs/SCRIPT.md)  
[**API Reference**](docs/API_REFERENCE.md)
