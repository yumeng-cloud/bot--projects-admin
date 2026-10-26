import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@acb/api-client': path.resolve(__dirname, '../packages/api-client/src/index.ts'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5174,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true },
      '/healthz': { target: 'http://127.0.0.1:8080', changeOrigin: true },
    },
  },
  preview: {
    host: '0.0.0.0',
    port: 5174,
  },
})
