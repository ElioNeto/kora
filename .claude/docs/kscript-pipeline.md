# Knowledge Base: Pipeline de Compilação KScript

Este documento descreve o funcionamento interno do compilador KScript → Go do Kora Engine. **Leia este documento antes de qualquer modificação em `compiler/`.**

---

## Visão Geral do Pipeline

KScript é uma linguagem **AOT (Ahead-of-Time)**: não existe VM, interpretação em runtime, ou JIT. Todo código KScript é convertido em Go puro antes do build do APK.

```
 arquivo.ks
     │
     ▼
 [ 1. Lexer ]           → []Token
     │
     ▼
 [ 2. Parser ]          → *ast.Program
     │
     ▼
 [ 3. Checker ]         → []TypeError (ou nil)
     │
     ▼
 [ 4. Transform ]       → asyncMap (mapeamento de corrotinas)
     │
     ▼
 [ 5. Emitter ]         → string (código Go)
     │
     ▼
  arquivo.ks.go         ← GERADO: nunca editar manualmente
```

**Entry point público:** `compiler.CompileFile(path, opts)` e `compiler.CompileSource(src, opts)` em `compiler/compiler.go`.

---

## Fase 1: Lexer (`compiler/lexer/`)

**Responsabilidade:** Tokenizar o source KScript em uma lista plana de tokens.

### Regras críticas
- Palavras-chave reservadas: `object`, `component`, `fn`, `async`, `await`, `emit`, `signal`, `spawn`, `cancel`, `loop`, `when`
- String interpolation: `"$var"` e `"$(expr)"` — o lexer expande em tempo de tokenização
- Null-safety tokens: `?.` (null-safe nav), `??` (null coalesce), `T?` (nullable type) são tokens distintos
- Comentários: `//` de linha e `/* */` de bloco — descartados pelo lexer
- Números: `int` (sem ponto), `float` (com ponto ou `e` expoente)

### Erros do Lexer
- Retorna `fmt.Errorf("lexer: %w", err)` com linha e coluna
- Caracteres inválidos geram `UnexpectedCharError{Line, Col, Char}`

---

## Fase 2: Parser (`compiler/parser/`)

**Responsabilidade:** Construir a Árvore Sintática Abstrata (AST) a partir dos tokens.

### Nós principais do AST (`compiler/ast/`)

| Nó AST | Representa |
|---------|------------|
| `*ast.Program` | Raíz — lista de top-level declarations |
| `*ast.ObjectDecl` | `object NomeObjeto { ... }` |
| `*ast.ComponentDecl` | `component NomeComponent { ... }` |
| `*ast.FnDecl` | Função/método (sync ou async) |
| `*ast.VarDecl` | Declaração de variável com tipo opcional |
| `*ast.CallExpr` | Chamada de função ou método built-in |
| `*ast.AwaitExpr` | Expressão `await expr` |
| `*ast.SignalDecl` | Declaração de sinal |
| `*ast.EmitStmt` | `emit NomeSinal(args)` |

### Regras do Parser
- Um arquivo `.ks` pode conter múltiplos `object` e `component` no top-level
- Funções dentro de `object/component` são métodos (vinculados ao receiver)
- Funções no top-level são funções livres (geradas como func Go global)
- `when` é um `switch` tipado — não tem fallthrough implícito

---

## Fase 3: Checker (`compiler/checker/`)

**Responsabilidade:** Verificar tipos, escopos e uso correto da API.

> ⚠️ **Área de dívida técnica (DEBT-001):** Genéricos `Array<T>` e `Map<K,V>` têm verificação parcial. Ao adicionar suporte a novos tipos genéricos, verificar se propagam corretamente pelo AST.

### Como o checker valida built-ins
O checker mantém um **registro de módulos built-in** (provavelmente em `builtins.go`). Cada entrada define:
- Nome do módulo (ex: `"Audio"`)
- Funções com assinaturas: parâmetros e tipo de retorno

Quando o KScript acessa `Audio.play("som.ogg")`, o checker:
1. Resolve `Audio` como módulo built-in
2. Verifica que `play` existe no registro
3. Valida que o argumento `"som.ogg"` é compatível com `string`

