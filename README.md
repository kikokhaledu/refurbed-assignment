# Refurbed Engineering - Senior Fullstack Assignment

This repository contains the backend plus the vue/tailwind frontend  for the assignment.

## Environment Configuration
- Default config lives in [.env.example](./.env.example).
- `make` commands automatically use:
1. `.env` (if present), otherwise
2. `.env.example`.
- This keeps one config model across backend, frontend, Docker Compose, and Playwright E2E (in prod we might want to change that so e2e uses diffrent values).
- Keep `VITE_API_BASE_URL` empty for same-host deployments so frontend derives backend URL from host + `VITE_BACKEND_PORT`.
- In Compose, `VITE_BACKEND_PORT` falls back to `BACKEND_PORT` (single source for backend port).
- Playwright uses dedicated local ports by default (`PW_BACKEND_PORT=18080`, `PW_FRONTEND_PORT=4173`) to avoid conflicts when `make up` is already running.

If you need custom values:
```bash
cp .env.example .env
```

## Documentation Index
- Assignment brief: [assignment.md](./assignment.md)
- Backend guide: [backend/README.md](./backend/README.md)
- Backend test matrix: [backend/TEST_MATRIX.md](./backend/TEST_MATRIX.md)
- Vue track overview: [assignment_vue/README.md](./assignment_vue/README.md)
- Vue frontend guide: [assignment_vue/frontend-vue/README.md](./assignment_vue/frontend-vue/README.md)
- Vue frontend E2E matrix: [assignment_vue/frontend-vue/PLAYWRIGHT_E2E_MATRIX.md](./assignment_vue/frontend-vue/PLAYWRIGHT_E2E_MATRIX.md)
- Vue frontend unit matrix: [assignment_vue/frontend-vue/UNIT_TEST_MATRIX.md](./assignment_vue/frontend-vue/UNIT_TEST_MATRIX.md)

## Repository Layout
- `backend/` - Go API (`GET /products`, caching, filtering, load-more pagination).
- `assignment_vue/frontend-vue/` - Vue + Tailwind frontend implementation.

## Quick Start (Backend Only)
```bash
cd backend
go run .
```

Backend default URL is `http://localhost:8080` (configurable via `BACKEND_HOST`/`BACKEND_PORT`).

## Quick Start (Vue Frontend Only)
```bash
cd assignment_vue/frontend-vue
npm install
npm run dev
```

Frontend default URL is `http://localhost:5173` (configurable via `FRONTEND_HOST`/`FRONTEND_PORT`).

## Docker Compose (Full Stack)
From repository root:

```bash
make up
```

Stop and remove containers:

```bash
make down
```

Notes:
- Compose file: [docker-compose.yml](./docker-compose.yml)
- Backend image definition: [backend/Dockerfile](./backend/Dockerfile)
- Frontend image definition: [assignment_vue/frontend-vue/Dockerfile](./assignment_vue/frontend-vue/Dockerfile)
- `backend/data` is mounted read-only into the container at `/app/data`.
- `make up`/`make down` read env values from `.env` (if present) or `.env.example` (fallback).

## Makefile Commands
Common developer commands are available from repository root:

```bash
make help
make up
make seed
make down
make reset
make lint
make fmt-check
make test-frontend
make frontend-lint
make frontend-check
make test-e2e
make e2e-install
make pre-push
```

`make up`  starts both backend and Vue frontend.

`make pre-push` runs:
- `make fmt-check`
- `make vet`
- `make lint`
- `make frontend-lint`
- `make frontend-check`
- `make test`
- `make test-e2e`
- `make compose-check`

`make pre-push` is check-only i like to do that helps in dev env.

Note: `make test-e2e` runs `make e2e-install` first, so first-time machines auto-install Playwright Chromium.
Note: E2E web servers are isolated from Compose default ports, so `make pre-push` can run even if `make up` is active.
Note: `make test-race` requires `CGO_ENABLED=1` and a working `gcc` toolchain on macOS/Linux/Windows (for example Xcode CLT, build-essential, or MSYS2/MinGW). It is intentionally not part of `make pre-push`.
Note: `make frontend-deps` tries `npm ci` first, then retries with `npx npm@10 ci`, and if that still fails it falls back to `npm ci` inside a `node:20-alpine` Docker container.

