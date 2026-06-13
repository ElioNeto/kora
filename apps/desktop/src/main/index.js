/**
 * Kora Editor - Electron Main Process
 *
 * ═══════════════════════════════════════════════════════════════════
 * ⚠️  LINUX — PREVENÇÃO DE CRASH DO DISPLAY SERVER
 * ═══════════════════════════════════════════════════════════════════
 *
 * No Linux (Pop!_OS, Ubuntu, etc.), Electron pode crashar o
 * X11/Wayland se o GPU process falhar — porque ele corrompe o
 * estado compartilhado do driver gráfico.
 *
 * SOLUÇÃO:
 *   1. Verificamos /dev/shm ANTES de iniciar (erro clássico)
 *   2. Forçamos --in-process-gpu: GPU roda no MESMO processo
 *      → se crashar, só derruba o app, não o display server
 *   3. Capturamos TODOS os erros sem propagar para o sistema
 *
 * Se ainda crashar, compile com:
 *   NODE_ENV=production electron . --no-sandbox --disable-gpu --disable-dev-shm-usage
 */
const { app, BrowserWindow, ipcMain, dialog, Menu, shell } = require('electron')
const path = require('path')
const fs = require('fs')

let mainWindow = null
let previewWindow = null

// ─── Verificação de /dev/shm (previne crash do display server) ────
try {
  fs.accessSync('/dev/shm', fs.constants.W_OK | fs.constants.X_OK)
} catch {
  console.error('╔══════════════════════════════════════════════════════════╗')
  console.error('║  ❌  /dev/shm INACCESSÍVEL                             ║')
  console.error('║                                                       ║')
  console.error('║  O Electron precisa de /dev/shm com permissão 1777.   ║')
  console.error('║  Execute no terminal:                                 ║')
  console.error('║                                                       ║')
  console.error('║    sudo chmod 1777 /dev/shm                           ║')
  console.error('║                                                       ║')
  console.error('║  Ou use --disable-dev-shm-usage (já configurado).     ║')
  console.error('║  Se o problema persistir, reinicie o computador.      ║')
  console.error('╚══════════════════════════════════════════════════════════╝')
}

// ─── Prevenção de crash do display server (Linux/NVIDIA/Wayland) ──
// CRÍTICO: --in-process-gpu faz o GPU rodar no mesmo processo.
// Se o driver de vídeo crashar, só derruba o app — NÃO o X11/Wayland.
app.disableHardwareAcceleration()
app.commandLine.appendSwitch('in-process-gpu')
app.commandLine.appendSwitch('disable-gpu')
app.commandLine.appendSwitch('disable-software-rasterizer')
app.commandLine.appendSwitch('no-sandbox')
app.commandLine.appendSwitch('disable-dev-shm-usage')
app.commandLine.appendSwitch('use-gl', 'swiftshader')
app.commandLine.appendSwitch('disable-gpu-sandbox')

const isDev = process.env.NODE_ENV === 'development' || (!app.isPackaged && process.env.NODE_ENV !== 'production')
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
        { label: 'Executar Jogo',  accelerator: 'F5',               click: () => mainWindow.webContents.send('menu:play') },
        { label: 'Parar Preview',  accelerator: 'F6',               click: () => { if (previewWindow) previewWindow.close() } },
        { type: 'separator' },
        { label: 'Exportar KScript', accelerator: 'CmdOrCtrl+E',    click: () => mainWindow.webContents.send('menu:export-ks') },
        { label: 'Exportar APK...',                                  click: () => mainWindow.webContents.send('menu:export-apk') }
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
            detail: 'Engine de jogos 2D para Android\nKScript\n\nFeito no Brasil 🇧🇷\nMIT License'
          })
        }
      ]
    }
  ]
  Menu.setApplicationMenu(Menu.buildFromTemplate(template))
}

// ─── Janela principal ────────────────────────────────────────────────────────
function createWindow() {
  try {
    mainWindow = new BrowserWindow({
      width: 1400, height: 900,
      minWidth: 800, minHeight: 600,
      frame: true,
      backgroundColor: '#1a1c1e',
      title: 'Kora Editor',
      show: false, // Só mostra depois de carregar para evitar flicker
      webPreferences: {
        nodeIntegration: false,
        contextIsolation: true,
        preload: path.join(__dirname, 'preload.js'),
        webviewTag: false,
        backgroundThrottling: false
      }
    })

    createMenu()

    if (rendererPath.startsWith('http://')) mainWindow.loadURL(rendererPath)
    else mainWindow.loadFile(rendererPath)

    // Só mostra a janela quando o conteúdo estiver pronto
    mainWindow.once('ready-to-show', () => {
      mainWindow.show()
      if (isDev) mainWindow.webContents.openDevTools()
    })

    const bounds = getWindowConfig('mainBounds')
    if (bounds) mainWindow.setBounds(bounds)

    mainWindow.on('resize', () => saveWindowConfig('mainBounds', mainWindow.getBounds()))
    mainWindow.on('maximize',   () => mainWindow.webContents.send('window:maximized'))
    mainWindow.on('unmaximize', () => mainWindow.webContents.send('window:unmaximized'))
    mainWindow.on('closed', () => { mainWindow = null })

    // Captura crash do processo de renderização sem derrubar o sistema
    // Nota: 'crashed' está deprecated, usar 'render-process-gone'
    mainWindow.webContents.on('render-process-gone', (_event, details) => {
      console.error(`⚠️ Renderer process crashed (reason: ${details.reason}). Recarregando...`)
      // Espera 1 segundo e recarrega — não deixa o crash se propagar
      setTimeout(() => {
        try { mainWindow?.webContents.reload() } catch {}
      }, 1000)
    })

    // Fallback para versões antigas do Electron
    mainWindow.webContents.on('crashed', () => {
      setTimeout(() => {
        try { mainWindow?.webContents.reload() } catch {}
      }, 1000)
    })

    // Captura erros não tratados
    mainWindow.webContents.on('unresponsive', () => {
      console.warn('⚠️ Renderer process unresponsive. Forçando reload...')
      setTimeout(() => {
        try { mainWindow?.webContents.reload() } catch {}
      }, 2000)
    })
  } catch (err) {
    console.error('❌ Erro ao criar janela principal:', err)
    // Não deixa o erro derrubar o display server — apenas loga e continua
  }
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
    backgroundColor: '#1a1c1e',
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

// ─── Captura global de erros (evita crash do display server) ─────
// CRÍTICO: NUNCA deixe erros não tratados propagarem para o sistema.
// No Linux, um segfault do Electron pode derrubar o X11/Wayland inteiro.
process.on('uncaughtException', (error) => {
  console.error('❌ [Kora] Erro não tratado capturado — não vai propagar:', error.message)
  if (error.stack) console.error(error.stack.split('\n').slice(0, 4).join('\n'))
  // Não chama process.exit() — apenas loga e continua.
  // Se o processo precisar morrer, que seja limpo, não por crash.
})

process.on('unhandledRejection', (reason) => {
  console.error('❌ [Kora] Promise rejeitada não tratada:', reason)
})

process.on('exit', (code) => {
  console.log(`👋 [Kora] Processo encerrado com código ${code}`)
})

// CRÍTICO: SIGSEGV (segmentation fault) geralmente é o GPU process
// quebrando. Com --in-process-gpu, isso mata só o app, não o display.
process.on('SIGSEGV', () => {
  console.error('❌ [Kora] SIGSEGV — crash de memória. Isso NÃO deveria')
  console.error('   acontecer com --in-process-gpu. Reporte o bug.')
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
