# Skill: audit-and-fix
## Auditar e corrigir qualidade de código em um arquivo Go ou KScript

### Objetivo
Rodar o pipeline completo de qualidade (fmt, vet, build, test) em um arquivo ou pacote específico do Kora Engine, identificar todos os problemas e corrigi-los automaticamente quando possível, reportando o que requer intervenção manual.

---

### Quando usar esta skill
- Antes de criar um Pull Request
- Após editar um arquivo sensível de `core/` ou `compiler/`
- Quando o CI estiver falhando
- Para limpar dívida técnica em um pacote específico

---

### Passos que a IA deve seguir

**Passo 1 — Identificar o alvo**
O usuário deve informar:
- Caminho do arquivo ou pacote (ex: `core/physics/`, `compiler/checker/checker.go`)
- Escopo: apenas linting, ou linting + testes?

**Passo 2 — Verificação de formatação (auto-corrigível)**
```bash
# Verificar quais arquivos estão fora do padrão gofmt
gofmt -l <caminho>/

# Corrigir automaticamente
gofmt -w <caminho>/
```
Se houver arquivos listados → corrigir. Reportar quais foram alterados.

**Passo 3 — Análise estática com go vet**
```bash
go vet ./<caminho>/...
```
Para cada erro reportado:
- Se for **shadow de variável**: renomear a variável shadoweada
- Se for **unreachable code**: remover o código morto
- Se for **printf format mismatch**: corrigir o formato
- Se for **compositelit sem chaves**: adicionar nomes de campos
- Se não souber corrigir: marcar com comentário `// TODO(audit): <descrição do problema>`

**Passo 4 — Verificar compilação do pacote**
```bash
go build ./<caminho>/...
```
Se falhar: identificar o erro, propor correção e re-rodar até compilar.

**Passo 5 — Rodar testes do pacote**
```bash
# Testes do pacote específico
go test -v -race ./<caminho>/...

# Se for um arquivo no compiler/
cd compiler && go test -v -race ./<subpacote>/...
```
Para cada teste falhando:
- Ler o output de erro completo
- Verificar se é falha de lógica ou de setup
- Se for setup (mock inválido, fixture ausente): corrigir o helper de teste
- Se for lógica: reportar para o usuário com análise do problema (não corrigir silenciosamente)

**Passo 6 — Verificações específicas do Kora**
Verificar manualmente no arquivo:
- [ ] Há `import "C"` (cgo proibido)?
- [ ] Há `localStorage` ou `sessionStorage` em arquivos `.js` do editor?
- [ ] Há edição manual de arquivos `*.ks.go`?
- [ ] Há `runtime.GC()` chamado dentro de `Update()` ou `Draw()` do Ebiten?
- [ ] Exported symbols sem comentário godoc?
- [ ] Type assertions sem verificação de `ok` (`v := x.(Tipo)`)?
- [ ] Erros retornados sem `%w` wrapping?

**Passo 7 — Relatório final**
Ao finalizar, produzir um relatório estruturado:

```
## Relatório de Auditoria: <caminho>

### ✅ Corrigido automaticamente
- [lista de correções aplicadas]

### ⚠️ Requer revisão manual
- [problema] em [arquivo:linha] — [motivo não pôde ser auto-corrigido]

### 📊 Resultado dos testes
- X/Y testes passando
- Falhas: [lista]

### 🎯 Próximos passos recomendados
- [ação prioritária]
```

---

### Comandos de Terminal Permitidos

```bash
# Formatação
gofmt -l <caminho>/
gofmt -w <caminho>/

# Análise estática
go vet ./<caminho>/...

# Build
go build ./<caminho>/...

# Testes com race detector
go test -v -race ./<caminho>/...

# Cobertura de testes
go test -cover ./<caminho>/...

# Testes do compilador KScript
cd compiler && go test -v -race ./...

# Verificação do projeto inteiro (para confirmar sem regressão)
go build ./...
go test ./...

# Lint via Makefile
make lint
make format
```

---

### Restrições
- **Nunca** modificar arquivos `*.ks.go` (gerados pelo transpiler)
- **Nunca** corrigir silenciosamente uma falha de lógica de teste — reportar ao usuário
- **Nunca** remover um teste que está falhando sem aprovação explícita
- **Nunca** usar `//nolint` para suprimir warnings sem justificativa no comentário

---

### Exemplo de invocação
> "Use a skill audit-and-fix no pacote `core/physics/`."  
> "Rode audit-and-fix em `compiler/checker/checker.go` e corrija tudo que for possível."
