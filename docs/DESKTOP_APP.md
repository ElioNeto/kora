# Kora Editor Desktop - v0.3.0

> O aplicativo desktop oficial da Kora Engine - Construído com Electron

## 📱 Visão Geral

Kora Editor Desktop é o editor visual completo para criar jogos 2D nativos para Android. Transforma o editor web em um aplicativo desktop com:

- **Acesso completo ao sistema de arquivos** - Abrir, salvar, importar assets com dialogs nativos
- **Menus nativos** - File, Edit, View, Help com atalhos completos
- **Preview integrado** - Teste cenas em tempo real com física 2D
- **Exportação APK** - Build direto para Android via gomobile
- **Assets persistentes** - IndexedDB salva assets entre sessões
- **Offline First** - Funciona completamente offline

## 🏗️ Arquitetura

### Estrutura de Processos

```
┌─────────────────────────────────────────────────────────┐
│              Kora Editor Desktop                          │
├─────────────────────────────────────────────────────────┤
│                  Main Process                             │
│              (Node.js + Electron)                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Window Manager           Menu Manager            │  │
│  │  File Dialogs             IPC Handlers           │  │
│  │  APK Build System         Auto-updater           │  │
│  └───────────────────────────────────────────────────┘  │
│  │                                                      │
│  │  ← IPC Bridge →                                      │
│  │                                                      │
├─────────────────────────────────────────────────────────┤
│              Renderer Process                           │
│              (Chrome/Blink)                              │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Editor UI (HTML/CSS/JS)                           │  │
│  │  ├─ Scene Canvas (2D)                              │  │
│  │  ├─ Asset Management Panel                         │  │
│  │  ├─ Entity Inspector                               │  │
│  │  ├─ Hierarchy Tree                                 │  │
│  │  └─ Console Panel                                  │  │
│  │                                                      │
│  │  Dependencies:                                     │  │
│  │  ├─ serializer.js   (JSON ↔ KScript)               │  │
│  │  ├─ assets-panel.js (Import, IndexedDB)            │  │
│  │  ├─ editor.js       (Core logic)                   │  │
│  │  └─ idb.js          (IndexedDB wrapper)            │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Componentes

| Componente | Localização | Descrição |
|------------|-------------|-----------|
| **Main Process** | `src/main/index.js` | Window, menu, IPC, file access |
| **Preload** | `src/main/preload.js` | Secure API bridge for renderer |
| **Renderer** | `src/renderer/index.html` | Editor UI rendered by Chrome |
| **Editor Core** | `../../editor/` | Código do editor web (importado) |
| **Compiler** | `../../compiler/` | KScript compiler (integrated) |

## ✨ Funcionalidades

### Editor Visual

- **Cena 2D** - Canvas com zoom, pan, grid
- **Zoom** - 10% a 400% com wheel ou menu
- **Grid** - Grid snapping optional
- **Bounds** - Visualização da área lógica (360x640)
- **Tools** - Select (V), Move (G), Scale (S)

### Sistema de Entidades

| Tipo | Descrição |
|------|-----------|
| **Sprite** | Imagem/texture (PNG, JPG, WebP) |
| **Tilemap** | Tilemap/grid com collision |
| **Camera** | Camera viewport |
| **AudioEmitter** | Source de áudio (OGG, WAV) |
| **Custom** | Entidade genérica |

### Assets Management

- **Importação** - File dialog ou drag-and-drop
- **Formatos** - PNG, JPG, JPEG, WebP, GIF, SVG, OGG, WAV, MP3, JSON
- **Persistência** - IndexedDB salva entre sessões
- **Thumbnail** - Preview 56x56px para imagens
- **Filters** - Todos, Imagens, Áudio, Tilemaps
- **Delete** - Botão × flutuante (hover no card)
- **Context Menu** - Direito no card: Spawn, Rename, Delete

### KScript Integration

- **Exportar** - Gera .ks executável
- **Compile direto** - Via Electron spawn
- **Preview** - Testa cena localmente

### Build System

- **APK Debug** - Build rápido, sem assinatura
- **APK Release** - Com keystore para produção
- **Console de Build** - Logs no painel
- **Log** - Build steps visíveis

## 📂 Estrutura de Arquivos

```
apps/desktop/
├── src/
│   ├── main/                    # Electron Main Process
│   │   ├── index.js             # Window, menu, IPC, file handlers
│   │   └── preload.js           # Secure bridge (contextBridge)
│   │
│   └── renderer/                # Renderer Process
│       ├── index.html           # Main HTML file
│       ├── assets/
│       │   └── style.css        # Editor styles
│       └── vite.config.js       # Vite configuration
│
├── package.json                 # Dependencies & build config
├── vite.config.js               # Vite build config
└── README.md
                    
