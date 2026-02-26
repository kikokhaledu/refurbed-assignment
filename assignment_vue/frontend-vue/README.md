# Frontend - Vue 3 + Tailwind CSS

This is the primary implemented frontend for this repository.

## Prerequisites
- Node.js 18+
- Backend API running (default: `http://localhost:8080`)

## Local Development
From this directory:

```bash
npm install
npm run dev
```

Frontend is available at `http://localhost:5173`.

## Environment Variables
- `VITE_API_BASE_URL` (optional): explicit backend base URL override.
- `VITE_BACKEND_PORT` (optional): backend port used by fallback API URL when `VITE_API_BASE_URL` is empty (default: `8080`).
  Fallback format is `<current-page-protocol>://<current-page-hostname>:<VITE_BACKEND_PORT>`.
- `FRONTEND_HOST` (optional): Vite host for dev server (default: `127.0.0.1`)
- `FRONTEND_PORT` (optional): Vite port for dev server (default: `5173`)

Example:

```bash
VITE_BACKEND_PORT=9090 FRONTEND_PORT=5173 npm run dev
```

Explicit cross-origin backend:

```bash
VITE_API_BASE_URL=http://localhost:9090 FRONTEND_PORT=5173 npm run dev
```

## Docker
This frontend has a Dockerfile at [Dockerfile](./Dockerfile) and is wired into root Compose.

From repository root:

```bash
make up
```

Then open:
- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

`make up` uses `.env` when present, otherwise falls back to root `.env.example`.

## Build
```bash
npm run build
```

## Lint
```bash
npm run lint
```

## Tests
```bash
npm run test:run
```

Playwright E2E:
```bash
npm run e2e:install
npm run e2e
```

Default Playwright ports are isolated from Compose runtime:
- `PW_BACKEND_PORT=18080`
- `PW_FRONTEND_PORT=4173`

Optional:
```bash
npm run e2e:headed
npm run e2e:ui
```

E2E coverage matrix:
- [PLAYWRIGHT_E2E_MATRIX.md](./PLAYWRIGHT_E2E_MATRIX.md)
- [UNIT_TEST_MATRIX.md](./UNIT_TEST_MATRIX.md)
- Current automated run status is tracked in the `Automation Status (Current)` section.


## Tech Stack
- Vue 3
- Tailwind CSS
- Vitest + Vue Test Utils (unit tests)
- Playwright (E2E tests: desktop + mobile Chromium)

## Implemented UI Features
- Debounced search auto-apply (500ms after typing stops).
- Draft-vs-applied filtering for non-search controls (`Apply` button inside filter panel), with `Reset` acting as an immediate clear+apply action so we reduce queries.
- Filters panel actions are position-aware: action controls stay at the top when visible, and switch to a sticky bottom action bar (with unapplied-changes warning) when the top controls scroll out of view.
- URL sync for applied query state (`search`, `category`, `brand`, `color`, `condition`, `bestseller`, `onSale`, `inStock`, `sort`, `minPrice`, `maxPrice`, `offset`).
- Applied URL state uses browser history entries so back/forward navigation restores prior results and filter selections.
- Filter panel with category, brand, colors, condition, bestseller, on-sale toggle, stock availability (`all`/`in stock`), and a dual-thumb single-track price slider with minimum 1 EUR thumb gap (when bounds allow) to prevent zero-width cross states.
- Price slider bounds are data-driven from backend response (`price_min`/`price_max`), not hardcoded.
- Color filter options are data-driven from backend `available_colors`.
- Brand filter options are data-driven from backend `available_brands`.
- Sort control exposes three base options (`popularity`, `price_asc`, `price_desc`) as a multi-select dropdown with green checkmarks; non-conflicting selections combine while preserving user pick order in backend sort values (for example `sort=price_asc,popularity` when price is picked first), contradictory price directions auto-resolve to one active direction, and selection is applied once when the dropdown closes.
- Sort dropdown supports staging: users can pick multiple options in one open interaction without reloading after each click; one apply/network refresh happens when the dropdown closes.
- Active filter chips allow single-filter removal and `Clear all` from the results area.
- On mobile, active filter chips use a compact summary row (`N filters applied`) with `Show/Hide` toggle and mobile clear-all action to avoid result-list crowding.
- App-level orchestration is now split with shared query/sort/filter utilities and a dedicated mobile-dialog composable to keep `App.vue` focused.
- Product card color swatches change the card image using backend `image_urls_by_color` mapping (fallback to `image_url` if a color mapping is missing).
- Product images include lazy loading, async decoding, explicit dimensions, and runtime fallback to a local placeholder image when a URL fails.
- Original strikethrough price is shown only for valid discounts between 1% and 99% to avoid invalid values for edge-case discount inputs.
- Out-of-stock colors are hidden from card swatches and backend-provided color filter options.
- Stock label updates per selected color (from `stock_by_color`).
- Load-more pagination (`limit`/`offset`) with append behavior.
- Loading skeletons, error state with retry, and empty state with clear filters.
- Dark mode toggle with persisted preference (`localStorage`).
- Mobile UX: filter modal with focus trap + `Esc` close + body scroll lock, plus floating actions (`Filters` and `Top` buttons).
- Minimal SEO baseline in HTML shell: page title, description meta tag, and basic OpenGraph tags.

## Production Notes (SEO)
- Current SEO setup is intentionally minimal for assignment scope.
- For production, add canonical URL strategy, structured data (JSON-LD), richer OG/Twitter tags, and consider SSR/prerender for crawler-friendly content.

## Note 
- images might not match the colors  since it was just for dev 
