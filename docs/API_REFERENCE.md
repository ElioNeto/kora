# API Reference - KScript Engine

Este documento lista todas as APIs disponíveis na engine Kora.

---

## Input

### Métodos

```kscript
// Pressão única (frame único)
Input.pressed(key: string): bool

// Tecla segurada
Input.down(key: string): bool

// Tecla liberada (frame único)
Input.released(key: string): bool

// Eixo do teclado/joystick
Input.axisX(): float  // -1.0 a 1.0
Input.axisY(): float  // -1.0 a 1.0

// Mouse
Input.mouseLeftDown(): bool
Input.mouseRightDown(): bool
Input.mouseMiddleDown(): bool
Input.mousePos(): Vector2
Input.mouseDelta(): Vector2

// Touch (mobile)
Input.touchCount(): int
Input.touchPos(index: int): Vector2
Input.touchDelta(index: int): Vector2

// Gamepad
Input.gamepadConnected(index: int): bool
Input.gamepadAxis(gamepad: int, axis: int): float
Input.gamepadButton(gamepad: int, button: int): bool
```

### Keys String

```
ArrowUp, ArrowDown, ArrowLeft, ArrowRight
Space, Enter, Escape, Tab
KeyA-KeyZ (A-Z)
Num0-Num9 (0-9)
F1-F12
```

---

## Physics

### Propriedades

```kscript
// Gravidade global
this.gravity: float  // pixels/segundo²

// Velocidade
this.velocity: Vector2

// Aceleração
this.acceleration: Vector2

// Colisão
this.solid: bool       // é solido?
this.trigger: bool     // trigger?
this.sensors: bool     // usar sensores?
```

### Métodos

```kscript
// Aplicar força (contínua)
this.applyForce(dir: Vector2, force: float)

// Aplicar impulso (instantâneo)
this.applyImpulse(dir: Vector2, impulse: float)

// Aplicar rotação
this.applyTorque(angle: float)

// Raycast (detecção de ray)
this.raycast(dir: Vector2, distance: float): RaycastResult | null

// Circlecast
this.circlecast(center: Vector2, radius: float): RaycastResult | null

// Colisão AABB
this.getOverlaps(): Array<Entity>
this.getCollisions(): Array<Entity>

// Chão/parede
this.onGround(): bool
this.onWall(): bool

// Gravidade engine-level
this.physics.gravity(): float
this.physics.setGravity(g: float)
```

### RaycastResult

```kscript
{
  entity: Entity,
  pos: Vector2,
  normal: Vector2,
  distance: float
}
```

---

## Transform

### Propriedades

```kscript
// Posição
this.pos: Vector2
this.x: float
this.y: float

// Tamanho
this.size: Vector2
this.width: float
this.height: float

// Rotação
this.rotation: float  // graus
this.rotationSpeed: float

// Escala
this.scale: Vector2
this.scaleX: float
this.scaleY: float

// Atributos
this.visible: bool
this.solid: bool
this.active: bool
this.locked: bool
```

### Métodos

```kscript
// Move para posição
this.moveTo(pos: Vector2, duration: float, easing?: Easing)

// Rotate para ângulo
this.rotateTo(angle: float, duration: float)

// Scale para tamanho
this.scaleTo(size: Vector2, duration: float)

// Snap (arredondar posição)
this.snapToGrid(gridSize: Vector2)

// Centrar/alinhar
this.centerIn(rect: Rect)
this.alignTo(target: Entity, axis: string)

// Atributo global
this.setGlobalProp(key: string, value: any)
this.getGlobalProp(key: string): any
```

---

## Sprite/Gfx

### Propriedades

```kscript
// Sprite
this.sprite: Image | null
this.animation: string

// Cor/alpha
this.color: Color
this.alpha: float  // 0.0 - 1.0

// Flip
this.flipX: bool
this.flipY: bool

// Pivot
this.pivot: Vector2
this.pivotX: float
this.pivotY: float

// Layer
this.zIndex: int
this.layer: string

// Visual
this.drawOrder: int
this.debugMode: bool
```