```

### Principais Arquivos

#### src/main/index.js
```javascript
// Window creation
createWindow()

// Menu template
createMenu()

// IPC handlers
ipcMain.handle('select-file', ...)
ipcMain.handle('save-scene', ...)
ipcMain.handle('build-apk', ...)
```

#### src/main/preload.js
```javascript
// Expose APIs to renderer
contextBridge.exposeInMainWorld('electronAPI', {
  selectFile: () => ipcRenderer.invoke(...),
  readFile: () => ipcRenderer.invoke(...),
  saveScene: () => ipcRenderer.invoke(...),
  buildAPK: () => ipcRenderer.invoke(...),
  on: (channel, func) => ipcRenderer.on(channel, fn),
})
```

## 🔧 APIs

### window.electronAPI

#### File System

```javascript
// Open file dialog
const filePaths = await window.electronAPI.selectFile({
  filters: [{ name: 'Kora Scenes', extensions: ['kora.json'] }],
  properties: ['openFile', 'multiSelections']
})

// Save dialog
const savePath = await window.electronAPI.saveFile({
  defaultPath: 'minha-cena.kora.json',
  filters: [{ name: 'Kora Scenes', extensions: ['kora.json'] }]
})

// Read text file
const { content, error } = await window.electronAPI.readFile('/path/to/file.txt')

// Write text file
const result = await window.electronAPI.writeFile('/path/to/file.txt', 'content')

// Read binary file (returns base64)
const { base64, error } = await window.electronAPI.readBinaryFile('/path/to/image.png')

// Check if file exists
const exists = await window.electronAPI.fileExists('/path/to/file.json')

// Special: Save scene Kora
const saveResult = await window.electronAPI.saveScene(
  sceneData,  // Object: { entities, meta }
  fileName     // String: 'scene.kora.json'
)
```

#### Application

```javascript
// Get app path
const userDataPath = await window.electronAPI.getAppPath('userData')
const logsPath = await window.electronAPI.getAppPath('logs')

// Operating system
await window.electronAPI.showItemInFolder('/path/to/file')
await window.electronAPI.openExternal('https://koraengine.dev')
```

#### Window Controls

```javascript
// Window states
await window.electronAPI.minimize()
await window.electronAPI.maximize()
await window.electronAPI.unmaximize()

// Check current state
const isMaximized = await window.electronAPI.isMaximized()
```

#### Build APK

```javascript
// Trigger APK build
const result = await window.electronAPI.buildAPK({
  androidHome: '/path/to/android/sdk',
  gradlePath: '/path/to/gradle',
  buildType: 'release'  // 'debug' or 'release'
})
```

### Eventos (IPC)

```javascript
// Listen to events from main process
const unsubscribe = window.electronAPI.on('menu:new-scene', () => {
  // Handler code
})

// Clear listener
unsubscribe()
```

## 🎨 Customization

### Custom Menu

Editor `src/main/index.js`:

```javascript
const menuTemplate = [
  {
    label: 'File',
    submenu: [
      {
        label: 'New Project',
        accelerator: 'CmdOrCtrl+N',
        click: () => mainWindow.webContents.send('menu:new-scene')
      },
      { type: 'separator' },
      {
        label: 'Export APK...',
        click: () => mainWindow.webContents.send('menu:export-apk')
      }
    ]
  },
  {
    label: 'View',
    submenu: [
      {
        label: 'Zoom In',
        accelerator: 'CmdOrCtrl+plus',
        click: () => {
          currentZoom = Math.min(3.0, currentZoom * 1.1)
          window.setZoomFactor(currentZoom)
        }
      }
    ]
  }
]

Menu.setApplicationMenu(Menu.buildFromTemplate(menuTemplate))
```

### Theme Toggle

O editor suporta dark/light themes baseado no sistema:

```javascript
// Theme detection
const osTheme = await window.electronAPI.getOSTheme() // 'dark' | 'light'

