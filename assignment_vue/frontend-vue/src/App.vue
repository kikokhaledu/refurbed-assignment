<template>
  <main data-testid="app-root" class="min-h-screen bg-slate-100 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
    <div class="mx-auto max-w-[1320px] px-4 py-8 sm:px-6 lg:px-8">
      <header class="space-y-4">
        <div class="flex items-center justify-between gap-4">
          <h1 class="text-2xl font-extrabold tracking-tight sm:text-3xl">Product discovery</h1>
          <div class="flex items-center gap-3">
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">Dark mode</span>
            <button
              type="button"
              role="switch"
              data-testid="theme-toggle"
              :aria-checked="isDarkMode ? 'true' : 'false'"
              class="relative h-7 w-14 overflow-hidden rounded-full border border-slate-300 bg-white transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900"
              :class="isDarkMode ? 'bg-slate-800' : 'bg-white'"
              @click="toggleDarkMode"
            >
              <span
                aria-hidden="true"
                class="absolute left-0.5 top-0.5 h-6 w-6 rounded-full bg-slate-200 shadow transition-transform duration-200 dark:bg-cyan-300"
                :class="isDarkMode ? 'translate-x-0' : 'translate-x-7'"
              ></span>
            </button>
          </div>
        </div>

        <SearchBar
          :model-value="draftFilters.search"
          :disabled="isLoading || isLoadingMore"
          @update:model-value="handleSearchUpdate"
        />
      </header>

      <div class="mt-6 grid gap-6 lg:grid-cols-[260px,1fr]">
        <aside class="hidden lg:block">
          <FiltersPanel
            :selected-categories="draftFilters.categories"
            :selected-brands="draftFilters.brands"
            :selected-colors="draftFilters.colors"
            :selected-conditions="draftFilters.conditions"
            :bestseller="draftFilters.bestseller"
            :on-sale="draftFilters.onSale"
            :in-stock="draftFilters.inStock"
            :min-price="draftFilters.minPrice"
            :max-price="draftFilters.maxPrice"
            :price-floor="priceFloor"
            :price-ceiling="priceCeiling"
            :category-options="categoryOptions"
            :brand-options="brandOptions"
            :color-options="colorOptions"
            :condition-options="conditionOptions"
            :validation-message="filterValidationMessage"
            :has-pending-changes="hasPendingChanges"
            :apply-disabled="isApplyDisabled"
            :disabled="isLoading"
            @update:selected-categories="handleCategoriesUpdate"
            @update:selected-brands="handleBrandsUpdate"
            @update:selected-colors="handleColorsUpdate"
            @update:selected-conditions="handleConditionsUpdate"
            @update:bestseller="handleBestsellerUpdate"
            @update:onSale="handleOnSaleUpdate"
            @update:in-stock="handleInStockUpdate"
            @update:min-price="handleMinPriceUpdate"
            @update:max-price="handleMaxPriceUpdate"
            @apply="applyFilters"
            @clear="resetDraftFilters"
          />
        </aside>

        <section data-testid="results-section" aria-live="polite" :aria-busy="isLoading ? 'true' : 'false'">
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p data-testid="products-count" class="text-3xl font-semibold tracking-tight text-slate-600 dark:text-slate-300">
              {{ total }} products found
            </p>
            <SortSelect
              :model-value="appliedFilters.sort"
              :disabled="isLoading || isLoadingMore"
              @update:model-value="handleSortUpdate"
            />
          </div>

          <ActiveFilterChips
            :chips="activeFilterChips"
            @remove-chip="removeAppliedChip"
            @clear-all="clearFiltersAndApply"
          />

          <div
            v-if="inlineErrorMessage"
            data-testid="inline-error-state"
            class="mb-4 flex items-center justify-between rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-800 dark:bg-rose-950 dark:text-rose-300"
            role="alert"
          >
            <span>{{ inlineErrorMessage }}</span>
            <button
              type="button"
              data-testid="inline-retry-button"
              class="rounded-md border border-rose-300 bg-white px-3 py-1 font-medium text-rose-700 transition hover:bg-rose-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-rose-500 dark:border-rose-700 dark:bg-rose-900 dark:text-rose-100 dark:hover:bg-rose-800"
              @click="retryFailedRequest"
            >
              Retry
            </button>
          </div>

          <section
            v-if="showFatalErrorState"
            data-testid="fatal-error-state"
            class="grid min-h-[360px] place-items-center rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm dark:border-slate-800 dark:bg-slate-900"
            aria-label="Error state"
          >
            <div>
              <p class="text-xl font-semibold text-slate-800 dark:text-slate-100">Could not load products</p>
              <p class="mt-2 text-slate-500 dark:text-slate-400">{{ errorMessage }}</p>
              <button
                type="button"
                data-testid="fatal-retry-button"
                class="mt-5 rounded-xl border border-slate-300 bg-white px-4 py-2 font-semibold text-slate-800 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
                @click="retryFailedRequest"
              >
                Retry
              </button>
            </div>
          </section>

          <section
            v-else-if="showEmptyState"
            data-testid="empty-state"
            class="grid min-h-[360px] place-items-center rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm dark:border-slate-800 dark:bg-slate-900"
            aria-label="Empty state"
          >
            <div>
              <p class="text-2xl font-semibold text-slate-800 dark:text-slate-100">No products found</p>
              <p class="mt-2 text-slate-500 dark:text-slate-400">
                We could not find products matching your current filters.
              </p>
              <button
                type="button"
                data-testid="empty-clear-filters-button"
                class="mt-5 rounded-xl border border-slate-300 bg-white px-4 py-2 font-semibold text-slate-800 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
                @click="clearFiltersAndApply"
              >
                Clear all filters
              </button>
            </div>
          </section>

          <ProductGrid
            v-else
            :products="products"
            :loading="isLoading"
            :skeleton-count="pageSize"
          />

          <LoadMoreButton
            :show="products.length > 0 && hasMore && !isLoading"
            :loading="isLoadingMore"
            :disabled="isLoadingMore"
            @click="loadMore"
          />
        </section>
      </div>

      <div
        v-if="isMobileFiltersOpen"
        id="filters-modal"
        data-testid="mobile-filters-modal"
        class="fixed inset-0 z-50 lg:hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="filters-modal-title"
      >
        <div
          data-testid="mobile-filters-backdrop"
          class="absolute inset-0 bg-slate-950/45"
          @click="closeMobileFilters"
        ></div>
        <div class="absolute inset-x-0 bottom-0 z-10 max-h-[88vh] overflow-y-auto px-4 pb-4">
          <section
            ref="mobileFiltersDialogRef"
            class="rounded-2xl"
            tabindex="-1"
          >
            <div class="mb-3 flex items-center justify-between">
              <h2 id="filters-modal-title" class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                Filters
              </h2>
              <button
                ref="mobileFiltersCloseButtonRef"
                type="button"
                data-testid="mobile-filters-close"
                class="rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
                @click="closeMobileFilters"
              >
                Close
              </button>
            </div>

            <FiltersPanel
              :selected-categories="draftFilters.categories"
              :selected-brands="draftFilters.brands"
              :selected-colors="draftFilters.colors"
              :selected-conditions="draftFilters.conditions"
              :bestseller="draftFilters.bestseller"
              :on-sale="draftFilters.onSale"
              :in-stock="draftFilters.inStock"
              :min-price="draftFilters.minPrice"
              :max-price="draftFilters.maxPrice"
              :price-floor="priceFloor"
              :price-ceiling="priceCeiling"
              :category-options="categoryOptions"
              :brand-options="brandOptions"
              :color-options="colorOptions"
              :condition-options="conditionOptions"
              :validation-message="filterValidationMessage"
              :has-pending-changes="hasPendingChanges"
              :apply-disabled="isApplyDisabled"
              :disabled="isLoading"
              @update:selected-categories="handleCategoriesUpdate"
              @update:selected-brands="handleBrandsUpdate"
              @update:selected-colors="handleColorsUpdate"
              @update:selected-conditions="handleConditionsUpdate"
              @update:bestseller="handleBestsellerUpdate"
              @update:onSale="handleOnSaleUpdate"
              @update:in-stock="handleInStockUpdate"
              @update:min-price="handleMinPriceUpdate"
              @update:max-price="handleMaxPriceUpdate"
              @apply="applyFilters"
              @clear="resetDraftFilters"
            />
          </section>
        </div>
      </div>
    </div>

    <div class="fixed bottom-4 right-4 z-40 flex flex-col gap-2 lg:hidden">
      <button
        ref="mobileFiltersTriggerRef"
        type="button"
        data-testid="mobile-filters-trigger"
        class="rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 shadow-lg transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
        :aria-expanded="isMobileFiltersOpen ? 'true' : 'false'"
        aria-controls="filters-modal"
        @click="toggleMobileFilters"
      >
        {{ isMobileFiltersOpen ? 'Hide filters' : 'Filters' }}
      </button>
      <button
        v-if="showScrollTopButton"
        type="button"
        data-testid="mobile-scroll-top"
        class="rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-800 shadow-lg transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
        @click="scrollToTop"
      >
        Top
      </button>
    </div>
  </main>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import ActiveFilterChips from './components/ActiveFilterChips.vue'
