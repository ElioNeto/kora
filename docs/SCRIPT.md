# KScript — Documentação da Linguagem

KScript é uma linguage de script estática, tipada, compilada para Go antes do build do APK. Não há VM — código KScript é transpilado para Go nativo.

## Visão Geral

KScript foi criado especificamente para desenvolvimento de games 2D. Sua sintaxe é inspirada no TypeScript, com foco em:

- **Tipagem estática** — erros detectados em tempo de compilação
- **Async/await nativo** — corrotinas para lógica não-blocking
- **Sistema de sinais** — eventos e callbacks reativos
- **Compilação AOT** — performance nativa, sem overhead de VM

## Exemplo Básico

```kscript
object Player {
  // Propriedades
  speed: float = 180.0
  hp: int = 5
  velocity: Vector2 = Vector2(0, 0)

  // Construtor
  async create() {
    // Inicialização assíncrona
    await wait(0.5)
    this.hp = 10
    spawn("ParticleEffect", this.pos + Vector2(0, -20))
  }

  // Atualização por frame
  update(dt: float) {
    // Movimento com input
    const ax = Input.axisX()
    const ay = Input.axisY()

    this.velocity.x = ax * this.speed
    this.velocity.y = ay * this.speed

    this.pos += this.velocity * dt
  }

  // Colisão
  onCollision(other: Entity, type: CollisionType) {
    switch (type) {
      case CollisionType.Collide:
        takeDamage(1)
      case CollisionType.Overlap:
        collect(other)
    }
  }

  // Métodos privados
  fn takeDamage(amount: int) {
    this.hp -= amount
    await flashRed(0.1)
    if (this.hp <= 0) {
      emit(this, "dead")
      destroyAsync(this)
    }
  }
}
```

---

## Sistema de Tipos

### Tipos Primitivos

| Tipo | Descrição | Exemplo |
|------|-----------|---------|
| `bool` | Booleano | `true`, `false` |
| `int` | Inteiro (32-bit) | `42`, `-10`, `0` |
| `float` | Número de ponto flutuante | `3.14`, `-0.5`, `1.0e-3` |
| `string` | String UTF-8 | `"hello"`, `"valor"` |
| `void` | Sem valor | `return valor` |

### Tipos Estruturados

| Tipo | Descrição | Exemplo |
|------|-----------|---------|
| `Vector2` | Vetor 2D (x, y) | `Vector2(10, 20)` |
| `Vector3` | Vetor 3D (x, y, z) | `Vector3(0, 1, 0)` |
| `Color` | Cor RGBA | `Color(255, 0, 0, 255)` |
| `Rect` | Retângulo | `Rect(0, 0, 100, 100)` |
| `Array<T>` | Array genérico | `Array<int>[1, 2, 3]` |
| `Map<K, V>` | Map genérico | `Map<string, Entity>` |

### Tipos de Referência

```kscript
// Declaração de tipo nomeado
type Health = int
type HealthPotion = object {
  amount: int
}

// Tipagem opcional
var maybeValue: string? = null
if (maybeValue != null) {
  log(maybeValue.value) // acesso seguro
}
```

### Enums

```kscript
enum Direction {
  Up,
  Down,
  Left,
  Right
}

enum CollisionType {
  Collide,
  Overlap,
  Enter,
  Exit
}

// Uso
var dir: Direction = Direction.Right
if (dir == Direction.Up) {
  // lógica
}
```

---

## Estrutura Básica

### Objetos (Classes)

```kscript
object Player {
  // Campos
  x: float = 0
  y: float = 0

  // Lifecycle hooks
  async create() { /* inicia */ }
  void update(dt: float) { /* por frame */ }
  void draw() { /* render */ }
  void onDestroy() { /* cleanup */ }

  // Eventos
  onInput(key: string, action: string) { /* input */ }
  onCollision(other: Entity, type: CollisionType) { /* colisão */ }
}
```

### Components

```kscript
component Healthbar {
  target: Entity

  create() {
    this.target = null
  }

  update(dt: float) {
    if (this.target != null) {
      this.pos = this.target.pos + Vector2(0, 50)
      this.width = this.target.width
    }
  }
}
```

### Scripts Globais

```kscript
// Main.kscript (ponto de entrada)
async main() {
  createPlayer()
  await wait(10.0)
  gameOver()
}

fn createPlayer() {
  createEntity("Player", Vector2(0, 0))
}
```

---

## Variáveis e Constantes

