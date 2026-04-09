# CLAUDE.md — Guia de Contexto para o Kora Engine

Este arquivo é lido automaticamente pelo Claude ao trabalhar neste repositório.
Ele fornece contexto crítico sobre arquitetura, convenções e decisões de design.

---

## O que é o Kora

Kora é uma **game engine 2D para Android** com editor visual desktop.
O objetivo é ser "o GameMaker brasileiro": criar jogos Android sem precisar
configurar Android Studio, NDK ou Gradle manualmente.

**Stack principal:**
- Runtime: Go 1.22 + Ebiten v2 (2D rendering, áudio, input)
- Exportação Android: `golang.org/x/mobile` + `gomobile`
- Linguagem de script: **KScript** (transpila para Go antes do build APK)
- Editor: HTML/CSS/JS puro (sem framework), roda no browser ou Electron

---

## Estrutura do Repositório

```
kora/
├── main.go                  # Entry point do runtime
├── go.mod                   # Módulo: github.com/ElioNeto/kora
│
├── core/                    # Engine core (Go)
│   ├── engine/              # Loop principal, inicialização Ebiten
│   ├── scene/               # SceneManager, Entity, Transition
│   ├── physics/             # Física 2D AABB, gravidade, colisão
│   ├── audio/               # AudioPlayer, SoundEmitter, loop/fade
│   ├── input/               # InputManager, VirtualPad, touch/swipe
│   ├── render/              # Sprite renderer, câmera, tilemap
│   └── async/               # Scheduler de corrotinas (async/await KScript)
│
├── compiler/                # Compilador KScript → Go
│   ├── lexer/               # Tokenização
│   ├── parser/              # AST
│   ├── ast/                 # Nós do AST
│   ├── checker/             # Type checker
│   ├── transform/           # Otimizações AST
│   ├── emitter/             # Code generation Go
│   └── compiler.go          # Orquestrador
│
├── runner/                  # Integração runtime + scripts compilados
├── android/                 # Manifests, build scripts, Gradle wrapper
├── editor/                  # Editor visual (HTML/JS)
│   ├── index.html           # Shell do editor
│   ├── editor.js            # Lógica principal (scene, inspetor, hierarquia)
│   ├── serializer.js        # JSON ↔ cena, exportação KScript
│   ├── preview-panel.js     # Preview in-browser do jogo
│   ├── assets-panel.js      # Painel de assets drag-and-drop
│   └── style.css            # Design system (dark/light mode)
│
├── templates/               # Projetos de exemplo (platformer, topdown, puzzle)
├── examples/                # Jogos de demonstração
├── docs/                    # Documentação
│   ├── SCRIPT.md            # Referência completa da linguagem KScript
│   ├── API_REFERENCE.md     # API do engine (Physics, Input, Audio, Camera...)
│   ├── ARCHITECTURE.md      # Decisões de arquitetura
│   ├── EDITOR_GUIDE.md      # Como usar o editor
│   └── kscript-spec.md      # Spec formal da gramática KScript
└── cmd/                     # CLI: kora build, kora run, kora export
```

---

## A Linguagem KScript

KScript é **transpilada para Go** (AOT), não interpretada. Não há VM.
Sua sintaxe é inspirada no TypeScript com elementos de jogos.

### Palavras-chave principais
```
object, component, fn, async, await, emit, signal
var, const, type, enum, if, else, when, for, while
return, break, continue, spawn, cancel, loop
```

### Lifecycle hooks de objetos
```kscript
object Player {
  async create()          // inicialização (pode ser async)
  update(dt: float)       // chamado todo frame
  draw()                  // render customizado
  onDestroy()             // cleanup
  onCollision(other, type)
  onInput(key, action)
}
```

### Regras críticas do KScript
- Tipos primitivos: `bool`, `int`, `float`, `string`, `void`
- Tipos engine: `Vector2`, `Vector3`, `Color`, `Rect`, `Array<T>`, `Map<K,V>`
- Campos privados: prefixo `_` (ex: `_state`)
- Constantes: `UPPER_SNAKE_CASE`
- Métodos e propriedades: `camelCase`
- Objects/types/enums: `PascalCase`
- String interpolation: `"$variavel"` ou `"$(expr)"`
- Null safety: `?.` (null-safe access), `??` (null coalesce), `T?` (nullable type)

### Módulos built-in expostos ao KScript
| Módulo | Responsabilidade |
|--------|------------------|
| `Input` | Teclado, touch, gamepad, joystick virtual |
| `Audio` | play, stop, loop, fade, volume por canal |
| `Camera` | follow, zoom, shake, bounds |
| `Asset` | load, cache, preload de sprites/áudio |
| `System` | width, height, time, delta, random |
| `Physics` | gravity, raycast, solid/trigger |
| `Scene` | load, reload, additive, transições |
| `Entity` | get, create, destroy, getOverlaps |

---

## Editor Visual (editor/)

O editor roda como **HTML estático** — sem servidor, sem build step.
Todas as dependências são CDN. A cena é salva como `.kora.json`.

