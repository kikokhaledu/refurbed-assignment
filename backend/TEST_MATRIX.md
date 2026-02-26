# Backend Test Coverage Matrix

Status legend:
- `Covered`: directly verified by one or more tests.
- `Partially covered`: key path is tested, but some branches or operational concerns are not.
- `Not covered`: no direct automated test in current suite.

| Area | Status | Notes |
|---|---|---|
| Query parsing defaults, validation, boundaries | Covered | `query_test.go` covers defaults, strict validation, invalid inputs, boundaries, and degenerate token input. |
| Filtering behavior (`search`, `category`, `brand`, `condition`, `color`, `bestseller`, `onSale`, `inStock`, `minStock`, price bounds) | Covered | `service_test.go` has scenarios for filter combinations (including brand), inclusive bounds, and color-scoped stock semantics. |
| Integer point-price semantics for slider-style requests (`minPrice=n&maxPrice=n`) | Covered | `query_test.go`, `service_test.go`, and `http_test.go` verify integer upper-bound expansion to include cent prices (`n..n.99`). |
| Pagination semantics (`limit`, `offset`, load-more shape) | Covered | `service_test.go` validates paging behavior including `offset > total` response semantics. |
| Aggregation/merge correctness from two sources | Covered | `service_test.go` validates merge output, duplicate/empty IDs, price calculation, stock/image normalization behavior. |
| Price computation precision | Covered | `TestDiscountedPriceCents_RoundsAtCentPrecision`. |
| Cache TTL, refresh, stale fallback, anti-stampede, wait cancellation | Covered | `service_test.go` includes TTL hit/miss, stale-on-error, single refresh fan-in, and cancellation while waiting. |
| Sorting modes (`sort=popularity`, `sort=price_asc`, `sort=price_desc`) plus non-contradicting multi-sort combinations and non-fatal popularity source failure | Covered | `service_test.go` and `query_test.go` cover accepted sort modes, combined ordering behavior, conflict rejection, and popularity-source fallback. |
| Repository file loading (missing file, malformed JSON, context cancel, null/missing scalar behavior) | Covered | `repository_test.go`. |
| HTTP handler method validation, bad query, success path, internal error JSON, CORS OPTIONS | Covered | `http_test.go`. |
| CORS behavior for non-OPTIONS requests | Covered | `http_test.go` validates GET header behavior, and `main_test.go` validates middleware-wrapped `/health` GET. |
| Popularity data normalization edge cases (duplicate IDs, empty IDs, invalid rank values) | Partially covered | Happy path and source failure are covered; invalid ranking payload branches are not directly unit-tested. |
| Cached snapshot facet reuse (`available_colors`, `available_brands`, `price_min`, `price_max`) | Covered | `service.go` precomputes facets in `buildProductSnapshot`; query tests and service tests exercise stable response facets through repeated requests. |
| `/health` endpoint behavior | Covered | `main_test.go` validates GET 200 payload and method-not-allowed behavior. |
| Server bootstrap and graceful shutdown wiring (`main.go`, signal handling, timeout config) | Partially covered | `main_test.go` verifies handler bootstrap/route registration and middleware stack; process-level signal/shutdown wiring remains untested. |
| Logging middleware output format/content | Covered | `main_test.go` captures logs and asserts method/path entries for `/health` requests. |
| Load/performance/soak behavior | Not covered | No benchmark or load-test suite in repository. |
| Race detector execution in CI-like environment | Partially covered | Code is race-conscious and `go test -race` is documented, but compiler/runtime availability can block it in some environments. |

## Additional Edge-Case Gaps (Current)

| Edge Case | Status | Notes |
|---|---|---|
| Repeated singleton params (`bestseller`, `inStock`, `onSale`, `minPrice`, `maxPrice`, `minStock`, `limit`, `offset`) | Covered | `TestParseProductQuery_RejectsRepeatedSingletonParams` verifies strict rejection of repeated singleton values. |
| Empty singleton params (`bestseller=`, `inStock=`, `onSale=`, `minPrice=`, `maxPrice=`, `minStock=`, `limit=`, `offset=`) | Covered | `TestParseProductQuery_RejectsEmptySingletonParams` verifies strict rejection instead of silent fallback. |
| Case-insensitive sort parsing (`sort=POPULARITY`, `sort=PRICE_DESC`) | Covered | `TestParseProductQuery_SortModes` validates mixed/upper-case normalization for sort values. |
| Repeated conflicting sort params (`sort=price_asc&sort=price_desc`) | Covered | `TestParseProductQuery_ConflictingPriceSortsRejected` verifies contradictory price directions are rejected. |
| Integer point-price filter (`minPrice=n&maxPrice=n`) includes decimal prices in that euro bucket | Covered | `TestParseProductQuery_IntegerMaxPriceExpandsToEuroCeiling`, `TestProductService_IntegerPricePointIncludesCentPrices`, and `TestProductHandler_IntegerPointPriceBucketIncludesCentPrices`. |
| Unknown color filter against `stock_by_color` map | Partially covered | Color-scoped stock logic is covered, but unknown-color behavior is not isolated as a dedicated case. |
| Dataset with `stock_by_color` present but `colors` missing/empty | Partially covered | Normalization behavior exists; no focused test for this shape. |
| Dataset with `image_urls_by_color` present but `colors` missing/empty | Partially covered | Merge logic supports adding colors from image map keys; no dedicated edge test. |
| `available_colors` when `stock_by_color` absent and aggregate `stock <= 0` | Not covered | Out-of-stock color exclusion is tested for `stock_by_color`; aggregate-stock-only path is not isolated. |
| `limit` clamping from service sanitization (direct service call with oversized limit) | Not covered | Query parser rejects `>100`, but direct `ProductQuery` service path clamp is not tested. |
| `offset < 0` sanitization in service (direct service call) | Not covered | Parser rejects negative offsets; service-level sanitization path is not directly tested. |
| Popularity rank tie-break determinism (same rank -> name -> id) | Not covered | Popularity sorting is tested, but explicit tie-break rules are not isolated in a dedicated assertion. |
| Invalid popularity data fallback path (`duplicate id`, `rank <= 0`, empty id) | Not covered | Branch exists (logs and continues), but no direct test through query/service path. |
| Cached product immutability under caller mutation attempts (`Colors`, `StockByColor`, `ImageURLsByColor`) | Covered | `TestProductService_QueryProductsResponseIsImmutableFromCallerMutations` verifies response-level deep clone safety across cache hits. |
| `/products` response `Content-Type` contract on both success and errors | Partially covered | Body/status behavior is asserted; explicit header assertions are limited. |
| CORS headers on non-OPTIONS product requests | Covered | `http_test.go` and `main_test.go` assert CORS headers on GET requests. |
| `/health` endpoint wrapped with middleware stack (`withCORS`, `withLogging`) | Covered | `main_test.go` exercises `buildServerHandler` and asserts both CORS and logging behavior. |
| Shutdown path in `main.go` when context is canceled | Not covered | Graceful shutdown wiring is present but untested. |
