# E2E Test Matrix

This matrix covers end-to-end behavior for the implemented Vue frontend in `assignment_vue/frontend-vue`.

## Scope
- App shell and rendering
- Data fetch lifecycle
- Search and filter behavior (draft vs applied)
- URL state sync/hydration
- Product card interactions
- Sorting and pagination
- Error and empty states
- Mobile modal and floating actions
- Theme persistence
- A11y-critical interactions

## Environment Assumptions
- Defaults are loaded from root `.env.example` (or `.env` when present).
- Playwright runtime defaults:
  - Backend: `http://127.0.0.1:18080` (`PW_BACKEND_HOST`/`PW_BACKEND_PORT`)
  - Frontend: `http://127.0.0.1:4173` (`PW_FRONTEND_HOST`/`PW_FRONTEND_PORT`)
- Seed data from `backend/data/metadata.json`, `backend/data/details.json`, `backend/data/popularity.json`
- Desktop and mobile viewports are both tested

## Suggested Playwright Projects
- `chromium-desktop` (e.g. `1280x900`)
- `chromium-mobile` (e.g. Pixel 7 viewport + touch)

## Test Matrix

| ID | Area | Scenario | Type | Priority | Expected Result |
|---|---|---|---|---|---|
| FE-E2E-001 | Boot | Initial page load | Happy | P0 | Page renders heading, search input, product count, grid/cards, filter panel (desktop). |
| FE-E2E-002 | Boot | Initial load shows skeletons then products | Happy | P1 | Skeletons visible while request in-flight, replaced by cards when response arrives. |
| FE-E2E-003 | Boot | Document title updates after load | Happy | P2 | Title becomes `<n> products found | Refurbed` after loading. |
| FE-E2E-004 | Search | Typing updates input immediately | Happy | P1 | Input reflects typed value without waiting for network. |
| FE-E2E-005 | Search | Debounced request after 500ms idle | Happy | P0 | Request is sent only after typing pause; not one request per keystroke. |
| FE-E2E-006 | Search | Rapid typing triggers single effective result update | Edge | P1 | Final results match final query; no stale query remains visible. |
| FE-E2E-007 | Search | Search change resets pagination offset | Happy | P1 | URL `offset` removed/reset and list reloads from first page. |
| FE-E2E-008 | Search | Clearing search returns to full dataset | Happy | P1 | Full list/count restored. |
| FE-E2E-009 | Filters | Draft filters do not auto-apply (non-search fields) | Happy | P0 | Changing category/color/price/etc does not fetch until `Apply`. |
| FE-E2E-010 | Filters | Unapplied warning appears after draft edits | Happy | P1 | Warning text appears when draft differs from applied. |
| FE-E2E-011 | Filters | Apply commits draft and fetches | Happy | P0 | Results and URL update only after `Apply`. |
| FE-E2E-012 | Filters | Reset clears and applies immediately | Happy | P1 | Reset clears draft+applied filter state and immediately refreshes results/URL. |
| FE-E2E-013 | Filters | Remove single active chip | Happy | P1 | Chip removed, corresponding filter removed, results/URL refreshed. |
| FE-E2E-014 | Filters | Clear all chips | Happy | P1 | All chips removed and filters reset to default applied state. |
| FE-E2E-015 | Filters | Clear all filters button from empty state | Happy | P0 | All filters cleared and full results restored. |
| FE-E2E-016 | Filters | Category multi-select apply | Happy | P1 | Applied results satisfy selected categories only. |
| FE-E2E-017 | Filters | Condition multi-select apply | Happy | P1 | Applied results satisfy selected conditions only. |
| FE-E2E-018 | Filters | Color multi-select apply | Happy | P1 | Applied results include products matching any selected color. |
| FE-E2E-019 | Filters | Bestseller radio apply | Happy | P1 | `bestseller=true` applied correctly; results and URL reflect it. |
| FE-E2E-020 | Filters | On-sale radio apply | Happy | P1 | `onSale=true` applied correctly; only discounted products returned. |
| FE-E2E-021 | Filters | In-stock radio apply | Happy | P1 | `inStock=true` applied correctly; out-of-stock excluded. |
| FE-E2E-022 | Price | Slider min/max visual updates | Happy | P1 | Min/max badges and highlighted range update while sliding. |
| FE-E2E-023 | Price | Min thumb cannot pass max thumb | Edge | P1 | Min emitted/applied value is clamped to keep at least 1 EUR gap when possible. |
| FE-E2E-024 | Price | Max thumb cannot go below min thumb | Edge | P1 | Max emitted/applied value is clamped to keep at least 1 EUR gap when possible. |
| FE-E2E-025 | Price | Slider bounds come from API price_min/price_max | Happy | P0 | Floor/ceiling labels and slider range reflect backend bounds, not hardcoded values. |
| FE-E2E-073 | Price | Crossing min over max keeps non-zero range and still returns data | Edge | P1 | Forced cross clamps to valid range (`min < max`) and matching results are returned. |
| FE-E2E-074 | Price | Crossing at upper bound still preserves minimum slider gap | Edge | P2 | Even at upper bound, forced cross keeps a non-zero range instead of collapsing to a point. |
| FE-E2E-026 | Price | Invalid price combination blocked from apply | Edge | P1 | Apply remains disabled and validation message shown for invalid price state. |
| FE-E2E-027 | Sorting | Sort default + staged apply behavior | Happy | P1 | Default sort is empty and URL has no `sort` param; multi-select changes are staged while dropdown is open and applied when it closes. |
| FE-E2E-028 | Sorting | Sort by popularity | Happy | P0 | Request includes `sort=popularity`; order reflects popularity ranking. |
| FE-E2E-029 | Sorting | Removing sort chip resets sort | Happy | P1 | `sort` removed from URL and results revert to default order. |
| FE-E2E-071 | Sorting | Multi-sort with active filters | Happy | P1 | Combined sort (`sort=popularity,price_asc`) keeps active filters and remains deterministic. |
| FE-E2E-072 | URL Sync | Conflicting price sort params normalize safely | Edge | P1 | `sort=price_asc&sort=price_desc` resolves to one effective price direction in UI state. |
| FE-E2E-030 | URL Sync | Applied filters reflected in URL query params | Happy | P0 | URL includes applied keys only (`search`, filters, `sort`, `offset`). |
| FE-E2E-031 | URL Sync | Page load hydrates state from URL | Happy | P0 | Controls/results initialize correctly from query string. |
| FE-E2E-032 | URL Sync | Browser back/forward rehydrates state | Happy | P1 | `popstate` updates filters + products to previous URL state. |
| FE-E2E-033 | URL Sync | Unknown/invalid URL values degrade safely | Edge | P2 | App does not crash; unsupported values (including invalid `minPrice`/`maxPrice`) normalize to defaults. |
| FE-E2E-034 | Pagination | Load more appends items | Happy | P0 | New page items appended, previous items retained, count consistent. |
| FE-E2E-035 | Pagination | Load more updates offset in URL | Happy | P1 | URL `offset` increments per page loaded. |
| FE-E2E-036 | Pagination | Load more hidden when `has_more=false` | Happy | P1 | Button disappears/disabled when no more items. |
| FE-E2E-037 | Pagination | Load more disabled while loading | Edge | P1 | No duplicate concurrent load-more requests from repeated clicks. |
| FE-E2E-038 | Product Card | Card shows required fields | Happy | P0 | Image, name, current price, discount badge (if discounted), bestseller badge (if true), color swatches, stock hint. |
| FE-E2E-039 | Product Card | Original price shown only when discounted | Happy | P1 | Strikethrough original price appears only if `discount_percent > 0`. |
| FE-E2E-040 | Product Card | Color selection switches image | Happy | P0 | Clicking a swatch changes displayed product image by selected color mapping. |
| FE-E2E-041 | Product Card | Missing color image mapping falls back to base image | Edge | P1 | Uses `image_url` when selected color missing in `image_urls_by_color`. |
| FE-E2E-042 | Product Card | Broken image URL falls back to placeholder | Edge | P1 | Fallback placeholder image shown on image error. |
| FE-E2E-043 | Product Card | Out-of-stock colors hidden from swatches | Happy | P0 | Colors with zero stock not shown as selectable swatches. |
| FE-E2E-044 | Product Card | Stock label updates by selected color | Happy | P1 | Stock text changes with swatch selection (`n in stock (Color)`). |
| FE-E2E-045 | Product Card | Color selection ring visible in light/dark themes | Happy | P2 | Active swatch is visually distinguishable in both themes. |
| FE-E2E-046 | Colors | Filter color options come from API `available_colors` | Happy | P0 | Color filter list reflects backend-supplied available colors. |
| FE-E2E-047 | Colors | Selected colors remain available after refreshes | Edge | P2 | Selected color tokens persist in options list while selected. |
| FE-E2E-076 | Brands | Filter brand options come from API `available_brands` | Happy | P0 | Brand filter list reflects backend-supplied available brands. |
| FE-E2E-077 | Brands | Brand filter apply narrows results and syncs URL/chips | Happy | P1 | Applied brand appears in request/URL/chips and results match selected brand(s). |
| FE-E2E-048 | Errors | Inline error on load-more failure | Edge | P1 | Error banner appears above list when list exists and request fails. |
| FE-E2E-049 | Errors | Fatal error state when initial load fails | Edge | P0 | Full error state with retry button displayed when no items loaded. |
| FE-E2E-050 | Errors | Retry from inline/fatal states | Happy | P1 | Retry triggers appropriate request and recovers UI on success. |
| FE-E2E-051 | Empty State | Empty results UI shown with guidance | Happy | P0 | Empty state text and clear-all action displayed for zero matches. |
| FE-E2E-052 | Theme | Dark mode toggle works | Happy | P1 | Theme class toggles and visual palette changes. |
| FE-E2E-053 | Theme | Theme preference persists across reload | Happy | P1 | `localStorage` preference reapplied on refresh. |
| FE-E2E-054 | Responsive | Desktop shows sidebar filters | Happy | P0 | Sidebar visible at desktop breakpoint, mobile floating button hidden. |
| FE-E2E-055 | Responsive | Mobile shows floating filters button | Happy | P0 | Floating filters button visible on mobile. |
| FE-E2E-056 | Responsive | Mobile filters open as modal | Happy | P0 | Bottom-sheet/modal appears with overlay and close control. |
| FE-E2E-057 | Responsive | Mobile overlay click closes modal | Happy | P1 | Clicking backdrop closes modal. |
| FE-E2E-058 | Responsive | Mobile `Esc` closes modal | Happy | P1 | Escape key closes modal. |
| FE-E2E-059 | Responsive | Mobile focus trap in modal | Edge | P0 | `Tab`/`Shift+Tab` cycles focus inside modal while open. |
| FE-E2E-060 | Responsive | Focus returns to trigger after modal close | Edge | P1 | Focus restored to filters trigger button on close. |
| FE-E2E-061 | Responsive | Body scroll lock while modal open | Edge | P1 | Background page scrolling disabled while modal is open. |
| FE-E2E-062 | Floating UX | Scroll-to-top button appears on mobile after threshold | Happy | P1 | `Top` button appears after scrolling down. |
| FE-E2E-063 | Floating UX | Scroll-to-top button scrolls page to top | Happy | P1 | Smooth scroll reaches near top and button hides again. |
| FE-E2E-075 | Mobile UX | Active chips are collapsed by default with summary/toggle | Happy | P1 | Mobile shows compact `N filters applied` row, supports Show/Hide and clear-all without layout overload. |
| FE-E2E-064 | A11y | Landmark/ARIA attributes present | Happy | P1 | Main/section semantics, `aria-live`, `aria-busy`, dialog attrs, pressed states exist. |
| FE-E2E-065 | A11y | Keyboard access for filter controls and chips | Happy | P1 | All controls operable via keyboard (tab + space/enter). |
| FE-E2E-066 | A11y | Search input has accessible label | Happy | P1 | Search field has associated accessible name (`label`/`sr-only`). |
| FE-E2E-067 | A11y | Product image alt text includes product context | Happy | P1 | Alt includes product name and selected color context when available. |
| FE-E2E-068 | Draft/Apply UX | Sticky bottom apply bar appears when top controls out of view | Happy | P1 | Bottom sticky action bar appears after scrolling filter panel. |
| FE-E2E-069 | Draft/Apply UX | Unapplied warning shown in sticky bar when pending | Happy | P1 | Warning moves to sticky action area when top controls not visible. |
| FE-E2E-070 | Stability | App does not throw uncaught errors during major flows | Edge | P0 | No uncaught exceptions for search/filter/sort/pagination/theme/mobile modal flows. |