import FiltersPanel from './components/FiltersPanel.vue'
import LoadMoreButton from './components/LoadMoreButton.vue'
import ProductGrid from './components/ProductGrid.vue'
import SearchBar from './components/SearchBar.vue'
import SortSelect from './components/SortSelect.vue'
import { useMobileFilterDialog } from './composables/useMobileFilterDialog'
import { useTheme } from './composables/useTheme'
import {
  CATEGORY_OPTIONS,
  CONDITION_OPTIONS,
  DEFAULT_BACKEND_PORT,
  MOBILE_SCROLL_TOP_THRESHOLD,
  PAGE_SIZE,
  SEARCH_DEBOUNCE_MS,
  THEME_STORAGE_KEY,
} from './constants/app'
import {
  areFilterStatesEqual,
  assignFilterState,
  buildFilterValidationMessage,
  createDefaultFilters,
  formatTokenLabel,
  normalizeBestsellerValue,
  normalizeFilterState,
  normalizeInStockValue,
  normalizeOnSaleValue,
  normalizeTokenList,
  sortLabel,
} from './utils/filterState'
import { buildURLSearchParams, parseOffset, parseURLState as parseURLStateFromSearch } from './utils/productQueryState'
import { normalizeSortValue } from './utils/sort'

function parsePositiveInt(rawValue, fallback) {
  const parsed = Number.parseInt(String(rawValue ?? ''), 10)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return parsed
}