### Tipos KScript → Go (mapeamento interno)
| KScript | Go emitido |
|---------|------------|
| `string` | `string` |
| `int` | `int` |
| `float` | `float64` |
| `bool` | `bool` |
| `void` | — (sem retorno) |
| `string?` | `*string` ou `(string, bool)` |
| `Array<string>` | `[]string` |
| `Map<string, int>` | `map[string]int` |
| `Vector2` | `struct{ X, Y float64 }` do core |

---

## Fase 4: Transform (`compiler/transform/`)

**Responsabilidade:** Analisar funções `async` e construir o `asyncMap` que o emitter usa para gerar corrotinas Go.

### Como `async/await` é transpilado
KScript:
```kscript
object Player {
  async create() {
    await Asset.load("player.png")
    spawn patrol()
  }
}
```

Go gerado:
```go
func (p *Player) Create() {
    koraAsync.Schedule(func() {
        <-koraAsset.Load("player.png")  // channel para await
        koraAsync.Spawn(p.patrol)
    })
}
```

- `async fn` → função Go que agenda no `core/async/scheduler`
- `await expr` → receive em channel `<-ch`
- `spawn fn()` → `koraAsync.Spawn(fn)`
- `cancel task` → `koraAsync.Cancel(task)`

> **Regra de ouro:** Nunca bloquear a goroutine do game loop do Ebiten. Toda operação lenta é agendada no scheduler do `core/async/`.

---

## Fase 5: Emitter (`compiler/emitter/`)

**Responsabilidade:** Receber o AST verificado + asyncMap e produzir código Go válido como string.

> ⚠️ **Área de alto risco:** Erros no emitter geram Go inválido **silenciosamente** — o compilador não detecta. Sempre rodar `go build ./...` após qualquer mudança no emitter.

### Convenções do código gerado
- Prefixo `kora` em todas as variáveis de runtime injetadas: `koraAudio`, `koraPhysics`, `koraScene`, etc.
- Lifecycle hooks mapeados: `create` → `Create()`, `update` → `Update(dt float64)`, `draw` → `Draw(screen *ebiten.Image)`
- Objects KScript → structs Go com receiver pointer: `type Player struct { ... }`
- Fields privados (`_field`) → campos Go minúsculos (`field`)
- Constantes (`MAX_SPEED`) → `const MaxSpeed = ...` (PascalCase Go)

### Estrutura do arquivo `.ks.go` gerado
```go
// Code generated by KScript compiler. DO NOT EDIT.
package game

import (
    "github.com/ElioNeto/kora/core/audio"
    "github.com/ElioNeto/kora/core/async"
    // ... demais imports necessários
)

// --- gerado a partir de Player.ks ---
type Player struct { ... }
func (p *Player) Create() { ... }
func (p *Player) Update(dt float64) { ... }
```

---

## Como Adicionar um Novo Built-in (Resumo)

Ver skill completa em `.claude/skills/add-kscript-builtin.md`. Em resumo:
1. Implementar runtime em `core/<modulo>/`
2. Registrar assinatura no checker (`compiler/checker/builtins.go` ou equiv.)
3. Adicionar case no emitter para gerar `kora<Modulo>.FuncName(...)`
4. Criar fixture de teste `compiler/testdata/<Modulo>_test.ks`
5. Rodar `go build ./...` para validar o `.ks.go` gerado
6. Atualizar `docs/API_REFERENCE.md`

---

## Erros Comuns e Como Diagnosticar

| Sintoma | Causa provável | Como investigar |
|---------|---------------|------------------|
| `go build` falha em `*.ks.go` | Emitter gerou Go inválido | Ver saída do emitter; checar `case` correspondente em `emitter/` |
| Checker aceita tipo errado | Built-in não registrado ou registro incorreto | Verificar `compiler/checker/builtins.go` |
| `await` não gera channel | Transform não mapeou a função como async | Verificar `asyncMap` em `compiler/transform/` |
| Frame drop no jogo | Função `async` bloqueia o game loop | Não usar blocking calls — usar `core/async/scheduler` |
| `*.ks.go` desatualizado | Build antigo, não rodou `kora build` | Rodar `go run ./cmd/kora build` |