### Métodos

```kscript
// Animação
this.animPlay(name: string)
this.animStop()
this.animFade(name: string, duration: float)
this.animIsPlaying(): bool

// Setar sprite
this.setSprite(asset: Image)
this.loadSprite(path: string)

// Frame
this.currentFrame(): int
this.setFrame(index: int)
this.getTotalFrames(): int

// Tamanho de sprite
this.getSpriteSize(): Vector2
```

### Easing Functions

```kscript
// Predefinidos
Easing.Linear
Easing.Quad.In
Easing.Quad.Out
Easing.Quad.InOut
Easing.Cubic.In
Easing.Cubic.Out
Easing.Cubic.InOut
Easing.Quart.In
Easing.Quart.Out
Easing.Quart.InOut
Easing.Quint.In
Easing.Quint.Out
Easing.Quint.InOut
Easing.Sine.In
Easing.Sine.Out
Easing.Sine.InOut
Easing.Expo.In
Easing.Expo.Out
Easing.Expo.InOut
Easing.Circ.In
Easing.Circ.Out
Easing.Circ.InOut
Easing.Back.In
Easing.Back.Out
Easing.Back.InOut
Easing.Bounce.In
Easing.Bounce.Out
Easing.Bounce.InOut
Easing.Elastic.In
Easing.Elastic.Out
Easing.Elastic.InOut
```

---

## Tween

### Métodos

```kscript
// Tween para destino
await tween(obj: any, props: Object, duration: float, easing: Easing = Easing.Linear)

// Tween com loop
loopTween(obj: any, props: Object, duration: float, loopCount: int, easing: Easing)

// Parar tween
tweenStop(obj: any, propName: string)

// Todos tweens parados
tweenStopAll(obj: any)
```

### Exemplos

```kscript
// Move em 1 segundo
await tween(this, { x: 100 }, 1.0)

// Fade out
await tween(this, { alpha: 0.0 }, 0.5)

// Rotate 360
await tween(this, { rotation: 360 }, 2.0)

// Multi-prop
await all(
  tween(this, { x: 100 }, 1.0, Easing.Quad.Out),
  tween(this, { y: 100 }, 1.0, Easing.Quad.Out)
)
```

---

## Asset Loader

### Métodos Estáticos

```kscript
// Carregar asset síncrono
Asset.load(path: string): any

// Cache de asset
Asset.cache(key: string, path: string): any

// Carregar múltiplos
Asset.all(...paths: Array<string>): Promise<Array<any>>

// Carregar async (Promise)
Asset.loadAsync(path: string): Promise<any>

// Verificar carregamento
Asset.loaded(path: string): bool

// Limpar cache
Asset.clear(key: string)
Asset.clearAll()

// Listar assets
Asset.list(): Array<string>

// Preload (pré-carrega)
Asset.preload(paths: Array<string>): Promise<void>
```

### Tipos

```kscript
// Imagem
Asset.load("player.png")  // Image

// Audio
Asset.load("jump.wav")    // AudioBuffer

// Tilemap
Asset.load("level.json")  // TilemapData

// Tileset
Asset.load("tiles.png")   // TilesetImage
```

---

## Audio

### Métodos

```kscript
// SFX Play
Audio.play(sound: string): AudioInstance

// Stop
Audio.stop(sound: string)
Audio.stopAll()

// Pause/Resume
Audio.pause()
Audio.resume()

// Volume
Audio.volume: float  // 0.0 - 1.0
Audio.setVolume(level: float)

// Loop
Audio.playLoop(bgm: string)
Audio.stopLoop()

// Fade
Audio.fadeOut(sound: string, duration: float)
Audio.fadeTo(sound: string, target: float, duration: float)

// Canais
Audio.playChannel(name: string, sound: string, channel: int)
Audio.setChannelVolume(channel: int, vol: float)
```

### AudioInstance

```kscript
{
  current: float,      // posição atual
  total: float,        // duração total
  loop: bool,
  speed: float         // speed de playback
}
```

---

## Camera

### Métodos