function buildFallbackApiBaseUrl(port) {
  if (typeof window === 'undefined') {
    return `http://localhost:${port}`
  }

  const url = new URL(window.location.origin)
  url.port = String(port)
  return url.origin
}

const configuredApiBaseUrl = String(import.meta.env.VITE_API_BASE_URL ?? '').trim()
const fallbackBackendPort = parsePositiveInt(import.meta.env.VITE_BACKEND_PORT, DEFAULT_BACKEND_PORT)
const fallbackApiBaseUrl = buildFallbackApiBaseUrl(fallbackBackendPort)
const apiBaseUrl = configuredApiBaseUrl || fallbackApiBaseUrl

const pageSize = PAGE_SIZE
const searchDebounceMs = SEARCH_DEBOUNCE_MS
const categoryOptions = CATEGORY_OPTIONS
const conditionOptions = CONDITION_OPTIONS

const colorOptions = ref([])
const brandOptions = ref([])
const priceFloor = ref(0)
const priceCeiling = ref(0)

const {
  isMobileFiltersOpen,
  mobileFiltersDialogRef,
  mobileFiltersCloseButtonRef,
  mobileFiltersTriggerRef,
  closeMobileFilters,
  toggleMobileFilters,
} = useMobileFilterDialog()

const { isDarkMode, initializeTheme, toggleDarkMode } = useTheme(THEME_STORAGE_KEY)