```kscript
// Variáveis
var x: int = 10
var nome: string = "Kora"

// Variáveis locais (inferência de tipo)
var count = 5 // tipo: int
var speed = 180.5 // tipo: float

// Constantes
const MAX_HEALTH: int = 100
const PI: float = 3.14159

// Reatribuição
x = 20 // OK
const y: int = 30
y = 40 // ERRO!

// Tipo implícito
var arr = Array<string>["a", "b", "c"]
```

### Scoping

```kscript
var globalVar = 10

function outer() {
  var outerVar = 20

  function inner() {
    var innerVar = 30
    log(globalVar) // OK
    log(outerVar)  // OK
    log(innerVar)  // OK
  }
}
```

---

## Funções

### Declaração

```kscript
// Função básica
fn add(a: int, b: int): int {
  return a + b
}

// Função com body em uma linha
fn square(x: float): float => x * x

// Func arrow
const multiply = fn(a: int, b: int) {
  return a * b
}
```

### Parâmetros Opcionais e Default

```kscript
fn greet(name: string, greeting: string = "Hello") {
  log("$greeting, $name!")
}

greet("Maria") // "Hello, Maria!"
greet("João", "Oi") // "Oi, João!"
```

### Parâmetros Variádicos

```kscript
fn sum(...numbers: float): float {
  var total: float = 0
  for (n in numbers) {
    total += n
  }
  return total
}

sum(1, 2, 3) // 6
sum(1.5, 2.5, 3.5) // 7.5
```

---

## Controle de Fluxo

### Condicionais

```kscript
// If/else
if (health > 50) {
  log("Saúde boa")
} else if (health > 20) {
  log("Saúde média")
} else {
  log("Saúde crítica")
}

// When (switch moderno)
fn getDirectionName(dir: Direction): string {
  return when (dir) {
    Direction.Up => "Cima",
    Direction.Down => "Baixo",
    Direction.Left => "Esquerda",
    Direction.Right => "Direita",
  }
}

// Guard conditions
when (value) {
  case if (v) v > 100 => "Muito grande",
  case if (v) v > 50 => "Grande",
  default => "Pequeno"
}
```

### Loops

```kscript
// For tradicional
for (var i: int = 0; i < 10; i++) {
  log(i)
}

// For-in (Array)
for (val in [1, 2, 3, 4, 5]) {
  log(val)
}

// For-of (Map)
var map = Map<string, int>{"a": 1, "b": 2}
for (k, v in map) {
  log("$k: $v")
}

// Loop infinito com break
var pos: Vector2 = Vector2(0, 0)
while (true) {
  pos.x += 1
  if (pos.x > 100) break
}

// Continue
for (i in [1, 2, 3, 4, 5]) {
  if (i % 2 == 0) continue // pula pares
  log(i) // 1, 3, 5
}

// Loops com delay (async)
for (i in 0..10) { // 0 up to 9
  log(i)
  await wait(0.1)
}

for (i in 0..<10) { // 0 inclusive, 10 exclusive (igual acima)
  log(i)
}
```

---

## Expressões

### String Interpolation

```kscript
var name = "Mundo"
log("Hello, $name!") // "Hello, Mundo!"

// Expressões
var x = 10
var y = 20
log("A soma é $x + $y = $x + $y") // "A soma é 10 + 20 = 30"

// Blocos
log("Total: $(x * 2 + y)") // "Total: 40"
```

### Operadores

| Operador | Descrição | Exemplo |
|----------|-----------|---------|
| `+ - * /` | Aritmética | `5 + 3` |
| `%` | Módulo | `10 % 3` |
| `**` | Potência | `2 ** 8` |
| `== !=` | Igualdade | `a == b` |
| `< > <= >=` | Comparação | `a < b` |
| `&& ||` | Lógico | `a && b` |
| `!` | Negar | `!a` |
| `&&` | Bitwise AND | `a & b` |
| `\|\|` | Bitwise OR | `a | b` |
| `<< >>` | Deslocamento | `x << 2` |
| `?.` | Null-safe | `obj?.prop` |
| `??` | Null-coalesce | `a ?? b` |

### Lambdas e Arrow Functions

```kscript
// Arrow implícita
const squares = [1, 2, 3].map(x => x * x) // [1, 4, 9]

// Com blocos
const filtered = [1, 2, 3, 4, 5]
  .filter(x => {
    return x > 2
  })

// Com parêntesis (multi-parâmetro)
const add = (a: int, b: int) => a + b

// Closure
fn makeMultiplier(multiplier: float) {
  return fn(x: float) {
    return x * multiplier
  }
}

const double = makeMultiplier(2)
log(double(5)) // 10
```

---

## Async/Await