```kscript
// Camera principal
Camera.main(): Camera

// Seguir entidade
cam.follow(entity: Entity, speed: float)
cam.stopFollow()

// Movimentar
cam.move(pos: Vector2, duration: float)
cam.shake(amplitude: float, duration: float)

// Zoom
cam.zoom(target: float, duration: float)
cam.resetZoom()

// Filtros
cam.filter(color: Color, alpha: float)
cam.clearFilter()

//Viewport
cam.setViewport(x: float, y: float, w: float, h: float)
```

### Propriedades

```kscript
cam.x: float
cam.y: float
cam.zoom: float
cam.width: float
cam.height: float
```

---

## Tilemap

### Métodos

```kscript
// Criar tilemap
const tm = new Tilemap(image: Image, tileWidth: int, tileHeight: int)

// Adicionar layer
tm.addLayer(name: string, data: Array<Array<int>>)

// Set tile no grid
tm.setTile(x: int, y: int, layer: string, tileId: int)

// Get tile
tm.getTile(x: int, y: int, layer: string): int | null

// Collision
tm.isSolid(x: int, y: int): bool
tm.getCollisionLayer(): string

// Draw
tm.draw(cam: Camera)

// Raycast
tm.raycast(start: Vector2, dir: Vector2): TileResult | null
```

### Tile Result

```kscript
{
  x: int,
  y: int,
  layer: string,
  tileId: int,
  normal: Vector2
}
```

---

## System/Browser

### System

```kscript
// Tempo
System.time: float    // segundos desde início
System.frame: int     // frame atual
System.fps: int       // frames por segundo
System.delta: float   // delta time entre frames

// Screen
System.width: float
System.height: float
System.devicePixelRatio: float

// Center
System.center(): Vector2
System.left(): float
System.right(): float
System.top(): float
System.bottom(): float

// Rand
random(min: float, max: float): float
randomVector2(min: Vector2, max: Vector2): Vector2
choice<T>(arr: Array<T>): T

// Quit
quit()

// Info
System.info(): Object
```

### Device Info

```string
System.info() retorna:
{
  platform: "Web",
  device: "Desktop",
  userAgent: string,
  lang: string,
  width: int,
  height: int,
  pixelRatio: float,
  touchSupport: bool
}
```

---

## Console/Log/Debug

### Logs

```kscript
log(msg: string)  // Info
warn(msg: string) // Warning
error(msg: string) // Error

// Com vars
log("Health: $hp, Position: $pos")
```

### Debug Draw

```kscript
// Rect
debug.drawRect(pos: Vector2, size: Vector2, color: Color, duration: float)

// Line
debug.drawLine(pos1: Vector2, pos2: Vector2, color: Color, duration: float)

// Circle
debug.drawCircle(pos: Vector2, radius: float, color: Color, duration: float)

// Text
debug.drawText(text: string, pos: Vector2, fontSize: float, color: Color)

// Clear
debug.clear()

// FPS
debug.printFrameRate()
```

### Color

```kscript
// Construtor
Color(r: int, g: int, b: int, a: int = 255)

// Nômades
Color.WHITE = Color(255, 255, 255, 255)
Color.BLACK = Color(0, 0, 0, 255)
Color.RED = Color(255, 0, 0, 255)
// ... e mais

// Hex
Color.fromHex("#FF0000")

// Convert para hex
color.toHex(): string
color.toRGBA(): Object
```

---

## Entity/Scene

### Gerenciamento

```kscript
// Criar
createEntity(name: string, pos: Vector2): Entity

// Buscar
Entity.get(name: string): Entity | null
Entity.getAll(type: string): Array<Entity>

// Destroy
destroy(entity: Entity)
destroyAsync(entity: Entity)
destroyAll(type: string)

// Query
Entity.query(criteria: Object): Array<Entity>
```

### Signals

```kscript
// Emitir
emit(obj: any, signal: string, ...args)

// Esperar
await signal(obj: any, signal: string): any

// Wait any
await any(signal1, signal2, ...)

// Wait all
await all(signal1, signal2, ...)
```

