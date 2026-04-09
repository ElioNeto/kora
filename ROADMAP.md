# 🗺️ Kora Engine — Roadmap

> Engine 2D para criação de jogos Android, programada em **KScript**.
> Este roadmap organiza o desenvolvimento em 4 milestones progressivas, da fundação até uma engine completa e publicável.

---

## v0.1 — MVP: Runtime + Editor Base

> **Objetivo:** Ter um jogo simples rodando no editor e exportável como APK de debug para Android.
> Esta é a fundação sobre a qual todo o resto será construído.

| Status | Issue | Descrição |
|--------|-------|-----------|
| 🔲 | [#33](https://github.com/ElioNeto/kora/issues/33) | Sistema de Nós (Nodes) como blocos de construção da engine |
| 🔲 | [#34](https://github.com/ElioNeto/kora/issues/34) | Sistema de Cenas (Árvore de Nós) com composição e herança |
| 🔲 | [#35](https://github.com/ElioNeto/kora/issues/35) | SceneTree — Loop principal e ciclo de vida dos nós |
| 🔲 | [#10](https://github.com/ElioNeto/kora/issues/10) | Sistema de câmera 2D — follow, zoom, limites de mundo |
| 🔲 | [#22](https://github.com/ElioNeto/kora/issues/22) | Motor de Física 2D integrado |
| 🔲 | [#37](https://github.com/ElioNeto/kora/issues/37) | Tipos de corpos físicos: RigidBody2D, CharacterBody2D, StaticBody2D e Area2D |
| 🔲 | [#16](https://github.com/ElioNeto/kora/issues/16) | Sistema de Sprites — Editor e importação |
| 🔲 | [#11](https://github.com/ElioNeto/kora/issues/11) | Sistema de animação de sprites — frame strip, Timeline, eventos |
| 🔲 | [#5](https://github.com/ElioNeto/kora/issues/5)  | Painel de código KScript com syntax highlight |
| 🔲 | [#4](https://github.com/ElioNeto/kora/issues/4)  | CI/CD — GitHub Actions gera APK a cada push |
| 🔲 | [#6](https://github.com/ElioNeto/kora/issues/6)  | Loja de templates — plataforma, top-down, puzzle |

---

## v0.2 — Beta: Editor Completo + Ferramentas de Criação

> **Objetivo:** Editor maduro com todas as ferramentas que um desenvolvedor precisa para criar jogos completos e bem polidos.

| Status | Issue | Descrição |
|--------|-------|-----------|
| 🔲 | [#17](https://github.com/ElioNeto/kora/issues/17) | Sistema de Objetos com Eventos e Ações via KScript |
| 🔲 | [#18](https://github.com/ElioNeto/kora/issues/18) | Editor de Salas (Rooms) — Construção de níveis |
| 🔲 | [#23](https://github.com/ElioNeto/kora/issues/23) | Sistema de Tiles — Auto-tile e pincéis de blocos |
| 🔲 | [#38](https://github.com/ElioNeto/kora/issues/38) | AnimationPlayer — Animação de propriedades de nós via timeline |
| 🔲 | [#19](https://github.com/ElioNeto/kora/issues/19) | Editor de Sequências (cutscenes e animações complexas) |
| 🔲 | [#21](https://github.com/ElioNeto/kora/issues/21) | Mixer de Áudio — Efeitos sonoros e música |
| 🔲 | [#20](https://github.com/ElioNeto/kora/issues/20) | Depurador (Debugger) — Inspeção em tempo real |
| 🔲 | [#36](https://github.com/ElioNeto/kora/issues/36) | AutoLoad (Singletons) — Estado global acessível em qualquer cena |
| 🔲 | [#32](https://github.com/ElioNeto/kora/issues/32) | Integração nativa com Git para controle de versão |

---

## v0.3 — Publicação Android + Monetização

> **Objetivo:** Pipeline completo para publicar jogos na Google Play Store e monetizá-los com IAP, anúncios, analytics e push notifications.

| Status | Issue | Descrição |
|--------|-------|-----------|
| 🔲 | [#25](https://github.com/ElioNeto/kora/issues/25) | Exportação para Android (APK de debug e AAB assinado para produção) |
| 🔲 | [#28](https://github.com/ElioNeto/kora/issues/28) | Monetização — Compras no aplicativo (IAP) e anúncios |
| 🔲 | [#29](https://github.com/ElioNeto/kora/issues/29) | Analytics — Coleta de dados de comportamento do jogador |
| 🔲 | [#30](https://github.com/ElioNeto/kora/issues/30) | Push Notifications para engajamento de jogadores |

---

## v1.0 — Engine Completa: Recursos Avançados

> **Objetivo:** Funcionalidades avançadas que diferenciam a Kora, tornando-a capaz de produzir jogos Android visualmente ricos, com IA, física avançada e atualizações eficientes.

| Status | Issue | Descrição |
|--------|-------|-----------|
| 🔲 | [#40](https://github.com/ElioNeto/kora/issues/40) | Sistema de Partículas 2D (CPU e GPU) |
| 🔲 | [#41](https://github.com/ElioNeto/kora/issues/41) | Iluminação 2D Dinâmica com sombras |
| 🔲 | [#42](https://github.com/ElioNeto/kora/issues/42) | Navegação 2D com Pathfinding para IA de inimigos e NPCs |
| 🔲 | [#43](https://github.com/ElioNeto/kora/issues/43) | Pseudo-3D e Paralaxe para profundidade em cenas 2D |
| 🔲 | [#39](https://github.com/ElioNeto/kora/issues/39) | Esqueletos 2D (Skeleton2D) com animação esquelética e IK |
| 🔲 | [#24](https://github.com/ElioNeto/kora/issues/24) | Suporte a Shaders (efeitos visuais via GPU) |
| 🔲 | [#44](https://github.com/ElioNeto/kora/issues/44) | Shader Baker — Pré-compilação de shaders no build Android |
| 🔲 | [#45](https://github.com/ElioNeto/kora/issues/45) | Delta Patching — Patches de atualização incrementais para Android |
| 🔲 | [#31](https://github.com/ElioNeto/kora/issues/31) | Networking — Multiplayer, lobbies e salvamento em nuvem |

---

## Legenda

| Ícone | Significado |
|-------|-------------|
| 🔲 | Não iniciado |
| 🔄 | Em progresso |
| ✅ | Concluído |

---

## Sobre a Kora

**Kora** é uma engine 2D focada exclusivamente em jogos Android, usando **KScript** como linguagem de programação nativa. O objetivo é oferecer uma experiência de desenvolvimento simples e direta — do editor ao APK — sem precisar configurar SDKs externos ou lidar com complexidade desnecessária.

- 🔗 [Issues abertas](https://github.com/ElioNeto/kora/issues)
- 📋 [Milestones](https://github.com/ElioNeto/kora/milestones)
- 📖 [Documentação](https://github.com/ElioNeto/kora/tree/main/docs)
