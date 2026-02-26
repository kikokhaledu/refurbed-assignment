import { defineConfig, devices } from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const isCI = Boolean(process.env.CI)
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const repoRoot = path.resolve(__dirname, '../..')

function parseDotEnv(content) {
  const parsed = {}
  for (const line of content.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      continue
    }

    const firstEquals = trimmed.indexOf('=')
    if (firstEquals === -1) {
      continue
    }

    const key = trimmed.slice(0, firstEquals).trim()
    let value = trimmed.slice(firstEquals + 1).trim()
    if (!key) {
      continue
    }

    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }

    parsed[key] = value
  }

  return parsed
}

function loadRootEnv() {
  const envPath = path.join(repoRoot, '.env')
  const examplePath = path.join(repoRoot, '.env.example')

  let loaded = {}
  if (fs.existsSync(examplePath)) {
    loaded = { ...loaded, ...parseDotEnv(fs.readFileSync(examplePath, 'utf8')) }
  }
  if (fs.existsSync(envPath)) {
    loaded = { ...loaded, ...parseDotEnv(fs.readFileSync(envPath, 'utf8')) }
  }

  return { ...loaded, ...process.env }
}

function toPositiveInt(rawValue, fallback) {
  const parsed = Number.parseInt(String(rawValue ?? ''), 10)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return parsed
}

const rootEnv = loadRootEnv()
const backendHost = rootEnv.PW_BACKEND_HOST || '127.0.0.1'
const backendPort = toPositiveInt(rootEnv.PW_BACKEND_PORT, 18080)
const frontendHost = rootEnv.PW_FRONTEND_HOST || '127.0.0.1'
const frontendPort = toPositiveInt(rootEnv.PW_FRONTEND_PORT, 4173)

const backendBaseURL = `http://${backendHost}:${backendPort}`
const frontendBaseURL = `http://${frontendHost}:${frontendPort}`
const playwrightAPIBaseURL = rootEnv.PW_VITE_API_BASE_URL || backendBaseURL

process.env.PW_BACKEND_BASE_URL = backendBaseURL

export default defineConfig({
  testDir: './e2e',
  timeout: 45_000,
  expect: {
    timeout: 8_000,
  },
  fullyParallel: true,
  workers: isCI ? 1 : undefined,
  retries: isCI ? 1 : 0,
  reporter: isCI ? [['list'], ['html', { open: 'never' }]] : [['list']],
  use: {
    baseURL: frontendBaseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium-desktop',
      testMatch: /.*\.desktop\.spec\.js/,
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 900 },
      },
    },
    {
      name: 'chromium-mobile',
      testMatch: /.*\.mobile\.spec\.js/,
      use: {
        ...devices['Pixel 7'],
      },
    },
  ],
  webServer: [
    {
      command: 'go run .',
      cwd: '../../backend',
      url: `${backendBaseURL}/health`,
      timeout: 90_000,
      reuseExistingServer: false,
      env: {
        ...process.env,
        BACKEND_HOST: backendHost,
        BACKEND_PORT: String(backendPort),
        BACKEND_DATA_DIR: rootEnv.BACKEND_DATA_DIR || 'data',
        BACKEND_CACHE_TTL_SECONDS: rootEnv.BACKEND_CACHE_TTL_SECONDS || '30',
        BACKEND_CORS_ALLOW_ORIGIN: rootEnv.BACKEND_CORS_ALLOW_ORIGIN || '*',
      },
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: `npm run dev -- --host ${frontendHost} --port ${frontendPort} --strictPort`,
      url: frontendBaseURL,
      timeout: 90_000,
      reuseExistingServer: false,
      env: {
        ...process.env,
        FRONTEND_HOST: frontendHost,
        FRONTEND_PORT: String(frontendPort),
        VITE_API_BASE_URL: playwrightAPIBaseURL,
      },
      stdout: 'pipe',
      stderr: 'pipe',
    },
  ],
})
