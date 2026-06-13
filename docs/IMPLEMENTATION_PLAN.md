# Kora Engine — Plano de Implementação

> **32 issues abertas** | Prioridades: 3 P0, 6 P1, 17 P2, 6 P3
> **Estimativa total:** ~4-6 meses (dev full-time)

---

## 🧭 Visão Geral da Arquitetura

```
┌──────────────────────────────────────────────────────────────────┐
│                        KORA ENGINE                               │
├──────────────────────┬───────────────────────────────────────────┤
│   ▲ GO EDITOR        │   ▲ RUNTIME (Go + Ebitengine)            │
│   (cmd/kora-editor)  │   (core/)                                │
│                      │                                           │
│   ATUAL:             │   Node2D tree → Scene → SceneTree        │
│   • Entidades flat   │   Sprite2D, Camera2D, PhysicsBody2D...   │
│   • Read-only        │   Render, Audio, Particles, Shaders...   │
│   • Sem asset        │   Compiler KScript → Go → binary         │
│   • Sem runtime      │                                           │
│                      │                                           │
│   FUTURO:            │   ←── BRIDGE ──→                         │
│   Editor visual      │   Editor ↔ Runtime unificados            │
│   conectado ao       │   Dados de cena compartilhados           │
│   runtime Go         │   Preview in-Editor                      │
└──────────────────────┴───────────────────────────────────────────┘
```

**Problema estrutural:** Editor Go e runtime têm modelos de dados SEPARADOS.
- Editor: `SceneEntity` (flat list, JSON)
- Runtime: `Node2D` (tree, Go structs)
- **Não compartilham código**

**Solução:** Criar uma **bridge** que converte SceneEntity ↔ Node2D tree,
permitindo que o editor edite cenas que o runtime executa diretamente.

---

## 📊 Mapa de Dependências

```
P0 ──► P1 ──► P2 ──► P3
┌─────────────────────────────────────────────────────────────────┐
│ #68 Preview panel          ──► #82 Sprite Editor               │
│ desconectado do runtime         (precisa de preview pra testar) │
│                              ──► #89 Tilemap Editor            │
│                                  (precisa de runtime bridge)    │
│                                                                │
│ #72 Hot-reload KScript      ──► #71 KScript LSP                │
│ #84 Intellisense                (LSP alimenta o intellisense)  │
│                                  #62 for-in/switch/enums       │
│                                  (compilador precisa ser        │
│                                   completo antes do LSP)       │
│                                                                │
│ #57 Timeline animação       ──► #83 Animator Editor            │
│                                  (timeline é base do animator)  │
│                                                                │
│ #82 Sprite Editor            ──► #85 Particle Editor           │
│ #89 Tilemap Editor              (precisa de asset pipeline)    │
│                                  #86 Shader Editor             │
│                                  #87 Audio Editor              │
│                                  #88 Vampire Survivors Example │
│                                  (precisa dos editores prontos) │
│                                                                │
│ #90 Localization             ──► #91 Dialogue System           │
│                                  (diálogo precisa de i18n)      │
│                                                                │
│ #92 Physics Debug            ──► #93 Behavior Tree             │
│                                  (debug ajuda a criar BT)       │
│                                                                │
│ #75 Android back button      ──► #74 Android IME               │
│ #61 Gamepad                  ──► #60 Touch gestures            │
│ #76 Multi-resolution                                          │
│                                                               │
│ #64 Save/Load               ──► #80 Keyboard shortcuts         │
│ #63 Camera preview            (atalhos dependem de ter funções) │
│ #65 kora-build                                               │
│ #66 docs/roadmap                                             │
│                                                               │
│ #28 Monetization  ──►  #29 Analytics  ──►  #30 Push Notif.    │
│ #94 ECS RFC                                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📋 Fase 1 — Fundação (P0/P1) • ~4-6 semanas

### 1.1 Bridge Editor ↔ Runtime (#68 - P0)
**Dependências:** Nenhuma (é a base)
**Arquivos:** `cmd/kora-editor/main.go`, `core/scene/loader.go`, `core/scene/node_entity.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 1.1.1 | Criar `core/editor/bridge.go` — converte `SceneEntity` → `Node2D` tree |
| 1.1.2 | Mapear tipos: sprite→Sprite2D, camera→Camera2D, tilemap→Tilemap2D, audio→AudioPlayer2D |
| 1.1.3 | Implementar `Node2D` → `SceneEntity` (serialização reversa) |
| 1.1.4 | Editor carrega scene JSON e exibe via runtime renderer (não mais desenho manual) |
| 1.1.5 | Preview in-editor: botão "Play" roda o runtime na mesma janela |
| 1.1.6 | Testar: criar cena no editor → rodar preview → ver sprites reais |

