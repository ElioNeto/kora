/**
 * Kora Editor - Electron Main Process
 * Handles native functionality, file system, and IPC communication
 */
const { app, BrowserWindow, ipcMain, dialog, Menu, shell } = require('electron')
const path = require('path')
const fs = require('fs')

let mainWindow = null
let rendererPath = null

// Configurar caminhos
const isDev = process.env.NODE_ENV === 'development' || !app.isPackaged

if (isDev) {
  rendererPath = 'http://localhost:5173'
} else {
  rendererPath = path.join(__dirname, '../dist/index.html')
}

// Configurar menu
function createMenu() {
  const template = [
    {
      label: 'Arquivo',
      submenu: [
        {
          label: 'Nova Cena',
          accelerator: 'CmdOrCtrl+N',
          click: () => mainWindow.webContents.send('menu:new-scene')
        },
        { type: 'separator' },
        {
          label: 'Abrir Cena',
          accelerator: 'CmdOrCtrl+O',
          click: () => mainWindow.webContents.send('menu:open-scene')
        },
        {
          label: 'Salvar Cena',
          accelerator: 'CmdOrCtrl+S',
          click: () => mainWindow.webContents.send('menu:save-scene')
        },
        {
          label: 'Salvar Como...',
          accelerator: 'CmdOrCtrl+Shift+S',
          click: saveSceneAs
        },
        { type: 'separator' },
        {
          label: 'Exportar KScript',
          accelerator: 'CmdOrCtrl+E',
          click: () => mainWindow.webContents.send('menu:export-ks')
        },
        {
          label: 'Exportar APK...',
          click: exportAPK
        },
        { type: 'separator' },
        {
          label: 'Sair',
          accelerator: 'CmdOrCtrl+Q',
          click: () => app.quit()
        }
      ]
    },
    {
      label: 'Editar',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectAll' }
      ]
    },
    {
      label: 'Visual',
      submenu: [
        {
          label: 'Zoom In',
          accelerator: 'CmdOrCtrl+1',
          click: () => {
            if (mainWindow.getZoomFactor() < 3) {
              mainWindow.setZoomFactor(mainWindow.getZoomFactor() * 1.1)
            }
          }
        },
        {
          label: 'Zoom Out',
          accelerator: 'CmdOrCtrl+-',
          click: () => {
            if (mainWindow.getZoomFactor() > 0.3) {
              mainWindow.setZoomFactor(mainWindow.getZoomFactor() / 1.1)
            }
          }
        },
        {
          label: 'Zoom Reset',
          accelerator: 'CmdOrCtrl+0',
          click: () => mainWindow.setZoomFactor(1)
        },
        { type: 'separator' },
        {
          label: 'Ativar/Desativar Developer Tools',
          accelerator: 'F12',
          click: () => {
            if (mainWindow) {
              mainWindow.webContents.toggleDevTools()
            }
          }
        }
      ]
    },
    {
      label: 'Ajuda',
      submenu: [
        {
          label: 'Documentação',
          click: () => shell.openExternal('https://koraengine.dev/docs')
        },
        {
          label: 'Reportar Problema',
          click: () => shell.openExternal('https://github.com/koraengine/desktop/issues')
        },
        { type: 'separator' },
        {
          label: 'Sobre Kora Editor',
          click: () => {
            dialog.showMessageBox(mainWindow, {
              type: 'info',
              title: 'Sobre',
              message: 'Kora Editor v0.1.0',
              detail: 'Engine de jogos 2D para Android\nCom KScript\n\nMIT License'
            })
          }
        }
      ]
    }
  ]

  const menu = Menu.buildFromTemplate(template)
  Menu.setApplicationMenu(menu)
}

// Criar janela principal
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 800,
    minHeight: 600,
    frame: !app.isPackaged,
    backgroundColor: '#0f1117',
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
      webviewTag: true
    }
  })

  // Configurar menu
  createMenu()

  if (rendererPath.startsWith('http://')) {
    mainWindow.loadURL(rendererPath)
  } else {
    mainWindow.loadFile(rendererPath)
  }

  // Desenvolvimento - abrir DevTools
  if (isDev) {
    mainWindow.webContents.openDevTools()
  }

  // Restaurar tamanho da janela
  const bounds = getMainWindowBounds()
  if (bounds) {
    mainWindow.setBounds(bounds)
  }

  // Salvar bounds ao fechar
  mainWindow.on('resize', () => {
    saveMainWindowBounds(mainWindow.getBounds())
  })

  // Prevenir redimensionamento fora do mínimo
  mainWindow.on('maximize', () => {
    mainWindow.webContents.send('window:maximized')
  })

  mainWindow.on('unmaximize', () => {
    mainWindow.webContents.send('window:unmaximized')
  })

  mainWindow.on('close', (event) => {
    if (app.isPackaged) {
      event.preventDefault()
      dialog.showMessageBox(mainWindow, {
        type: 'question',
        buttons: ['Sim', 'Não'],
        defaultId: 1,
        title: 'Sair do Kora Editor?',
        message: 'Tem certeza que deseja sair?',
        detail: 'Salve seu progresso antes de fechar!'
      }).then(result => {
        if (result.response === 0) {
          app.exit(0)
        }
      })
    } else {
      // Em desenvolvimento, permitir fechar sem confirmação
      mainWindow = null
      app.quit()
    }
  })
}

