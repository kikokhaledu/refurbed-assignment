# Frontend Unit Test Matrix

Status legend:
- `Covered`: directly asserted by one or more unit tests.
- `Partially covered`: core path is tested, but some branches/states are not directly asserted.
- `Not covered`: no direct unit test in current suite.

## Coverage Summary

| Area | Status | Evidence |
|---|---|---|
| Filters panel slider clamp rules (minimum 1 EUR gap when possible) | Covered | `src/components/__tests__/FiltersPanel.spec.js` (`clamps minimum...`, `clamps maximum...`) |
| Filters panel slider shared scale (same floor/ceiling for both thumbs) | Covered | `FiltersPanel.spec.js` (`keeps both sliders on the same floor/ceiling scale`) |
| Filters panel slider initialization and bounds re-sync on API bound updates | Covered | `FiltersPanel.spec.js` (`initializes slider thumbs...`, `re-syncs slider thumbs...`) |
| Filters panel brand multi-select event emission | Covered | `FiltersPanel.spec.js` (`emits selected brands updates when a brand is toggled`) |
| Sticky apply bar visibility behavior | Covered | `FiltersPanel.spec.js` (`shows bottom sticky actions when top controls scroll out of view`) |
| Product card image fallback and color-switch behavior | Covered | `src/components/__tests__/ProductCard.spec.js` |
| Product card out-of-stock swatch visibility rules | Covered | `ProductCard.spec.js` |
| Sort dropdown multi-select and contradictory sort auto-resolution | Covered | `src/components/__tests__/SortSelect.spec.js` |
| Active filter chips remove/clear actions plus mobile collapse toggle behavior | Covered | `src/components/__tests__/ActiveFilterChips.spec.js` |
| URL parser handling of invalid `minPrice`/`maxPrice` values | Covered | `src/utils/productQueryState.spec.js` |
| App-level URL sync/history/search debounce integration | Not covered | Covered by Playwright E2E; no direct `App.vue` unit tests yet. |
| Mobile dialog focus trap + body scroll lock | Not covered | Covered by Playwright E2E; no dedicated component-level unit tests. |
| Theme persistence (`localStorage`) | Not covered | Covered by Playwright E2E; no dedicated composable unit test. |

## Current Gaps Worth Adding Next

| Gap | Priority | Reason |
|---|---|---|
| `App.vue` integration unit tests for applied-vs-draft state transitions | P1 | Would reduce reliance on E2E for core orchestration logic. |
| `useTheme` composable unit tests | P2 | Would directly cover persistence and class toggling logic. |
| Focus-trap utility unit tests (mobile modal key handling) | P2 | Would catch keyboard regressions faster than full E2E runs. |