## Automation Status (Current)

Status legend:
- `Automated`: covered by passing Playwright tests.
- `Not Automated`: not implemented in the current E2E suite.
- `N/A`: scenario is not reachable in current UI implementation.

Last run: `npm run e2e` -> **29 passed, 0 failed**.

| IDs | Status | Evidence |
|---|---|---|
| FE-E2E-001, FE-E2E-003, FE-E2E-054 | Automated | `e2e/app.desktop.spec.js` test: `initial shell and title` |
| FE-E2E-002 | Automated | `e2e/app.desktop.spec.js` test: `initial skeletons are shown while first request is pending` |
| FE-E2E-004, FE-E2E-005, FE-E2E-006, FE-E2E-007, FE-E2E-008 | Automated | `e2e/app.desktop.spec.js` test: `search is immediate in input and debounced in network` |
| FE-E2E-009, FE-E2E-010, FE-E2E-011, FE-E2E-012 | Automated | `e2e/app.desktop.spec.js` test: `draft filters require apply and reset applies immediately` |
| FE-E2E-013, FE-E2E-014 | Automated | `e2e/app.desktop.spec.js` test: `chips can remove one filter and clear all` |
| FE-E2E-016, FE-E2E-017, FE-E2E-018, FE-E2E-019, FE-E2E-020, FE-E2E-021, FE-E2E-030 | Automated | `e2e/app.desktop.spec.js` test: `applied filters are reflected in request and URL` |
| FE-E2E-022, FE-E2E-023, FE-E2E-024, FE-E2E-025 | Automated | `e2e/app.desktop.spec.js` test: `price slider updates badges, clamps, and uses API bounds` |
| FE-E2E-073 | Automated | `e2e/app.desktop.spec.js` test: `crossing min over max keeps a valid range and still returns data` |
| FE-E2E-074 | Automated | `e2e/app.desktop.spec.js` test: `crossing at upper bound keeps a minimum 1 EUR gap` |
| FE-E2E-026 | N/A | Invalid min/max is clamped by dual-slider UI; an invalid draft state cannot be produced via controls. |
| FE-E2E-027, FE-E2E-028, FE-E2E-029 | Automated | `e2e/app.desktop.spec.js` test: `sorting options can be applied and removed` |
| FE-E2E-071 | Automated | `e2e/app.desktop.spec.js` test: `multi sorting works with active filters and preserves filter query params` |
| FE-E2E-072 | Automated | `e2e/app.desktop.spec.js` test: `conflicting sort params in URL normalize safely to a single effective sort` |
| FE-E2E-015 | Automated | `e2e/app.desktop.spec.js` test: `empty state is shown and clear-all restores full results` |
| FE-E2E-031, FE-E2E-033 | Automated | `e2e/app.desktop.spec.js` test: `URL hydration works and invalid values degrade safely` |
| FE-E2E-032 | Automated | `e2e/app.desktop.spec.js` test: `browser back and forward rehydrate filter state` |
| FE-E2E-034, FE-E2E-035, FE-E2E-036, FE-E2E-037 | Automated | `e2e/app.desktop.spec.js` test: `load more appends results once and updates offset` |
| FE-E2E-038, FE-E2E-039, FE-E2E-040, FE-E2E-043, FE-E2E-044 | Automated | `e2e/app.desktop.spec.js` test: `product cards show badges, swatches, stock by color, and image switching` |
| FE-E2E-041 | Automated | `e2e/app.desktop.spec.js` test: `missing color image mapping falls back to base image` |
| FE-E2E-042 | Automated | `e2e/app.desktop.spec.js` test: `broken image URL falls back to local placeholder` |
| FE-E2E-045, FE-E2E-052, FE-E2E-053 | Automated | `e2e/app.desktop.spec.js` test: `dark mode toggles, persists, and selected swatch remains visually active` |
| FE-E2E-046, FE-E2E-047 | Automated | `e2e/app.desktop.spec.js` test: `color filter options come from API and selected value persists` |
| FE-E2E-076, FE-E2E-077 | Automated | `e2e/app.desktop.spec.js` test: `brand filter options come from API and apply correctly` |
| FE-E2E-048, FE-E2E-049, FE-E2E-050 | Automated | `e2e/app.desktop.spec.js` test: `error states and retry flows work for initial load and load more` |
| FE-E2E-051 | Automated | `e2e/app.desktop.spec.js` test: `empty state is shown and clear-all restores full results` |
| FE-E2E-055, FE-E2E-056 | Automated | `e2e/app.mobile.spec.js` test: `floating filters button opens modal` |
| FE-E2E-057 | Automated | `e2e/app.mobile.spec.js` test: `clicking modal backdrop closes filters modal` |
| FE-E2E-058, FE-E2E-059, FE-E2E-060, FE-E2E-061 | Automated | `e2e/app.mobile.spec.js` test: `escape closes modal, focus is trapped/restored, and body scroll is locked` |
| FE-E2E-062, FE-E2E-063 | Automated | `e2e/app.mobile.spec.js` test: `floating top button appears on scroll and returns to top` |
| FE-E2E-075 | Automated | `e2e/app.mobile.spec.js` test: `mobile active filters chips are collapsed by default with summary and toggle` |
| FE-E2E-064, FE-E2E-065, FE-E2E-066, FE-E2E-067, FE-E2E-068, FE-E2E-069, FE-E2E-070 | Automated | `e2e/app.desktop.spec.js` test: `a11y and sticky apply UX stay healthy under keyboard flow` |

## Nice-to-Have Extended E2E Cases

| ID | Scenario | Priority | Expected Result |
|---|---|---|---|
| FE-E2E-X01 | Verify request cancellation behavior on rapid state changes | P2 | Latest request wins and UI does not flash stale data. |
| FE-E2E-X02 | Network throttling simulation for loading UX timings | P2 | Skeleton/error/inline states behave consistently under slow network. |
| FE-E2E-X03 | Cross-browser project parity (WebKit/Firefox) | P3 | Core flows pass across engines. |
| FE-E2E-X04 | Visual snapshot tests for critical layouts (desktop/mobile) | P3 | No major layout regressions in card grid/filter modal. |