### Lifecycle

```kscript
// Hooks (em objetos)
onCreate(): void
onDestroy(): void
onUpdate(dt: float): void
onDraw(): void

// Eventos
onInput(key: string, action: string): void
onCollision(other: Entity, type: CollisionType): void
onHit(damage: int): void
onDeath(): void
onEnter(trigger: string): void
onExit(trigger: string): void
```

---

## Time/Task

### Wait

```kscript
// Tempo
await wait(seconds: float)

// Frames
await waitFrames(count: int)

// Sleep
await sleep(fn: Function, delay: float): Promise<...>
```

### Task

```kscript
// Spawn task
spawn(fn: Function): Task

// Cancel
cancel(task: Task)

// Check status
task.isCancelled(): bool
task.isDone(): bool
```

---

## Utils

### String

```kscript
"string".length
"string".upper()
"string".lower()
"string".trim()
"string".startsWith(prefix: string): bool
"string".endsWith(suffix: string): bool
"string".replace(old: string, new: string): string
"string".split(separator: string): Array<string>
"string".pad(width: int, char: string = " "): string
```

### Math

```kscript
Math.abs(x: float): float
Math.round(x: float): int
Math.floor(x: float): int
Math.ceil(x: float): int
Math.max(a: float, b: float): float
Math.min(a: float, b: float): float
Math.clamp(x: float, min: float, max: float): float
Math.lerp(a: float, b: float, t: float): float
Math.dist(p1: Vector2, p2: Vector2): float
Math.angle(v1: Vector2, v2: Vector2): float

// Lerp manual
lerp(start: float, end: float, t: float): float
```

### Vector2

```kscript
// Criação
Vector2(x: float, y: float)
Vector2.fromAngle(angle: float, length: float)

// Atributos
v.x, v.y
v.length, v.lengthSquared()

// Operations
v + other
v - other
v * scalar
v / scalar
v * other  // dot product
-v        // negate

// Methods
v.normalize(): Vector2
v.clampLength(min: float, max: float): Vector2
v.rotate(angle: float): Vector2
v.angle(): float
v.perpendicular(): Vector2
v.project On(vec: Vector2): Vector2
```

---

## Tipos de Dados

### Array

```kscript
Array<T>

// Create
arr = Array<int>[1, 2, 3]
arr = Array<string>()
arr = Array<int>.repeat(10, 0)

// Methods
arr.length
arr.push(item: T): void
arr.pop(): T
arr.remove(index: int): T
arr.includes(item: T): bool
arr.find(fn: (T) -> bool): T | null
arr.map(fn: (T) -> U): Array<U>
arr.filter(fn: (T) -> bool): Array<T>
arr.reduce(fn: (acc: U, item: T) -> U, init: U): U
arr.sort(fn: (a: T, b: T) -> int)
arr.reverse()
arr.clear()
arr.first(): T
arr.last(): T
```

### Map

```kscript
Map<K, V>

// Create
m = Map<string, int>{"a": 1, "b": 2}
m = Map<K, V>()

// Methods
m.size
m.has(key: K): bool
m.get(key: K): V | null
m.set(key: K, value: V): void
m.delete(key: K): bool
m.keys(): Array<K>
m.values(): Array<V>
m.entries(): Array<[K, V]>
m.clear()
```

---

## Colisão Types

```kscript
enum CollisionType {
  Collide,   // colisão física (não atravessa)
  Overlap,   // sobreposição (atravessa)
  Enter,     // começou a sobrepor
  Exit       // terminou sobrepor
}
```

---

## Constants

```kscript
CollisionType.Collide
CollisionType.Overlap
CollisionType.Enter
CollisionType.Exit

Easing.Linear      // Linear
Easing.Quad.In     // Quadratic in
Easing.Quad.Out    // Quadratic out
Easing.Quad.InOut  // Quadratic in-out
// ... mais easing

Direction.Up
Direction.Down
Direction.Left
Direction.Right
```

---

**Referência completa da engine Kora** - para uso em scripts KScript.