const draftFilters = reactive(createDefaultFilters())
const appliedFilters = reactive(createDefaultFilters())

const products = ref([])
const total = ref(0)
const hasMore = ref(false)
const currentOffset = ref(0)

const isLoading = ref(false)
const isLoadingMore = ref(false)
const errorMessage = ref('')
const lastFailedRequest = ref('reload')
const showScrollTopButton = ref(false)

let activeController = null
let searchApplyTimer = null

const inlineErrorMessage = computed(() => {
  if (products.value.length === 0 || !errorMessage.value) {
    return ''
  }
  return errorMessage.value
})

const showFatalErrorState = computed(() => {
  return products.value.length === 0 && !isLoading.value && errorMessage.value.length > 0
})

const showEmptyState = computed(() => {
  return !isLoading.value && !errorMessage.value && products.value.length === 0
})

const filterValidationMessage = computed(() => buildFilterValidationMessage(draftFilters))

const hasPendingChanges = computed(() => {
  return !areFilterStatesEqual(normalizeFilterState(draftFilters), normalizeFilterState(appliedFilters))
})

const isApplyDisabled = computed(() => {
  return (
    !hasPendingChanges.value ||
    Boolean(filterValidationMessage.value) ||
    isLoading.value ||
    isLoadingMore.value
  )
})

const activeFilterChips = computed(() => {
  const normalized = normalizeFilterState(appliedFilters)
  const chips = []

  if (normalized.search) {
    chips.push({ id: 'search', field: 'search', label: `Search: "${normalized.search}"` })
  }

  for (const category of normalized.categories) {
    chips.push({
      id: `category:${category}`,
      field: 'categories',
      value: category,
      label: `Category: ${formatTokenLabel(category)}`,
    })
  }

  for (const brand of normalized.brands) {
    chips.push({
      id: `brand:${brand}`,
      field: 'brands',
      value: brand,
      label: `Brand: ${formatTokenLabel(brand)}`,
    })
  }

  for (const color of normalized.colors) {
    chips.push({
      id: `color:${color}`,
      field: 'colors',
      value: color,
      label: `Color: ${formatTokenLabel(color)}`,
    })
  }

  for (const condition of normalized.conditions) {
    chips.push({
      id: `condition:${condition}`,
      field: 'conditions',
      value: condition,
      label: `Condition: ${formatTokenLabel(condition)}`,
    })
  }

  if (normalized.bestseller === 'true') {
    chips.push({ id: 'bestseller', field: 'bestseller', label: 'Bestseller only' })
  }

  if (normalized.onSale === 'true') {
    chips.push({ id: 'onSale', field: 'onSale', label: 'On sale only' })
  }

  if (normalized.inStock === 'true') {
    chips.push({ id: 'inStock', field: 'inStock', label: 'In stock only' })
  }

  if (normalized.minPrice) {
    chips.push({ id: 'minPrice', field: 'minPrice', label: `Min price: EUR ${normalized.minPrice}` })
  }

  if (normalized.maxPrice) {
    chips.push({ id: 'maxPrice', field: 'maxPrice', label: `Max price: EUR ${normalized.maxPrice}` })
  }

  if (normalized.sort) {
    chips.push({ id: 'sort', field: 'sort', label: `Sort: ${sortLabel(normalized.sort)}` })
  }

  return chips
})