### Espera e Corrotinas

```kscript
// Espera tempo
async function delayedTask() {
  log("Iniciando...")
  await wait(2.0) // espera 2 segundos
  log("Feito!")
}

// Espera frames
async function frameTask() {
  for (var i: int = 0; i < 60; i++) {
    log("Frame $i")
    await waitFrames(1) // espera 1 frame
  }
}

// Espera sinal
async function waitForEnemyDefeat(enemy: Entity) {
  await signal(enemy, "defeated")
  log("Inimigo derrotado!")
}

// Race (qualquer um completa primeiro)
async function raceExample() {
  const task1 = fetchFromServer()
  const task2 = loadFromCache()
  const result = await race(task1, task2)
  log("Resultado: $"result)
}

// All (todos devem completar)
async function loadEverything() {
  const [sprites, audio, level] = await all(
    loadSprites(),
    loadAudio(),
    loadLevel()
  )
  startGame(sprites, audio, level)
}

// Cancelar task
var task: Task | null = null

async function startAnimation() {
  task = spawn(async () => {
    for (var i: int = 0; i < 100; i++) {
      await wait(0.016)
      animate(i)
    }
  })
}

function stopAnimation() {
  if (task != null) {
    cancel(task!)
  }
}
```

### Async em Métodos

```kscript
object Player {
  async create() {
    await wait(0.5)
    this.loadSprites()
    await this.spawnParticles()
  }

  async loadSprites() {
    this.sprite = await Asset.load("player.png")
  }
}
```

---

## Sistema de Sinais

Sinais são eventos reativos entre objetos e entidades.

### Emitir Sinais

```kscript
// Emitir sinal
emit(this, "hit", damage)
emit(this, "death")
```

### Escutar Sinais

```kscript
// On-sign (decorator/método)
object Enemy {
  onHit(amount: int) {
    this.health -= amount
    emit(this, "damageTaken", amount)
  }

  onDeath() {
    emit(this, "dead")
    dropLoot()
  }
}

// Esperar sinal
async function waitUntilDead(enemy: Enemy) {
  await signal(enemy, "dead")
  log("Inimigo morreu!")
}

// Sinais múltiplos
async function waitForEvents() {
  await any(signal(enemy, "dead"), signal(enemy, "flee"))
  log("Inimigo eliminado")
}
```

### Lista de Sinais Padrão

| Sinal | Quando |
|-------|--------|
| `create` | Entidade é criada |
| `destroy` | Entidade é destruída |
| `hit` | Entidade recebe dano |
| `dead` | Entidade morre |
| `enter` | Entra em trigger |
| `exit` | Sai de trigger |
| `overlap` | Começa sobreposição |

---

## API do Engine

### Input

```kscript
// Pressão de tecla
if (Input.pressed("Space")) {
  jump()
}

// Tecla segurada
if (Input.down("ArrowRight")) {
  moveRight()
}

// Tecla liberada
if (Input.released("E")) {
  interact()
}

// Eixo (joystick/gamepad)
const axisX = Input.axisX()
const axisY = Input.axisY()

// Mouse
if (Input.mouseDown(MouseButton.Left)) {
  const pos = Input.mousePos()
  log("Posição: $pos")
}

// Touch (mobile)
Input.touchCount() // número de toques
Input.touchPos(0) // posição do primeiro toque
```

### Entity e Transform

```kscript
// Criar entidade
const player = createEntity("Player", Vector2(100, 200))

// Acessar this (no método do objeto)
this.pos // posição
this.size // tamanho
this.rotation // rotação em graus
this.visible // visibilidade

// Atributos
this.health = 100
this.speed = 180

// Movimento
const vel: Vector2 = this.velocity * dt
this.pos += vel

// Colisão
const overlaps: Array<Entity> = this.getOverlaps()
const collision = this.detectCollision(other)
```

### Sprite e Render

```kscript
// Definir sprite
this.sprite = Asset.load("player.png")

// Animação
animationPlayer.play("idle")
animationPlayer.play("run")
animationPlayer.fadeTo("jump", 0.2)

// Tamanho
this.width = 32
this.height = 32

// Cor/efet
this.alpha = 1.0
this.color = Color(255, 255, 255, 255)

// Flip
this.flipX = true
this.flipY = false
```

### Tween/Animação

