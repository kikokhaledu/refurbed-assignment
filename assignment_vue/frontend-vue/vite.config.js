import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

function parsePort(rawValue, fallback) {
  const parsed = Number.parseInt(String(rawValue ?? ''), 10)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return parsed
}

const frontendHost = process.env.FRONTEND_HOST || '127.0.0.1'
const frontendPort = parsePort(process.env.FRONTEND_PORT, 5173)

export default defineConfig({
  plugins: [vue()],
  server: {
    host: frontendHost,
    port: frontendPort,
  },
})