### 1.2 Timeline de Animação no Editor (#57 - P1)
**Dependências:** Nenhuma (runtime já tem AnimationPlayer)
**Arquivos:** `cmd/kora-editor/main.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 1.2.1 | Adicionar painel de timeline na UI do editor |
| 1.2.2 | Implementar playhead com drag |
| 1.2.3 | Botões play/pause/stop/loop |
| 1.2.4 | Keyframe: inserir/remover com clique |
| 1.2.5 | Tracks: position, rotation, scale, alpha |
| 1.2.6 | Preview da animação no viewport |
| 1.2.7 | Salvar/carregar animação como `.kora.anim` |

### 1.3 Hot-Reload KScript (#72 - P1)
**Dependências:** 1.1 (precisa do runtime bridge)
**Arquivos:** `cmd/kora-editor/main.go`, `compiler/compiler.go`, `core/scene/kscript.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 1.3.1 | Watcher de arquivos `.ks` no diretório do projeto |
| 1.3.2 | Ao detectar mudança: recompilar KScript → Go |
| 1.3.3 | Recarregar plugin Go compilado (ou reiniciar cena) |
| 1.3.4 | Mostrar erros de compilação no console do editor |
| 1.3.5 | Feedback visual (compile OK/FAIL) |

### 1.4 Preview de Câmera e Gizmos (#63 - P2)
**Dependências:** 1.1
**Arquivos:** `cmd/kora-editor/main.go`, `core/render/camera.go`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 1.4.1 | Mostrar frustum da câmera no viewport |
| 1.4.2 | Gizmos de transform (setas de posição, círculo de rotação, handles de scale) |
| 1.4.3 | Grid overlay configurável (tamanho, cor, snap) |

---

## 📋 Fase 2 — Editores Núcleo (P1/P2) • ~5-7 semanas

### 2.1 Sprite Editor (#82 - P1)
**Dependências:** 1.1
**Arquivos:** `cmd/kora-editor/main.go` + novos: `core/editor/sprite_editor.go`
**Esforço:** 2.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 2.1.1 | Importar imagem (PNG/JPEG) para o editor |
| 2.1.2 | Corte por grid (linhas × colunas) — slice tool |
| 2.1.3 | Corte manual (drag para selecionar região) |
| 2.1.4 | Editor visual de pivot (crosshair drag) + presets |
| 2.1.5 | Hitbox editor (retângulo de colisão com handles) |
| 2.1.6 | Preview: zoom, grid overlay, checkerboard bg |
| 2.1.7 | Frame animation preview (se spritesheet) |
| 2.1.8 | Salvar como `.kora.sprite` (JSON com metadados) |
| 2.1.9 | Integração com runtime: Sprite2D carrega `.kora.sprite` |

### 2.2 Tilemap Editor (#89 - P2)
**Dependências:** 2.1 (precisa do sprite editor para tileset)
**Arquivos:** `cmd/kora-editor/main.go` + novos: `core/editor/tilemap_editor.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 2.2.1 | Tile palette a partir de tileset (spritesheet) |
| 2.2.2 | Brush tools: pencil, rectangle fill, bucket |
| 2.2.3 | Pintar com clique/arrasto, snap ao grid |
| 2.2.4 | Múltiplas layers (background/midground/foreground) |
| 2.2.5 | Auto-tile com regras de adjacência (bitmask) |
| 2.2.6 | Collision painting (pintar colisão diretamente) |
| 2.2.7 | Salvar como `.kora.tilemap` |
| 2.2.8 | Runtime: Tilemap carrega `.kora.tilemap` |

### 2.3 Animator Editor (#83 - P2)
**Dependências:** 1.2 (timeline), 2.1 (sprites)
**Arquivos:** `cmd/kora-editor/main.go` + novos: `core/editor/animator_editor.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 2.3.1 | Timeline completa com zoom horizontal |
| 2.3.2 | Múltiplas tracks (position, rotation, scale, alpha, custom) |
| 2.3.3 | Easing editor visual (dropdown + preview de curva) |
| 2.3.4 | Keyframe interpolation (linear, step, bezier) |
| 2.3.5 | Animation blending (crossfade) preview |
| 2.3.6 | Frame-by-frame animation (troca de frames do spritesheet) |
| 2.3.7 | Salvar como `.kora.anim` |

