import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api':     process.env.BACKEND_URL || 'http://localhost:8080',
      '/uploads': process.env.BACKEND_URL || 'http://localhost:8080',
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  }
})
