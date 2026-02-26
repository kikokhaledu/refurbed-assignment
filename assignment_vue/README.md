# Vue Track Overview

This folder contains the Vue frontend track.

## Structure
```text
assignment_vue/
`-- frontend-vue/    # Vue 3 + Vite + Tailwind implementation
```

Backend is shared from the repository root in `backend/`.

## Run Options

### Local Development
1. Start backend:
```bash
cd backend
go run .
```
2. Start Vue frontend:
```bash
cd assignment_vue/frontend-vue
npm install
npm run dev
```

### Docker Compose (recommended)
From repository root:
```bash
make up
```

Compose configuration is read from `.env` (if present) or `.env.example` (fallback).

## URLs
- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

## Related Docs
- Vue frontend details: [frontend-vue/README.md](./frontend-vue/README.md)
- Vue frontend E2E matrix: [frontend-vue/PLAYWRIGHT_E2E_MATRIX.md](./frontend-vue/PLAYWRIGHT_E2E_MATRIX.md)
- Vue frontend unit matrix: [frontend-vue/UNIT_TEST_MATRIX.md](./frontend-vue/UNIT_TEST_MATRIX.md)
- Backend details: [../backend/README.md](../backend/README.md)