### Estado global (`state` em editor.js)
```js
state = {
  entities: [],          // Array de entidades da cena atual
  selected: null,        // ID da entidade selecionada
  tool: 'select',        // 'select' | 'move' | 'scale'
  cam: { x, y, zoom },  // Câmera do editor
  meta: { name, version, logicalW: 360, logicalH: 640 },
  dirty: false,          // há alterações não salvas?
}
```

### Resolução lógica padrão
**360×640** (retrato Android). Todo posicionamento de entidade usa coordenadas
do mundo (0,0 = centro da cena). Conversão `worldToScreen` / `screenToWorld`
usa o `cam.zoom` e `cam.{x,y}`.

### Formato `.kora.json`
```json
{
  "meta": { "name": "string", "version": 1, "logicalW": 360, "logicalH": 640 },
  "entities": [
    {
      "id": 1,
      "name": "Player",
      "type": "sprite",        // sprite | tilemap | camera | audio | custom
      "x": 0, "y": 0,
      "w": 48, "h": 48,
      "rotation": 0,
      "visible": true,
      "locked": false,
      "color": "#00e5a0",
      "assetId": "asset_...",  // opcional
      "script": ""             // código KScript inline
    }
  ]
}
```

---

## Convenções de Código Go

### Pacotes
- Um pacote por diretório, nome = nome do diretório
- Exports com `PascalCase`, internos com `camelCase`
- Interfaces com sufixo na descrição, não em `I` prefixo (ex: `AudioPlayer`, não `IAudioPlayer`)

### Testes
- Todo arquivo `foo.go` deve ter `foo_test.go` correspondente
- Use `testing.T` padrão, sem frameworks externos
- Nomes: `TestNomeDaFuncao_cenario` (ex: `TestPhysics_AABBCollision`)

### Ebiten
- O game loop é `Update()` + `Draw()` chamados pelo Ebiten
- `Update()` roda a 60 TPS fixo; `dt = 1.0/60.0` (use `ebiten.ActualTPS()` para debug)
- Nunca bloqueie a goroutine do game loop — use o scheduler de `core/async/`
- Rendering via `*ebiten.Image`; nunca use `image.RGBA` direto para render final

### gomobile / Android export
- O entry point Android é `android/main.go` (gerado pelo build script)
- `build.sh debug` → APK não assinado para testes
- `build.sh release` → AAB assinado com keystore (não commitar keystore no repo)
- Target SDK: 34 (Android 14), minSDK: 24 (Android 7)

---

## Issues Abertas (backlog)

| # | Título | Prioridade |
|---|--------|------------|
| [#5](https://github.com/ElioNeto/kora/issues/5) | Editor KScript (CodeMirror 6) | Alta |
| [#4](https://github.com/ElioNeto/kora/issues/4) | CI/CD APK via GitHub Actions | Alta |
| [#10](https://github.com/ElioNeto/kora/issues/10) | Câmera 2D (follow, zoom, bounds) | Média |
| [#11](https://github.com/ElioNeto/kora/issues/11) | Animação de sprites (spritesheet, Timeline) | Média |
| [#6](https://github.com/ElioNeto/kora/issues/6) | Templates plataforma / top-down / puzzle | Baixa |

---

## Decisões de Design (ADRs resumidos)

### Por que Ebiten e não SDL/OpenGL direto?
Ebiten abstrai a camada gráfica de forma idiomática em Go, tem suporte
nativo a `gomobile` e export Android/iOS, e possui áudio integrado via
`ebitengine/oto`. Evita dependências de C/cgo complexas.

### Por que KScript transpila para Go (AOT) e não tem VM?
Performance nativa em Android sem overhead de interpretação. Go compila
para ARM via `gomobile`, gerando binários pequenos (~8–12 MB base).
Scripts KScript são verificados em tempo de compilação, não em runtime.

### Por que o editor é HTML estático e não Electron/desktop app?
Reduz a barreira de entrada: qualquer browser serve. Electron pode ser
adicionado como wrapper depois (ver `docs/DESKTOP_APP.md`) sem mudar
a lógica do editor. A serialização `.kora.json` é independente de plataforma.

### Resolução lógica 360×640
Representa a maioria dos smartphones Android em portrait. O runtime
escala para a resolução real do dispositivo mantendo aspect ratio (letterbox).
No editor, a mesma resolução lógica é usada para garantir fidelidade no preview.

---

## O que NÃO fazer

- **Não editar** `*.ks.go` (arquivos gerados pelo transpiler KScript)
- **Não commitar** `*.keystore`, `.env`, `android/signing.properties`
- **Não usar** `localStorage` ou `sessionStorage` no editor (roda em iframe sandbox)
- **Não usar** `cgo` no core — todo o código deve compilar com `GOARCH=arm64 GOOS=android` via gomobile
- **Não hardcodar** resolução no editor — sempre usar `state.meta.logicalW/H`
- **Não quebrar** a API pública de `core/physics`, `core/audio`, `core/input`, `core/scene`
  sem atualizar `docs/API_REFERENCE.md` e os testes correspondentes