---

## 📋 Fase 3 — DevEx do KScript (P1/P2) • ~5-6 semanas

### 3.1 Compilador: for-in, switch, enums (#62 - P2)
**Dependências:** Nenhuma (compilador já existe)
**Arquivos:** `compiler/parser/`, `compiler/ast/`, `compiler/checker/`, `compiler/emitter/`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 3.1.1 | `for item in collection { }` — iteração |
| 3.1.2 | `switch (expr) { case 1: ... default: ... }` |
| 3.1.3 | `enum Color { Red, Green, Blue }` |
| 3.1.4 | Type aliases: `type MyType = int` |
| 3.1.5 | Testes para cada nova feature |

### 3.2 KScript Language Server (#71 - P1)
**Dependências:** 3.1 (compilador completo)
**Arquivos:** Novos: `compiler/lsp/` ou ferramenta separada
**Esforço:** 3 semanas

| Tarefa | Descrição |
|--------|-----------|
| 3.2.1 | Implementar protocolo LSP (initialize, textDocument/completion, textDocument/hover, textDocument/definition, textDocument/diagnostic) |
| 3.2.2 | Cache de AST para resposta rápida |
| 3.2.3 | Autocomplete: palavras-chave, runtime API, símbolos do projeto |
| 3.2.4 | Hover: tipo da expressão, documentação |
| 3.2.5 | Diagnostics: erros de sintaxe/tipo em tempo real |
| 3.2.6 | Go-to-definition: navegar até declaração |
| 3.2.7 | Testar integração com VS Code (via extensão) |

### 3.3 KScript Intellisense no Editor (#84 - P1)
**Dependências:** 3.2 (LSP)
**Arquivos:** `cmd/kora-editor/main.go` + `core/editor/code_editor.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 3.3.1 | Editor de código embutido no editor Go (Syntax highlighting, line numbers) |
| 3.3.2 | Conectar ao LSP: autocomplete popup, diagnostics inline |
| 3.3.3 | Aba de script por entidade |
| 3.3.4 | Snippets para padrões KScript |

### 3.4 CLI `kora-build` (#65 - P3)
**Dependências:** 3.1 (compilador completo)
**Arquivos:** Novo: `cmd/kora-build/main.go`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 3.4.1 | Comando `kora-build game.ks --target desktop` |
| 3.4.2 | Compilar KScript → Go → `go build` → binário |
| 3.4.3 | Suporte a Windows/Mac/Linux (cross-compile) |
| 3.4.4 | Suporte a Android (gomobile) |
| 3.4.5 | Asset bundling (empacotar assets no binário) |

---

## 📋 Fase 4 — Editores de Conteúdo (P2) • ~5-6 semanas

### 4.1 Particle Editor (#85 - P2)
**Dependências:** Nenhuma (runtime já tem Particles2D)
**Arquivos:** `cmd/kora-editor/main.go` + `core/editor/particle_editor.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 4.1.1 | Viewport de preview com play/pause/restart |
| 4.1.2 | Sliders para: emission rate, lifetime, speed, gravity, spread, angle |
| 4.1.3 | Color gradient editor (múltiplos stops) |
| 4.1.4 | Size-over-life curve editor |
| 4.1.5 | Blend mode selector (add, multiply, normal) |
| 4.1.6 | Spawn shape: point, circle, rectangle |
| 4.1.7 | 6+ presets (fire, smoke, explosion, sparkle, rain, magic) |
| 4.1.8 | Salvar como `.kora.particles` |

### 4.2 Shader Editor (#86 - P2)
**Dependências:** Nenhuma (runtime já tem Kage shaders)
**Arquivos:** `cmd/kora-editor/main.go` + `core/editor/shader_editor.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 4.2.1 | Editor de código Kage com syntax highlighting |
| 4.2.2 | Preview ao vivo: shader aplicado a sprite de teste |
| 4.2.3 | Detecção automática de uniforms → sliders/color pickers |
| 4.2.4 | Biblioteca de shaders (10+ exemplos: grayscale, blur, glow, CRT, wave, glitch) |
| 4.2.5 | Hot-reload — preview atualiza ao digitar |
| 4.2.6 | Salvar como `.kora.shader` |

### 4.3 Audio Editor (#87 - P2)
**Dependências:** Nenhuma (runtime já tem mixer multi-bus)
**Arquivos:** `cmd/kora-editor/main.go` + `core/editor/audio_editor.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 4.3.1 | Importar clips OGG/WAV/MP3 |
| 4.3.2 | Waveform preview + play/stop |
| 4.3.3 | Mixer multi-bus: Master, Music, SFX, Voice com volume/mute/solo |
| 4.3.4 | VU meter (nível em tempo real) |
| 4.3.5 | Spatial audio setup: posição 2D, range visual, attenuation curve |
| 4.3.6 | Salvar como `.kora.audio` |

