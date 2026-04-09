/**
 * Kora Editor - Preload Script
 * Fornece API segura para o renderer acessar funcionalidades nativas
 */

const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('electronAPI', {
  // Sistema de arquivos
  selectDirectory: () => ipcRenderer.invoke('select-directory'),
  selectFile: (options) => ipcRenderer.invoke('select-file', options),
  saveFile: (options) => ipcRenderer.invoke('save-file', options),
  readFile: (filePath) => ipcRenderer.invoke('read-file', filePath),
  readBinaryFile: (filePath) => ipcRenderer.invoke('read-binary-file', filePath),
  writeFile: (filePath, content) => ipcRenderer.invoke('write-file', filePath, content),
  fileExists: (filePath) => ipcRenderer.invoke('file-exists', filePath),

  // Caminhos do sistema
  getAppPath: (name) => ipcRenderer.invoke('get-app-path', name),

  // Sistema operacional
  showItemInFolder: (filePath) => ipcRenderer.invoke('show-item-in-folder', filePath),
  openExternal: (url) => ipcRenderer.invoke('open-external', url),

  // Save scene específico
  saveScene: (sceneData, fileName) => ipcRenderer.invoke('save-scene', { sceneData, fileName }),

  // Preview / teste do jogo
  openPreview: (htmlContent) => ipcRenderer.invoke('open-preview', htmlContent),

  // Build APK
  buildAPK: (config) => ipcRenderer.invoke('build-apk', config),

  // Janela
  minimize: () => ipcRenderer.invoke('window-minimize'),
  maximize: () => ipcRenderer.invoke('window-maximize'),
  unmaximize: () => ipcRenderer.invoke('window-unmaximize'),
  isMaximized: () => ipcRenderer.invoke('window-is-maximized'),
  close: () => ipcRenderer.invoke('window-close'),

  // Eventos do main process → renderer
  on: (channel, func) => {
    const validChannels = [
      'menu:new-scene',
      'menu:open-scene',
      'menu:save-scene',
      'menu:export-ks',
      'window:maximized',
      'window:unmaximized'
    ]
    if (validChannels.includes(channel)) {
      const subscription = (event, ...args) => func(...args)
      ipcRenderer.on(channel, subscription)
      return () => ipcRenderer.removeListener(channel, subscription)
    }
  },

  getVersion: () => ipcRenderer.invoke('get-version'),
  getPlatform: () => process.platform
})

contextBridge.exposeInMainWorld('electronVersion', {
  chrome: process.versions.chrome,
  node: process.versions.node,
  electron: process.versions.electron
})
