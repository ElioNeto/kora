# Kora Editor Desktop

> O editor visual desktop oficial da Kora Engine - Criado com Electron

## 📱 Sobre

Kora Editor Desktop é o aplicativo desktop completo para criação de jogos com a engine Kora. Baseado no editor web original, oferece:

- ✅ **Acesso completo ao sistema de arquivos** - Abrir, salvar, importar assets
- ✅ **Menus nativos** - File, Edit, View, Help com atalhos completos
- ✅ **Preview integrado** - Teste cenas em tempo real com física
- ✅ **Exportação APK** - Build direto para Android
- ✅ **Persistência** - Assets salvos no IndexedDB entre sessões
- ✅ **Offline** - Funciona sem internet

## 🎯 Funcionalidades

### Editor Visual
- **Cena 2D** - Canvas com zoom, pan e grid
- **Entidades** - Sprite, Tilemap, Camera, AudioEmitter, Custom
- **Inspector** - Edite propriedades em tempo real
- **Hierarquia** - Tree de entidades com visibilidade
- **Console** - Logs e mensagens de debug

### Assets Management
- **Importar** - PNG, JPG, WebP, GIF, OGG, WAV, MP3
- **Drag & Drop** - Do sistema ou entre abas
- **Thumbnail** - Preview de imagens no painel
- **Delete** - Botão × flutuante nos cards
- **Filtros** - Imagens, Áudio, Tilemaps
- **Persistência** - IndexedDB (salva entre sessões)

### KScript
- **Exportar** - Gera .ks executável
- **Integração** - Compilador embutido
- **Preview** - Roda cena localmente antes de exportar

### Build APK
- **Build Debug** - Rápido, sem assinatura
- **Build Release** - Com keystore para produção
- **Log** - Console de build integrado

## 🚀 Instalação

### Pré-requisitos

- **Node.js 18+**
- **npm** ou **yarn**
- **Go 1.22+** (para build APK)

### Desenvolvimento

```bash
# Clone o repositório
git clone https://github.com/koraengine/kora.git
cd kora/apps/desktop

# Instale dependências
npm install

# Inicie em modo desenvolvimento
npm run dev
```

### Build

```bash
# Build para distribuição
npm run build

# Platform-specific
npm run build:win   # Windows (.exe, portable)
npm run build:mac   # macOS (.dmg, .app)
npm run build:linux # Linux (.AppImage, .deb)
```

Os artefatos são gerados em `dist-electron/`.

## 📖 Uso

### Criando uma Cena

1. **Novo Projeto**: `File → New` ou `Ctrl+N`
2. **Importar Assets**: Botão `+ Importar` ou arraste arquivos
3. **Adicionar Entidade**: Arraste do Assets para o Canvas
4. **Posicionar**: Clique e arraste no canvas
5. **Editar**: Use o Inspector à direita
6. **Salvar**: `Ctrl+S` (.kora.json)
7. **Preview**: Pressione `F5` ou clique "Preview"
8. **Exportar KScript**: Botão `.ks`

### Atalhos

| Tecla | Ação |
|-------|------|
| `Ctrl+N` | Nova cena |
| `Ctrl+O` | Abrir cena |
| `Ctrl+S` | Salvar cena |
| `F5` | Preview |
| `Delete` | Excluir entidade |
| `V` | Tool: Select |
| `G` | Tool: Move |
| `F` | Zoom Fit |
| `F12` | DevTools |

### Menus

```
Arquivo
├── Nova Cena (Ctrl+N)
├── Abrir (Ctrl+O)
├── Salvar (Ctrl+S)
├── Salvar Como... (Ctrl+Shift+S)
├── Exportar KScript (.ks)
├── Exportar APK...
└── Sair

Editar
├── Undo (Ctrl+Z)
├── Redo (Ctrl+Y)
├── Cut (Ctrl+X)
├── Copy (Ctrl+C)
├── Paste (Ctrl+V)
└── Select All

Exibir
├── Zoom In (Ctrl+=)
├── Zoom Out (Ctrl+-)
├── Zoom Reset (Ctrl+0)
└── Toggle DevTools (F12)

Ajuda
├── Documentação
├── Reportar Problema
└── Sobre Kora Editor
```

## 🔧 APIs Electron

O aplicativo expõe APIs via `window.electronAPI`:

### Sistema de Arquivos

```javascript
// Abrir arquivo
const filePaths = await window.electronAPI.selectFile({
  filters: [{ name: 'Images', extensions: ['png', 'jpg'] }],
  properties: ['openFile', 'multiSelections']
})

// Salvar arquivo
const savePath = await window.electronAPI.saveFile({
  defaultPath: 'minha-cena.kora.json',
  filters: [{ name: 'Kora Scenes', extensions: ['kora.json'] }]
})

// Ler conteúdo
const { content, error } = await window.electronAPI.readFile(filePath)

// Escrever conteúdo
await window.electronAPI.writeFile(filePath, content)

// Binary (base64)
const { base64, error } = await window.electronAPI.readBinaryFile(filePath)
```

