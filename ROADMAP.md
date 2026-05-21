# 🗺️ Kora Engine — Roadmap

> Engine 2D para criação de jogos Android, programada em **KScript**.
> Este roadmap organiza o desenvolvimento em milestones, do MVP até uma engine completa.

---

## ✅ v0.1 — MVP: Runtime + Editor Base

> **Objetivo:** Ter um jogo simples rodando no editor e exportável como APK.
> **Status: ✅ Concluído**

| Status | Issue | Descrição |
|--------|-------|-----------|
| ✅ | [#33](https://github.com/ElioNeto/kora/issues/33) | Sistema de Nós (Node2D) |
| ✅ | [#34](https://github.com/ElioNeto/kora/issues/34) | Sistema de Cenas |
| ✅ | [#35](https://github.com/ElioNeto/kora/issues/35) | SceneTree |
| ✅ | [#10](https://github.com/ElioNeto/kora/issues/10) | Câmera 2D (follow, zoom, shake) |
| ✅ | [#22](https://github.com/ElioNeto/kora/issues/22) | Física 2D (AABB, gravidade) |
| ✅ | [#37](https://github.com/ElioNeto/kora/issues/37) | Corpos físicos: RigidBody2D, CharacterBody2D, StaticBody2D, Area2D |
| ✅ | [#16](https://github.com/ElioNeto/kora/issues/16) | Sistema de Sprites (Sprite2D, spritesheet, animator) |
| ✅ | [#11](https://github.com/ElioNeto/kora/issues/11) | Animação de sprites (frame strip + keyframe) |
| ✅ | [#5](https://github.com/ElioNeto/kora/issues/5) | Painel de código KScript com syntax highlight |
| ✅ | [#4](https://github.com/ElioNeto/kora/issues/4) | CI/CD — GitHub Actions |
| 🔲 | [#6](https://github.com/ElioNeto/kora/issues/6) | Loja de templates |

---

## ✅ v0.2 — Beta: Editor + Ferramentas de Criação

> **Objetivo:** Editor maduro com ferramentas completas para criação de jogos.
> **Status: ✅ Concluído**

| Status | Issue | Descrição |
|--------|-------|-----------|
| ✅ | [#17](https://github.com/ElioNeto/kora/issues/17) | Sistema de Objetos + Eventos KScript |
| ✅ | [#18](https://github.com/ElioNeto/kora/issues/18) | Editor de Salas (grid snapping, layers, multi-select) |
| ✅ | [#23](https://github.com/ElioNeto/kora/issues/23) | Sistema de Tiles (auto-tile, collision layers) |
| ✅ | [#38](https://github.com/ElioNeto/kora/issues/38) | AnimationPlayer (keyframe com easing) |
| ✅ | [#19](https://github.com/ElioNeto/kora/issues/19) | CutscenePlayer (12 tipos de ação, timeline) |
| ✅ | [#21](https://github.com/ElioNeto/kora/issues/21) | Mixer de Áudio (multi-bus, som espacial) |
| ✅ | [#20](https://github.com/ElioNeto/kora/issues/20) | Debugger (FPS, tasks, node tree) |
| ✅ | [#36](https://github.com/ElioNeto/kora/issues/36) | AutoLoad (Singletons) |
| ✅ | [#32](https://github.com/ElioNeto/kora/issues/32) | Git Panel (status, stage, commit, diff) |

---

## ✅ v0.3 — Runtime Features Avançadas

> **Objetivo:** Funcionalidades avançadas de runtime para jogos com qualidade profissional.
> **Status: ✅ Concluído**

| Status | Issue | Descrição |
|--------|-------|-----------|
| ✅ | [#25](https://github.com/ElioNeto/kora/issues/25) | Exportação Android (APK debug + AAB release) |
| ✅ | [#40](https://github.com/ElioNeto/kora/issues/40) | Sistema de Partículas (CPU, burst/continuous, blend modes) |
| ✅ | [#41](https://github.com/ElioNeto/kora/issues/41) | Iluminação 2D Dinâmica (PointLight, DirectionalLight, sombras) |
| ✅ | [#42](https://github.com/ElioNeto/kora/issues/42) | Pathfinding A* (NavigationRegion2D, NavigationAgent2D) |
| ✅ | [#43](https://github.com/ElioNeto/kora/issues/43) | Parallax (ParallaxBackground, camadas com scroll independente) |
| ✅ | [#39](https://github.com/ElioNeto/kora/issues/39) | Skeleton2D (Bone2D, CCD IK, rest pose) |
| ✅ | [#24](https://github.com/ElioNeto/kora/issues/24) | Shaders (ShaderManager, ShaderNode, Kage) |
| 🔲 | [#31](https://github.com/ElioNeto/kora/issues/31) | Networking / Multiplayer |

---

## 🚧 v1.0 — Engine Completa (Em andamento)

> **Objetivo:** Engine polida, documentada e pronta para publicação.
> Abaixo estão as features implementadas e planejadas.

### ✅ Implementado

| Issue | Descrição |
|-------|-----------|
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

### 🔲 Planejado

| Issue | Descrição | Prioridade |
|-------|-----------|------------|
| [#57](https://github.com/ElioNeto/kora/issues/57) | Timeline de animação no editor | Alta |
| [#59](https://github.com/ElioNeto/kora/issues/59) | Render targets (pós-processamento) | Média |
| [#61](https://github.com/ElioNeto/kora/issues/61) | Suporte a gamepad | Média |
| [#62](https://github.com/ElioNeto/kora/issues/62) | for-in, switch, enums no KScript | Média |
| [#79](https://github.com/ElioNeto/kora/issues/79) | Sprite batching | Média |
| [#80](https://github.com/ElioNeto/kora/issues/80) | Atalhos de teclado customizáveis | Média |
| [#65](https://github.com/ElioNeto/kora/issues/65) | Comando kora-build | Baixa |
| [#66](https://github.com/ElioNeto/kora/issues/66) | Atualizar documentação | Baixa |
| [#81](https://github.com/ElioNeto/kora/issues/81) | Pipeline de integração E2E | Baixa |

### 📌 Externos (requerem serviços terceiros)

| Issue | Descrição | Motivo |
|-------|-----------|--------|
| [#28](https://github.com/ElioNeto/kora/issues/28) | Monetização (IAP + anúncios) | Requer Google Play Billing |
| [#29](https://github.com/ElioNeto/kora/issues/29) | Analytics | Requer backend de coleta |
| [#30](https://github.com/ElioNeto/kora/issues/30) | Push Notifications | Requer Firebase Cloud Messaging |
| [#31](https://github.com/ElioNeto/kora/issues/31) | Networking / Multiplayer | Requer servidor dedicado |

---

## Legenda

| Ícone | Significado |
|-------|-------------|
| ✅ | Concluído |
| 🔄 | Em progresso |
| 🔲 | Não iniciado |

---

## 📊 Métricas do Projeto

- **Pacotes Go**: 17+
- **Linhas de código**: ~25.000+ (Go + JS)
- **Testes**: 500+ testes + 30+ benchmarks
- **Cobertura**: Todos os pacotes com testes
- **Dependência externa**: Apenas Ebitengine v2.7

---

## Sobre a Kora

**Kora** é uma engine 2D focada em jogos Android, usando **KScript** como linguagem de programação nativa. O objetivo é oferecer uma experiência de desenvolvimento simples e direta — do editor ao APK — sem precisar configurar SDKs externos.

- 🔗 [Issues abertas](https://github.com/ElioNeto/kora/issues)
- 📋 [Milestones](https://github.com/ElioNeto/kora/milestones)
- 📖 [Documentação](https://github.com/ElioNeto/kora/tree/main/docs)
