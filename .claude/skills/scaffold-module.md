# Skill: scaffold-module
## Criar um novo módulo do Kora Engine

### Objetivo
Criar um novo módulo (pacote Go) dentro de `core/` seguindo estritamente a estrutura de pastas, nomenclatura e padrões do Kora Engine. Garante que todo módulo novo já nasça com arquivo principal, interface pública, e arquivo de testes correspondente.

---

### Quando usar esta skill
- Criar um novo subsistema no engine (ex: `core/particles/`, `core/network/`, `core/save/`)
- Criar um novo passo no pipeline do compilador (ex: `compiler/optimizer/`)
- Criar um novo comando CLI em `cmd/`

---

### Passos que a IA deve seguir

**Passo 1 — Coletar informações**
Perguntar ao usuário:
1. Nome do módulo (ex: `particles`) — será o nome do pacote e do diretório
2. Qual área: `core/`, `compiler/` ou `cmd/`?
3. Qual é a responsabilidade principal (1 frase)?
4. Quais structs/tipos públicos este módulo vai expor?

**Passo 2 — Validar nomenclatura**
- Nome do diretório: `snake_case` (ex: `save_system` → pasta `save_system/`)
- Nome do pacote Go: igual ao diretório sem underscores (`savesystem`)
- Structs exportadas: `PascalCase` (ex: `SaveManager`, `SaveSlot`)
- Interfaces: sem prefixo `I` (ex: `Saver`, não `ISaver`)
- Constantes: `UPPER_SNAKE_CASE`
- Campos privados: `camelCase` prefixo sem underscore

**Passo 3 — Criar a estrutura de arquivos**

Criar os seguintes arquivos:

```
core/<nome>/
├── <nome>.go          ← implementação principal
├── <nome>_test.go     ← testes obrigatórios (Go stdlib testing)
└── doc.go             ← comentário de pacote (godoc)
```

**Passo 4 — Conteúdo padrão de cada arquivo**

`core/<nome>/doc.go`:
```go
// Package <nome> <descrição de uma linha do módulo>.
package <nome>
```

`core/<nome>/<nome>.go`:
```go
package <nome>

import (
    // stdlib
    // externos (github.com/...)
    // internos (github.com/ElioNeto/kora/...)
)

// <NomeStruct> <descrição pública obrigatória>.
type <NomeStruct> struct {
    // campos privados aqui
}

// New<NomeStruct> cria e retorna uma nova instância de <NomeStruct>.
func New<NomeStruct>() *<NomeStruct> {
    return &<NomeStruct>{}
}
```

`core/<nome>/<nome>_test.go`:
```go
package <nome>

import "testing"

func TestNew<NomeStruct>(t *testing.T) {
    s := New<NomeStruct>()
    if s == nil {
        t.Fatal("New<NomeStruct>() retornou nil")
    }
}
```

**Passo 5 — Se for módulo `core/`, integrar ao engine**
- Verificar se `core/engine/engine.go` precisa importar o novo módulo
- Se sim, adicionar o import e inicializar no método `New()` ou `Init()`
- **Nunca** adicionar `import "C"` (proibido cgo)

**Passo 6 — Verificar compilação**
Rodar os comandos de verificação abaixo e corrigir qualquer erro antes de finalizar.

---

### Comandos de Terminal Permitidos

```bash
# Verificar se o módulo compila
go build ./core/<nome>/...

# Rodar os testes do novo módulo
go test -v ./core/<nome>/...

# Verificar vet no pacote
go vet ./core/<nome>/...

# Verificar formatação
gofmt -l ./core/<nome>/

# Corrigir formatação automaticamente
gofmt -w ./core/<nome>/

# Verificar que o projeto inteiro ainda compila
go build ./...

# Rodar todos os testes para garantir sem regressão
go test ./...
```

---

### Restrições
- **Proibido** usar `import "C"` ou qualquer cgo
- **Proibido** usar frameworks de teste externos (testify, gomock, etc.)
- **Obrigatório** que todo arquivo `.go` tenha comentário godoc no exported symbol
- **Obrigatório** criar `_test.go` correspondente junto com o arquivo principal
- Imports Go organizados em 3 grupos: stdlib / externos / internos

---

### Exemplo de invocação
> "Use a skill scaffold-module para criar o módulo `particles` em `core/` responsável por gerenciar sistemas de partículas 2D, com a struct `ParticleEmitter`."