### Janelas

```javascript
// Minimizar
await window.electronAPI.minimize()

// Maximizar
await window.electronAPI.maximize()

// Restaurar
await window.electronAPI.unmaximize()

// Checar estado
const isMaximized = await window.electronAPI.isMaximized()
```

### Eventos

```javascript
// Ouvir eventos
const unsubscribe = window.electronAPI.on('menu:new-scene', () => {
  console.log('Novo projeto solicitado')
  // Handler
})

// Limpar listener
unsubscribe()
```

### Apps Data

```javascript
// User data path
const userData = await window.electronAPI.getAppPath('userData')

// Logs path
const logs = await window.electronAPI.getAppPath('logs')
```

## 🏗️ Arquitetura

```
apps/desktop/
├── src/
│   ├── main/                # Electron Main Process
│   │   ├── index.js         # Janela, menu, IPC handlers
│   │   └── preload.js       # Bridge seguro para renderer
│   │
│   └── renderer/            # Electron Renderer
│       ├── index.html       # UI principal
│       ├── assets/
│       │   └── style.css    # Estilos do editor
│       └── vite.config.js   # Config de build
│
├── package.json             # Dependencies & scripts
├── vite.config.js           # Vite configuration
└── README.md
```

### Processos

**Main Process** (Node.js)
- Gerencia janela principal
- Menus nativos do SO
- File dialogs do sistema
- IPC communication
- Build APK (via spawn)

**Renderer Process** (Web)
- UI do editor (via HTML/CSS/JS)
- Canvas 2D rendering
- Asset management
- State management
- Integração com Electron APIs

### IPC (Inter-Process Communication)

```javascript
// Main → Renderer
mainWindow.webContents.send('menu:new-scene', data)

// Renderer → Main
window.electronAPI.selectFile(options).then(path => { ... })
```

## 🔨 Build System

### Scripts NPM

| Comand | Descrição |
|--------|-----------|
| `npm run dev` | Inicia app em modo dev |
| `npm run build` | Build all platforms |
| `npm run build:win` | Build Windows |
| `npm run build:mac` | Build macOS |
| `npm run build:linux` | Build Linux |
| `npm run preview` | Preview build |

### Electron Builder Config

```json
{
  "appId": "com.koraengine.editor",
  "productName": "Kora Editor",
  "files": ["dist/**/*", "src/main/**/*"],
  "extraResources": [{
    "from": "../editor",
    "to": "editor",
    "filter": ["**/*"]
  }],
  "win": {
    "target": ["nsis", "portable"]
  },
  "mac": {
    "target": ["dmg", "zip"],
    "category": "public.app-category.games"
  },
  "linux": {
    "target": ["AppImage", "deb"]
  }
}
```

## 📦 Pack Content

O app inclui:

- Editor completo (do `editor/` directory)
- Assets panel com IndexedDB
- Serializer (JSON ↔ KScript)
- Preview panel com physics
- Estilos CSS

Todas as dependências são empacotadas no binário final.

## 🐛 Troubleshooting

### App não abre

```bash
# Verificar Node
node -v  # deve ser >= 18

# Reinstalar
rm -rf node_modules
npm install

# Rebuild
npm run build
```

### Arquivos não salvam

```bash
# Verificar permissões
ls -la ~/.config/com.koraengine.editor/

# Resetar config
rm ~/.config/com.koraengine.editor/config.json
```

### Assets não aparecem

```javascript
// Check console logs
// Error pode ser: file access denied
// Ou: IndexedDB quota exceeded
```

### Build falha

```bash
# Deps
apt-get install libsecret-1-dev  # Linux

# macOS
xcode-select --install

# Windows
# Instale Visual Studio Build Tools
```

## 🚧 Roadmap

### v0.2
- [ ] KScript editor com syntax highlighting
- [ ] Plugin system
- [ ] Asset optimization (sprite sheets)
- [ ] Version control (Git integration)

### v0.3
- [ ] Multi-window (Inspector separado)
- [ ] Cloud sync (Google Drive, OneDrive)
- [ ] Asset import wizard
- [ ] Scene template browser

### v1.0
- [ ] Auto-update mechanism
- [ ] Marketplace de assets/templates
- [ ] Tutorial integrado
- [ ] Analytics (opt-in)

## 📚 Recursos

- [Documento Principal](../../README.md)
- [KScript Guide](../../docs/SCRIPT.md)
- [Editor Guide](../../docs/EDITOR_GUIDE.md)
- [Architecture](../../docs/ARCHITECTURE.md)

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/koraengine/kora/issues)
- **Docs**: [Documentation](https://koraengine.dev/docs)
- **Email**: support@koraengine.dev

## 📄 License

MIT License — see [../../LICENSE](../../LICENSE)

---

**Kora Editor Desktop** - Crie jogos incríveis para Android com poder nativo!