---

## 📋 Fase 5 — Sistemas Avançados (P2) • ~5-7 semanas

### 5.1 Physics Debug Visualization (#92 - P2)
**Dependências:** 1.1 (precisa do runtime bridge)
**Arquivos:** `core/physics/debug.go` + `cmd/kora-editor/main.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 5.1.1 | Desenhar AABB de todos os corpos (cores por tipo) |
| 5.1.2 | Desenhar collision shapes (rect/circle outlines) |
| 5.1.3 | Desenhar contatos (pontos vermelhos) + normais (setas) |
| 5.1.4 | Desenhar joints (linhas/molas) |
| 5.1.5 | Desenhar raycasts (verde hit, vermelho miss) |
| 5.1.6 | Desenhar spatial hash grid |
| 5.1.7 | Toggle F2 no editor/runtime |
| 5.1.8 | Zero overhead quando desligado |

### 5.2 Localization (i18n) (#90 - P2)
**Dependências:** Nenhuma
**Arquivos:** `core/i18n/manager.go`, `core/node/label.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 5.2.1 | TranslationManager: carregar `.po`/`.json`/`.kora.locale` |
| 5.2.2 | API KScript: `tr("key", params)`, `trn("key", count)` |
| 5.2.3 | Pluralização com regras por idioma |
| 5.2.4 | Detecção de idioma do SO |
| 5.2.5 | Label i18n (auto-update ao trocar idioma) |
| 5.2.6 | Hot-reload de traduções |

### 5.3 Dialogue System (#91 - P2)
**Dependências:** 5.2 (i18n para textos traduzíveis)
**Arquivos:** `core/dialogue/manager.go`, `core/dialogue/nodes.go`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 5.3.1 | DialogueManager com suporte a branching (árvore) |
| 5.3.2 | Node types: text, choices, condition, event, jump, random |
| 5.3.3 | Variáveis de história (flags bool + contadores) |
| 5.3.4 | Formato `.kora.dialogue` (JSON/YAML) |
| 5.3.5 | API KScript: `Dialogue.load("id").start()` |
| 5.3.6 | Typing effect, emotes, voice clips |

### 5.4 Behavior Tree (#93 - P2)
**Dependências:** Nenhuma
**Arquivos:** `core/ai/behavior_tree.go` + `core/editor/bt_editor.go`
**Esforço:** 2.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 5.4.1 | Composite: Sequence, Selector, RandomSelector, Parallel |
| 5.4.2 | Decorator: Invert, Repeat, UntilFail, Cooldown |
| 5.4.3 | Action/condition nodes + custom nodes em KScript |
| 5.4.4 | Blackboard (memória compartilhada) |
| 5.4.5 | Graph editor visual (nós arrastáveis, conexões) |
| 5.4.6 | Step-by-step preview no editor |
| 5.4.7 | API KScript: `BehaviorTree.fromFile("patrol.bt")` |
| 5.4.8 | Salvar como `.kora.bt` |

---

## 📋 Fase 6 — Input, Save/Load e Exemplo (P2) • ~4-5 semanas

### 6.1 Save/Load de Jogo (#64 - P2)
**Dependências:** Nenhuma
**Arquivos:** `core/save/manager.go`, `core/save/slot.go`
**Esforço:** 1.5 semanas

| Tarefa | Descrição |
|--------|-----------|
| 6.1.1 | SaveSlot: salvar em JSON/binário |
| 6.1.2 | Múltiplos slots (até 10) |
| 6.1.3 | API KScript: `Save.write("slot1", data)`, `Save.read("slot1")` |
| 6.1.4 | Save/Load UI (tela de seleção de slot) |
| 6.1.5 | Auto-save configurável |

### 6.2 Gamepad (#61 - P2) + Touch Gestures (#60 - P2)
**Dependências:** Nenhuma
**Arquivos:** `core/input/`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 6.2.1 | Gamepad: detectar conexão, ler axis/buttons |
| 6.2.2 | Mapeamento de gamepad para actions |
| 6.2.3 | API KScript: `Input.gamepadAxis(index)`, `Input.gamepadPressed("A")` |
| 6.2.4 | Touch: swipe detection (direção + distância) |
| 6.2.5 | Touch: pinch (zoom gesture) |
| 6.2.6 | Touch: tap longo (hold), double tap |
| 6.2.7 | Multi-touch (até 5 pontos) |

