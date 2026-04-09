/**
 * Kora Editor - Electron Main Process
 */
const { app, BrowserWindow, ipcMain, dialog, Menu, shell } = require('electron')
const path = require('path')
const fs = require('fs')

let mainWindow = null
let previewWindow = null

const isDev = process.env.NODE_ENV === 'development' || !app.isPackaged
const rendererPath = isDev ? 'http://localhost:5173' : path.join(__dirname, '../dist/index.html')

// ─── Menu nativo ────────────────────────────────────────────────────────────
function createMenu() {
  const template = [
    {
      label: 'Arquivo',
      submenu: [
        { label: 'Nova Cena',      accelerator: 'CmdOrCtrl+N',       click: () => mainWindow.webContents.send('menu:new-scene') },
        { type: 'separator' },
        { label: 'Abrir Cena...',  accelerator: 'CmdOrCtrl+O',       click: () => mainWindow.webContents.send('menu:open-scene') },
        { label: 'Salvar Cena',   accelerator: 'CmdOrCtrl+S',       click: () => mainWindow.webContents.send('menu:save-scene') },
        { label: 'Salvar Como...', accelerator: 'CmdOrCtrl+Shift+S', click: () => mainWindow.webContents.send('menu:save-scene-as') },
        { type: 'separator' },
        { label: 'Exportar KScript', accelerator: 'CmdOrCtrl+E',    click: () => mainWindow.webContents.send('menu:export-ks') },
        { label: 'Exportar APK...',                                  click: () => mainWindow.webContents.send('menu:export-apk') },
        { type: 'separator' },
        { label: 'Sair', accelerator: 'CmdOrCtrl+Q', click: () => app.quit() }
      ]
    },
    {
      label: 'Editar',
      submenu: [
        { role: 'undo',      label: 'Desfazer' },
        { role: 'redo',      label: 'Refazer' },
        { type: 'separator' },
        { role: 'cut',       label: 'Recortar' },
        { role: 'copy',      label: 'Copiar' },
        { role: 'paste',     label: 'Colar' },
        { role: 'selectAll', label: 'Selecionar Tudo' },
        { type: 'separator' },
        { label: 'Duplicar Entidade',  accelerator: 'CmdOrCtrl+D', click: () => mainWindow.webContents.send('menu:duplicate-entity') },
        { label: 'Excluir Entidade',   accelerator: 'Delete',       click: () => mainWindow.webContents.send('menu:delete-entity') }
      ]
    },
    {
      label: 'Exibir',
      submenu: [
        { label: 'Cena',           accelerator: 'CmdOrCtrl+1', click: () => mainWindow.webContents.send('menu:tab-scene') },
        { label: 'Assets',         accelerator: 'CmdOrCtrl+2', click: () => mainWindow.webContents.send('menu:tab-assets') },
        { type: 'separator' },
        { label: 'Zoom In',        accelerator: 'CmdOrCtrl+Equal', click: () => mainWindow.webContents.send('menu:zoom-in') },
        { label: 'Zoom Out',       accelerator: 'CmdOrCtrl+-',     click: () => mainWindow.webContents.send('menu:zoom-out') },
        { label: 'Encaixar',       accelerator: 'CmdOrCtrl+0',     click: () => mainWindow.webContents.send('menu:zoom-fit') },
        { type: 'separator' },
        { label: 'DevTools',       accelerator: 'F12',              click: () => mainWindow.webContents.toggleDevTools() },
        { label: 'Recarregar',     accelerator: 'CmdOrCtrl+R',     click: () => mainWindow.webContents.reload() }
      ]
    },
    {
      label: 'Jogo',
      submenu: [
        { label: 'Testar Jogo',    accelerator: 'F5',               click: () => mainWindow.webContents.send('menu:play') },
        { label: 'Parar Preview',  accelerator: 'F6',               click: () => { if (previewWindow) previewWindow.close() } }
      ]
    },
    {
      label: 'Ajuda',
      submenu: [
        { label: 'Documentação',   click: () => shell.openExternal('https://github.com/ElioNeto/kora') },
        { label: 'Reportar Problema', click: () => shell.openExternal('https://github.com/ElioNeto/kora/issues') },
        { type: 'separator' },
        {
          label: 'Sobre Kora Editor',
          click: () => dialog.showMessageBox(mainWindow, {
            type: 'info', title: 'Sobre',
            message: 'Kora Editor v0.1.0',
            detail: 'Engine de jogos 2D para Android\nKScript\n\nMIT License'
          })
        }
      ]
    }
  ]
  Menu.setApplicationMenu(Menu.buildFromTemplate(template))
}

// ─── Janela principal ────────────────────────────────────────────────────────
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400, height: 900,
    minWidth: 800, minHeight: 600,
    frame: true,
    backgroundColor: '#0f1117',
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
      webviewTag: false
    }
  })

  createMenu()

  if (rendererPath.startsWith('http://')) mainWindow.loadURL(rendererPath)
  else mainWindow.loadFile(rendererPath)

  if (isDev) mainWindow.webContents.openDevTools()

  const bounds = getWindowConfig('mainBounds')
  if (bounds) mainWindow.setBounds(bounds)

  mainWindow.on('resize', () => saveWindowConfig('mainBounds', mainWindow.getBounds()))
  mainWindow.on('maximize',   () => mainWindow.webContents.send('window:maximized'))
  mainWindow.on('unmaximize', () => mainWindow.webContents.send('window:unmaximized'))
  mainWindow.on('closed', () => { mainWindow = null })
}

