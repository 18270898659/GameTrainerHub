import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    },
    extensions: ['.ts', '.vue', '.tsx', '.mjs', '.js', '.jsx', '.json']
  },
  server: {
    port: 34115,
    strictPort: true
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})

