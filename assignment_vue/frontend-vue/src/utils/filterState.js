import { normalizeSortValue, sortValueLabel } from './sort'

export function createDefaultFilters() {
  return {
    search: '',
    categories: [],
    brands: [],
    colors: [],
    conditions: [],
    bestseller: 'all',
    onSale: 'all',
    inStock: 'all',
    sort: '',
    minPrice: '',
    maxPrice: '',
  }
}

export function normalizeBestsellerValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

export function normalizeOnSaleValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

export function normalizeInStockValue(rawValue) {
  return rawValue === 'true' ? 'true' : 'all'
}

export function normalizeTokenList(values) {
  return Array.from(
    new Set(
      values
        .map((value) => value.toLowerCase().trim())
        .filter(Boolean),
    ),
  ).sort()
}

export function normalizeFilterState(source, options = {}) {
  const strictSort = options.strictSort === true
  return {
    search: String(source.search ?? '').trim(),
    categories: normalizeTokenList(source.categories ?? []),
    brands: normalizeTokenList(source.brands ?? []),
    colors: normalizeTokenList(source.colors ?? []),
    conditions: normalizeTokenList(source.conditions ?? []),
    bestseller: normalizeBestsellerValue(source.bestseller),
    onSale: normalizeOnSaleValue(source.onSale),
    inStock: normalizeInStockValue(source.inStock),
    sort: normalizeSortValue(source.sort, { strict: strictSort }),
    minPrice: String(source.minPrice ?? '').trim(),
    maxPrice: String(source.maxPrice ?? '').trim(),
  }
}

export function assignFilterState(target, source) {
  target.search = source.search
  target.categories = [...source.categories]
  target.brands = [...source.brands]
  target.colors = [...source.colors]
  target.conditions = [...source.conditions]
  target.bestseller = source.bestseller
  target.onSale = source.onSale
  target.inStock = source.inStock
  target.sort = source.sort
  target.minPrice = source.minPrice
  target.maxPrice = source.maxPrice
}

export function areFilterStatesEqual(left, right) {
  return JSON.stringify(left) === JSON.stringify(right)
}

function parsePrice(rawValue) {
  const normalized = String(rawValue ?? '').trim()
  if (!normalized) {
    return null
  }

  const parsed = Number(normalized)
  if (!Number.isFinite(parsed)) {
    return Number.NaN
  }
  return parsed
}

export function buildFilterValidationMessage(source) {
  const min = parsePrice(source.minPrice)
  const max = parsePrice(source.maxPrice)

  if (Number.isNaN(min)) {
    return 'Minimum price must be a valid number.'
  }
  if (Number.isNaN(max)) {
    return 'Maximum price must be a valid number.'
  }
  if (min !== null && min < 0) {
    return 'Minimum price must be greater than or equal to 0.'
  }
  if (max !== null && max < 0) {
    return 'Maximum price must be greater than or equal to 0.'
  }
  if (min !== null && max !== null && min > max) {
    return 'Minimum price cannot be greater than maximum price.'
  }

  return ''
}

export function sanitizePriceInput(rawValue) {
  const normalized = String(rawValue ?? '').trim()
  if (!normalized) {
    return ''
  }

  const parsed = Number(normalized)
  if (!Number.isFinite(parsed) || parsed < 0) {
    return ''
  }

  return normalized
}

export function sanitizePriceRange(minPriceRaw, maxPriceRaw) {
  const minPrice = sanitizePriceInput(minPriceRaw)
  const maxPrice = sanitizePriceInput(maxPriceRaw)

  if (minPrice && maxPrice && Number(minPrice) > Number(maxPrice)) {
    return { minPrice: '', maxPrice: '' }
  }

  return { minPrice, maxPrice }
}

export function sortLabel(sortValue) {
  return sortValueLabel(sortValue, 'Default')
}

export function formatTokenLabel(token) {
  return token
    .split('-')
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' ')
}
