# 🗺️ Kora Engine — Roadmap

> Motor 2D desktop-first com runtime Go, linguagem KScript e ferramentas CLI.
> Este roadmap organiza o desenvolvimento em três grandes milestones, do runtime desktop ao ecossistema completo.

---

## 🧭 Direção do Projeto

A partir de 2025, a Kora Engine **migrou para desktop-first**. As mudanças estratégicas:

- **Go como centro** — runtime, compilador, CLI e futuro editor serão todos em Go
- **Desktop-native** — Windows, macOS e Linux como alvos primários; Android via gomobile é secundário
- **CLI-first** — todo o fluxo de desenvolvimento via terminal, sem dependência de navegador
- **Editor nativo Go** — substituição gradual do editor HTML/JS por um editor desktop escrito em Go (Ebitengine + interface IMGUI)
- **KScript como linguagem primária** — compilador próprio maduro, sem VM, sem runtime extra

> O editor web legado (`editor/`) permanece para referência, mas não receberá novas funcionalidades.
> Issues de store, multiplayer e cloud foram movidas para v3.0.

---

## 🚧 v1.0 — Runtime Desktop (Em andamento)

> **Objetivo:** Runtime desktop estável, ferramentas CLI completas, exportação nativa para Windows/Mac/Linux.
> **Status:** 🚧 Em desenvolvimento ativo

### ✅ Implementado