```kscript
// Tween de propriedades
await tween(this, { x: 100 }, 1.0) // move em 1s
await tween(this, { alpha: 0.0 }, 0.5) // fade out
await tween(this, { rotation: 360 }, 2.0) // rotate

// Com easing
await tween(this, { y: 100 }, 1.5, Easing.OutQuad)

// Sequencia
await all(
  tween(this, { x: 100 }, 1.0),
  tween(this, { y: 100 }, 1.0)
)

// Loop
loop(async () => {
  await tween(this, { scale: 1.2 }, 0.3)
  await tween(this, { scale: 1.0 }, 0.3)
}, "pulse")
```

### Física

```kscript
// Gravidade
this.gravity = 980 // pixels/segundo²

// Velocidade
this.velocity = Vector2(100, 10)

// Colisão
this.solid = true // é solido?
this.trigger = true // é trigger?

// Aplicar força
this.applyForce(Vector2(0, -500))
this.applyImpulse(Vector2(100, 0))

// Detecção
const hit = this.raycast(Vector2(0, -1), 50)
if (hit) {
  log("Colidiu em: $"hit.pos)
}
```

### Asset Loader

```kscript
// Carregar asset síncrono
const sprite = Asset.load("player.png")

// Asset com cache
const bgm = Asset.cache("music.wav", "bgm")

// Carregar múltiplos
const [sprites, audio] = await Asset.all(
  "player_idle.png",
  "player_run.png",
  "music.wav"
)

// Verificar carregamento
if (_asset.loaded("player.png")) {
  draw(_asset("player.png"))
}
```

### Log e Debug

```kscript
// Mensagens
log("Jogo iniciado")
log("Player hp: $hp") // com interpolação

// Erros
error("Falta de saúde!")

// Debug
debug.drawRect(pos, size, Color(0, 255, 0, 128))
debug.drawLine(start, end, Color(255, 0, 0, 255))
debug.drawCircle(center, radius, Color(0, 0, 255, 128))
debug.printFrameRate()

// Console do editor
console.log("Debug info")
```

### Audio

```kscript
// SFX
Audio.play("jump.wav")
Audio.stop("jump.wav")
Audio.fadeOut("jump.wav", 0.5)

// Música
Audio.playLoop("bgm.wav")
Audio.pause()
Audio.resume()
Audio.volume = 0.5

// Canais
Audio.playChannel("sfx_jump", "jump.wav", 1)
```

### Camera

```kscript
// Câmera atual
const cam = Camera.main()

// Seguir entidade
cam.follow(this, 0.2) // 0.2s de delay

// Mover camera
cam.move(Vector2(100, 0), 1.0)

// Zoom
cam.zoom(1.5)
```

### System

```kscript
// Tempo
const dt = now().delta // delta time entre frames
const now = System.time // tempo em segundos

// Screen
const screenW = System.width
const screenH = System.height
const screenCenter = System.center

// Random
const x = random(0, 100)
const v = randomVector2(10, 20)
const choice = choice([1, 2, 3]) // aleatório

// Game Loop
if (Input.pressed("Escape")) {
  quit() // fecha jogo
}
```

---

## Exemplos Completos

### Player Controller

```kscript
object Player {
  runSpeed: float = 200
  jumpForce: float = -400
  gravity: float = 980
  canDoubleJump: bool = false

  async create() {
    this.gravity = this.physics.gravity()
  }

  update(dt: float) {
    // Movimento horizontal
    const move = Input.axisX()
    this.velocity.x = move * this.runSpeed

    // Pulo
    if (Input.justPressed("Space")) {
      if (this.onGround()) {
        this.velocity.y = this.jumpForce
        this.canDoubleJump = true
        Audio.play("jump.wav")
      } else if (this.canDoubleJump) {
        this.velocity.y = this.jumpForce * 0.8
        this.canDoubleJump = false
        Audio.play("jump.wav")
      }
    }

    // Rotação visual (apontar direção)
    if (move != 0) {
      this.flipX = move < 0
    }

    // Limite de tela
    if (this.pos.x < 0) this.pos.x = 0
    if (this.pos.x > System.width) this.pos.x = System.width
  }

  onInput(key: string, action: string) {
    if (key == "E" && action == "pressed") {
      emit(this, "interact")
    }
  }
}
```

### Inimigo AI Simples

