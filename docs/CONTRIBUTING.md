# Guia de Contribuição — Kora Engine

Bem-vindo ao Kora! Este guia explica como contribuir para o projeto.

## Como Contribuir

1. **Reporte bugs**: Abra uma issue descrevendo o problema
2. **Sugira features**: Abra uma issue com a proposta
3. **Faça um fork**: Clone o repositório e crie uma branch
4. **Submeta um PR**: Faça um pull request com suas mudanças

## Estrutura do Código

```
kora/
├── editor/         # Editor visual (HTML/JS puro)
├── compiler/       # KScript compiler
├── core/           # Runtime Go (render, physics, scene, async)
├── android/        # Build pipeline para APK
└── examples/       # Cenas de exemplo
```

## Desenvolvimento — Editor

Para desenvolver o editor:

```bash
# Iniciar servidor local
cd editor
python -m http.server 8080
# http://localhost:8080
```

### Componentes do Editor

| Arquivo | Descrição |
|---------|-----------|
| `index.html` | DOM principal |
| `editor.js` | Lógica do editor (entidades, canvas, render) |
| `assets-panel.js` | Importação, IndexedDB, drag-drop |
| `serializer.js` | Serialização JSON ↔ KScript |
| `idb.js` | Wrapper IndexedDB para persistência |
| `preview.html` | Runtime de preview com física |
| `style.css` | Estilos |

## Compilador KScript

Para testar mudanças no compiler:

```bash
cd compiler
go test ./...
go build -o ../bin/kora-compiler .
```

## Runtime Go

Para testar o runtime:

```bash
go test ./core/...
go run main.go examples/
```

## Convenções de Commit

Usamos convenções semânticas:

```
feat(editor): assets panel com drag-and-drop (closes #3)
fix(compiler): resolver bug com async em loop
docs: atualizar README com quickstart
refactor(core): melhorar scheduler de tasks
```

Formatação: `<tipo>(<escopo>): <mensagem>`

**Tipos**: `feat`, `fix`, `docs`, `refactor`, `chore`, `test`

## Testes

```bash
# Todos os testes
go test ./...

# Editor (verificar console)
abrir editor/index.html no navegador

# Compiler
cd compiler && go test -v
```

## Pull Requests

1. Branch isolada: `git checkout -b feat/asset-import`
2. Commit descritivo: `git commit -m "feat(assets): importar png, jpg, webp"`
3. PR para `main`: "feat(editor): assets panel com..."
4. Linkar issues: "(closes #X)" na mensagem

## Código de Conduta

Seja respeitoso e profissional. Colaboração é a chave!

---

**Obrigado por contribuir!**
