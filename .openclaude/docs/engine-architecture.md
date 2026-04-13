# Knowledge Base: Arquitetura do Kora Engine

Este documento descreve a arquitetura interna do runtime do Kora Engine. **Leia antes de modificar qualquer módulo em `core/` ou a integração com Ebiten.**

---

## O Game Loop Ebiten

O Ebiten é o coeur do runtime. Ele gerencia a janela, o rendering e o loop de jogo.

```
 Ebiten.Run(game)
       │
       ├────► Update()   ← chamado a 60 TPS fixo (1/60s por tick)
       │              Responsabilidade: lógica do jogo, física, input, async
       │
       └────► Draw()     ← chamado a cada frame (desacoplado do TPS)
                     Responsabilidade: APENAS rendering, NUNCA lógica
```

### Regras absolutas do game loop

1. **`Update()` é síncrono e deve retornar em < 1ms.** Qualquer operação lenta (I/O, rede, assets) deve ser delegada ao `core/async/scheduler`.
2. **`Draw()` é read-only.** Nunca modificar estado do jogo dentro de `Draw()`. Apenas leitura de estado + render.
3. **Nunca usar `sync.Mutex` no caminho crítico** de `Update()` ou `Draw()` — pode causar deadlock com o scheduler interno do Ebiten.
4. **`runtime.GC()` explícito é proibido** no game loop — o GC do Go já está afinado para Ebiten.
5. **`dt = 1.0/60.0`** é o delta time fixo. Para debug de TPS real: `ebiten.ActualTPS()`.

---

## Módulo `core/async/` — Scheduler de Corrotinas

O scheduler é o único mecanismo seguro para operações assíncronas dentro do jogo.

### API do Scheduler (interna Go)
```go
// Agendar uma goroutine gerenciada
koraAsync.Spawn(func() { ... })

// Cancelar uma task (retorna token para cancel)
task := koraAsync.Spawn(fn)
koraAsync.Cancel(task)

// Sleep sem bloquear o game loop
koraAsync.Wait(120) // 120 ticks = 2 segundos a 60 TPS
```

### ⚠️ Dívida técnica (DEBT-002)
O scheduler **não tem limite** de goroutines concorrentes. Em jogos com muitas entidades, `spawn` sem `cancel()` correspondente pode causar **memory leak**. Regra: todo `spawn` com loop infinito DEVE ter condição de saída ou `cancel()` explícito.

---

## Módulo `core/physics/` — Física AABB

### O que AABB faz (e o que não faz)
- **Faz:** Detecção de colisão entre retângulos axis-aligned (sem rotação)
- **Não faz:** Física com rotação, colisões circulares, corpo rígido completo

> ⚠️ **Dívida técnica (DEBT-005):** O campo `rotation` das entidades é visual apenas. O colisor sempre usa o bounding box axis-aligned, independente da rotação visual. Não implementar jogo que dependa de colisão rotacionada.

### Estrutura do colisor
```go
type AABB struct {
    X, Y float64   // posição (centro ou canto superior esquerdo? verificar impl)
    W, H float64   // dimensões
}

// Tipos de colisor
const (
    ColliderSolid   // bloqueia movimento
    ColliderTrigger // detecta overlap sem bloquear
)
```

### Gravíade e velocidade
Física aplicada por tick:
```
nova_vel_y = vel_y + gravidade * dt
nova_pos_y = pos_y + nova_vel_y * dt
```
A gravidade é configurável por cena via `Physics.gravity`.

---

## Módulo `core/scene/` — Gerenciamento de Cenas

### Tipos de transição

| Tipo | Comportamento |
|------|---------------|
| `Scene.load(name)` | Substitui cena atual; destrói todas as entidades |
| `Scene.reload()` | Reinicia a cena atual do zero |
| `Scene.additive(name)` | Carrega cena em paralelo sem destruir a atual |

### Ciclo de vida de uma cena
```
load("NomeDaCena")
    │
    ├─► Chama onDestroy() em todas as entidades da cena anterior
    ├─► Cancela todas as tasks do scheduler associadas à cena anterior
    ├─► Limpa o EntityManager
    ├─► Instancia entidades da nova cena a partir do .kora.json
    └─► Chama create() em todas as novas entidades
```