| Issue | Descrição |
|---|---|
| [#33](https://github.com/ElioNeto/kora/issues/33) | Sistema de Nós (Node2D) |
| [#34](https://github.com/ElioNeto/kora/issues/34) | Sistema de Cenas |
| [#35](https://github.com/ElioNeto/kora/issues/35) | SceneTree |
| [#10](https://github.com/ElioNeto/kora/issues/10) | Câmera 2D (follow, zoom, shake) |
| [#22](https://github.com/ElioNeto/kora/issues/22) | Física 2D (AABB, gravidade) |
| [#37](https://github.com/ElioNeto/kora/issues/37) | Corpos físicos (RigidBody2D, CharacterBody2D, StaticBody2D, Area2D) |
| [#16](https://github.com/ElioNeto/kora/issues/16) | Sistema de Sprites (Sprite2D, spritesheet, animator) |
| [#11](https://github.com/ElioNeto/kora/issues/11) | Animação de sprites (frame strip + keyframe) |
| [#17](https://github.com/ElioNeto/kora/issues/17) | Sistema de Objetos + Eventos KScript |
| [#18](https://github.com/ElioNeto/kora/issues/18) | Editor de Salas (grid, snap, layers) |
| [#23](https://github.com/ElioNeto/kora/issues/23) | Sistema de Tiles (auto-tile, collision layers) |
| [#38](https://github.com/ElioNeto/kora/issues/38) | AnimationPlayer (keyframe com easing) |
| [#19](https://github.com/ElioNeto/kora/issues/19) | CutscenePlayer (12 ações, timeline) |
| [#21](https://github.com/ElioNeto/kora/issues/21) | Mixer de Áudio (multi-bus, som espacial) |
| [#20](https://github.com/ElioNeto/kora/issues/20) | Debugger (FPS, tasks, node tree) |
| [#36](https://github.com/ElioNeto/kora/issues/36) | AutoLoad (Singletons) |
| [#5](https://github.com/ElioNeto/kora/issues/5) | Painel de código KScript com syntax highlight |
| [#25](https://github.com/ElioNeto/kora/issues/25) | Exportação Android (APK debug + AAB release) |
| [#40](https://github.com/ElioNeto/kora/issues/40) | Sistema de Partículas (CPU, burst/contínuo, blend modes) |
| [#41](https://github.com/ElioNeto/kora/issues/41) | Iluminação 2D Dinâmica (PointLight, DirectionalLight, sombras) |
| [#42](https://github.com/ElioNeto/kora/issues/42) | Pathfinding A\* (NavigationRegion2D, NavigationAgent2D) |
| [#43](https://github.com/ElioNeto/kora/issues/43) | Parallax (ParallaxBackground, camadas com scroll) |
| [#39](https://github.com/ElioNeto/kora/issues/39) | Skeleton2D (Bone2D, CCD IK, rest pose) |
| [#24](https://github.com/ElioNeto/kora/issues/24) | Shaders (ShaderManager, ShaderNode, Kage) |
| [#48](https://github.com/ElioNeto/kora/issues/48) | Bridge Entity↔Node2D |
| [#49](https://github.com/ElioNeto/kora/issues/49) | Renderização real de sprites |
| [#50](https://github.com/ElioNeto/kora/issues/50) | Bitmap font + DebugTextAt |
| [#51](https://github.com/ElioNeto/kora/issues/51) | Unificação PhysicsBody2D ↔ PhysicsWorld |
| [#52](https://github.com/ElioNeto/kora/issues/52) | Áudio real no AudioPlayer2D |
| [#53](https://github.com/ElioNeto/kora/issues/53) | UI System (Label, Button, Panel) |
| [#54](https://github.com/ElioNeto/kora/issues/54) | Prefabs (templates reutilizáveis) |
| [#55](https://github.com/ElioNeto/kora/issues/55) | SpatialHash (broad-phase colisão) |
| [#56](https://github.com/ElioNeto/kora/issues/56) | Asset Manager (ref counting, async) |
| [#58](https://github.com/ElioNeto/kora/issues/58) | Joints (Distance, Spring, Pin) |
| [#67](https://github.com/ElioNeto/kora/issues/67) | Unificação engine/runner |
| [#69](https://github.com/ElioNeto/kora/issues/69) | Frustum culling tilemaps |
| [#70](https://github.com/ElioNeto/kora/issues/70) | Object Pool genérico |
| [#73](https://github.com/ElioNeto/kora/issues/73) | Benchmark tests (30+) |
| [#77](https://github.com/ElioNeto/kora/issues/77) | 28 funções de easing |
| [#78](https://github.com/ElioNeto/kora/issues/78) | CCD (Continuous Collision Detection) |
| [#4](https://github.com/ElioNeto/kora/issues/4) | CI/CD — GitHub Actions |
| — | CLI `kora-run` (compilador/executor) |
| — | CLI `kora-android` (entry point gomobile) |
| — | Makefile com comandos de build, teste, bench |

### 🔲 Planejado para v1.0

| Issue | Descrição | Prioridade |
|---|---|---|
| [#79](https://github.com/ElioNeto/kora/issues/79) | Sprite batching para GPU | Alta |
| — | CLI `kora-build` (empacotador desktop) | Alta |
| — | Exportação nativa Windows (exe) | Alta |
| — | Exportação nativa macOS (app bundle) | Alta |
| — | Exportação nativa Linux (binário) | Alta |
| — | Editor Go (versão初期 baseada em Ebitengine) | Média |
| [#59](https://github.com/ElioNeto/kora/issues/59) | Render targets (pós-processamento) | Média |
| [#61](https://github.com/ElioNeto/kora/issues/61) | Suporte a gamepad | Média |
| [#62](https://github.com/ElioNeto/kora/issues/62) | for-in, switch, enums no KScript | Média |
| [#57](https://github.com/ElioNeto/kora/issues/57) | Timeline de animação | Média |
| [#80](https://github.com/ElioNeto/kora/issues/80) | Atalhos de teclado customizáveis | Baixa |
| [#81](https://github.com/ElioNeto/kora/issues/81) | Pipeline de integração E2E | Baixa |
| — | Documentação completa da runtime | Baixa |
| — | Jogos exemplo (platformer, top-down) | Baixa |

---

## 🔄 v2.0 — Editor Go & Ecossistema (Planejado)

> **Objetivo:** Editor desktop nativo Go, pipeline de assets, distribuição de templates.
> **Status:** 🔲 Aguardando conclusão da v1.0

| Item | Descrição | Prioridade |
|---|---|---|
| Editor Go (Ebitengine + IMGUI) | Substituir editor HTML/JS por editor nativo | Alta |
| Asset Pipeline | Importação, conversão e otimização de assets | Alta |
| Scene Editor | Viewport com arrastar e soltar, hierarquia, inspetor | Alta |
| KScript IDE | Editor de código com autocomplete, diagnóstico, snippets | Alta |
| Gerenciador de Projetos | Criar, abrir, gerenciar projetos Kora | Média |
| Template System | Sistema de templates de projetos (local) | Média |
| Sprite Atlas | Empacotamento automático de sprites | Média |
| Build Profiles | Configurações de exportação por plataforma | Média |
| Plugin System | API de plugins para extensão do editor | Baixa |

---

## ☁️ v3.0 — Cloud & Multiplayer (Futuro)

> **Objetivo:** Loja de templates, multiplayer, serviços online e monetização.
> **Status:** 🔲 Sem previsão

| Issue | Descrição | Motivo |
|---|---|---|
| [#6](https://github.com/ElioNeto/kora/issues/6) | Loja de templates | Requer backend e marketplace |
| [#31](https://github.com/ElioNeto/kora/issues/31) | Networking / Multiplayer | Requer servidor dedicado |
| [#28](https://github.com/ElioNeto/kora/issues/28) | Monetização (IAP + anúncios) | Requer Google Play / loja |
| [#29](https://github.com/ElioNeto/kora/issues/29) | Analytics | Requer backend de coleta |
| [#30](https://github.com/ElioNeto/kora/issues/30) | Push Notifications | Requer Firebase / serviço externo |
| — | Leaderboards / Matchmaking | Requer servidor dedicado |
| — | Cloud Save | Requer backend com autenticação |
| — | Asset Store | Requer marketplace e curadoria |

---

## 📊 Métricas do Projeto

- **Pacotes Go**: 17+
- **Linhas de código**: ~25.000+ (Go + KScript)
- **Testes**: 500+ testes + 30+ benchmarks
- **Cobertura**: Todos os pacotes com testes
- **Dependência externa**: Apenas Ebitengine v2.7

---

## Legenda

| Ícone | Significado |
|---|---|
| ✅ | Concluído |
| 🚧 | Em progresso |
| 🔄 | Planejado (próximo) |
| 🔲 | Futuro (sem previsão) |

---

## Links

- 🔗 [Issues abertas](https://github.com/ElioNeto/kora/issues)
- 📋 [Milestones](https://github.com/ElioNeto/kora/milestones)
- 📖 [Documentação](https://github.com/ElioNeto/kora/tree/main/docs)

---

**Kora Engine** — Motor 2D desktop-first. Go. KScript. CLI. Performance nativa.