watch(
  [total, isLoading],
  () => {
    if (typeof document === 'undefined') {
      return
    }

    if (isLoading.value) {
      document.title = 'Loading products...'
      return
    }
    document.title = `${total.value} products found | Refurbed`
  },
  { immediate: true },
)

function parseURLState() {
  if (typeof window === 'undefined') {
    return { filters: normalizeFilterState(createDefaultFilters()), offset: 0 }
  }

  return parseURLStateFromSearch(window.location.search, pageSize)
}

function syncURLState(historyMode = 'replace') {
  if (typeof window === 'undefined') {
    return
  }

  const params = buildURLSearchParams(appliedFilters, {
    includeOffset: currentOffset.value > 0,
    offset: currentOffset.value,
    pageSize,
  })

  const query = params.toString()
  const nextURL = query ? `${window.location.pathname}?${query}` : window.location.pathname
  const currentURL = `${window.location.pathname}${window.location.search}`
  if (currentURL === nextURL) {
    return
  }

  if (historyMode === 'push') {
    window.history.pushState(null, '', nextURL)
    return
  }

  window.history.replaceState(null, '', nextURL)
}

function buildApiQuery(offset) {
  return buildURLSearchParams(appliedFilters, {
    includeLimit: true,
    includeOffset: true,
    pageSize,
    offset,
  })
}

function extractColorOptionsFromPayload(payload) {
  if (Array.isArray(payload?.available_colors)) {
    return normalizeTokenList(payload.available_colors)
  }

  if (Array.isArray(payload?.items)) {
    return normalizeTokenList(
      payload.items.flatMap((item) => {
        if (Array.isArray(item?.colors)) {
          return item.colors
        }
        return []
      }),
    )
  }

  return []
}

function extractBrandOptionsFromPayload(payload) {
  if (Array.isArray(payload?.available_brands)) {
    return normalizeTokenList(payload.available_brands)
  }

  if (Array.isArray(payload?.items)) {
    return normalizeTokenList(
      payload.items
        .map((item) => item?.brand)
        .filter(Boolean),
    )
  }

  return []
}

function syncColorOptionsFromPayload(payload) {
  const available = extractColorOptionsFromPayload(payload)
  const selected = normalizeTokenList([...draftFilters.colors, ...appliedFilters.colors])
  const merged = normalizeTokenList([...available, ...selected])
  colorOptions.value = merged
}

function syncBrandOptionsFromPayload(payload) {
  const available = extractBrandOptionsFromPayload(payload)
  const selected = normalizeTokenList([...draftFilters.brands, ...appliedFilters.brands])
  const merged = normalizeTokenList([...available, ...selected])
  brandOptions.value = merged
}

function syncPriceBoundsFromPayload(payload) {
  const parsedMin = Number(payload?.price_min)
  const parsedMax = Number(payload?.price_max)

  if (!Number.isFinite(parsedMin) || !Number.isFinite(parsedMax) || parsedMax < parsedMin) {
    return
  }

  priceFloor.value = Math.max(0, Math.floor(parsedMin))
  priceCeiling.value = Math.max(priceFloor.value, Math.ceil(parsedMax))
}

