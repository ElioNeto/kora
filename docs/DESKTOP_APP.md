# Kora Editor Desktop

## Visão

Kora Editor Desktop é o aplicativo desktop oficial da Engine Kora. Transforma o editor web em um aplicativo nativo usando Electron, proporcionando melhor desempenho, acesso ao sistema de arquivos e funcionalidades desktop completas.

## Arquitetura

```
Kora Editor Desktop
├── Main Process (Node.js/Electron)
│   ├── Gerenciamento de janelas
│   ├── Sistema de arquivos
│   ├── Menu nativo
│   └── IPC communication
│
├── Renderer Process (Electron)
│   ├── UI do Editor
│   ├── Canvas 2D
│   └── Renderização dos assets
│
└── Shared
    ├── editor/ (código existente)
    └── compiler/ (compilador KScript)
```

## Estrutura do Projeto

```
apps/desktop/
├── src/
│   ├── main/
│   │   ├── index.js          # Electron main process
│   │   └── preload.js        # Preload script (API bridge)
│   └── renderer/
│       ├── index.html        # Electron window
│       ├── assets/
│       │   └── style.css     # Editor styles
│       └── vite.config.js    # Vite config
├── package.json
├── vite.config.js
├── src/renderer/index.html   # Main renderer file
└── .gitignore
```

## Funcionalidades

### Sistema de Arquivos
- Abrir/Salvar cenas (.kora.json)
- Importar assets (PNG, JPG, WebP, OGG, WAV)
- Exportar KScript (.ks)
- Exportar APK (integrado com build system)

### Menu Nativo
- **Arquivo**: Nova, Abrir, Salvar, Exportar
- **Editar**: Undo, Redo, Cut, Copy, Paste
- **Exibir**: Zoom, Developer Tools
- **Ajuda**: Documentação, Reportar Problema, Sobre

### Window Controls
- Minimize/Maximize/Close nativos do sistema
- Salvar bounds da janela entre sessões
- Suporte multi-monitor

### Integração
- Compilador KScript
- Build system Android
- Asset cache (IndexedDB)

## Build & Develop

### Pré-requisitos
- Node.js 18+
- npm ou yarn
- Electron global (opcional para develop)

### Instalação

```bash
cd apps/desktop
npm install
```

### Desenvolvimento

```bash
npm run dev
```

Inicia o Electron app em modo desenvolvimento com DevTools.

### Build

```bash
# Build all platforms
npm run build

# Windows
npm run build:win

# macOS
npm run build:mac

# Linux
npm run build:linux
```

### Instalação no Dispositivo Android

```bash
# Após build APK
adb install dist-electron/Kora-Editor-[version].apk
```

## API Electron

### System Files

```javascript
// Open dialog
await window.electronAPI.selectFile({
  filters: [{ name: 'JSON', extensions: ['kora.json'] }]
})

// Save dialog
await window.electronAPI.saveFile({
  defaultPath: 'cena.kora.json',
  filters: [{ name: 'JSON', extensions: ['kora.json'] }]
})

// Read file
const { content } = await window.electronAPI.readFile(filePath)

// Write file
await window.electronAPI.writeFile(filePath, content)

// Binary file
const { base64 } = await window.electronAPI.readBinaryFile(filePath)
```

### Application

```javascript
// Get app path
await window.electronAPI.getAppPath('userData')

// Show in folder
await window.electronAPI.showItemInFolder(filePath)

// Open external
await window.electronAPI.openExternal('https://koraengine.dev')

// Window controls
await window.electronAPI.minimize()
await window.electronAPI.maximize()
await window.electronAPI.unmaximize()
```

### Events

```javascript
// Listen to events
const unsubscribe = window.electronAPI.on('menu:new-scene', () => {
  // Handler
})

unsubscribe() // Clean up
```

## Customization

### Themes
O editor suporta theme dark/light baseado na configuração do sistema.

### Menus
Customize menus em `src/main/index.js`:

```javascript
const template = [
  {
    label: 'Meu Menu',
    submenu: [
      { label: 'Opção 1', click: () => {} },
      { label: 'Opção 2', click: () => {} }
    ]
  }
]
Menu.buildFromTemplate(template)
```

### Plugins
Future: Sistema de plugins via JavaScript/TypeScript.

## Roadmap

### v0.2
- [ ] Plugin system
- [ ] Asset import wizard
- [ ] Scene template browser
- [ ] Cloud sync (Google Drive, OneDrive)

### v0.3
- [ ] Asset optimization (sprite sheets, compression)
- [ ] Version control integration (Git)
- [ ] Multi-window (Inspector, Console separados)
- [ ] Dark/Light theme toggle

### v1.0
- [ ] Auto-update mechanism
- [ ] Analytics (opt-in)
- [ ] Marketplace de assets
- [ ] Tutorial integrado

## Troubleshooting

### Build fails
```bash
# Clear and reinstall
rm -rf node_modules
npm install
```

### App won't start
```bash
# Check logs
electron --enable-logging

# Reset config
rm ~/.config/kora-editor/config.json
```

### Assets don't load
Check console for file access errors:
```javascript
console.error('File access:', err.message)
```

## License

MIT License - See [LICENSE](../../LICENSE)