// Gerenciar bounds da janela
function saveMainWindowBounds(bounds) {
  const configPath = path.join(app.getPath('userData'), 'config.json')
  try {
    const config = {
      bounds: {
        width: bounds.width,
        height: bounds.height,
        x: bounds.x,
        y: bounds.y,
        maximized: bounds.width === screen.getPrimaryDisplay().bounds.width &&
                   bounds.height === screen.getPrimaryDisplay().bounds.height
      }
    }
    fs.writeFileSync(configPath, JSON.stringify(config), 'utf8')
  } catch (err) {
    console.error('Erro ao salvar bounds:', err)
  }
}

function getMainWindowBounds() {
  const configPath = path.join(app.getPath('userData'), 'config.json')
  try {
    if (fs.existsSync(configPath)) {
      const config = JSON.parse(fs.readFileSync(configPath, 'utf8'))
      return config.bounds
    }
  } catch (err) {
    console.error('Erro ao ler bounds:', err)
  }
  return null
}

// Handlers de IPC
ipcMain.handle('select-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory']
  })
  return result.canceled ? null : result.filePaths[0]
})

ipcMain.handle('select-file', async (event, options) => {
  const defaultProps = {
    properties: ['openFile'],
    filters: [{ name: 'All Files', extensions: ['*'] }]
  }
  const result = await dialog.showOpenDialog(mainWindow, {
    ...defaultProps,
    ...options
  })
  return result.canceled ? null : result.filePaths[0]
})

ipcMain.handle('save-file', async (event, options) => {
  const defaultProps = {
    properties: ['saveFile'],
    filters: [{ name: options.filterName || 'All Files', extensions: [options.filterExt || '*'] }]
  }
  const result = await dialog.showSaveDialog(mainWindow, {
    ...defaultProps,
    defaultPath: options.defaultPath || 'untitled',
    ...options
  })
  return result.canceled ? null : result.filePath
})

ipcMain.handle('read-file', async (event, filePath) => {
  try {
    const content = fs.readFileSync(filePath, 'utf8')
    return { content, error: null }
  } catch (err) {
    return { content: null, error: err.message }
  }
})

ipcMain.handle('write-file', async (event, filePath, content) => {
  try {
    fs.writeFileSync(filePath, content, 'utf8')
    return { success: true, error: null }
  } catch (err) {
    return { success: false, error: err.message }
  }
})

ipcMain.handle('read-binary-file', async (event, filePath) => {
  try {
    const buffer = fs.readFileSync(filePath)
    const base64 = buffer.toString('base64')
    return { base64, error: null }
  } catch (err) {
    return { base64: null, error: err.message }
  }
})

ipcMain.handle('file-exists', async (event, filePath) => {
  try {
    fs.accessSync(filePath)
    return true
  } catch {
    return false
  }
})

ipcMain.handle('get-app-path', async (event, appPath) => {
  return app.getPath(appPath)
})

ipcMain.handle('show-item-in-folder', async (event, filePath) => {
  shell.showItemInFolder(filePath)
})

ipcMain.handle('open-external', async (event, url) => {
  shell.openExternal(url)
})

// Handlers de Exportação
ipcMain.handle('build-apk', async (event, config) => {
  const { androidHome, gradlePath, buildType } = config
  return {
    success: false,
    message: 'Build APK requer configuração do ambiente Android'
  }
})

// Salvar cena
ipcMain.handle('save-scene', async (event, { sceneData, fileName }) => {
  const filePath = await dialog.showSaveDialog(mainWindow, {
    title: 'Salvar Cena',
    defaultPath: fileName || 'cena.kora.json',
    filters: [{ name: 'Cenas Kora', extensions: ['kora.json'] }]
  })

  if (filePath.canceled) return { success: false }

  try {
    fs.writeFileSync(filePath.filePath, JSON.stringify(sceneData, null, 2), 'utf8')
    return { success: true, filePath: filePath.filePath }
  } catch (err) {
    return { success: false, error: err.message }
  }
})

// Funções para menu
async function saveSceneAs() {
  mainWindow.webContents.send('menu:save-scene', true)
}

async function exportAPK() {
  // Placeholder para build APK
  // Na prática, este comando executaria ./android/build.sh release
  dialog.showMessageBox(mainWindow, {
    type: 'info',
    title: 'Exportar APK',
    message: 'A exportação APK será integrada com o build system em breve.',
    buttons: ['OK']
  })
}

// Eventos do app
app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow()
  }
})

// Prevenir múltiplas instâncias
const gotTheLock = app.requestSingleInstanceLock()

if (!gotTheLock) {
  app.quit()
} else {
  app.on('second-instance', () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore()
      mainWindow.focus()
    }
  })
}