// Apply theme
document.documentElement.setAttribute(
  'data-theme',
  theme === 'dark' ? 'dark' : 'light'
)
```

### Plugin System (v0.2+)

Plugin hooks disponibles:

```javascript
// Hook: beforeExport
window.electronAPI.on('pre-export', (sceneData) => {
  // Modify sceneData before export
  return sceneData
})

// Hook: afterImport
window.electronAPI.on('post-import', (filePath) => {
  console.log(`Importado: ${filePath}`)
})
```

## 🚀 Build & Install

### Desenvolvimento

```bash
# Instale dependências
npm install

# Modo dev (com DevTools)
npm run electron:dev

# Apenas dev server
npm run dev
```

### Build

```bash
# Build all
npm run electron:build

# Platform-specific
npm run electron:build:win    # Windows (.exe, portable)
npm run electron:build:mac    # macOS (.dmg, .app)
npm run electron:build:linux  # Linux (.AppImage, .deb)
```

### Distribuição

Os builds gerados ficam em `dist-electron/`:

```
dist-electron/
├── Kora Editor Setup x64.exe    # Windows
├── Kora Editor.dmg              # macOS
└── Kora-Editor_0.3.0_amd64.deb  # Linux
```

### Instalação no Android

```bash
# Após build APK
cd android/app/build/outputs/apk/release

# Instalar em dispositivo
adb install kora-release.apk

# Instalar com overwite
adb install -r kora-release.apk

# Lista de pacotes instalados
adb shell pm list packages | grep kora
```

## 📋 Roadmap de Features

### v0.3 (Current)
- [x] Electron wrapper
- [x] Native menus
- [x] File dialogs
- [x] Asset management
- [x] IndexedDB persistence
- [x] KScript export
- [x] Build system integration
- [ ] KScript editor (in progress)

### v0.4
- [ ] KScript syntax highlighting
- [ ] Plugin system
- [ ] Asset optimization (sprite sheets)
- [ ] Version control (Git integration)
- [ ] Multi-window support

### v0.5
- [ ] Asset import wizard
- [ ] Scene template browser
- [ ] Cloud sync (Google Drive, OneDrive)
- [ ] Collaboration mode

### v1.0
- [ ] Auto-updater
- [ ] Marketplace de assets/templates
- [ ] Tutorial integrado
- [ ] Analytics (opt-in)
- [ ] Full IDE (debugger, profiler)

## 🔨 Troubleshooting

### App não inicia

```bash
# Verificar Node version
node -v  # Mínimo 18

# Reinstalar
rm -rf node_modules package-lock.json
npm install

# Limpar cache Electron
rm -rf ~/.cache/electron

# Logs
electron --enable-logging 2>&1 | tee electron.log
```

### Arquivos não salvam

```bash
# Verificar permissões
ls -la ~/.config/com.koraengine.editor/

# Reset config
rm ~/.config/com.koraengine.editor/config.json
rm ~/.config/com.koraengine.editor/*.db

# Restart app
```

### Assets não aparecem

```bash
# Verificar IndexedDB quota
# No DevTools Console:
indexedDB.databases()

# Limpar IndexedDB
indexedDB.deleteDatabase('kora-editor')
location.reload()

# Verificar console log
console.error('Asset load error:', err)
```

### Build APK falha

```bash
# Verificar ANDROID_HOME
echo $ANDROID_HOME

# Instalar gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Instalar Android SDK tools
sdkmanager "platform-tools" "platforms;android-33"

# Build manual
cd android
chmod +x build.sh
./build.sh debug
```

## 📚 Recursos

- [**Documentação Completa**](../../README.md)
- [**KScript Language**](./SCRIPT.md)
- [**API Reference**](./API_REFERENCE.md)
- [**Editor Guide**](./EDITOR_GUIDE.md)
- [**Architecture**](./ARCHITECTURE.md)
- [**Contributing**](./CONTRIBUTING.md)

## 💬 Suporte

- **GitHub Issues**: [Reportar Bug](https://github.com/koraengine/kora/issues)
- **Documentation**: [docs.koraengine.dev](https://docs.koraengine.dev)
- **Email**: support@koraengine.dev

## 📄 License

MIT License — see [../../LICENSE](../../LICENSE)

---

**Kora Editor Desktop v0.3.0**

Crie jogos incríveis para Android com editor visual e performance nativa!
