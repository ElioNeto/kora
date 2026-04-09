import { defineConfig } from 'vite'
import path from 'path'

export default defineConfig({
  base: './',
  root: 'src/renderer',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      entry: {
        index: path.resolve(__dirname, 'src/renderer/index.html')
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src/renderer')
    }
  },
  css: {
    modules: {
      localsConvention: 'camelCase'
    }
  },
  server: {
    port: 5173,
    open: false
  }
})
