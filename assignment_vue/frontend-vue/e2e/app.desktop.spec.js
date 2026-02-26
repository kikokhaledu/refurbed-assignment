import { expect, test } from '@playwright/test'
import {
  applyFiltersFromPanel,
  cardByName,
  clearDraftFromPanel,
  gotoAndWaitForInitialData,
  productCards,
  readProductsCount,
  waitForCards,
  waitForProductsResponse,
} from './helpers'

const backendBaseURL = process.env.PW_BACKEND_BASE_URL ?? 'http://127.0.0.1:8080'

function expectNonDecreasing(values) {
  for (let index = 1; index < values.length; index += 1) {
    expect(values[index]).toBeGreaterThanOrEqual(values[index - 1])
  }
}

function expectNonIncreasing(values) {
  for (let index = 1; index < values.length; index += 1) {
    expect(values[index]).toBeLessThanOrEqual(values[index - 1])
  }
}

function bucketForPrice(price) {
  if (!Number.isFinite(price)) {
    return null
  }
  return Math.max(0, Math.floor(price))
}

async function readSelectedSortModes(page) {
  const rawValue = await page.getByTestId('sort-select').getAttribute('data-sort-value')
  if (!rawValue) {
    return []
  }
  return rawValue
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)
}

async function setSortModes(page, modes) {
  const desired = Array.from(new Set(modes))
  const trigger = page.getByTestId('sort-select')

  async function ensureDropdownOpen() {
    if ((await trigger.getAttribute('aria-expanded')) === 'true') {
      return
    }
    await trigger.click()
    await expect(page.getByTestId('sort-dropdown')).toBeVisible()
  }

  async function clickSortOption(mode) {
    await ensureDropdownOpen()
    await page.getByTestId(`sort-option-${mode}`).click()
  }

  let current = await readSelectedSortModes(page)

  for (const mode of current) {
    if (desired.includes(mode)) {
      continue
    }
    await clickSortOption(mode)
    current = await readSelectedSortModes(page)
  }

  for (const mode of desired) {
    if (current.includes(mode)) {
      continue
    }
    await clickSortOption(mode)
    current = await readSelectedSortModes(page)
  }

  await ensureDropdownOpen()
  await trigger.click()
  await expect(page.getByTestId('sort-dropdown')).toHaveCount(0)
}

