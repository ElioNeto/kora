/**
 * Kora Editor - Preload Script
 * Fornece API segura para o renderer acessar funcionalidades nativas
 */

const { contextBridge, ipcRenderer } = require('electron')

// Expor APIs seguras ao contexto do renderer
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

  // Build APK
  buildAPK: (config) => ipcRenderer.invoke('build-apk', config),

  // Eventos do renderer
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

  // Versão do app
  getVersion: () => ipcRenderer.invoke('get-version'),
  getPlatform: () => process.platform
})

// Expor versão
contextBridge.exposeInMainWorld('electronVersion', {
  app: app.getVersion(),
  chrome: process.versions.chrome,
  node: process.versions.node
})