```kscript
object Enemy {
  patrolSpeed: float = 50
  attackRange: float = 80
  hp: int = 3

  state: string = "patrol"
  patrolTarget: Vector2 | null = null

  async create() {
    this.patrol()
  }

  update(dt: float) {
    switch (this.state) {
      case "patrol":
        this.patrolUpdate(dt)
      case "chase":
        this.chaseUpdate(dt)
      case "attack":
        this.attackUpdate(dt)
    }

    // Reacender patrolling se player sair de visão
    if (this.state == "patrol" && !this.hasPlayerInSight()) {
      this.state = "idle"
    }
  }

  fn patrol() {
    this.patrolTarget = Vector2(
      this.pos.x + random(-100, 100),
      this.pos.y
    )
    this.state = "patrol"
  }

  fn patrolUpdate(dt: float) {
    if (this.patrolTarget != null) {
      const dir = this.patrolTarget - this.pos
      const dist = dir.length()

      if (dist < 5) {
        this.patrol()
        return
      }

      this.velocity = dir.normalized() * this.patrolSpeed
      this.pos += this.velocity * dt
    }
  }

  hasPlayerInSight(): bool {
    const player = Entity.get("Player")
    if (player == null) return false

    const dist = (player.pos - this.pos).length()
    return dist < this.attackRange
  }

  onHit(damage: int) {
    this.hp -= damage
    emit(this, "hit")

    // Flash
    await tween(this, { alpha: 0.3 }, 0.1)
    await tween(this, { alpha: 1.0 }, 0.1)

    if (this.hp <= 0) {
      emit(this, "dead")
      destroyAsync(this)
    }
  }
}
```

---

## Convenções e Boas Práticas

### Nomes

```kscript
// camelCase para propriedades e métodos
var playerHealth = 10
fn update() { }
fn takeDamage(amount) { }

// PascalCase para enums e types
enum Direction { Up, Down }
type Health = int

// snake_case para constantes
const MAX_PLAYERS = 4
const GRAVITY_Y = -980

// Prefixo this para propriedades de objeto
this.health // não: health (pode confundir com parâmetros)
```

### Organização

```kscript
// Ordem em objetos:
// 1. Campos públicos
// 2. Campos privados
// 3. Lifecycle hooks (create, update, destroy)
// 4. Métodos públicos
// 5. Métodos privados
// 6. Event handlers (onXxx)

object Player {
  // Campos
  health: int = 100
  speed: float = 180

  // Privados
  _state: string = "idle"
  _lastHitTime: float = 0

  // Lifecycle
  async create() { }
  update(dt: float) { }
  onDestroy() { }

  // Públicos
  fn takeDamage(amount: int) { }
  fn heal(amount: int) { }

  // Privados
  fn _updateState() { }

  // Events
  onHit(amount: int) { }
}
```

### Async

```kscript
// Use async para ops com wait
async function loadData() {
  await waitFrames(30) // carrega enquanto renderiza
  process(data)
}

// Prefira await sobre callbacks
// Antigo: onLoad(result)
// Novo: const result = await waitFor("load")

// Cancelle tasks pendentes
fn startLongTask() {
  if (this._task != null) {
    cancel(this._task!)
  }
  this._task = spawn(async () => {
    await wait(10)
  })
}
```

---

## Referências Rápidas

### Sintaxe

```kscript
// Declaração
var x: int = 10
const PI: float = 3.14

// Condicionais
if (cond) { }
switch (val) { case 1 => ... }
when (val) { case X => ... }

// Loops
for (i = 0; i < n; i++) { }
for (v in arr) { }
for (k, v in map) { }
while (true) { }
loop(async fn() { }, "name")

// Funções
fn name(args): type { }
fn name(args) => body  // arrow

// Async/await
await wait(1.0)
await all(task1, task2)
await race(task1, task2)

// Expressões
"string $var"  // interpolação
obj?.prop  // null-safe
a ?? b  // null-coalesce
```

### Métodos Comuns de Entity

```kscript
this.pos          // Vector2 - posição
this.size         // Vector2 - tamanho
this.rotation     // float - rotação em graus
this.velocity     // Vector2 - velocidade
this.acceleration // Vector2 - aceleração

this.create(child)          // cria filho
this.destroy()              // destrói entidade
this.destroyAsync()         // destrói async
this.getOverlaps()          // Array<Entity>
this.raycast(dir, dist)     // RaycastResult | null
this.emit(signal, data)     // emite sinal
signal(this, signal)        // espera sinal
```

---

## Glossário

| Termo | Definição |
|-------|-----------|
| **Entity** | Objeto no mundo do jogo (com posição, tamanho, etc) |
| **Component** | Sistema modular ligado a entidades |
| **Signal/Emit** | Sistema reativo de eventos |
| **Task** | Task async gerenciada pelo scheduler |
| **Tween** | Interpolação automática de propriedades |
| **KScript** | Linguagem compilada para Go |
| **AOT** | Ahead-Of-Time compilation (Go) |

---

**KScript** — compile para Go, rode nativo, desenvolva com TypeScript syntax.