> **Atenção:** `additive` não cancela tasks da cena base. Garanta `cancel()` manual ao descarregar cenas aditivas.

---

## Módulo `core/render/` — Renderer 2D

### Coordenadas do mundo vs. tela
- **Coordenadas do mundo:** (0,0) = centro da cena, Y aumenta para baixo
- **Resolução lógica padrão:** 360×640 (portrait Android)
- **Conversão:** `worldToScreen(wx, wy, cam)` aplica zoom e offset da câmera

### Pipeline de render por frame
```
Draw(screen *ebiten.Image)
    │
    ├─► Limpar background
    ├─► Ordenar entidades por Z-order (layer)
    ├─► Para cada entidade visível:
    │       └─► DrawImage com transformação de câmera
    └─► UI e overlays (layer mais alto)
```

### Regra de rendering
- `*ebiten.Image` é o único tipo de buffer aceito
- Nunca usar `image.RGBA` diretamente para render final
- Sprites são carregados via `core/asset/` com cache automático

---

## Módulo `core/input/` — Input e VirtualPad

### Fontes de input suportadas
- **Teclado:** Via `ebiten.IsKeyPressed(key)`
- **Touch:** Gestos tap, swipe, multitouch
- **VirtualPad:** D-pad virtual renderizado na tela (padrão Android)
- **Gamepad:** Via Ebiten gamepad API

### VirtualPad
O VirtualPad é um componente renderizado que simula controles na tela touch. Sua configuração (tamanho, posição, botões) é definida no `.kora.json` da cena. **Não hardcodar posições do VirtualPad** — usar as coordenadas lógicas 360×640.

---

## Editor Visual — Integração com o Runtime

O editor é **completamente desacoplado** do runtime. A comunicação acontece apenas via formato `.kora.json`:

```
Editor (HTML/JS)
     │
     └─► salva .kora.json
                  │
                  └─► Runtime Go lê .kora.json
                               │
                               └─► Instancia entidades + executa KScript compilado
```

### ⚠️ Dívida técnica (DEBT-003)
A serialização em `editor/serializer.js` é **síncrona** para cenas grandes (>500 entidades) e pode travar a UI do browser. Não adicionar operações pesadas síncronas neste arquivo.

### Backward compatibility do `.kora.json`
O formato `.kora.json` deve ser **sempre backward-compatible**. Novos campos devem ter valores padrão. O campo `"version"` no meta identifica a versão do schema:
- `version: 1` — formato atual
- Ao incrementar, adicionar migration no serializer

---

## Android Export Pipeline

```
bash build.sh debug|release
    │
    ├─► gomobile bind -target android ./cmd/kora
    ├─► Gera .aar (Android Archive)
    ├─► Gradle assembles APK/AAB
    └─► debug: APK não assinado | release: AAB assinado com keystore
```

### Requisitos obrigatórios para o build Android funcionar
1. Go 1.22+ instalado
2. `gomobile` instalado: `go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init`
3. Android SDK + NDK na variável `ANDROID_HOME`
4. Para release: `android/signing.properties` configurado (NÃO commitar)

### Target SDK
- **Target:** Android 14 (API 34)
- **Min:** Android 7 (API 24)
- **ABI:** `arm64-v8a` (principal), `x86_64` (emulador)

---

## Dependências Externas Autorizadas

| Dependência | Versão | Motivo |
|-------------|--------|--------|
| `github.com/hajimehoshi/ebiten/v2` | v2.7.5 | Renderer, áudio, input, game loop |
| `github.com/ebitengine/gomobile` | indireta | Cross-compilation Android |
| `github.com/ebitengine/oto/v3` | indireta | Áudio (dep do Ebiten) |
| `golang.org/x/image` | indireta | Imagens (dep do Ebiten) |

**Regra:** Novas dependências diretas precisam ser Go puro (sem cgo). Qualquer adicional deve ser discutido e adicionado ao `go.mod` com justificativa no PR.