### 6.3 Multi-resolução (#76 - P2) + Atalhos Customizáveis (#80 - P2)
**Dependências:** 6.2
**Arquivos:** `core/runner/`, `core/input/actions.go`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 6.3.1 | Suporte a aspect ratio diferentes (letterbox/pillarbox) |
| 6.3.2 | Configuração de resolução em `runner.Config` |
| 6.3.3 | Sistema de atalhos configuráveis (JSON) |
| 6.3.4 | Painel de configuração de atalhos no editor |

### 6.4 Android: Back Button (#75) + IME (#74)
**Dependências:** Nenhuma
**Arquivos:** `cmd/kora-android/main.go`, `core/input/`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 6.4.1 | Capturar back button do Android → evento KScript |
| 6.4.2 | IME (teclado virtual) para entrada de texto |
| 6.4.3 | API KScript: `Input.backPressed()`, `Input.showKeyboard()` |

### 6.5 Vampire Survivors Example (#88 - P2)
**Dependências:** Todas as fases anteriores
**Arquivos:** `examples/vampire-survivors/*`
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 6.5.1 | Estrutura do projeto (scenes/entities/weapons) |
| 6.5.2 | Player movement + auto-attack |
| 6.5.3 | 5+ tipos de inimigos com BT |
| 6.5.4 | 5+ armas (projétil, área, laser, escudo, aura) |
| 6.5.5 | Sistema de level-up com upgrades |
| 6.5.6 | Hordas progressivas (spawn rate aumenta) |
| 6.5.7 | Partículas (explosão, level-up, trail) |
| 6.5.8 | Iluminação dinâmica (PointLight2D no player) |
| 6.5.9 | Shader de dano na tela (pós-processamento) |
| 6.5.10 | Áudio (BGM + SFX com mixer multi-bus) |
| 6.5.11 | UI completa (HP, level, timer, game over) |
| 6.5.12 | Transições de cena (menu → jogo → game over) |

---

## 📋 Fase 7 — Polimento & Futuro (P3) • ~3-4 semanas

### 7.1 E2E Pipeline (#81 - P3)
**Arquivos:** `scripts/e2e/`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 7.1.1 | Cena de teste em JSON |
| 7.1.2 | Testar serialização ↔ desserialização idêntica |
| 7.1.3 | Testar runner.Game.Run() em CI (xvfb) |

### 7.2 Atualizar Documentação (#66 - P3)
**Arquivos:** `docs/*`
**Esforço:** 1 semana

| Tarefa | Descrição |
|--------|-----------|
| 7.2.1 | Atualizar ROADMAP.md |
| 7.2.2 | Completar API_REFERENCE.md com todas as novas APIs |
| 7.2.3 | Guias de uso para cada editor |
| 7.2.4 | Exemplos de código KScript documentados |

### 7.3 Cloud Features (#28, #29, #30 - P3)
**Arquivos:** Novos
**Esforço:** 2 semanas

| Tarefa | Descrição |
|--------|-----------|
| 7.3.1 | Plugin system para extensões |
| 7.3.2 | IAP (Google Play Billing) |
| 7.3.3 | Analytics (eventos de gameplay) |
| 7.3.4 | Push Notifications (Firebase) |

### 7.4 ECS RFC (#94 - P3)
**Arquivos:** `docs/adr/ecs.md`
**Esforço:** 1 semana (discussão/POC)

| Tarefa | Descrição |
|--------|-----------|
| 7.4.1 | Protótipo ECS em Go (World, Entity, Component, System, Query) |
| 7.4.2 | Benchmark comparativo Node2D vs ECS (10k entidades) |
| 7.4.3 | Proposta de integração com runtime existente |

---

## 📅 Cronograma Visual