test.describe('Desktop Product Discovery', () => {
  test('FE-E2E-001 FE-E2E-003 FE-E2E-054 initial shell and title', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await expect(page.getByRole('heading', { name: 'Product discovery' })).toBeVisible()
    await expect(page.getByTestId('search-input')).toBeVisible()
    await waitForCards(page, 1)
    await expect(page.getByTestId('filters-panel')).toBeVisible()
    await expect(page.getByTestId('mobile-filters-trigger')).not.toBeVisible()

    const count = await readProductsCount(page)
    expect(count).toBeGreaterThan(0)
    await expect(page).toHaveTitle(new RegExp(`^${count} products found \\| Refurbed$`))
  })

  test('FE-E2E-002 initial skeletons are shown while first request is pending', async ({ page }) => {
    let delayedInitialRequest = false
    let releaseInitialResponse = () => {}
    const initialResponseGate = new Promise((resolve) => {
      releaseInitialResponse = resolve
    })

    await page.route('**/products?*', async (route) => {
      const url = new URL(route.request().url())
      if (!delayedInitialRequest && url.searchParams.get('offset') === '0') {
        delayedInitialRequest = true
        const upstream = await route.fetch()
        const upstreamBody = await upstream.text()
        await initialResponseGate
        await route.fulfill({ response: upstream, body: upstreamBody })
        return
      }
      await route.continue()
    })

    await page.goto('/')
    await expect.poll(() => delayedInitialRequest).toBeTruthy()
    await expect(page.getByTestId('results-section')).toHaveAttribute('aria-busy', 'true')
    await expect(page.getByTestId('product-skeleton').first()).toBeVisible()
    releaseInitialResponse()
    await waitForCards(page, 1)
  })

  test('FE-E2E-004 FE-E2E-005 FE-E2E-006 FE-E2E-007 FE-E2E-008 search is immediate in input and debounced in network', async ({ page }) => {
    const requestUrls = []

    await page.route('**/products?*', async (route) => {
      requestUrls.push(route.request().url())
      await route.continue()
    })

    await gotoAndWaitForInitialData(page)
    const baselineRequests = requestUrls.length

    const searchInput = page.getByTestId('search-input')
    await searchInput.type('iph', { delay: 30 })
    await searchInput.fill('ipad')
    await expect(searchInput).toHaveValue('ipad')

    const debouncedSearchResponse = waitForProductsResponse(page, (url) => url.searchParams.get('search') === 'ipad')
    const requestDispatchedTooEarly = await waitForProductsResponse(
      page,
      (url) => url.searchParams.get('search') === 'ipad',
      250,
    ).then(
      () => true,
      () => false,
    )

    expect(requestDispatchedTooEarly).toBe(false)

    await debouncedSearchResponse

    expect(requestUrls.length).toBeGreaterThanOrEqual(baselineRequests + 1)
    await expect.poll(() => new URL(page.url()).searchParams.get('search')).toBe('ipad')
    await expect.poll(() => new URL(page.url()).searchParams.get('offset')).toBeNull()

    await expect(productCards(page)).toHaveCount(1)
    await expect(page.getByRole('heading', { name: 'iPad Air' })).toBeVisible()

    const clearResponse = waitForProductsResponse(page, (url) => !url.searchParams.has('search'))
    await searchInput.fill('')
    await clearResponse

    await expect(productCards(page)).toHaveCount(6)
  })

  test('FE-E2E-009 FE-E2E-010 FE-E2E-011 FE-E2E-012 draft filters require apply and reset applies immediately', async ({ page }) => {
    const requestUrls = []
    await page.route('**/products?*', async (route) => {
      requestUrls.push(route.request().url())
      await route.continue()
    })

    await gotoAndWaitForInitialData(page)
    const requestsBeforeDraftChange = requestUrls.length

    await page.getByTestId('filter-category-smartphones').check()
    await expect(page.getByTestId('filter-unapplied-warning-top')).toBeVisible()
    expect(new URL(page.url()).searchParams.getAll('category')).toHaveLength(0)

    const autoAppliedBeforeApply = await waitForProductsResponse(
      page,
      (url) => url.searchParams.getAll('category').includes('smartphones'),
      700,
    ).then(
      () => true,
      () => false,
    )
    expect(autoAppliedBeforeApply).toBe(false)
    expect(requestUrls.length).toBe(requestsBeforeDraftChange)

    const categoryResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('category').includes('smartphones')
    })
    await applyFiltersFromPanel(page)
    const categoryPayload = await (await categoryResponse).json()

    expect(categoryPayload.items.length).toBeGreaterThan(0)
    expect(categoryPayload.items.every((item) => item.category === 'smartphones')).toBeTruthy()
    expect(new URL(page.url()).searchParams.getAll('category')).toEqual(['smartphones'])
    await expect(page.getByTestId('active-chip-category:smartphones')).toBeVisible()

    await page.getByTestId('filter-condition-new').check()
    const resetResponse = waitForProductsResponse(page, (url) => {
      return !url.searchParams.has('category') && !url.searchParams.has('condition')
    })
    await clearDraftFromPanel(page)
    await resetResponse

    await expect(page.getByTestId('filter-category-smartphones')).not.toBeChecked()
    await expect(page.getByTestId('filter-condition-new')).not.toBeChecked()
    await expect(page.getByTestId('active-chip-category:smartphones')).toHaveCount(0)
    expect(new URL(page.url()).searchParams.getAll('category')).toHaveLength(0)
    expect(new URL(page.url()).searchParams.getAll('condition')).toHaveLength(0)
  })

  test('FE-E2E-013 FE-E2E-014 chips can remove one filter and clear all', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('filter-bestseller-true').check()
    await page.getByTestId('filter-onsale-true').check()
    const appliedResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.get('bestseller') === 'true' && url.searchParams.get('onSale') === 'true'
    })
    await applyFiltersFromPanel(page)
    await appliedResponse

    await expect(page.getByTestId('active-chip-bestseller')).toBeVisible()
    await expect(page.getByTestId('active-chip-onSale')).toBeVisible()

    const removeOneResponse = waitForProductsResponse(page, (url) => !url.searchParams.has('onSale'))
    await page.getByTestId('active-chip-onSale').click()
    await removeOneResponse
    await expect(page.getByTestId('active-chip-onSale')).toHaveCount(0)
    await expect(page.getByTestId('active-chip-bestseller')).toBeVisible()

    const clearAllResponse = waitForProductsResponse(page, (url) => url.searchParams.toString() === 'limit=6&offset=0')
    await page.getByTestId('active-chip-clear-all').click()
    await clearAllResponse

    await expect(page.getByTestId('active-filter-chips')).toHaveCount(0)
    expect(new URL(page.url()).searchParams.toString()).toBe('')
  })

  test('FE-E2E-016 FE-E2E-017 FE-E2E-018 FE-E2E-019 FE-E2E-020 FE-E2E-021 FE-E2E-030 applied filters are reflected in request and URL', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('filter-category-smartphones').check()
    await page.getByTestId('filter-category-accessories').check()
    await page.getByTestId('filter-condition-refurbished').check()
    await page.getByTestId('filter-color-blue').click()
    await page.getByTestId('filter-bestseller-true').check()
    await page.getByTestId('filter-onsale-true').check()
    await page.getByTestId('filter-instock-true').check()

    const responsePromise = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.get('bestseller') === 'true' &&
        url.searchParams.get('onSale') === 'true' &&
        url.searchParams.get('inStock') === 'true' &&
        url.searchParams.getAll('category').length > 0 &&
        url.searchParams.getAll('condition').includes('refurbished') &&
        url.searchParams.getAll('color').includes('blue')
      )
    })
    await applyFiltersFromPanel(page)
    const payload = await (await responsePromise).json()

    expect(payload.items.length).toBeGreaterThan(0)
    for (const item of payload.items) {
      expect(['smartphones', 'accessories']).toContain(item.category)
      expect(item.condition).toBe('refurbished')
      expect(item.colors).toContain('blue')
      expect(item.bestseller).toBe(true)
      expect(item.discount_percent).toBeGreaterThan(0)
      expect(item.stock).toBeGreaterThan(0)
    }

    const url = new URL(page.url())
    expect(url.searchParams.getAll('category').sort()).toEqual(['accessories', 'smartphones'])
    expect(url.searchParams.getAll('condition')).toEqual(['refurbished'])
    expect(url.searchParams.getAll('color')).toEqual(['blue'])
    expect(url.searchParams.get('bestseller')).toBe('true')
    expect(url.searchParams.get('onSale')).toBe('true')
    expect(url.searchParams.get('inStock')).toBe('true')
  })

  test('FE-E2E-022 FE-E2E-023 FE-E2E-024 FE-E2E-025 price slider updates badges, clamps, and uses API bounds', async ({ page }) => {
    const firstResponse = await gotoAndWaitForInitialData(page)
    let initialPayload = null
    if (firstResponse) {
      initialPayload = await firstResponse.json()
    } else {
      const fallbackResponse = await page.request.get(`${backendBaseURL}/products?limit=6&offset=0`)
      initialPayload = await fallbackResponse.json()
    }

    const floor = Math.max(0, Math.floor(initialPayload.price_min))
    const ceiling = Math.max(floor, Math.ceil(initialPayload.price_max))

    await expect(page.getByTestId('price-min-badge')).toContainText(`Min EUR ${floor}`)
    await expect(page.getByTestId('price-max-badge')).toContainText(`Max EUR ${ceiling}`)
    await expect(page.getByTestId('price-slider-min')).toHaveValue(String(floor))
    await expect(page.getByTestId('price-slider-max')).toHaveValue(String(ceiling))
    await expect(page.getByText(`EUR ${floor}`).first()).toBeVisible()
    await expect(page.getByText(`EUR ${ceiling}`).first()).toBeVisible()

    const clampedCrossMin = ceiling > floor ? ceiling - 1 : ceiling

    await page.getByTestId('price-slider-min').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, ceiling + 25)
    await expect(page.getByTestId('price-min-badge')).toContainText(`Min EUR ${clampedCrossMin}`)

    await page.getByTestId('price-slider-max').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, floor - 25)
    await expect(page.getByTestId('price-max-badge')).toContainText(`Max EUR ${ceiling}`)
    await expect(page.getByTestId('price-min-badge')).toContainText(`Min EUR ${clampedCrossMin}`)

    const responsePromise = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.get('minPrice') === String(clampedCrossMin) &&
        url.searchParams.get('maxPrice') === String(ceiling)
      )
    })
    await applyFiltersFromPanel(page)
    await responsePromise
  })

  test('FE-E2E-073 crossing min over max keeps a valid range and still returns data', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const fullDatasetResponse = await page.request.get(`${backendBaseURL}/products?limit=100&offset=0`)
    const fullDatasetPayload = await fullDatasetResponse.json()
    const floor = Math.max(0, Math.floor(Number(fullDatasetPayload.price_min)))
    const candidate = fullDatasetPayload.items.find((item) => {
      const bucket = bucketForPrice(Number(item.price))
      return bucket !== null && bucket > floor
    })
    expect(candidate).toBeTruthy()

    const bucket = bucketForPrice(Number(candidate?.price))
    expect(Number.isFinite(bucket)).toBeTruthy()
    const expectedMin = Math.max(floor, bucket - 1)

    await page.getByTestId('price-slider-max').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, bucket)
    await page.getByTestId('price-slider-min').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, expectedMin)
    await page.getByTestId('price-slider-min').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, bucket + 10)

    await expect(page.getByTestId('price-min-badge')).toContainText(`Min EUR ${expectedMin}`)
    await expect(page.getByTestId('price-max-badge')).toContainText(`Max EUR ${bucket}`)

    const responsePromise = waitForProductsResponse(page, (url) => {
      return url.searchParams.get('minPrice') === String(expectedMin) && url.searchParams.get('maxPrice') === String(bucket)
    })
    await applyFiltersFromPanel(page)
    const payload = await (await responsePromise).json()

    expect(payload.items.length).toBeGreaterThan(0)
    for (const item of payload.items) {
      const price = Number(item.price)
      expect(price).toBeGreaterThanOrEqual(expectedMin)
      expect(price).toBeLessThan(bucket + 1)
    }
  })

  test('FE-E2E-074 crossing at upper bound keeps a minimum 1 EUR gap', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const initialResponse = await page.request.get(`${backendBaseURL}/products?limit=6&offset=0`)
    const initialPayload = await initialResponse.json()
    const floor = Math.max(0, Math.floor(Number(initialPayload.price_min)))
    const ceiling = Math.max(floor, Math.ceil(Number(initialPayload.price_max)))
    const expectedMin = ceiling > floor ? ceiling - 1 : ceiling

    await page.getByTestId('price-slider-max').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, ceiling)
    await page.getByTestId('price-slider-min').evaluate((element, value) => {
      element.value = String(value)
      element.dispatchEvent(new Event('input', { bubbles: true }))
    }, ceiling + 20)

    await expect(page.getByTestId('price-min-badge')).toContainText(`Min EUR ${expectedMin}`)
    await expect(page.getByTestId('price-max-badge')).toContainText(`Max EUR ${ceiling}`)

    const responsePromise = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.get('minPrice') === String(expectedMin) &&
        url.searchParams.get('maxPrice') === String(ceiling)
      )
    })
    await applyFiltersFromPanel(page)
    await responsePromise
  })

  test('FE-E2E-027 FE-E2E-028 FE-E2E-029 sorting options can be applied and removed', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await expect.poll(() => readSelectedSortModes(page)).toEqual([])
    expect(new URL(page.url()).searchParams.has('sort')).toBeFalsy()

    const popularityResponse = waitForProductsResponse(page, (url) => url.searchParams.get('sort') === 'popularity')
    await setSortModes(page, ['popularity'])
    const popularityPayload = await (await popularityResponse).json()

    await expect(page.getByTestId('active-chip-sort')).toBeVisible()
    const ranks = popularityPayload.items.map((item) => item.popularity_rank ?? Number.MAX_SAFE_INTEGER)
    const sortedRanks = [...ranks].sort((a, b) => a - b)
    expect(ranks).toEqual(sortedRanks)
    expect(new URL(page.url()).searchParams.get('sort')).toBe('popularity')

    const priceAscResponse = waitForProductsResponse(page, (url) => url.searchParams.get('sort') === 'price_asc')
    await setSortModes(page, ['price_asc'])
    const priceAscPayload = await (await priceAscResponse).json()
    expect(new URL(page.url()).searchParams.get('sort')).toBe('price_asc')
    expectNonDecreasing(priceAscPayload.items.map((item) => item.price))

    const priceDescResponse = waitForProductsResponse(page, (url) => url.searchParams.get('sort') === 'price_desc')
    await setSortModes(page, ['price_desc'])
    const priceDescPayload = await (await priceDescResponse).json()
    expect(new URL(page.url()).searchParams.get('sort')).toBe('price_desc')
    expectNonIncreasing(priceDescPayload.items.map((item) => item.price))

    const comboResponse = waitForProductsResponse(page, (url) => url.searchParams.get('sort') === 'popularity,price_asc')
    await setSortModes(page, ['popularity', 'price_asc'])
    const comboPayload = await (await comboResponse).json()
    expect(new URL(page.url()).searchParams.get('sort')).toBe('popularity,price_asc')
    const comboRanks = comboPayload.items.map((item) => item.popularity_rank ?? Number.MAX_SAFE_INTEGER)
    expect(comboRanks).toEqual([...comboRanks].sort((a, b) => a - b))

    const removeSortResponse = waitForProductsResponse(page, (url) => !url.searchParams.has('sort'))
    await page.getByTestId('active-chip-sort').click()
    await removeSortResponse

    await expect.poll(() => readSelectedSortModes(page)).toEqual([])
    expect(new URL(page.url()).searchParams.has('sort')).toBeFalsy()
  })

  test('FE-E2E-071 multi sorting works with active filters and preserves filter query params', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('filter-condition-refurbished').check()
    await page.getByTestId('filter-onsale-true').check()

    const filteredResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('condition').includes('refurbished') && url.searchParams.get('onSale') === 'true'
    })
    await applyFiltersFromPanel(page)
    const filteredPayload = await (await filteredResponse).json()
    expect(filteredPayload.items.length).toBeGreaterThanOrEqual(2)

    const comboSortResponse = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.get('sort') === 'popularity,price_asc' &&
        url.searchParams.getAll('condition').includes('refurbished') &&
        url.searchParams.get('onSale') === 'true'
      )
    })
    await setSortModes(page, ['popularity', 'price_asc'])
    const comboSortPayload = await (await comboSortResponse).json()
    const ranks = comboSortPayload.items.map((item) => item.popularity_rank ?? Number.MAX_SAFE_INTEGER)
    expect(ranks).toEqual([...ranks].sort((a, b) => a - b))
    expect(new URL(page.url()).searchParams.get('sort')).toBe('popularity,price_asc')
    expect(new URL(page.url()).searchParams.get('onSale')).toBe('true')
    expect(new URL(page.url()).searchParams.getAll('condition')).toEqual(['refurbished'])

    const priceDescResponse = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.get('sort') === 'price_desc' &&
        url.searchParams.getAll('condition').includes('refurbished') &&
        url.searchParams.get('onSale') === 'true'
      )
    })
    await setSortModes(page, ['price_desc'])
    const priceDescPayload = await (await priceDescResponse).json()
    expectNonIncreasing(priceDescPayload.items.map((item) => item.price))
    expect(new URL(page.url()).searchParams.get('sort')).toBe('price_desc')
    expect(new URL(page.url()).searchParams.get('onSale')).toBe('true')
    expect(new URL(page.url()).searchParams.getAll('condition')).toEqual(['refurbished'])
  })

  test('FE-E2E-072 conflicting sort params in URL normalize safely to a single effective sort', async ({ page }) => {
    const firstResponse = await gotoAndWaitForInitialData(
      page,
      '/?sort=price_asc&sort=price_desc&condition=refurbished&onSale=true',
    )

    await expect.poll(() => readSelectedSortModes(page)).toEqual(['price_asc'])
    await expect.poll(() => new URL(page.url()).searchParams.getAll('sort')).toEqual(['price_asc'])

    let payload = firstResponse ? await firstResponse.json() : null
    if (!payload) {
      const fallbackResponse = await page.request.get(
        `${backendBaseURL}/products?limit=6&offset=0&sort=price_asc&condition=refurbished&onSale=true`,
      )
      payload = await fallbackResponse.json()
    }

    if (Array.isArray(payload.items) && payload.items.length > 1) {
      expectNonDecreasing(payload.items.map((item) => item.price))
    }
  })

  test('FE-E2E-031 FE-E2E-033 URL hydration works and invalid values degrade safely', async ({ page }) => {
    await gotoAndWaitForInitialData(
      page,
      '/?search=ipad&category=tablets&condition=refurbished&color=blue&bestseller=true&onSale=true&inStock=true&sort=popularity&minPrice=200&maxPrice=800',
    )

    await expect(page.getByTestId('search-input')).toHaveValue('ipad')
    await expect(page.getByTestId('filter-category-tablets')).toBeChecked()
    await expect(page.getByTestId('filter-condition-refurbished')).toBeChecked()
    await expect(page.getByTestId('filter-color-blue')).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByTestId('filter-bestseller-true')).toBeChecked()
    await expect(page.getByTestId('filter-onsale-true')).toBeChecked()
    await expect(page.getByTestId('filter-instock-true')).toBeChecked()
    await expect.poll(() => readSelectedSortModes(page)).toEqual(['popularity'])

    await gotoAndWaitForInitialData(page, '/?bestseller=invalid&onSale=wrong&inStock=nope&sort=unknown&offset=abc&minPrice=oops&maxPrice=-20')
    await expect(page.getByTestId('fatal-error-state')).toHaveCount(0)
    await waitForCards(page, 1)
    await expect(page.getByTestId('search-input')).toHaveValue('')
    await expect.poll(() => readSelectedSortModes(page)).toEqual([])
    await expect.poll(() => new URL(page.url()).searchParams.get('minPrice')).toBeNull()
    await expect.poll(() => new URL(page.url()).searchParams.get('maxPrice')).toBeNull()
  })

  test('FE-E2E-032 browser back and forward rehydrate filter state', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('filter-category-smartphones').check()
    await expect(page.getByTestId('filter-category-smartphones')).toBeChecked()
    const firstApplyResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('category').includes('smartphones') && !url.searchParams.has('condition')
    })
    await applyFiltersFromPanel(page)
    await firstApplyResponse
    await expect.poll(() => new URL(page.url()).searchParams.getAll('category')).toEqual(['smartphones'])
    await expect.poll(() => new URL(page.url()).searchParams.getAll('condition')).toEqual([])

    const firstURL = new URL(page.url())
    expect(firstURL.searchParams.getAll('category')).toEqual(['smartphones'])
    expect(firstURL.searchParams.getAll('condition')).toEqual([])

    await page.getByTestId('filter-condition-refurbished').check()
    await expect(page.getByTestId('filter-condition-refurbished')).toBeChecked()
    const secondApplyResponse = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.getAll('category').includes('smartphones') &&
        url.searchParams.getAll('condition').includes('refurbished')
      )
    })
    await applyFiltersFromPanel(page)
    await secondApplyResponse

    await expect.poll(() => new URL(page.url()).searchParams.getAll('condition')).toEqual(['refurbished'])
    const secondURL = new URL(page.url())
    expect(secondURL.searchParams.getAll('category')).toEqual(['smartphones'])

    const backResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('category').includes('smartphones') && !url.searchParams.has('condition')
    })
    await page.goBack()
    await backResponse

    await expect(page.getByTestId('filter-category-smartphones')).toBeChecked()
    await expect(page.getByTestId('filter-condition-refurbished')).not.toBeChecked()
    expect(new URL(page.url()).search).toBe(firstURL.search)

    const forwardResponse = waitForProductsResponse(page, (url) => {
      return (
        url.searchParams.getAll('category').includes('smartphones') &&
        url.searchParams.getAll('condition').includes('refurbished')
      )
    })
    await page.goForward()
    await forwardResponse

    await expect(page.getByTestId('filter-category-smartphones')).toBeChecked()
    await expect(page.getByTestId('filter-condition-refurbished')).toBeChecked()
    expect(new URL(page.url()).search).toBe(secondURL.search)
  })

  test('FE-E2E-034 FE-E2E-035 FE-E2E-036 FE-E2E-037 load more appends results once and updates offset', async ({ page }) => {
    let offsetSixRequests = 0
    let releaseOffsetSixResponse = () => {}
    const offsetSixResponseGate = new Promise((resolve) => {
      releaseOffsetSixResponse = resolve
    })

    await page.route('**/products?*', async (route) => {
      const url = new URL(route.request().url())
      if (url.searchParams.get('offset') === '6') {
        offsetSixRequests += 1
        if (offsetSixRequests === 1) {
          await offsetSixResponseGate
        }
      }
      await route.continue()
    })

    await gotoAndWaitForInitialData(page)
    await expect(productCards(page)).toHaveCount(6)
    await expect(page.getByTestId('load-more-button')).toBeVisible()

    const loadMoreResponse = waitForProductsResponse(page, (url) => url.searchParams.get('offset') === '6')
    await page.evaluate(() => {
      const button = document.querySelector('[data-testid="load-more-button"]')
      if (!(button instanceof HTMLButtonElement)) {
        return
      }

      button.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      button.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      button.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    })
    await expect.poll(() => offsetSixRequests).toBe(1)
    releaseOffsetSixResponse()
    await loadMoreResponse

    expect(offsetSixRequests).toBe(1)
    await expect(productCards(page)).toHaveCount(8)
    expect(new URL(page.url()).searchParams.get('offset')).toBe('6')
    await expect(page.getByTestId('load-more-button')).toHaveCount(0)
  })

  test('FE-E2E-038 FE-E2E-039 FE-E2E-040 FE-E2E-043 FE-E2E-044 product cards show badges, swatches, stock by color, and image switching', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const iphoneCard = cardByName(page, 'iPhone 12')
    await expect(iphoneCard).toBeVisible()
    await expect(iphoneCard.getByText('Bestseller')).toBeVisible()
    await expect(iphoneCard.getByText('-25%')).toBeVisible()
    await expect(iphoneCard.locator('img')).toBeVisible()
    await expect(iphoneCard.getByTestId('product-stock-p1')).toContainText('12 in stock (Blue)')

    await expect(iphoneCard.getByTestId('product-swatch-p1-red')).toHaveCount(0)
    await expect(iphoneCard.getByTestId('product-swatch-p1-blue')).toBeVisible()
    await expect(iphoneCard.getByTestId('product-swatch-p1-green')).toBeVisible()

    const image = iphoneCard.getByTestId('product-image-p1')
    const beforeSrc = await image.getAttribute('src')
    await iphoneCard.getByTestId('product-swatch-p1-green').click()
    await expect(iphoneCard.getByTestId('product-stock-p1')).toContainText('22 in stock (Green)')
    await expect.poll(async () => image.getAttribute('src')).not.toBe(beforeSrc)

    await expect(iphoneCard.locator('p.line-through')).toHaveCount(1)
    const watchCard = cardByName(page, 'Apple Watch Series 7')
    await expect(watchCard.locator('p.line-through')).toHaveCount(0)
  })

  test('FE-E2E-041 missing color image mapping falls back to base image', async ({ page }) => {
    let patched = false
    await page.route('**/products?*', async (route) => {
      const response = await route.fetch()
      const url = new URL(route.request().url())
      if (patched || url.searchParams.get('offset') !== '0') {
        await route.fulfill({ response })
        return
      }

      const payload = await response.json()
      if (payload.items?.[0]) {
        payload.items[0].image_url = '/product-placeholder.svg'
        payload.items[0].image_urls_by_color = {
          blue: 'https://example.com/blue-only-image.jpg',
        }
        payload.items[0].colors = ['blue', 'green']
        payload.items[0].stock_by_color = { blue: 3, green: 2 }
      }

      patched = true
      await route.fulfill({ response, json: payload })
    })

    await gotoAndWaitForInitialData(page)

    const iphoneCard = cardByName(page, 'iPhone 12')
    await iphoneCard.getByTestId('product-swatch-p1-green').click()
    const image = iphoneCard.getByTestId('product-image-p1')
    await expect.poll(async () => image.getAttribute('src')).toContain('/product-placeholder.svg')
  })

  test('FE-E2E-042 broken image URL falls back to local placeholder', async ({ page }) => {
    let patched = false
    await page.route('**/products?*', async (route) => {
      const response = await route.fetch()
      const url = new URL(route.request().url())
      if (patched || url.searchParams.get('offset') !== '0') {
        await route.fulfill({ response })
        return
      }

      const payload = await response.json()
      if (payload.items?.[0]) {
        payload.items[0].image_url = '/does-not-exist.png'
        payload.items[0].image_urls_by_color = {}
        payload.items[0].colors = []
      }

      patched = true
      await route.fulfill({ response, json: payload })
    })

    await gotoAndWaitForInitialData(page)
    const firstImage = page.locator('[data-testid^="product-image-"]').first()
    await expect.poll(async () => firstImage.getAttribute('src')).toContain('/product-placeholder.svg')
  })

  test('FE-E2E-046 FE-E2E-047 color filter options come from API and selected value persists', async ({ page }) => {
    const response = await gotoAndWaitForInitialData(page)
    let payload = null
    if (response) {
      payload = await response.json()
    } else {
      const fallbackResponse = await page.request.get(`${backendBaseURL}/products?limit=6&offset=0`)
      payload = await fallbackResponse.json()
    }
    const availableColors = [...payload.available_colors].sort()

    const colorIds = await page.locator('[data-testid^="filter-color-"]').evaluateAll((elements) => {
      return elements
        .map((element) => element.getAttribute('data-testid') ?? '')
        .filter(Boolean)
        .map((id) => id.replace('filter-color-', ''))
        .sort()
    })

    expect(colorIds).toEqual(availableColors)

    await page.getByTestId('filter-color-orange').click()
    await page.getByTestId('filter-category-smartphones').check()
    const responsePromise = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('color').includes('orange') && url.searchParams.getAll('category').includes('smartphones')
    })
    await applyFiltersFromPanel(page)
    await responsePromise

    await expect(page.getByTestId('filter-color-orange')).toBeVisible()
    await expect(page.getByTestId('filter-color-orange')).toHaveAttribute('aria-pressed', 'true')
  })

  test('FE-E2E-076 FE-E2E-077 brand filter options come from API and apply correctly', async ({ page }) => {
    const response = await gotoAndWaitForInitialData(page)
    let payload = null
    if (response) {
      payload = await response.json()
    } else {
      const fallbackResponse = await page.request.get(`${backendBaseURL}/products?limit=6&offset=0`)
      payload = await fallbackResponse.json()
    }

    const availableBrands = Array.isArray(payload.available_brands) ? [...payload.available_brands].sort() : []
    expect(availableBrands.length).toBeGreaterThan(0)

    const brandIds = await page.locator('[data-testid^="filter-brand-"]').evaluateAll((elements) => {
      return elements
        .map((element) => element.getAttribute('data-testid') ?? '')
        .filter(Boolean)
        .map((id) => id.replace('filter-brand-', ''))
        .sort()
    })
    expect(brandIds).toEqual(availableBrands)

    const selectedBrand = availableBrands[0]
    await page.getByTestId(`filter-brand-${selectedBrand}`).check()
    const responsePromise = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('brand').includes(selectedBrand)
    })
    await applyFiltersFromPanel(page)
    const filteredPayload = await (await responsePromise).json()

    expect(filteredPayload.items.length).toBeGreaterThan(0)
    expect(filteredPayload.items.every((item) => item.brand === selectedBrand)).toBeTruthy()
    await expect(page.getByTestId(`active-chip-brand:${selectedBrand}`)).toBeVisible()
  })

  test('FE-E2E-048 FE-E2E-049 FE-E2E-050 error states and retry flows work for initial load and load more', async ({ page }) => {
    let initialFailed = false
    await page.route('**/products?*', async (route) => {
      const url = new URL(route.request().url())
      if (!initialFailed && url.searchParams.get('offset') === '0') {
        initialFailed = true
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'intentional initial error' }),
        })
        return
      }
      await route.continue()
    })

    await page.goto('/')
    await expect(page.getByTestId('fatal-error-state')).toBeVisible()

    const recoverResponse = waitForProductsResponse(page, (url) => url.searchParams.get('offset') === '0')
    await page.getByTestId('fatal-retry-button').click()
    await recoverResponse
    await waitForCards(page, 1)

    let loadMoreFailures = 0
    await page.route('**/products?*', async (route) => {
      const url = new URL(route.request().url())
      if (url.searchParams.get('offset') === '6' && loadMoreFailures === 0) {
        loadMoreFailures += 1
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'intentional load-more error' }),
        })
        return
      }
      await route.continue()
    })

    await expect(page.getByTestId('load-more-button')).toBeVisible()
    await page.getByTestId('load-more-button').click()
    await expect(page.getByTestId('inline-error-state')).toBeVisible()
    await expect(productCards(page)).toHaveCount(6)

    const retryResponse = waitForProductsResponse(page, (url) => url.searchParams.get('offset') === '6')
    await page.getByTestId('inline-retry-button').click()
    await retryResponse
    await expect(productCards(page)).toHaveCount(8)
  })

  test('FE-E2E-015 FE-E2E-051 empty state is shown and clear-all restores full results', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('filter-category-desktops').check()
    await page.getByTestId('filter-condition-new').check()
    const emptyResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('category').includes('desktops') && url.searchParams.getAll('condition').includes('new')
    })
    await applyFiltersFromPanel(page)
    await emptyResponse

    await expect(page.getByTestId('empty-state')).toBeVisible()
    await expect(page.getByText('No products found')).toBeVisible()
    await expect(page.getByText('We could not find products matching your current filters.')).toBeVisible()

    const clearResponse = waitForProductsResponse(page, (url) => !url.searchParams.has('category') && !url.searchParams.has('condition'))
    await page.getByTestId('empty-clear-filters-button').click()
    await clearResponse
    await waitForCards(page, 1)
  })

  test('FE-E2E-052 FE-E2E-053 FE-E2E-045 dark mode toggles, persists, and selected swatch remains visually active', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const toggle = page.getByTestId('theme-toggle')
    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-checked', 'true')
    await expect.poll(() => page.evaluate(() => document.documentElement.classList.contains('dark'))).toBeTruthy()
    await expect.poll(() => page.evaluate(() => window.localStorage.getItem('refurbed-theme'))).toBe('dark')

    const swatch = page.getByTestId('product-swatch-p1-blue')
    await expect(swatch).toHaveClass(/ring-2/)

    await page.reload()
    await waitForCards(page, 1)
    await expect(page.getByTestId('theme-toggle')).toHaveAttribute('aria-checked', 'true')
    await expect.poll(() => page.evaluate(() => window.localStorage.getItem('refurbed-theme'))).toBe('dark')
    await expect(page.getByTestId('product-swatch-p1-blue')).toHaveClass(/ring-2/)
  })

  test('FE-E2E-064 FE-E2E-065 FE-E2E-066 FE-E2E-067 FE-E2E-068 FE-E2E-069 FE-E2E-070 a11y and sticky apply UX stay healthy under keyboard flow', async ({ page }) => {
    const uncaughtErrors = []
    page.on('pageerror', (error) => {
      uncaughtErrors.push(error.message)
    })

    await gotoAndWaitForInitialData(page)

    await expect(page.locator('main')).toBeVisible()
    await expect(page.getByTestId('results-section')).toHaveAttribute('aria-live', 'polite')
    await expect(page.getByRole('searchbox', { name: 'Search products' })).toBeVisible()

    const iphoneImage = page.getByTestId('product-image-p1')
    await expect(iphoneImage).toHaveAttribute('alt', /iPhone 12 in blue/i)
    await page.getByTestId('product-swatch-p1-green').click()
    await expect(iphoneImage).toHaveAttribute('alt', /iPhone 12 in green/i)

    const categoryCheckbox = page.getByTestId('filter-category-smartphones')
    await categoryCheckbox.focus()
    await page.keyboard.press('Space')
    await expect(categoryCheckbox).toBeChecked()

    await page.mouse.wheel(0, 2000)
    await expect(page.getByTestId('filters-bottom-actions')).toBeVisible()
    await expect(page.getByTestId('filter-unapplied-warning-bottom')).toBeVisible()

    const responsePromise = waitForProductsResponse(page, (url) => url.searchParams.getAll('category').includes('smartphones'))
    await page.getByTestId('filters-apply-bottom').focus()
    await page.keyboard.press('Enter')
    await responsePromise

    await expect(uncaughtErrors).toEqual([])
  })
})
