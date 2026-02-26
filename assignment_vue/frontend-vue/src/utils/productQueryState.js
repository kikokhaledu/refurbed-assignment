import { PAGE_SIZE } from '../constants/app'
import {
  createDefaultFilters,
  normalizeBestsellerValue,
  normalizeFilterState,
  normalizeInStockValue,
  normalizeOnSaleValue,
  normalizeTokenList,
  sanitizePriceRange,
} from './filterState'

function parseParamList(params, key) {
  return normalizeTokenList(
    params
      .getAll(key)
      .flatMap((value) => value.split(','))
      .map((value) => value.toLowerCase().trim())
      .filter(Boolean),
  )
}

export function parseOffset(rawValue, pageSize = PAGE_SIZE) {
  const parsed = Number.parseInt(rawValue ?? '', 10)
  if (Number.isNaN(parsed) || parsed < 0) {
    return 0
  }
  return Math.floor(parsed / pageSize) * pageSize
}

export function parseURLState(search, pageSize = PAGE_SIZE) {
  const params = new URLSearchParams(search)
  const nextFilters = createDefaultFilters()

  nextFilters.search = params.get('search')?.trim() ?? ''
  nextFilters.categories = parseParamList(params, 'category')
  nextFilters.brands = parseParamList(params, 'brand')
  nextFilters.colors = parseParamList(params, 'color')
  nextFilters.conditions = parseParamList(params, 'condition')
  nextFilters.bestseller = normalizeBestsellerValue(params.get('bestseller'))
  nextFilters.onSale = normalizeOnSaleValue(params.get('onSale'))
  nextFilters.inStock = normalizeInStockValue(params.get('inStock'))
  nextFilters.sort = params.getAll('sort')

  const { minPrice, maxPrice } = sanitizePriceRange(
    params.get('minPrice'),
    params.get('maxPrice'),
  )
  nextFilters.minPrice = minPrice
  nextFilters.maxPrice = maxPrice

  return {
    filters: normalizeFilterState(nextFilters, { strictSort: true }),
    offset: parseOffset(params.get('offset'), pageSize),
  }
}

export function appendAppliedFiltersToParams(params, filters) {
  if (filters.search) {
    params.set('search', filters.search)
  }

  for (const category of filters.categories) {
    params.append('category', category)
  }

  for (const brand of filters.brands) {
    params.append('brand', brand)
  }

  for (const color of filters.colors) {
    params.append('color', color)
  }

  for (const condition of filters.conditions) {
    params.append('condition', condition)
  }

  if (filters.bestseller !== 'all') {
    params.set('bestseller', filters.bestseller)
  }

  if (filters.onSale !== 'all') {
    params.set('onSale', filters.onSale)
  }

  if (filters.inStock !== 'all') {
    params.set('inStock', filters.inStock)
  }

  if (filters.sort) {
    params.set('sort', filters.sort)
  }

  if (filters.minPrice) {
    params.set('minPrice', filters.minPrice)
  }

  if (filters.maxPrice) {
    params.set('maxPrice', filters.maxPrice)
  }

  return params
}

export function buildURLSearchParams(filters, options = {}) {
  const {
    includeLimit = false,
    includeOffset = false,
    pageSize = PAGE_SIZE,
    offset = 0,
  } = options

  const params = new URLSearchParams()

  if (includeLimit) {
    params.set('limit', String(pageSize))
  }
  if (includeOffset) {
    params.set('offset', String(offset))
  }

  return appendAppliedFiltersToParams(params, filters)
}