## What I Would Improve/Change For Production
- Replace in-memory cache with Redis for shared cache across instances, stronger invalidation options, and better horizontal scaling.
- Add explicit stale-while-revalidate semantics with cache metadata/headers and stronger refresh observability.
- Move from file-based sources to resilient upstream APIs or persisted storage with schema validation and contract checks.
- Introduce a search service (Elasticsearch/OpenSearch/Solr) only when catalog size and product discovery complexity justify it (fuzzy matching, relevance ranking, faceting, typo tolerance).
- Add authentication/rate limiting where required, plus stricter CORS policy per environment.
- Move from basic standard-library logging to structured, leveled production logging (for example `slog`/`zap`) with request/trace correlation IDs and JSON output.
- Add end-to-end monitoring stack: structured logs, metrics dashboards, alerting, and distributed tracing telemetry.
- Extend CI quality gates with race tests in Linux CI runner, coverage threshold enforcement, and artifact/image scanning.
- Expand SEO beyond the current baseline with canonical URLs per route/state, structured data (JSON-LD), richer OpenGraph/Twitter tags, and prerender/SSR if crawlability becomes a requirement.

## Notes On Architecture, Decisions Or Other Comments
- The backend is organized into separable layers for parsing (`query.go`), data loading (`repository.go`), business logic/cache (`service.go`), and HTTP transport (`http.go`).
- Product data from two sources is merged once per cache refresh cycle and then filtered/paginated in memory per request.
- Cache snapshots precompute dataset facets (`available_colors`, `available_brands`, `price_min`, `price_max`) once per refresh and reuse them on cache hits.
- Response contract uses a paginated envelope (`items`, `total`, `limit`, `offset`, `has_more`) and also exposes `available_colors`/`available_brands` for data-driven filter options.
- Product payload supports `image_urls_by_color` for strict per-color image rendering in the frontend.
- Product payload supports `stock_by_color` for color-level stock representation.
- In addition to assignment-required filters, the API supports richer filtering for `category`, `brand`, `condition`, `onSale`, `inStock`, and `minStock` (backend-level support).
- Vue frontend uses debounced search auto-apply (500ms), explicit apply for non-search filters (with `Reset` as immediate clear+apply), a multi-select sort dropdown with three base modes (`sort=popularity`, `sort=price_asc`, `sort=price_desc`) that combines only non-conflicting modes while preserving user pick order and applies once on dropdown close, and mobile-friendly floating actions for filters + scroll-to-top.
- Frontend app orchestration is split with shared constants (`sort`/`color`/app settings), shared query+filter utilities, and a dedicated mobile-dialog composable to keep `App.vue` focused and easier to maintain.
- Sort dropdown interaction is staged: users can choose multiple sort options in one open menu session with no per-click reload, then a single refresh occurs when the menu closes.
- Sort edge-case behavior is documented and tested: UI URL normalization keeps one effective price direction, and backend API rejects contradictory price directions if both are sent.
- Applied state changes are written with browser-history entries (`pushState`) so back/forward navigation restores previous filter states.
- Invalid URL price params  degrade safely (`minPrice`/`maxPrice` sanitize to defaults instead of hard-failing initial load).
- Filter panel actions are position-aware: top controls are used when visible, and a sticky bottom action bar appears after scrolling past them (including unapplied-changes warning).
- Active filter chips support per-filter removal and clear-all behavior.
- Price range slider bounds are sourced from backend dataset-level price bounds (`price_min`/`price_max`) instead of static UI constants.
- Mobile filter UX  uses an accessible modal flow (focus trap, `Esc` close, body scroll lock).
- Product cards use explicit color-to-image mapping (`image_urls_by_color`) with fallback placeholder handling for broken image URLs.
- Frontend includes a minimal SEO baseline (`title`, `meta description`, basic OpenGraph tags) in the HTML shell.
- Query parsing is strict (allowlist + explicit validation + singleton-param repetition/empty-value rejection) to fail fast on malformed inputs.
- Price calculation uses cent-based arithmetic internally to reduce floating-point drift.
- Documented backend tradeoffs: refresh work is detached from request cancellation and stale cache can be served on refresh failure to prioritize availability.
- Source JSON ingestion currently tolerates missing/`null` scalar fields via zero-value defaults (assignment pragmatism), while type mismatches still fail decoding for production this should be backed by stricter schema validation and data-quality alerting.
- Dev ergonomics are supported with Docker Compose + Makefile commands for consistent local setup.

## Final Thoughts
- Current implementation intentionally prioritizes clarity, correctness, and testability over premature complexity.
- For this assignment scope, in-memory filtering and caching are sufficient and avoid overengineering.
- SEO is intentionally baseline-level for this challenge; the production path is documented above.
