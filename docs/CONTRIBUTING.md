# Guia de Contribuição — Kora Engine

Bem-vindo ao Kora! Este guia explica como contribuir com a engine, editor desktop e compilador KScript.

## Como Contribuir

1. **Reporte bugs**: Abra uma issue descrevendo o problema
2. **Sugira features**: Abra uma issue com a proposta
3. **Faça um fork**: Clone o repositório e crie uma branch
4. **Submeta um PR**: Faça um pull request com suas mudanças

## Estrutura do Projeto

```
kora/
├── editor/              # Editor web (HTML/JS puro) - desenvolvimento rápido
├── apps/
│   └── desktop/         # Editor Desktop (Electron) - aplicativo nativo
│       ├── src/
│       │   ├── main/    # Processo principal Electron
│       │   └── renderer/# Interface do editor
│       ├── package.json # Dev Dependencies
│       └── vite.config.js
├── compiler/            # KScript → Go compiler
├── core/                # Runtime Go
│   ├── render/         # 2D renderer
│   ├── physics/        # AABB collision
│   ├── input/          # Input system
│   ├── async/          # Task scheduler
│   └── assets/         # Asset loader
├── cmd/                 # CLI commands
├── android/             # Build pipeline APK (gomobile)
├── examples/            # Cenas e jogos de exemplo
└── docs/                # Documentação
```

## Desenvolvimento

### Editor Web (Rápido Prototyping)

```bash
# Iniciar servidor local
cd editor
python -m http.server 8080
# Acesse: http://localhost:8080

# Ou usar Node.js
npx serve editor --port 8080
```

### Editor Desktop (Produção)

```bash
# Instalare dependências
cd apps/desktop
npm install

# Modo desenvolvimento (com DevTools)
npm run dev

# Build para distribuição
npm run build
npm run build:win  # Windows
npm run build:mac  # macOS
npm run build:linux # Linux
```

### Componentes do Editor

| Arquivo/Categoria | Descrição |
|------------------|-----------|
| `editor/editor.js` | Core do editor (state, render, entities) |
| `editor/assets-panel.js` | Importação, IndexedDB, drag-drop |
| `editor/serializer.js` | Serialização JSON ↔ KScript |
| `editor/idb.js` | Wrapper IndexedDB |
| `editor/preview-panel.js` | Preview com physics |
| `editor/style.css` | Estilos globais |
| `apps/desktop/src/main/` | Electron main process (menus, filesystem) |
| `apps/desktop/src/renderer/` | UI renderizada (HTML/CSS) |

### Compilador KScript

```bash
cd compiler
go test ./...          # Run tests
go build -o ../bin/kora-compiler .  # Build
```

Para testar compilação:
```bash
../bin/kora-compiler exemplo.ks > generated.go
go run generated.go
```

### Runtime Go

```bash
# Tests
go test ./core/...

# Run example
go run main.go examples/

# Build runtime
go build -o bin/kora-runtime ./cmd
```

## Desktop App - Desenvolvimento Electron

### Arquitetura

```
Main Process (Node.js)
├── janela principal
├── menu nativo
├── file dialogs
└── IPC handlers

Renderer Process
└── Editor UI (HTML/CSS/JS)
```

### APIs Disponíveis (window.electronAPI)

```javascript
// Sistema de arquivos
await window.electronAPI.selectFile(options)
await window.electronAPI.saveFile(options)
const { content } = await window.electronAPI.readFile(path)
await window.electronAPI.writeFile(path, content)

// Save cena
await window.electronAPI.saveScene(data, filename)

// Events
window.electronAPI.on('menu:new-scene', handler)

// Window
await window.electronAPI.minimize()
await window.electronAPI.maximize()
```

## Convenções de Commit

Usamos **Conventional Commits**:

```
tipo(escopo): mensagem descritiva (closes #X)
```

**Exemplos**:
```
feat(editor): adicionar drag-drop de assets para canvas
fix(desktop): corrigir crash ao abrir arquivo binário
docs: atualizar README com comando de build
refactor(compiler): melhorar parser de operadores
chore: atualizar dependências Node
test(core): adicionar test de physics engine
```

**Tipos**:
- `feat`: nova funcionalidade
- `fix`: correção de bug
- `docs`: atualização de documentação
- `refactor`: reestruturação de código
- `chore`: manutenção, dependências
- `test`: adição de testes

**Escopos**:
- `editor`: editor web
- `desktop`: desktop Electron
- `compiler`: KScript compiler
- `runtime`: runtime Go
- `core`: core da engine
- `android`: build Android
- `docs`: documentação

## Testes

```bash
# Todos os testes Go
go test ./... -v

# Testes do compiler
cd compiler && go test -v

# Editor (manual)
abrir editor/index.html ou apps/desktop

# Testes JS (futuro)
npm test  # em apps/desktop
```

## Code Review Checklist

Para PRs, verifique:

- [ ] Código segue padrões Go/JavaScript
- [ ] Testes adicionados se aplicável
- [ ] Documentação atualizada
- [ ] Changelog ou mensagem de release
- [ ] Sem breaking changes (ou justificado)
- [ ] Compatibilidade retroativa mantida

## Build & Release

### Build Desktop

```bash
cd apps/desktop

# All platforms
npm run build

# Platform-specific
npm run build:win   # nsis, portable
npm run build:mac   # dmg, zip
npm run build:linux # AppImage, deb
```

### Build APK

```bash
# Requer ANDROID_HOME configurado
export ANDROID_HOME=$HOME/Android/Sdk

# Debug (rápido)
./android/build.sh debug

# Release (com assinatura)
./android/build.sh release
```

## Publicação

### Releases Desktop

```bash
# Na pasta apps/desktop
npm run build

# Push artefatos para GitHub Releases
# via electron-builder configurado
```

### Versões

- **0.x.x**: Pré-release (breaking changes possíveis)
- **1.0.0+**: Release estável

Formato: `MAJOR.MINOR.PATCH`

## Contribuindo com KScript

Para adicionar funcionalidades à linguagem:

1. **Add parser rule** em `compiler/parser/`
2. **Add type checker** em `compiler/checker/`
3. **Add Go emitter** em `compiler/emitter/`
4. **Add API doc** em `docs/SCRIPT.md`
5. **Add test case** em `compiler/parser/testdata/`

## Recursos em Desenvolvimento

| Feature | Componente | Status |
|---------|------------|--------|
| Asset optimization | editor | v1.1 |
| Plugin system | desktop | v1.2 |
| KScript debugger | compiler | v1.3 |
| Multi-window | desktop | planned |
| Asset marketplace | desktop | roadmap |

## Código de Conduta

Seja respeitoso e profissional. Colaboração é a chave!

Veja [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) para detalhes.

---

**Obrigado por contribuir com o Kora Engine!** 🎉