```
FASE 1 — Fundação (P0/P1)      ████████████████████░░░░ 6 sem
  1.1 Bridge Editor↔Runtime        ██████████░░░░░ 2 sem
  1.2 Timeline Anim                ██████████░░░░░ 2 sem
  1.3 Hot-Reload KScript           ███████░░░░░░░░ 1.5 sem
  1.4 Camera Preview + Gizmos      █████░░░░░░░░░░ 1 sem

FASE 2 — Editores Núcleo (P1/P2) ██████████████████████░░ 7 sem
  2.1 Sprite Editor                █████████████░░░ 2.5 sem
  2.2 Tilemap Editor               ██████████░░░░░ 2 sem
  2.3 Animator Editor              ██████████░░░░░ 2 sem

FASE 3 — DevEx KScript (P1/P2)   ████████████████████░░░░ 6 sem
  3.1 Compiler: for-in/switch      ██████████░░░░░ 2 sem
  3.2 KScript LSP                  ███████████████░░ 3 sem
  3.3 Intellisense no Editor       ███████░░░░░░░░ 1.5 sem
  3.4 CLI kora-build               █████░░░░░░░░░░ 1 sem

FASE 4 — Editores Conteúdo (P2)  ██████████████████░░░░░░ 5.5 sem
  4.1 Particle Editor              ██████████░░░░░ 2 sem
  4.2 Shader Editor                ██████████░░░░░ 2 sem
  4.3 Audio Editor                 ███████░░░░░░░░ 1.5 sem

FASE 5 — Sistemas Avançados (P2) ██████████████████████░░ 7 sem
  5.1 Physics Debug                ███████░░░░░░░░ 1.5 sem
  5.2 Localization                 ███████░░░░░░░░ 1.5 sem
  5.3 Dialogue System              ██████████░░░░░ 2 sem
  5.4 Behavior Tree                █████████████░░ 2.5 sem

FASE 6 — Input/Save/Exemplo (P2) ██████████████████░░░░░░ 6 sem
  6.1 Save/Load                    ███████░░░░░░░░ 1.5 sem
  6.2 Gamepad + Touch              ██████████░░░░░ 2 sem
  6.3 Multi-res + Shortcuts        █████░░░░░░░░░░ 1 sem
  6.4 Android Back + IME           █████░░░░░░░░░░ 1 sem
  6.5 Vampire Survivors Example    ██████████░░░░░ 2 sem

FASE 7 — Polimento & Futuro (P3) ██████████████░░░░░░░░ 4 sem
  7.1 E2E Pipeline                 █████░░░░░░░░░░ 1 sem
  7.2 Documentação                 █████░░░░░░░░░░ 1 sem
  7.3 Cloud Features               ██████████░░░░░ 2 sem
  7.4 ECS RFC                      █████░░░░░░░░░░ 1 sem
```

**Total estimado: ~42 semanas (~10 meses) — dev full-time**
**Com equipe de 2 devs: ~5-6 meses**
**Com equipe de 3 devs: ~4 meses**

---

## 🏗️ Arquivos Novos por Fase

| Fase | Arquivos Novos |
|------|---------------|
| 1 | `core/editor/bridge.go` |
| 2 | `core/editor/sprite_editor.go`, `core/editor/tilemap_editor.go`, `core/editor/animator_editor.go` |
| 3 | `compiler/lsp/server.go`, `compiler/lsp/handler.go`, `core/editor/code_editor.go`, `cmd/kora-build/main.go` |
| 4 | `core/editor/particle_editor.go`, `core/editor/shader_editor.go`, `core/editor/audio_editor.go` |
| 5 | `core/physics/debug.go`, `core/i18n/manager.go`, `core/dialogue/manager.go`, `core/dialogue/nodes.go`, `core/ai/behavior_tree.go`, `core/editor/bt_editor.go` |
| 6 | `core/save/manager.go`, `core/save/slot.go`, `examples/vampire-survivors/*` (12+ arquivos) |
| 7 | `scripts/e2e/`, `docs/adr/ecs.md` |

---

## ⚡ Recomendações de Ordem

### Se você tem 1 mês:
1. 🔴 Bridge Editor↔Runtime (#68) — desbloqueia tudo
2. 🔴 Timeline de animação (#57) — base visual
3. 🔴 Sprite Editor (#82) — editor mais usado
4. 🟡 Hot-Reload KScript (#72) — produtividade

### Se você tem 3 meses:
Tudo acima + Tilemap (#89), Animator (#83), Compiler (#62), Physics Debug (#92)

### Se você tem 6 meses:
Tudo acima + LSP (#71), Intellisense (#84), Particle (#85), Shader (#86), Audio (#87), Save/Load (#64), Gamepad (#61), Vampire Survivors (#88)

---

## 🔗 Links

- [Issues abertas](https://github.com/ElioNeto/kora/issues)
- [ROADMAP.md](ROADMAP.md)
- [README.md](../README.md)
- [API Reference](API_REFERENCE.md)
- [Guia KScript](SCRIPT.md)