async function fetchProductsPage(offset, signal) {
  const query = buildApiQuery(offset)
  const response = await fetch(`${apiBaseUrl}/products?${query.toString()}`, {
    method: 'GET',
    signal,
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    let message = 'Failed to load products.'
    try {
      const payload = await response.json()
      if (payload?.error) {
        message = payload.error
      }
    } catch {
      // Ignore malformed error payload.
    }
    throw new Error(message)
  }

  return response.json()
}

async function hydrateToOffset(targetOffset = 0, options = {}) {
  const historyMode = options.historyMode === 'push' ? 'push' : 'replace'
  const validationMessage = buildFilterValidationMessage(appliedFilters)
  if (validationMessage) {
    errorMessage.value = validationMessage
    products.value = []
    total.value = 0
    hasMore.value = false
    currentOffset.value = 0
    syncURLState(historyMode)
    return
  }

  if (activeController) {
    activeController.abort()
  }

  const controller = new AbortController()
  activeController = controller

  lastFailedRequest.value = 'reload'
  isLoading.value = true
  errorMessage.value = ''
  products.value = []
  total.value = 0
  hasMore.value = false
  currentOffset.value = 0

  let pageOffset = 0
  const normalizedTarget = parseOffset(targetOffset, pageSize)

  try {
    while (true) {
      const payload = await fetchProductsPage(pageOffset, controller.signal)
      syncColorOptionsFromPayload(payload)
      syncBrandOptionsFromPayload(payload)
      syncPriceBoundsFromPayload(payload)

      if (pageOffset === 0) {
        products.value = payload.items
      } else {
        products.value = [...products.value, ...payload.items]
      }

      total.value = payload.total
      hasMore.value = payload.has_more
      currentOffset.value = pageOffset

      if (pageOffset >= normalizedTarget || !payload.has_more) {
        break
      }

      pageOffset += pageSize
    }

    syncURLState(historyMode)
  } catch (error) {
    if (controller.signal.aborted) {
      return
    }

    errorMessage.value = error instanceof Error ? error.message : 'Failed to load products.'
    products.value = []
    total.value = 0
    hasMore.value = false
    currentOffset.value = 0
    syncURLState(historyMode)
  } finally {
    if (activeController === controller) {
      activeController = null
    }
    if (!controller.signal.aborted) {
      isLoading.value = false
    }
  }
}

async function loadMore() {
  if (isLoading.value || isLoadingMore.value || !hasMore.value) {
    return
  }

  if (activeController) {
    activeController.abort()
  }

  const controller = new AbortController()
  activeController = controller

  lastFailedRequest.value = 'load-more'
  isLoadingMore.value = true
  errorMessage.value = ''

  const nextOffset = currentOffset.value + pageSize

  try {
    const payload = await fetchProductsPage(nextOffset, controller.signal)
    syncColorOptionsFromPayload(payload)
    syncBrandOptionsFromPayload(payload)
    syncPriceBoundsFromPayload(payload)

    products.value = [...products.value, ...payload.items]
    total.value = payload.total
    hasMore.value = payload.has_more
    currentOffset.value = nextOffset

    syncURLState('push')
  } catch (error) {
    if (controller.signal.aborted) {
      return
    }
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load products.'
  } finally {
    if (activeController === controller) {
      activeController = null
    }
    if (!controller.signal.aborted) {
      isLoadingMore.value = false
    }
  }
}

function applyNormalizedFilters(nextFilters, historyMode = 'replace') {
  const normalized = normalizeFilterState(nextFilters)
  assignFilterState(draftFilters, normalized)
  assignFilterState(appliedFilters, normalized)
  currentOffset.value = 0
  void hydrateToOffset(0, { historyMode })
}

function applyFilters() {
  if (isApplyDisabled.value) {
    return
  }

  clearSearchApplyTimer()
  closeMobileFilters()
  applyNormalizedFilters(draftFilters, 'push')
}

function handleSearchUpdate(value) {
  draftFilters.search = value
  scheduleSearchApply()
}

function handleCategoriesUpdate(value) {
  draftFilters.categories = normalizeTokenList(value)
}

function handleBrandsUpdate(value) {
  draftFilters.brands = normalizeTokenList(value)
}

function handleColorsUpdate(value) {
  draftFilters.colors = normalizeTokenList(value)
}

function handleConditionsUpdate(value) {
  draftFilters.conditions = normalizeTokenList(value)
}

function handleBestsellerUpdate(value) {
  draftFilters.bestseller = normalizeBestsellerValue(value)
}

function handleOnSaleUpdate(value) {
  draftFilters.onSale = normalizeOnSaleValue(value)
}

function handleInStockUpdate(value) {
  draftFilters.inStock = normalizeInStockValue(value)
}

function handleSortUpdate(value) {
  const normalizedSort = normalizeSortValue(value)
  if (normalizedSort === appliedFilters.sort) {
    return
  }

  draftFilters.sort = normalizedSort
  appliedFilters.sort = normalizedSort
  currentOffset.value = 0
  void hydrateToOffset(0, { historyMode: 'push' })
}

function handleMinPriceUpdate(value) {
  draftFilters.minPrice = value
}

function handleMaxPriceUpdate(value) {
  draftFilters.maxPrice = value
}

function removeAppliedChip(chip) {
  const nextFilters = normalizeFilterState(appliedFilters)

  switch (chip.field) {
    case 'search':
      nextFilters.search = ''
      break
    case 'categories':
      nextFilters.categories = nextFilters.categories.filter((value) => value !== chip.value)
      break
    case 'brands':
      nextFilters.brands = nextFilters.brands.filter((value) => value !== chip.value)
      break
    case 'colors':
      nextFilters.colors = nextFilters.colors.filter((value) => value !== chip.value)
      break
    case 'conditions':
      nextFilters.conditions = nextFilters.conditions.filter((value) => value !== chip.value)
      break
    case 'bestseller':
      nextFilters.bestseller = 'all'
      break
    case 'onSale':
      nextFilters.onSale = 'all'
      break
    case 'inStock':
      nextFilters.inStock = 'all'
      break
    case 'sort':
      nextFilters.sort = ''
      break
    case 'minPrice':
      nextFilters.minPrice = ''
      break
    case 'maxPrice':
      nextFilters.maxPrice = ''
      break
    default:
      return
  }

  clearSearchApplyTimer()
  applyNormalizedFilters(nextFilters, 'push')
}

function resetDraftFilters() {
  clearSearchApplyTimer()
  closeMobileFilters()
  applyNormalizedFilters(createDefaultFilters(), 'push')
}

function clearFiltersAndApply() {
  resetDraftFilters()
}

function clearSearchApplyTimer() {
  if (searchApplyTimer !== null) {
    clearTimeout(searchApplyTimer)
    searchApplyTimer = null
  }
}

function scheduleSearchApply() {
  clearSearchApplyTimer()
  const normalizedSearch = String(draftFilters.search ?? '').trim()

  if (normalizedSearch === appliedFilters.search) {
    return
  }

  searchApplyTimer = setTimeout(() => {
    searchApplyTimer = null
    draftFilters.search = normalizedSearch
    appliedFilters.search = normalizedSearch
    currentOffset.value = 0
    void hydrateToOffset(0, { historyMode: 'push' })
  }, searchDebounceMs)
}

function retryFailedRequest() {
  if (lastFailedRequest.value === 'load-more' && products.value.length > 0) {
    void loadMore()
    return
  }

  void hydrateToOffset(currentOffset.value)
}

function handleWindowScroll() {
  showScrollTopButton.value = window.scrollY > MOBILE_SCROLL_TOP_THRESHOLD
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handlePopState() {
  clearSearchApplyTimer()
  const { filters, offset } = parseURLState()
  assignFilterState(draftFilters, filters)
  assignFilterState(appliedFilters, filters)
  void hydrateToOffset(offset)
}

onMounted(() => {
  initializeTheme()
  const { filters, offset } = parseURLState()
  assignFilterState(draftFilters, filters)
  assignFilterState(appliedFilters, filters)
  void hydrateToOffset(offset)

  if (typeof window !== 'undefined') {
    handleWindowScroll()
    window.addEventListener('scroll', handleWindowScroll, { passive: true })
    window.addEventListener('popstate', handlePopState)
  }
})

onBeforeUnmount(() => {
  clearSearchApplyTimer()

  if (typeof window !== 'undefined') {
    window.removeEventListener('scroll', handleWindowScroll)
    window.removeEventListener('popstate', handlePopState)
  }

  if (activeController) {
    activeController.abort()
  }
})
</script>