// ─── Janela de preview do jogo ───────────────────────────────────────────────
function openPreviewWindow(htmlContent) {
  if (previewWindow && !previewWindow.isDestroyed()) {
    previewWindow.focus()
    previewWindow.webContents.loadURL(`data:text/html;charset=utf-8,${encodeURIComponent(htmlContent)}`)
    return
  }

  previewWindow = new BrowserWindow({
    width: 400, height: 720,
    title: 'Kora — Testar Jogo',
    backgroundColor: '#000',
    resizable: true,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true
    }
  })

  previewWindow.setMenu(null)
  previewWindow.loadURL(`data:text/html;charset=utf-8,${encodeURIComponent(htmlContent)}`)
  if (isDev) previewWindow.webContents.openDevTools({ mode: 'detach' })
  previewWindow.on('closed', () => { previewWindow = null })
}

// ─── Config persistência ────────────────────────────────────────────────────
const configPath = () => path.join(app.getPath('userData'), 'config.json')

function saveWindowConfig(key, value) {
  try {
    let cfg = {}
    if (fs.existsSync(configPath())) cfg = JSON.parse(fs.readFileSync(configPath(), 'utf8'))
    cfg[key] = value
    fs.writeFileSync(configPath(), JSON.stringify(cfg), 'utf8')
  } catch {}
}

function getWindowConfig(key) {
  try {
    if (fs.existsSync(configPath())) {
      const cfg = JSON.parse(fs.readFileSync(configPath(), 'utf8'))
      return cfg[key] || null
    }
  } catch {}
  return null
}

// ─── IPC Handlers ────────────────────────────────────────────────────────────
ipcMain.handle('select-directory', async () => {
  const r = await dialog.showOpenDialog(mainWindow, { properties: ['openDirectory'] })
  return r.canceled ? null : r.filePaths[0]
})

ipcMain.handle('select-file', async (event, options = {}) => {
  const r = await dialog.showOpenDialog(mainWindow, {
    properties: options.properties || ['openFile'],
    filters: options.filters || [{ name: 'All Files', extensions: ['*'] }]
  })
  if (r.canceled) return null
  return (options.properties || []).includes('multiSelections') ? r.filePaths : r.filePaths[0]
})

ipcMain.handle('save-file', async (event, options = {}) => {
  const r = await dialog.showSaveDialog(mainWindow, {
    defaultPath: options.defaultPath || 'untitled',
    filters: options.filters || [{ name: 'All Files', extensions: ['*'] }]
  })
  return r.canceled ? null : r.filePath
})

ipcMain.handle('read-file', async (event, filePath) => {
  try { return { content: fs.readFileSync(filePath, 'utf8'), error: null } }
  catch (err) { return { content: null, error: err.message } }
})

ipcMain.handle('write-file', async (event, filePath, content) => {
  try { fs.writeFileSync(filePath, content, 'utf8'); return { success: true } }
  catch (err) { return { success: false, error: err.message } }
})

ipcMain.handle('read-binary-file', async (event, filePath) => {
  try { return { base64: fs.readFileSync(filePath).toString('base64'), error: null } }
  catch (err) { return { base64: null, error: err.message } }
})

ipcMain.handle('file-exists', async (event, filePath) => {
  try { fs.accessSync(filePath); return true } catch { return false }
})

ipcMain.handle('get-app-path', async (event, name) => app.getPath(name))
ipcMain.handle('show-item-in-folder', async (event, fp) => shell.showItemInFolder(fp))
ipcMain.handle('open-external', async (event, url) => shell.openExternal(url))
ipcMain.handle('get-version', async () => app.getVersion())

ipcMain.handle('window-minimize',   async () => mainWindow?.minimize())
ipcMain.handle('window-maximize',   async () => mainWindow?.maximize())
ipcMain.handle('window-unmaximize', async () => mainWindow?.unmaximize())
ipcMain.handle('window-is-maximized', async () => mainWindow?.isMaximized() ?? false)
ipcMain.handle('window-close',      async () => mainWindow?.close())

ipcMain.handle('save-scene', async (event, { sceneData, fileName }) => {
  const r = await dialog.showSaveDialog(mainWindow, {
    title: 'Salvar Cena',
    defaultPath: fileName || 'cena.kora.json',
    filters: [{ name: 'Cenas Kora', extensions: ['kora.json'] }]
  })
  if (r.canceled) return { success: false }
  try {
    fs.writeFileSync(r.filePath, sceneData, 'utf8')
    return { success: true, filePath: r.filePath }
  } catch (err) {
    return { success: false, error: err.message }
  }
})

ipcMain.handle('open-preview', async (event, htmlContent) => {
  openPreviewWindow(htmlContent)
  return { success: true }
})

ipcMain.handle('build-apk', async (event, config) => {
  return { success: false, message: 'Build APK requer configuração do ambiente Android' }
})

// ─── Ciclo de vida do app ────────────────────────────────────────────────────
app.whenReady().then(createWindow)

app.on('window-all-closed', () => { if (process.platform !== 'darwin') app.quit() })
app.on('activate', () => { if (BrowserWindow.getAllWindows().length === 0) createWindow() })

const gotTheLock = app.requestSingleInstanceLock()
if (!gotTheLock) {
  app.quit()
} else {
  app.on('second-instance', () => {
    if (mainWindow) { if (mainWindow.isMinimized()) mainWindow.restore(); mainWindow.focus() }
  })
}
