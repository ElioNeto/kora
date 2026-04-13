# Skill: add-kscript-builtin
## Adicionar um novo módulo built-in exposto ao KScript

### Objetivo
Esta é a tarefa mais complexa e recorrente do Kora Engine: **expandir a API disponível nos scripts KScript** adicionando um novo módulo built-in (ex: `Network`, `Storage`, `Tween`). O processo envolve 4 camadas que precisam ser atualizadas em sincronia: o runtime Go, o checker de tipos do compilador, o emitter de código e a documentação da API.

> ⚠️ **Área de alto risco**: erros no checker ou emitter geram código Go inválido silenciosamente. Seguir cada passo nesta ordem é obrigatório.

---

### Quando usar esta skill
- Adicionar um novo módulo global ao KScript (ex: `Network.get(url)`, `Storage.save(key, value)`)
- Expor uma funcionalidade existente do Go runtime para os scripts de jogo
- Atualizar a assinatura de uma função built-in já existente

---

### Passos que a IA deve seguir

**Passo 1 — Coletar especificação do módulo**
Perguntar ao usuário:
1. Nome do módulo (ex: `Storage`) — será `MODULO_NOME` neste guia
2. Quais funções expor? Para cada uma: nome, parâmetros com tipos KScript, tipo de retorno
3. Há estado interno (singleton) ou é stateless?

Exemplo de spec:
```
Storage.save(key: string, value: string): void
Storage.load(key: string): string?
Storage.delete(key: string): bool
Storage.exists(key: string): bool
```

**Passo 2 — Implementar o runtime Go em `core/`**

Criar `core/<nome_minusculo>/<nome_minusculo>.go`:
```go
package <nome_minusculo>

// <NomeModulo> implementa o módulo built-in <NomeModulo> do KScript.
type <NomeModulo> struct{}

// New<NomeModulo> retorna uma instância singleton de <NomeModulo>.
func New<NomeModulo>() *<NomeModulo> {
    return &<NomeModulo>{}
}

// <FuncName> <descrição da função>.
func (m *<NomeModulo>) <FuncName>(<params>) (<returnType>, error) {
    // implementação
}
```

Regras:
- **Zero cgo** — implementação Go puro
- Erros retornados como segundo valor de retorno
- Usar `fmt.Errorf("<nome>.<func>: %w", err)` para wrapping

**Passo 3 — Registrar no checker de tipos (`compiler/checker/`)**

Localizar o mapa de módulos built-in no checker (geralmente `builtins.go` ou similar) e adicionar a entrada do novo módulo com todas as funções e seus tipos KScript:

```go
// Em compiler/checker/builtins.go (ou arquivo equivalente)
"<NomeModulo>": BuiltinModule{
    Functions: map[string]FuncSignature{
        "<funcName>": {
            Params:  []KType{KTypeString, KTypeString},
            Returns: KTypeVoid,
        },
        // ... demais funções
    },
},
```

Mapeamento de tipos KScript → Go:
| KScript | Go |
|---------|----|
| `string` | `string` |
| `int` | `int` |
| `float` | `float64` |
| `bool` | `bool` |
| `void` | — (sem retorno) |
| `string?` | `(string, bool)` ou ponteiro |
| `Array<T>` | `[]T` |

**Passo 4 — Atualizar o emitter de código (`compiler/emitter/`)**

Localizar onde chamadas de módulos built-in são emitidas e adicionar o case para o novo módulo:

```go
// Em compiler/emitter/builtins_emitter.go (ou arquivo equivalente)
case "<NomeModulo>":
    return fmt.Sprintf("kora<NomeModulo>.%s(%s)", call.Method, argsCode)
```

Garantir que o import do runtime seja emitido nos arquivos `.ks.go` gerados:
```go
// Header de imports emitido
"github.com/ElioNeto/kora/core/<nome_minusculo>"
```

**Passo 5 — Criar fixtures de teste do compilador**

Criar arquivo `compiler/testdata/<NomeModulo>_test.ks` com exemplos válidos:
```kscript
object TestScene {
    async create() {
        <NomeModulo>.<funcName>("arg")
        var result = <NomeModulo>.<outraFunc>("key")
    }
}
```

**Passo 6 — Escrever testes para as 4 camadas**

```bash
# Testes do runtime Go
go test -v ./core/<nome_minusculo>/...

# Testes do checker (verificar que tipos são aceitos/rejeitados corretamente)
cd compiler && go test -v ./checker/...

# Testes do emitter (verificar que o código Go gerado compila)
cd compiler && go test -v ./emitter/...

# Pipeline completo: transpilar o fixture e verificar que compila
go run ./cmd/kora compile compiler/testdata/<NomeModulo>_test.ks
go build ./... # garante que o .ks.go gerado é Go válido
```

**Passo 7 — Atualizar a documentação**

Atualizar `docs/API_REFERENCE.md` adicionando a seção do novo módulo:
```markdown
## <NomeModulo>

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|------------|
| `<NomeModulo>.<func>` | ... | ... | ... |
```

**Passo 8 — Verificação final**

```bash
# Build completo
go build ./...
cd compiler && go build ./...

# Todos os testes
go test ./...
cd compiler && go test ./...

# Sem regressão no formato
make format
make lint
```

---

### Comandos de Terminal Permitidos

```bash
# Build e testes do runtime
go build ./core/<nome>/...
go test -v -race ./core/<nome>/...

# Build e testes do compilador
cd compiler && go build ./...
cd compiler && go test -v -race ./...
cd compiler && go test -v ./checker/...
cd compiler && go test -v ./emitter/...

# Transpilar fixture de teste
go run ./cmd/kora compile compiler/testdata/<Modulo>_test.ks

# Build completo (verifica que *.ks.go gerado é Go válido)
go build ./...
go test ./...

# Formatação e lint
make format
make lint
gofmt -w ./core/<nome>/
go vet ./...
```

---

### Restrições
- **Proibido** usar cgo em qualquer arquivo do runtime
- **Nunca** editar arquivos `*.ks.go` manualmente — eles são gerados
- **Obrigatório** atualizar as 4 camadas juntas: runtime, checker, emitter, docs
- **Obrigatório** rodar `go build ./...` após geração do `*.ks.go` para validar
- Se o checker aceitar um tipo errado, o emitter gerará Go inválido silenciosamente — sempre testar ambos

---

### Exemplo de invocação
> "Use a skill add-kscript-builtin para adicionar o módulo `Storage` com as funções `save(key, value)`, `load(key)` e `delete(key)`."
> "Use add-kscript-builtin para expor o módulo `Tween` com `to(target, duration, props)` e `cancel(id)`."
