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

## Licença

MIT License — see [LICENSE](LICENSE).

---

## Quickstart — Rodando o Projeto

### Pré-requisitos

- **Node.js 18+** (para desenvolvimento local)
- **Go 1.22+** (para compilar binários e APKs)

### Editor Visual (Desenvolvimento)

O editor é uma aplicação web pura. Para rodar localmente:

```bash
# Usando Python
cd editor
python -m http.server 8080
# Acesse: http://localhost:8080

# OU usando Node.js
npx serve editor --port 8080

# OU abra diretamente index.html no navegador
```

### Importando Assets no Editor

1. Clique na aba **Assets** no topo
2. Clique em **+ Importar** ou arraste arquivos (PNG, JPG, WebP, OGG, WAV)
3. Assets são salvos automaticamente em **IndexedDB** (persistem entre sessões)
4. Arraste um asset do painel para o canvas → cria entidade automaticamente
5. Botão "×" aparece no hover → deleta o asset

### Exportar Cena

- **Salvar JSON**: CTRL+S (cena .kora.json)
- **Exportar KScript**: Botão ".ks" → gera código executável

### Rodar Cena Exportada

```bash
# Compile sua cena .ks
go run cmd/build.go -o jogo jogo.ks

# Ou rode direto com exemplo
go run main.go examples/
```

## Build — Criando APK

### Passo 1: Preparar ambiente Android

```bash
# Instalar gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Definir ANDROID_HOME (ajuste conforme seu sistema)
export ANDROID_HOME=$HOME/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin
```

### Passo 2: Build release

```bash
cd android
./build.sh release
```

Gera `kora-release.apk` na pasta `android/app/build/outputs/apk/release/`.

### Passo 3: Build debug (mais rápido, sem signing)

```bash
./build.sh debug
```

Gera `kora-debug.apk` na pasta `android/app/build/outputs/apk/debug/`.

### Troubleshooting

| Problema | Solução |
|----------|---------|
| `gomobile: command not found` | Execute `go install golang.org/x/mobile/cmd/gomobile@latest` |
| `ANDROID_HOME not set` | Adicione ao seu `.bashrc` ou `.zshrc` |
| `SDK tools missing` | Use `sdkmanager` do Android Studio |

## Estrutura do Projeto

```
kora/
├── editor/         # Editor visual web (HTML5/JS)
│   ├── index.html
│   ├── editor.js       # Lógica principal
│   ├── assets-panel.js # Import, IndexedDB, drag-drop
│   ├── serializer.js   # JSON ↔ KScript
│   ├── idb.js          # IndexedDB wrapper
│   ├── preview.html    # Runtime preview
│   └── style.css
├── compiler/         # KScript → Go compiler
├── core/             # Runtime Go: render, input, physics, scene
├── cmd/              # Comandos CLI
├── android/          # Build pipeline APK (gomobile)
├── examples/         # Cenas de exemplo
└── docs/             # Documentação
```

## Workflow Recomendado

1. **Criar cena**: Abra editor, importe sprites (aba Assets)
2. **Montar cena**: Arraste sprites → posicione → edite propriedades
3. **Salve**: CTRL+S (`.kora.json`)
4. **Exporte KScript**: Botão ".ks"
5. **Teste**: Abra Preview (F5) ou compile para APK

## Exemplo Rápido

```bash
# 1. Abra http://localhost:8080
# 2. Importe player.png na aba Assets
# 3. Arraste para o canvas
# 4. Dê duplo clique no player no inspector
# 5. Adicione script:
#
#   on Update(dt) {
#     self.x += 100 * dt
#   }
#
# 6. Exporte como jogo.ks
# 7. go run cmd/build.go -o jogo jogo.ks
# 8. ./android/build.sh release
```
