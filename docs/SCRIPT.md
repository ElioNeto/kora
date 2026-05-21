# KScript — A Referência Completa da Linguagem

> **Versão:** 0.1 (em desenvolvimento)
> **Compila para:** Go nativo (sem VM, sem interpretador)
> **Propósito:** Lógica de jogos 2D para o motor Kora

---

## Sumário

1. [Introdução](#1-introdução)
2. [Quick Start](#2-quick-start)
3. [Guia de Sintaxe](#3-guia-de-sintaxe)
4. [Sistema de Tipos](#4-sistema-de-tipos)
5. [Variáveis e Constantes](#5-variáveis-e-constantes)
6. [Operadores](#6-operadores)
7. [Controle de Fluxo](#7-controle-de-fluxo)
8. [Funções e Métodos](#8-funções-e-métodos)
9. [Objetos (Classes)](#9-objetos-classes)
10. [Programação Assíncrona](#10-programação-assíncrona)
11. [Sistema de Sinais](#11-sistema-de-sinais)
12. [API do Motor](#12-api-do-motor)
13. [Pipeline de Compilação](#13-pipeline-de-compilação)
14. [Padrões Comuns](#14-padrões-comuns)
15. [Solução de Problemas](#15-solução-de-problemas)
16. [Apêndices](#16-apêndices)

---

## 1. Introdução

### 1.1 O que é KScript?

KScript é uma **linguagem de script estaticamente tipada** projetada especificamente para o motor de jogos 2D **Kora**. Diferente de linguagens de script tradicionais (Lua, Python), KScript **não possui VM nem interpretador** — o código-fonte é compilado *ahead-of-time* (AOT) para Go nativo, fazendo parte do binário final do jogo.

```
┌──────────────┐     ┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌───────────┐
│ player.ks    │ ──► │ Lexer    │ ──► │ Parser   │ ──► │ Type Checker │ ──► │ Emitter   │
│ (KScript)    │     │ (tokens) │     │ (AST)    │     │ (semântica)  │     │ (Go code) │
└──────────────┘     └──────────┘     └──────────┘     └──────────────┘     └─────┬─────┘
                                                                                  │
                                                                          ┌───────▼───────┐
                                                                          │ go build      │
                                                                          │ → kora.apk    │
                                                                          └───────────────┘
```

### 1.2 Filosofia

| Princípio | Descrição |
|-----------|-----------|
| **Game-first** | Sintaxe e APIs pensadas para jogos 2D — Vector2, física, colisão, animação |
| **Async nativo** | `async`/`await` para lógica sequencial sem callbacks aninhados |
| **Type-safe** | Erros detectados em compilação, não em runtime |
| **Sem overhead** | Compilação AOT → Go nativo → performance comparável a C++ |
| **Sintaxe familiar** | Inspirada em TypeScript/Go para facilitar a adoção |

### 1.3 Hello World

```kscript
object Game {
  async create() {
    System.log("Olá, Kora!")
  }
}
```

Cada arquivo `.ks` contém uma ou mais declarações `object`. O motor Kora instancia automaticamente os objetos definidos e chama seus *lifecycle hooks* (`create`, `update`, `draw`, etc.).

> 💡 **Nota:** A extensão de arquivo é `.ks`. Não confundir com Kotlin Script (`.kts`).

---

## 2. Quick Start

### 2.1 Seu Primeiro Script

Crie `player.ks`:

```kscript
object Player {
  // Propriedades com tipo explícito e valor padrão
  var speed: float = 200.0
  var hp: int = 100
  var pos: Vec2 = Vec2(0, 0)

  // Chamado quando o objeto é criado
  async create() {
    System.log("Player criado!")
    await wait(0.5) // espera 0.5s
    this.hp = 100
  }

  // Chamado a cada frame (60 fps)
  update(dt: float) {
    const ax = Input.axisX()
    const ay = Input.axisY()
    this.pos.x += ax * this.speed * dt
    this.pos.y += ay * this.speed * dt
  }
}
```

### 2.2 Compilando e Executando

```bash
# Compilar um script .ks
kora build player.ks

# Executar o jogo (compila automaticamente se necessário)
kora run

# O código Go gerado fica em:
# gen/player.go
```

### 2.3 Estrutura de um Projeto

```
meu_jogo/
├── main.ks          # Ponto de entrada
├── player.ks        # Objetos do jogador
├── enemy.ks         # Objetos dos inimigos
├── hud.ks           # Interface
├── scenes/          # Cenas (opcional)
│   ├── level1.ks
│   └── menu.ks
└── kora.json        # Configuração do projeto
```

---

## 3. Guia de Sintaxe

### 3.1 Comentários

```kscript
// Comentário de linha — suportado

/* Comentário de bloco */  // ⚠️ Planejado - ainda não implementado
```

Apenas comentários de linha (`//`) são suportados na versão atual.

### 3.2 Identificadores

- Começam com letra (`a-z`, `A-Z`) ou underscore (`_`)
- Seguidos por letras, dígitos ou underscore
- **Case-sensitive**: `Player`, `player` e `PLAYER` são diferentes
- Convenção: `camelCase` para variáveis e métodos, `PascalCase` para objetos

```kscript
var playerName: string = "Herói"    // camelCase
object HealthBar { }                 // PascalCase
const MAX_ITEMS: int = 10           // SCREAMING_SNAKE_CASE para constantes
var _internalState: int = 0         // underscore para privado (convenção)
```

### 3.3 Terminadores de Declaração

KScript **não exige** ponto-e-vírgula (`;`). As declarações são separadas por **quebra de linha**. O ponto-e-vírgula é opcional e ignorado pelo parser quando presente.

```kscript
var x: int = 10   // OK
var y: int = 20;  // OK, ; opcional
```

### 3.4 Palavras Reservadas

```
object    func      async     await     return
if        else      for       while     break
continue  const     var       true      false
null      import    from      this      emit
```

> ⚠️ `for` é reconhecido pelo lexer mas sua sintaxe completa (`for (;;)`, `for-in`) **ainda não está implementada** no parser (v0.1).

### 3.5 Blocos e Escopo

Blocos são delimitados por chaves `{ }`. Cada bloco cria um novo escopo:

```kscript
object Player {
  var name: string = "Player" // escopo do objeto

  update(dt: float) {
    var x: int = 10           // escopo do método
    if (x > 5) {
      var y: int = 20         // escopo do if
      // x, y, name visíveis aqui
    }
    // x visível, y NÃO visível
  }
}
```

---

## 4. Sistema de Tipos

KScript possui **tipagem estática e forte**. O tipo de cada expressão é conhecido em tempo de compilação.

### 4.1 Tipos Primitivos

| Tipo    | Descrição               | Exemplos                     | Go equivalente |
|---------|-------------------------|------------------------------|----------------|
| `int`   | Inteiro 64-bit signed   | `42`, `-10`, `0`             | `int`          |
| `float` | Ponto flutuante 64-bit  | `3.14`, `-0.5`, `1e3`        | `float64`      |
| `bool`  | Booleano                | `true`, `false`              | `bool`         |
| `string`| String UTF-8 imutável   | `"hello"`, `"Kora"`          | `string`       |
| `void`  | Ausência de valor       | (usado como retorno vazio)   | —              |

**Detalhes importantes:**

```kscript
// Inteiros
var a: int = 42
var b = 10          // inferido como int

// Floats
var c: float = 3.14
var d = 2.5         // inferido como float
var e: float = 10   // OK: int pode ser usado como float (promoção)

// Booleanos
var active: bool = true
var done = false    // inferido como bool

// Strings
var name: string = "Kora Engine"
var empty = ""      // string vazia
```

### 4.2 Tipos Compostos (Motor)

| Tipo   | Descrição             | Exemplo de criação        | Campos         |
|--------|-----------------------|---------------------------|----------------|
| `Vec2` | Vetor 2D              | `Vec2(10, 20)`            | `.x`, `.y`     |
| `Color`| Cor RGBA              | `Color(1, 0, 0, 1)`       | `.r`, `.g`, `.b`, `.a` |

> ⚠️ `Vector2`, `Vector3`, `Rect` como tipos nomeados estão em planejamento.
> Na versão atual usa-se `Vec2`. A criação é feita via função construtora, não literal.

### 4.3 Tipos Array

```kscript
var arr: int[]      // array de ints

// Arrays são usados via APIs do motor, não com literais (⚠️ planejado)
var items: Entity[] = Scene.findByGroup("enemies")
```

Tipos array suportados:
- `int[]`
- `float[]`
- `string[]`
- `Entity[]`

### 4.4 Tipos Nomeados pelo Usuário

```kscript
// Cada objeto define um tipo
object Health {
  var amount: int = 100
}

object Player {
  var health: Health  // tipo Health
}
```

### 4.5 Tipos Especiais

| Tipo     | Descrição                    | Uso                          |
|----------|------------------------------|------------------------------|
| `Task`   | Representa uma operação async| Retorno de `wait()`, `race()`|
| `Entity` | Qualquer objeto no mundo     | `this`, parâmetros de evento |
| `any`    | Tipo coringa (interno)       | Resolução futura             |

### 4.6 Inferência de Tipos

Quando o tipo pode ser deduzido, a anotação é opcional:

```kscript
var x = 10        // int
var y = 3.14      // float
var name = "Kora" // string
var active = true // bool
```

A inferência funciona com expressões:

```kscript
var sum = a + b        // tipo deduzido de a + b
var result = calc()    // tipo deduzido do retorno de calc()
```

### 4.7 Type Aliases (🔲 Não implementado)

> Type aliases (`type Health = int`) estão planejados mas não implementados.

### 4.8 Enums (🔲 Não implementado)

> Enums (`enum Direction { Up, Down }`) estão planejados mas não implementados.
> Use constantes como alternativa:

```kscript
const DIR_UP: int = 0
const DIR_DOWN: int = 1
const DIR_LEFT: int = 2
const DIR_RIGHT: int = 3
```

### 4.9 Tipos Opcionais / Nullable (🔲 Não implementado)

> Tipos opcionais (`string?`, `int?`) e operadores null-safe (`?.`, `??`)
> estão planejados mas não implementados. Use `null` diretamente:

```kscript
var target: Entity = null
if (target != null) {
  // usa target
}
```

---

## 5. Variáveis e Constantes

### 5.1 Declaração com `var`

Variáveis declaradas com `var` são **mutáveis** (podem ser reatribuídas):

```kscript
var x: int = 10
x = 20  // OK

var y = 5  // inferido como int
y = y + 1  // OK
```

### 5.2 Declaração com `const`

Constantes são **imutáveis** — qualquer tentativa de reatribuição causa erro de compilação:

```kscript
const MAX_HP: int = 100
MAX_HP = 50  // ERRO: cannot assign to const "MAX_HP"

const PI = 3.14159  // inferido como float
```

### 5.3 Declaração com Tipo Explícito vs Inferido

```kscript
// Tipo explícito
var hp: int = 100

// Tipo inferido do valor inicial
var hp = 100  // int

// Tipo explícito sem valor inicial (inicializa com zero)
var name: string
name = "Kora"
```

### 5.4 Regras de Escopo

KScript usa escopo por bloco (`{ }`):

```kscript
object Player {
  var global: int = 0   // visível em todo o objeto

  update(dt: float) {
    var local: int = 5  // visível só neste método

    if (dt > 0.5) {
      var inner: int = 10  // visível só neste if
    }
    // inner NÃO existe aqui
  }
}
```

### 5.5 Sombramento (Shadowing)

Variáveis de escopos internos podem sombrear variáveis de escopos externos:

```kscript
var x: int = 10
if (true) {
  var x: int = 20  // sombreia a x externa
  // x == 20
}
// x == 10 (a externa)
```

---

## 6. Operadores

### 6.1 Tabela de Operadores

| Categoria   | Operador | Descrição              | Exemplo          |
|-------------|----------|------------------------|------------------|
| Aritméticos | `+`      | Adição / concatenação  | `a + b`          |
|             | `-`      | Subtração              | `a - b`          |
|             | `*`      | Multiplicação          | `a * b`          |
|             | `/`      | Divisão                | `a / b`          |
|             | `%`      | Módulo (resto)         | `a % b`          |
| Comparação  | `==`     | Igual a                | `a == b`         |
|             | `!=`     | Diferente de           | `a != b`         |
|             | `<`      | Menor que              | `a < b`          |
|             | `>`      | Maior que              | `a > b`          |
|             | `<=`     | Menor ou igual         | `a <= b`         |
|             | `>=`     | Maior ou igual         | `a >= b`         |
| Lógicos     | `&&`     | E lógico               | `a && b`         |
|             | `\|\|`   | Ou lógico              | `a \|\| b`       |
|             | `!`      | Negação                | `!a`             |
| Atribuição  | `=`      | Atribuição simples     | `a = b`          |
|             | `+=`     | Atribuição com soma    | `a += b`         |
|             | `-=`     | Atribuição com subtração| `a -= b`        |
| Acesso      | `.`      | Acesso a membro        | `this.x`         |
|             | `[]`     | Indexação              | `arr[0]`         |
| Chamada     | `()`     | Chamada de função      | `wait(1.0)`      |

### 6.2 Operadores (🔲 Planejados)

| Operador | Descrição           | Previsão |
|----------|---------------------|----------|
| `?.`     | Encadeamento null-safe | Futuro |
| `??`     | Null-coalescing      | Futuro   |
| `**`     | Potenciação          | Futuro   |
| `&` `\|` `^` `<<` `>>` | Bitwise | Futuro |
| `*=` `/=` `%=` | Atribuição composta adicional | Futuro |

### 6.3 Promoção Numérica

KScript promove automaticamente `int` para `float` em operações mistas:

```kscript
var a: int = 10
var b: float = 3.5
var c = a + b   // float (10 é promovido para 10.0)
var d = a * 2.0 // float
```

### 6.4 Concatenação de Strings

O operador `+` concatena strings:

```kscript
var msg: string = "Olá, " + "Mundo!" // "Olá, Mundo!"
```

### 6.5 Precedência de Operadores

| Precedência | Operadores                  | Associatividade |
|-------------|-----------------------------|-----------------|
| 1 (maior)   | `!` `-` (unário)            | Direita         |
| 2           | `*` `/` `%`                 | Esquerda        |
| 3           | `+` `-`                     | Esquerda        |
| 4           | `<` `>` `<=` `>=`           | Esquerda        |
| 5           | `==` `!=`                   | Esquerda        |
| 6           | `&&`                        | Esquerda        |
| 7 (menor)   | `\|\|`                      | Esquerda        |

Use parênteses para desambiguação:

```kscript
var result = (a + b) * c  // explícito
```

---

## 7. Controle de Fluxo

### 7.1 Condicionais: `if` / `else if` / `else`

```kscript
if (hp <= 0) {
  System.log("Morreu!")
} else if (hp < 30) {
  System.log("Vida crítica!")
} else {
  System.log("Vida ok")
}
```

A condição **deve** ser do tipo `bool`:

```kscript
if (hp) { }          // ERRO: hp é int, não bool
if (hp > 0) { }      // OK
```

### 7.2 `while`

Loop com condição verificada antes de cada iteração:

```kscript
var i: int = 0
while (i < 10) {
  System.log("i = $i")
  i += 1
}
```

Loop infinito com `break`:

```kscript
var pos: float = 0
while (true) {
  pos += 1.0
  if (pos >= 100.0) break
}
```

### 7.3 `break` e `continue`

```kscript
while (true) {
  if (done) break      // sai do loop
}

var i: int = 0
while (i < 10) {
  i += 1
  if (i % 2 == 0) continue  // pula pares
  System.log("ímpar: $i")
}
```

### 7.4 `for` (🔲 Parcialmente implementado)

> A palavra-chave `for` é reconhecida pelo lexer, mas a sintaxe completa
> `for (init; cond; post)` e `for-in` **ainda não estão implementadas** no parser.
>
> **Alternativa:** Use `while`:

```kscript
// Em vez de:
// for (var i = 0; i < 10; i++) { ... }

// Use:
var i: int = 0
while (i < 10) {
  // ...
  i += 1
}
```

> ⚠️ O suporte completo a `for` está em desenvolvimento.

### 7.5 `switch` / `when` (🔲 Não implementado)

> `switch` e `when` estão planejados mas não implementados.
> Use `if`/`else if` encadeados:

```kscript
var dir: int = DIR_UP
if (dir == DIR_UP) {
  // cima
} else if (dir == DIR_DOWN) {
  // baixo
}
```

### 7.6 Operador Ternário (🔲 Não implementado)

> O ternário `cond ? val1 : val2` está planejado mas não implementado.

---

## 8. Funções e Métodos

### 8.1 Declaração de Funções

Funções dentro de objetos são chamadas de **métodos**. São declaradas com `func` (ou sem a keyword, apenas o nome):

```kscript
object Player {
  // Método sem keyword `func`
  update(dt: float) {
    this.move(dt)
  }

  // Método com keyword `func`
  func move(dt: float) {
    this.x += this.speed * dt
  }
}
```

### 8.2 Parâmetros

```kscript
object Calculator {
  func add(a: int, b: int): int {
    return a + b
  }

  func greet(name: string) {
    System.log("Olá, " + name)
  }
}
```

### 8.3 Retorno

Use `return` para retornar um valor:

```kscript
func square(x: float): float {
  return x * x
}

func logAndExit(msg: string) {
  System.log(msg)
  return  // opcional quando tipo de retorno é vazio
}
```

### 8.4 Métodos Assíncronos

Métodos marcados com `async` podem usar `await`:

```kscript
async create() {
  await wait(1.0)
  System.log("Pronto após 1 segundo")
}
```

Veja [Programação Assíncrona](#10-programação-assíncrona) para detalhes.

### 8.5 Parâmetros com Valor Padrão (🔲 Não implementado)

> Parâmetros com valores padrão estão planejados.

### 8.6 Parâmetros Variádicos (🔲 Não implementado)

> Parâmetros variádicos (`...args`) estão planejados.

### 8.7 Arrow Functions / Lambdas (🔲 Não implementado)

> Arrow functions (`x => x * 2`) estão planejadas mas não implementadas.

---

## 9. Objetos (Classes)

### 9.1 Declaração de Objeto

Objetos são o principal mecanismo de organização de código em KScript. Eles são equivalentes a classes em outras linguagens.

```kscript
object Player {
  // Campos (propriedades)
  var speed: float = 200.0
  var hp: int = 100
  var pos: Vec2 = Vec2(0, 0)

  // Métodos
  update(dt: float) {
    this.move(dt)
  }

  func move(dt: float) {
    this.pos.x += this.speed * dt
  }
}
```

### 9.2 Campos (Propriedades)

Campos são declarados com `var` ou `const` dentro do corpo do objeto:

```kscript
object Enemy {
  var hp: int = 10         // mutável, com valor padrão
  var speed: float          // mutável, sem valor padrão (inicializa como 0)
  const MAX_HP: int = 10   // imutável (constante do objeto)
}
```

> Valores padrão podem ser literais ou expressões simples.

### 9.3 A Palavra-chave `this`

`this` refere-se à própria instância do objeto. É usado para acessar campos e métodos:

```kscript
object Player {
  var hp: int = 100

  func takeDamage(amount: int) {
    this.hp -= amount     // acessa campo hp
    this.flashRed()       // chama método flashRed
  }

  func flashRed() {
    // ...
  }
}
```

**Convenção:** Sempre use `this` para acessar campos do objeto, mesmo quando não ambíguo. Isso melhora a legibilidade.

### 9.4 Métodos

```kscript
object Animator {
  var t: float = 0.0

  // Método síncrono
  update(dt: float) {
    this.t += dt
  }

  // Método com retorno
  func getProgress(): float {
    return this.t
  }

  // Método async
  async fadeOut() {
    await tween(this, 0.3)
  }
}
```

### 9.5 Lifecycle Hooks

O motor Kora chama automaticamente certos métodos se eles estiverem definidos:

| Hook        | Assinatura                          | Quando chamado                 |
|-------------|-------------------------------------|--------------------------------|
| `create`    | `async create()`                    | Após o objeto ser instanciado  |
| `update`    | `update(dt: float)`                 | A cada frame (60fps)           |
| `draw`      | `draw(r: render.Renderer)`          | A cada frame, após update      |
| `destroy`   | `destroy()`                         | Quando o objeto é destruído    |

```kscript
object Player {
  async create() {
    System.log("Player nasceu!")
    await wait(0.5)
    this.hp = 100
  }

  update(dt: float) {
    this.pos.x += this.speed * dt
  }

  draw(r: render.Renderer) {
    // Desenho customizado
  }

  destroy() {
    System.log("Player destruído!")
  }
}
```

### 9.6 Event Hooks

Além dos lifecycle hooks, o motor chama métodos específicos em resposta a eventos:

```kscript
object Player {
  // Colisão
  onCollision(other: Entity, type: int) {
    System.log("Colidiu com: " + other.Tag)
  }

  // Input
  onInput(key: string, action: string) {
    if (key == "Space" && action == "pressed") {
      this.jump()
    }
  }

  // Touch (mobile)
  onTouch(pos: Vec2) {
    System.log("Toque em: " + pos.x + ", " + pos.y)
  }
}
```

> ⚠️ A assinatura exata e disponibilidade de hooks de evento dependem da
> versão do motor. Consulte a documentação do runtime para detalhes.

### 9.7 Herança (🔲 Não implementado)

> Herança entre objetos (`object Enemy : Character`) está planejada
> mas não implementada. Como alternativa, use composição:

```kscript
object Character {
  var hp: int = 100
  var speed: float = 100.0
}

object Player {
  var base: Character  // composição
}
```

### 9.8 Componentes (🔲 Planejado)

> O sistema de componentes está em planejamento. A ideia é permitir
> comportamentos modulares reutilizáveis entre objetos.

---

## 10. Programação Assíncrona

### 10.1 Visão Geral

KScript possui suporte nativo a `async`/`await`. Isso permite escrever lógica
sequencial não-bloqueante de forma legível:

```kscript
// Sem callbacks — parece código síncrono
async func showCutscene() {
  System.log("Iniciando cutscene...")
  await wait(2.0)          // espera 2 segundos REAIS
  System.log("Fade in...")
  await tween(this, 0.5)   // espera animação terminar
  System.log("Cutscene completa!")
}
```

### 10.2 Como Funciona

Cada `async` é compilado para uma **máquina de estados** em Go:

```
async create() {
  System.log("A")
  await wait(1.0)     ← Ponto de pausa (estado 0)
  System.log("B")
  await wait(0.5)     ← Ponto de pausa (estado 1)
  System.log("C")     ← Estado terminal
}
```

Vira (conceitualmente):

```
struct Create_Task {
  state: int
  subtask: Task
}

func Tick(dt float64) Status {
  switch state {
  case 0:
    log("A")
    subtask = Wait(1.0)
    state = 1
    return Running
  case 1:
    if subtask.Tick(dt) != Done { return Running }
    log("B")
    subtask = Wait(0.5)
    state = 2
    return Running
  case 2:
    if subtask.Tick(dt) != Done { return Running }
    log("C")
    return Done
  }
}
```

Isso significa que **async tasks são polled** — o motor chama `Tick(dt)` a cada
frame até que a task retorne `Done`. **Não há threads**, **não há preempção**.

### 10.3 Regras do Async

1. `await` **só pode aparecer** dentro de métodos marcados como `async`
2. Métodos `async` **não podem ser chamados como síncronos**
3. Tudo que segue um `await` executa **no frame seguinte** (ou quando a task completar)
4. Tasks não iniciam automaticamente — precisam ser *spawnadas* ou chamadas por outra async

```kscript
object Player {
  update(dt: float) {
    await wait(1.0)  // ERRO: update não é async
  }
}
```

### 10.4 Primitivas Async

#### `wait(seconds: float): Task`

Espera um número de **segundos** (tempo real, não frames):

```kscript
async create() {
  await wait(1.5)  // espera 1.5 segundos
}
```

#### `waitFrames(n: int): Task`

Espera um número de **frames** (a 60fps, 1 frame ≈ 16.6ms):

```kscript
async animate() {
  var i: int = 0
  while (i < 60) {
    // executa a cada frame por 60 frames
    i += 1
    await waitFrames(1)
  }
}
```

> ⚠️ `waitFrames` é reconhecido pelo checker mas depende da implementação
> no runtime.

#### `waitSignal(target: Entity, name: string): Task`

Espera até que um sinal seja emitido:

```kscript
async waitForDeath(enemy: Entity) {
  await waitSignal(enemy, "dead")
  System.log("Inimigo morreu!")
}
```

> ⚠️ Reconhecido pelo checker mas depende da implementação no runtime.

#### `tween(target: Entity, duration: float): Task` (⚠️ Planejado)

> `tween` está no checker mas a implementação completa de animação
> interpolada está em desenvolvimento.

#### `race(...tasks: Task[]): Task`

Retorna quando **qualquer** uma das tasks completa:

```kscript
async waitForFirst() {
  var result = await race(
    wait(3.0),
    waitSignal(enemy, "dead")
  )
  // Quem completar primeiro, vence
  System.log("Algo aconteceu!")
}
```

> ⚠️ Reconhecido pelo checker mas depende da implementação do runtime.

#### `all(...tasks: Task[]): Task`

Espera **todas** as tasks completarem:

```kscript
async loadGame() {
  await all(
    loadSprites(),
    loadAudio(),
    loadLevel()
  )
  System.log("Tudo carregado!")
}
```

> ⚠️ Reconhecido pelo checker mas depende da implementação do runtime.

#### `cancel(task: Task)`

Cancela uma task em execução:

```kscript
var myTask: Task = null

async startTask() {
  myTask = spawn(this.longRunningTask())
}

func stopTask() {
  if (myTask != null) {
    cancel(myTask)
    myTask = null
  }
}
```

> ⚠️ Reconhecido pelo checker mas depende da implementação do runtime.

### 10.5 `spawn`

> 🔲 Não implementado. Use chamada direta de métodos async.

### 10.6 Boas Práticas com Async

```kscript
// ✅ BOM: async para lógica sequencial com espera
async waitAndAttack() {
  await wait(0.5)
  this.attack()
}

// ✅ BOM: async para animações
async flashRed() {
  await tween(this, 0.1)
  await tween(this, 0.1)
}

// ❌ EVITE: async sem await (use síncrono)
async doNothing() {   // sem await - use método normal
  this.x += 1
}

// ❌ EVITE: loops que criam tasks infinitas sem condição de saída
async badLoop() {
  while (true) {
    await wait(0.1)
    // sem break = task eterna
  }
}
```

---

## 11. Sistema de Sinais

### 11.1 Visão Geral

Sinais são o sistema de eventos do KScript. Um objeto pode **emitir** um sinal
e outros objetos podem **esperar** por ele. É um mecanismo de comunicação
fracamente acoplado.

### 11.2 Emitir Sinais

Use a declaração `emit` para disparar um sinal:

```kscript
object Enemy {
  var hp: int = 10

  func takeDamage(amount: int) {
    this.hp -= amount
    if (this.hp <= 0) {
      emit "dead"     // emite o sinal "dead"
    }
  }
}
```

> ⚠️ Na versão atual, `emit` aceita apenas uma string literal com o nome do
> sinal. A emissão com dados adicionais está em planejamento.

**Como funciona:** O compilador traduz `emit "dead"` para `o.EmitSignal("dead")`
no Go gerado. Cabe ao runtime gerenciar a entrega do sinal.

### 11.3 Escutar Sinais

Objetos podem esperar sinais via `waitSignal`:

```kscript
object Player {
  async create() {
    // Espera o inimigo morrer
    await waitSignal(enemy, "dead")
    System.log("Inimigo morto! Player ganhou XP.")
  }
}
```

### 11.4 Ciclo de Vida dos Sinais

Sinais em KScript têm **duração de um frame**:

1. Sinal é emitido (`emit "dead"`)
2. Neste frame, todos os waiting tasks que esperam "dead" são notificados
3. No frame seguinte, o sinal "expira"

Isso significa que **escutar um sinal ANTES dele ser emitido** é necessário.
Você não pode "perder" um sinal se já estiver esperando por ele.

### 11.5 Padrão: Sinal + Async

Combine sinais com async para criar fluxos reativos:

```kscript
object HealthSystem {
  var hp: int = 10
  var maxHp: int = 10

  func takeDamage(amount: int) {
    this.hp -= amount
    emit "damage_taken"

    if (this.hp <= 0) {
      emit "death"
    }
  }

  // Escuta seus próprios sinais
  async watchHealth() {
    await waitSignal(this, "damage_taken")
    this.flashRed()
  }

  func flashRed() {
    // Efeito visual
  }
}
```

### 11.6 Limitações Atuais

- Sinais são apenas strings — sem dados acompanhando
- Sem sistema de `onSignal` declarativo (métodos automáticos)
- Sem escopo de sinal (global vs local)
- A implementação do runtime está em desenvolvimento

---

## 12. API do Motor

> Esta seção documenta as APIs nativas disponíveis em scripts KScript.
> Para funcionarem, o runtime (`core/`) precisa estar implementado.

### 12.1 Input

Namespace `Input` para entrada do jogador:

| Método             | Retorno  | Descrição                        |
|--------------------|----------|----------------------------------|
| `Input.axisX()`    | `float`  | Eixo horizontal (-1 a 1)         |
| `Input.axisY()`    | `float`  | Eixo vertical (-1 a 1)           |
| `Input.pressed(k)` | `bool`   | Tecla foi pressionada neste frame|
| `Input.held(k)`    | `bool`   | Tecla está sendo segurada        |
| `Input.released(k)`| `bool`   | Tecla foi liberada neste frame   |
| `Input.touchPos(i)`| `Vec2`   | Posição do toque i               |

```kscript
update(dt: float) {
  // Movimento por eixo (teclado / gamepad)
  var moveX = Input.axisX()
  var moveY = Input.axisY()

  // Ações discretas
  if (Input.pressed("Space")) {
    this.jump()
  }

  if (Input.held("Left")) {
    this.moveLeft()
  }

  if (Input.released("E")) {
    this.interact()
  }

  // Touch (mobile)
  var touchCount: int = Input.touchCount()
  if (touchCount > 0) {
    var pos = Input.touchPos(0)
  }
}
```

> ⚠️ Os nomes de teclas seguem o padrão do motor (ex: `"Space"`, `"ArrowLeft"`,
> `"KeyE"`). Consulte a documentação do runtime para a lista completa.

### 12.2 Audio

Namespace `Audio` para som e música:

| Método                    | Descrição                    |
|---------------------------|------------------------------|
| `Audio.play(name)`        | Toca som uma vez             |
| `Audio.stop(name)`        | Para um som específico       |
| `Audio.fadeOut(name, s)`  | Fade out em s segundos       |
| `Audio.playLoop(name)`    | Toca em loop (BGM)           |
| `Audio.pause()`           | Pausa todo áudio             |
| `Audio.resume()`          | Resume áudio                 |

```kscript
func playSFX() {
  Audio.play("sfx_jump")
}

func playMusic() {
  Audio.playLoop("bgm_level1")
}

func stopMusic() {
  Audio.fadeOut("bgm_level1", 1.0)
}
```

### 12.3 Scene

Namespace `Scene` para gerenciamento de cenas:

| Método                          | Retorno   | Descrição                    |
|---------------------------------|-----------|------------------------------|
| `Scene.spawn(name, pos)`        | `Entity`  | Cria entidade na cena        |
| `Scene.find(name)`              | `Entity`  | Busca entidade por nome      |
| `Scene.pause()`                 | —         | Pausa a cena                 |
| `Scene.resume()`                | —         | Resume a cena                |
| `Scene.isPaused()`              | `bool`    | Verifica se está pausada     |
| `Scene.changeScene(sceneName)`  | —         | Transiciona para outra cena  |
| `Scene.load(scenePath)`         | —         | Carrega uma cena             |
| `Scene.instantiate(objName)`    | `Entity`  | Instancia um objeto KScript  |

```kscript
async spawnEnemies() {
  var e1 = Scene.spawn("Enemy", Vec2(100, 100))
  var e2 = Scene.find("player")
  Scene.pause()
}

func changeLevel() {
  Scene.changeScene("level2")
}
```

**Mapeamento para Go:** O compilador traduz chamadas `Scene.*` para
métodos de `runner.GameTree()` ou `runner.GameSceneManager()`.

### 12.4 Physics

Namespace `Physics` para física e colisão:

| Método                                           | Retorno    | Descrição                    |
|--------------------------------------------------|------------|------------------------------|
| `Physics.setGravity(x, y)`                       | —          | Define gravidade mundial     |
| `Physics.raycast(fromX, fromY, toX, toY, mask)`  | `{hit, x, y, normalX, normalY}` | Raycast |
| `Physics.overlapRect(minX, minY, maxX, maxY, mask)` | `Entity[]` | Overlap test |

```kscript
func checkRaycast() {
  var hit = Physics.raycast(0, 0, 100, 0, 1)
  if (hit.hit) {
    System.log("Acertou em: " + hit.x + ", " + hit.y)
  }
}

func checkOverlaps() {
  var entities = Physics.overlapRect(0, 0, 200, 200, 1)
}
```

> ⚠️ A API de física depende do `PhysicsWorld` registrado via
> `compiler.RegisterPhysicsAPI()`.

### 12.5 Camera (⚠️ Planejado)

> A API de câmera está em planejamento. Funcionalidades esperadas:

| Método                | Descrição                    |
|-----------------------|------------------------------|
| `Camera.setZoom(z)`   | Define zoom da câmera        |
| `Camera.setPos(x, y)` | Move a câmera                |
| `Camera.shake(a, d)`  | Efeito de shake              |
| `Camera.follow(e, s)` | Segue uma entidade           |

### 12.6 Math

Namespace `Math` para funções matemáticas:

| Método                | Retorno  | Descrição               |
|-----------------------|----------|-------------------------|
| `Math.abs(x)`         | `float`  | Valor absoluto          |
| `Math.floor(x)`       | `float`  | Arredonda para baixo    |
| `Math.ceil(x)`        | `float`  | Arredonda para cima     |
| `Math.round(x)`       | `float`  | Arredonda               |
| `Math.sqrt(x)`        | `float`  | Raiz quadrada           |
| `Math.sin(x)`         | `float`  | Seno (radianos)         |
| `Math.cos(x)`         | `float`  | Cosseno (radianos)      |
| `Math.tan(x)`         | `float`  | Tangente (radianos)     |
| `Math.lerp(a, b, t)`  | `float`  | Interpolação linear     |
| `Math.clamp(v, min, max)` | `float` | Limita valor          |

```kscript
var clamped = Math.clamp(value, 0.0, 1.0)
var lerped = Math.lerp(0.0, 100.0, 0.5)  // 50.0
var root = Math.sqrt(144)                  // 12
```

### 12.7 System

Namespace `System` para utilitários gerais:

| Método                  | Descrição                        |
|-------------------------|----------------------------------|
| `System.log(msg)`       | Log informativo                  |
| `System.warn(msg)`      | Log de aviso                     |
| `System.error(msg)`     | Log de erro                      |
| `System.time()`         | Tempo decorrido em segundos      |
| `System.exit()`         | Sai do jogo                      |

```kscript
System.log("Jogo iniciado!")
var elapsed = System.time()
System.warn("Cuidado: vida baixa!")
System.error("Erro fatal!")
```

### 12.8 Entity Properties

Todo objeto KScript possui propriedades de entidade embutidas:

| Propriedade  | Tipo      | Descrição                    |
|--------------|-----------|------------------------------|
| `this.x`     | `float`   | Posição X                    |
| `this.y`     | `float`   | Posição Y                    |
| `this.VelX`  | `float`   | Velocidade X                 |
| `this.VelY`  | `float`   | Velocidade Y                 |
| `this.Alpha` | `float`   | Opacidade (0.0 a 1.0)        |
| `this.Visible`| `bool`   | Visibilidade                 |
| `this.Tag`   | `string`  | Identificador textual        |

```kscript
object Player {
  update(dt: float) {
    this.x += this.VelX * dt    // movimento
    this.Alpha = Math.lerp(this.Alpha, 1.0, dt * 5)  // fade in suave
  }
}
```

### 12.9 Asset Loading (⚠️ Planejado)

> O sistema de carregamento de assets está em desenvolvimento.

| Método                      | Descrição                    |
|-----------------------------|------------------------------|
| `Asset.load(path)`          | Carrega asset                |
| `Asset.unload(path)`        | Libera asset                 |
| `Asset.get(path)`           | Obtém asset carregado        |

### 12.10 Debug (⚠️ Planejado)

> Utilitários de debug estão em planejamento.

---

## 13. Pipeline de Compilação

### 13.1 Visão Geral

```
KScript Source (.ks)
       │
       ▼
┌──────────────┐
│    LEXER     │  Tokens: IDENT("Player"), INT(42), etc.
│  lexer/      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   PARSER     │  AST: Program → Objects → Methods → Stmts → Exprs
│  parser/     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   CHECKER    │  Type annotation, scope resolution, semantic errors
│  checker/    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  TRANSFORM   │  Async methods → state machines (for await points)
│ transform/   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   EMITTER    │  Go source code (structs, methods, tasks)
│  emitter/    │
└──────┬───────┘
       │
       ▼
   Go Code (.go)
       │
       ▼
   go build → kora.apk
```

### 13.2 Lexer (lexer/)

O lexer converte o texto fonte em uma sequência de tokens. Cada token tem
um tipo, um valor literal e posição (linha:coluna).

**Tokens suportados:**

| Categoria  | Tokens                                          |
|------------|-------------------------------------------------|
| Literais   | `IDENT`, `INT`, `FLOAT`, `STRING`, `BOOL`       |
| Palavras-chave | `object`, `func`, `async`, `await`, `return`, `if`, `else`, `for`, `while`, `break`, `continue`, `const`, `var`, `true`, `false`, `null`, `import`, `from`, `this`, `emit` |
| Pontuação  | `{` `}` `(` `)` `[` `]` `,` `.` `:` `;`       |
| Operadores | `=` `+` `-` `*` `/` `%` `!` `<` `>` `<=` `>=` `==` `!=` `&&` `||` `+=` `-=` |

```go
// Exemplo de lexer: "object Player { var hp: int = 5 }"
// Saída:
// object  (keyword)
// Player  (IDENT)
// {       (punctuation)
// var     (keyword)
// hp      (IDENT)
// :       (punctuation)
// int     (IDENT)
// =       (operator)
// 5       (INT)
// }       (punctuation)
```

### 13.3 Parser (parser/)

O parser constrói a Árvore Sintática Abstrata (AST) a partir dos tokens.

**Estrutura da AST (versão atual):**

```
Program
├── Imports[]: { Module, Symbols[] }
└── Objects[]: {
      Name: string
      Fields[]: { Name, Type, Default? }
      Methods[]: {
        Name, Async, Params[], Return?,
        Body: Stmt[]
      }
    }
```

**Statements:** `VarStmt`, `AssignStmt`, `ReturnStmt`, `ExprStmt`,
`IfStmt`, `WhileStmt`, `BreakStmt`, `ContinueStmt`, `AwaitStmt`, `EmitStmt`

**Expressions:** `IntLit`, `FloatLit`, `StringLit`, `BoolLit`, `NullLit`,
`Ident`, `ThisExpr`, `BinaryExpr`, `UnaryExpr`, `CallExpr`, `MemberExpr`,
`IndexExpr`

### 13.4 Checker (checker/)

O checker realiza análise semântica:

1. **Coleta nomes de objetos** (detecta duplicatas)
2. **Valida imports** (apenas módulo `"kora"` permitido)
3. **Verifica tipos** de campos, parâmetros e retornos
4. **Valida bodies de métodos**:
   - Variáveis não declaradas
   - Reatribuição a const
   - `await` fora de async
   - Tipos de condicionais (devem ser bool)
   - Número de argumentos em chamadas de engine API
5. **Resolve namespaces** (Input, Audio, Scene, Math)

**Engine API conhecida:**

| Função        | Parâmetros              | Retorno  |
|---------------|-------------------------|----------|
| `wait`        | `float`                 | `Task`   |
| `waitFrames`  | `int`                   | `Task`   |
| `waitSignal`  | `Entity, string`        | `Task`   |
| `tween`       | `Entity, any, float`    | `Task`   |
| `race`        | variadic                | `Task`   |
| `all`         | variadic                | `Task`   |
| `cancel`      | `Task`                  | `void`   |

### 13.5 Transform (transform/)

O transformador converte métodos `async` em máquinas de estados:

1. Divide o corpo do método em **estados** separados por `await`
2. Cada estado contém statements síncronos + uma expressão `await` opcional
3. Variáveis que "atravessam" await points são promovidas a **campos da struct** da task

```kscript
async example() {
  var x: float = 10
  await wait(0.5)     // estado 0
  var y: float = x + 5 // estado 1 (terminal)
}
```

Vira:

```
States[0]: { stmts: [var x],             await: wait(0.5) }
States[1]: { stmts: [var y, y = x + 5], await: nil (done) }
LiveVars: [x: float]
```

### 13.6 Emitter (emitter/)

O emissor gera código Go compilável.

**Mapeamento KScript → Go:**

| KScript           | Go                                    |
|-------------------|---------------------------------------|
| `object Player`   | `type Player struct { X, Y float64 ... }` |
| `var speed: float`| `Speed float64`                       |
| `async create()`  | `type Player_Create_Task struct { ... }` + `Tick(dt) Status` |
| `update(dt)`      | `func (o *Player) Update(dt float64)` |
| `this.x`          | `o.X`                                 |
| `emit "dead"`     | `o.EmitSignal("dead")`                |
| `Scene.pause()`   | `runner.GameTree().Pause`             |
| `wait(1.0)`       | `async.Wait(1.0)`                     |

**Struct gerada automaticamente para cada objeto:**

```go
type Player struct {
    // Entity built-in fields
    X, Y         float64
    VelX, VelY   float64
    Alpha        float32
    Visible      bool
    Tag          string
    alive        bool

    // User fields
    Hp    int
    Speed float64
}
```

**Interface mínima de Entity:**

```go
func (o *Player) IsAlive() bool { return o.alive }
func (o *Player) Destroy()      { o.alive = false }
```

---

## 14. Padrões Comuns

### 14.1 Player Controller (Plataforma)

```kscript
object Player {
  var speed: float = 200.0
  var jumpForce: float = 400.0
  var gravity: float = 980.0
  var isOnGround: bool = false

  update(dt: float) {
    // Movimento horizontal
    var moveX = Input.axisX()
    this.VelX = moveX * this.speed

    // Pulo
    if (Input.pressed("Space") && this.isOnGround) {
      this.VelY = -this.jumpForce
      this.isOnGround = false
    }

    // Gravidade
    this.VelY += this.gravity * dt

    // Atualiza posição
    this.x += this.VelX * dt
    this.y += this.VelY * dt
  }

  onCollision(other: Entity, type: int) {
    if (type == 0) { // COLLIDE
      this.isOnGround = true
      this.VelY = 0
    }
  }
}
```

### 14.2 Player Controller (Top-Down)

```kscript
object Player {
  var speed: float = 180.0
  var hp: int = 100

  update(dt: float) {
    var ax = Input.axisX()
    var ay = Input.axisY()

    // Movimento diagonal normalizado
    if (ax != 0 || ay != 0) {
      var len = Math.sqrt(ax * ax + ay * ay)
      ax = ax / len
      ay = ay / len
    }

    this.x += ax * this.speed * dt
    this.y += ay * this.speed * dt

    // Interação
    if (Input.pressed("KeyE")) {
      emit "interact"
    }
  }
}
```

### 14.3 Inimigo com Máquina de Estados

```kscript
object Enemy {
  var state: int = 0 // 0=idle, 1=patrol, 2=chase
  var hp: int = 3
  var speed: float = 60.0
  var patrolTimer: float = 0.0
  var dir: float = 1.0

  update(dt: float) {
    if (this.state == 0) { // idle
      this.patrolTimer += dt
      if (this.patrolTimer > 2.0) {
        this.state = 1
        this.patrolTimer = 0.0
        if (Math.random() > 0.5) {
          this.dir = 1.0
        } else {
          this.dir = -1.0
        }
      }
    } else if (this.state == 1) { // patrol
      this.x += this.dir * this.speed * dt
      this.patrolTimer += dt
      if (this.patrolTimer > 3.0) {
        this.state = 0
        this.patrolTimer = 0.0
      }
    }
  }

  func takeDamage(amount: int) {
    this.hp -= amount
    if (this.hp <= 0) {
      emit "dead"
      this.Destroy()
    }
  }
}
```

### 14.4 Coleta com Invulnerabilidade

```kscript
object Player {
  var hp: int = 100
  var invulnerable: bool = false
  var invulTimer: float = 0.0

  update(dt: float) {
    if (this.invulnerable) {
      this.invulTimer -= dt
      if (this.invulTimer <= 0.0) {
        this.invulnerable = false
      }
    }
  }

  func takeDamage(amount: int) {
    if (this.invulnerable) return

    this.hp -= amount
    this.invulnerable = true
    this.invulTimer = 1.0 // 1 segundo de invulnerabilidade

    // Piscar (alternar visibilidade) — exemplo conceitual
    // (precisa de async ou timer)

    if (this.hp <= 0) {
      emit "dead"
      this.Destroy()
    }
  }
}
```

### 14.5 Sistema de Score com Objeto Global

```kscript
object ScoreManager {
  var score: int = 0
  var highScore: int = 0

  func addPoints(amount: int) {
    this.score += amount
    if (this.score > this.highScore) {
      this.highScore = this.score
    }
    System.log("Score: " + this.score)
  }

  func reset() {
    this.score = 0
  }
}
```

### 14.6 Transição entre Cenas

```kscript
object SceneTransition {
  var fading: bool = false
  var alpha: float = 0.0
  var nextScene: string = ""

  func goToScene(sceneName: string) {
    this.nextScene = sceneName
    this.fading = true
    this.alpha = 0.0
  }

  update(dt: float) {
    if (!this.fading) return

    this.alpha += dt * 2.0 // fade speed
    this.Visible = true
    this.Alpha = this.alpha

    if (this.alpha >= 1.0) {
      Scene.changeScene(this.nextScene)
      this.fading = false
    }
  }
}
```

### 14.7 Projéteis

```kscript
object Bullet {
  var speed: float = 500.0
  var dirX: float = 0.0
  var dirY: float = -1.0
  var lifetime: float = 2.0
  var age: float = 0.0

  update(dt: float) {
    this.x += this.dirX * this.speed * dt
    this.y += this.dirY * this.speed * dt

    this.age += dt
    if (this.age >= this.lifetime) {
      this.Destroy()
    }
  }
}

object Player {
  func shoot() {
    var bullet = Scene.spawn("Bullet", Vec2(this.x, this.y))
    // Configurar direção baseada no alvo
    // (requer acesso ao objeto Bullet)
  }
}
```

### 14.8 Save/Load (⚠️ Conceitual)

```kscript
// ⚠️ Exemplo conceitual — APIs de save/load estão em planejamento
object SaveManager {
  func saveGame() {
    // Serializar estado do jogo
    // Escrever para disco
    System.log("Jogo salvo!")
  }

  func loadGame() {
    // Ler do disco
    // Restaurar estado
    System.log("Jogo carregado!")
  }
}
```

---

## 15. Solução de Problemas

### 15.1 Erros do Compilador

#### "unknown type"

```
Error: object Player: field "hp" has unknown type "health"
```

O tipo `health` não foi definido. Tipos válidos: `int`, `float`, `bool`,
`string`, `void`, `Vec2`, `Color`, `Task`, `Entity`, ou nome de outro objeto.
Tipos array: `int[]`, `float[]`, `string[]`, `Entity[]`.

#### "await used outside of an async method"

```
Error: `await` used outside of an async method
```

`await` só pode ser usado em métodos marcados com `async`:

```kscript
// ❌ ERRADO
update(dt: float) {
  await wait(1.0)
}

// ✅ CORRETO
async update(dt: float) {
  await wait(1.0)
}
```

#### "await expression must return Task"

```
Error: `await` expression must return Task, got float
```

Só é possível usar `await` com expressões que retornam `Task`:

```kscript
// ❌ ERRADO
await 42

// ✅ CORRETO
await wait(1.0)
```

#### "undeclared identifier"

```
Error: undeclared identifier "speed"
```

A variável `speed` não foi declarada no escopo atual:

```kscript
// ❌ ERRADO
func run() {
  this.x += speed * dt  // speed não declarada
}

// ✅ CORRETO
func run() {
  var speed: float = 100.0
  this.x += speed * dt
}

// Ou, se for campo do objeto:
func run() {
  this.x += this.speed * dt  // assume que speed é um campo
}
```

#### "cannot assign to const"

```
Error: cannot assign to const "MAX_HP"
```

Constantes não podem ser reatribuídas:

```kscript
const MAX_HP: int = 100
MAX_HP = 50  // ❌ ERRO
```

#### "if condition must be bool"

```
Error: if condition must be bool, got int
```

Condições de `if` e `while` precisam ser expressões booleanas:

```kscript
// ❌ ERRADO
if (hp) { }

// ✅ CORRETO
if (hp > 0) { }
```

#### "expects N argument(s), got M"

```
Error: wait() expects 1 argument(s), got 2
```

Número incorreto de argumentos para uma função da engine API:

```kscript
wait(1.0, 2.0)     // ❌ wait espera 1 argumento
wait(1.0)           // ✅
```

#### "unknown module"

```
Error: unknown module "npm" — only engine modules may be imported
```

Apenas o módulo `"kora"` pode ser importado atualmente:

```kscript
import { Input } from "kora"   // ✅
import { X } from "npm"        // ❌
```

#### "duplicate object declaration"

```
Error: duplicate object declaration: "Player"
```

Dois objetos com o mesmo nome foram declarados:

```kscript
object Player { }  // primeiro
object Player { }  // ❌ duplicata
```

### 15.2 Erros de Runtime (⚠️)

> Estes são problemas que podem ocorrer durante a execução do jogo,
> mas não são capturados pelo compilador por serem dinâmicos.

#### Sinal não recebido

Se um sinal é emitido mas ninguém está escutando, ele é perdido.
Certifique-se de que `waitSignal` é chamado **antes** do `emit`.

#### Task nunca completa

Se uma task async tem um loop infinito sem `break`, ela nunca retorna `Done`.
Isso pode travar o jogo.

#### Performance: update()

Evite operações pesadas dentro de `update()`:
- Não crie muitas tasks async
- Não faça alocações desnecessárias
- Não use loops muito longos

### 15.3 Funcionalidades Não Implementadas

| Funcionalidade      | Status     | Alternativa                      |
|---------------------|------------|----------------------------------|
| `for` loops         | 🔲 Parcial | Use `while`                      |
| `switch`/`when`     | 🔲         | Use `if`/`else`                  |
| Ternário            | 🔲         | Use `if`/`else`                  |
| Enums               | 🔲         | Use constantes                   |
| Type aliases        | 🔲         | Use o tipo diretamente           |
| Lambdas/arrow       | 🔲         | Use funções nomeadas             |
| Block comments      | 🔲         | Use `//`                         |
| Null-safe (`?.`)    | 🔲         | Use `if (x != null)`             |
| Null-coalescing     | 🔲         | Use `if`/`else`                  |
| Bitwise ops         | 🔲         | —                                |
| Herança             | 🔲         | Use composição                   |
| Componentes         | 🔲         | Use objetos separados            |
| Array/Map literais  | 🔲         | Use APIs do motor                |
| Parâmetros default  | 🔲         | Passe explicitamente             |
| Variadic params     | 🔲         | Passe argumentos individuais     |

### 15.4 Dicas de Performance

```kscript
// ✅ BOM: Variáveis locais para cálculos repetidos
update(dt: float) {
  var spd = this.speed  // cache local
  this.x += spd * dt
  this.y += spd * dt
}

// ✅ BOM: Evitar alocações em update()
update(dt: float) {
  // Vec2 é criado pelo motor, reutilize
  this.x += ax * this.speed * dt
}

// ✅ BOM: Condições curto-circuito
if (this.invulnerable && this.hp > 0) { }
```

---

## 16. Apêndices

### A: Gramática (EBNF)

> Notação EBNF simplificada para o estado atual da linguagem.

```
Program       = { ImportDecl | ObjectDecl }

ImportDecl    = "import" "{" Ident { "," Ident } "}" "from" String

ObjectDecl    = "object" Ident "{" { FieldDecl | MethodDecl } "}"

FieldDecl     = ("var" | "const") Ident [ ":" Type ] [ "=" Expr ] [ ";" ]

MethodDecl    = [ "async" ] [ "func" ] Ident "(" [ Params ] ")" [ ":" Type ] Block

Params        = Param { "," Param }
Param         = Ident ":" Type

Block         = "{" { Stmt } "}"

Stmt          = VarStmt | ReturnStmt | IfStmt | WhileStmt
              | BreakStmt | ContinueStmt | AwaitStmt | EmitStmt
              | ExprStmt | AssignStmt

VarStmt       = ("var" | "const") Ident [ ":" Type ] [ "=" Expr ] [ ";" ]
ReturnStmt    = "return" [ Expr ] [ ";" ]
IfStmt        = "if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ]
WhileStmt     = "while" "(" Expr ")" Block
BreakStmt     = "break" [ ";" ]
ContinueStmt  = "continue" [ ";" ]
AwaitStmt     = "await" Expr [ ";" ]
EmitStmt      = "emit" String [ ";" ]
AssignStmt    = Expr ("=" | "+=" | "-=") Expr [ ";" ]
ExprStmt      = Expr [ ";" ]

Expr          = BinaryExpr | UnaryExpr | PostfixExpr
BinaryExpr    = UnaryExpr { ("||" | "&&" | "==" | "!=" | "<" | ">" | "<=" | ">="
                           | "+" | "-" | "*" | "/" | "%") UnaryExpr }
UnaryExpr     = ("!" | "-") UnaryExpr | PostfixExpr
PostfixExpr   = PrimaryExpr { "." Ident | "(" [ Args ] ")" | "[" Expr "]" }
PrimaryExpr   = Int | Float | String | Bool | Null | Ident | "this" | "(" Expr ")"
Args          = Expr { "," Expr }

Type          = Ident [ "[" "]" ]
```

### B: Palavras Reservadas

```
object        func          async         await         return
if            else          for           while         break
continue      const         var           true          false
null          import        from          this          emit
```

> Estas palavras não podem ser usadas como identificadores.

### C: Tabela de Tipos

| Tipo KScript | Go gerado             | Descrição                 |
|--------------|-----------------------|---------------------------|
| `int`        | `int`                 | Inteiro 64-bit            |
| `float`      | `float64`             | Ponto flutuante           |
| `bool`       | `bool`                | Booleano                  |
| `string`     | `string`              | String UTF-8              |
| `void`       | (vazio)               | Sem retorno               |
| `Task`       | `async.Task`          | Task assíncrona           |
| `Entity`     | `scene.Entity`        | Entidade do jogo          |
| `Vec2`       | `render.Vec2`         | Vetor 2D                  |
| `int[]`      | `[]int`               | Array de ints             |
| `float[]`    | `[]float64`           | Array de floats           |
| `string[]`   | `[]string`            | Array de strings          |
| `Entity[]`   | `[]scene.Entity`      | Array de entidades        |
| `any`        | `interface{}`         | Tipo coringa              |

### D: Precedência de Operadores

| Nível | Operadores                | Associatividade |
|-------|---------------------------|-----------------|
| 6     | `!` `-` (unário)          | Direita         |
| 5     | `*` `/` `%`               | Esquerda        |
| 4     | `+` `-`                   | Esquerda        |
| 3     | `<` `>` `<=` `>=`         | Esquerda        |
| 2     | `==` `!=`                 | Esquerda        |
| 1     | `&&`                      | Esquerda        |
| 0     | `\|\|`                    | Esquerda        |

### E: Códigos de Erro

| Código | Significado                           |
|--------|---------------------------------------|
| E001   | Tipo desconhecido                     |
| E002   | Variável não declarada                |
| E003   | Reatribuição a const                  |
| E004   | await fora de método async            |
| E005   | await em expressão não-Task           |
| E006   | Condição não-booleana                 |
| E007   | Número de argumentos incorreto        |
| E008   | Módulo de import desconhecido         |
| E009   | Objeto duplicado                      |
| E010   | Sinal vazio no emit                   |

---

## Histórico de Revisões

| Versão | Data       | Descrição                               |
|--------|------------|-----------------------------------------|
| 0.1    | 2026-05-21 | Documentação inicial da linguagem       |

---

*KScript — compile para Go, rode nativo, desenvolva com sintaxe familiar.*
