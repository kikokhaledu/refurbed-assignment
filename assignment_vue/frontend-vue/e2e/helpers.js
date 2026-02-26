import { expect } from '@playwright/test'

export function productsRequestMatcher(url) {
  return (
    url.pathname === '/products' ||
    url.pathname.endsWith('/products')
  )
}

export function waitForProductsResponse(page, predicate = () => true, timeout = 15_000) {
  return page.waitForResponse((response) => {
    if (response.request().method() !== 'GET') {
      return false
    }

    let url
    try {
      url = new URL(response.url())
    } catch {
      return false
    }

    if (!productsRequestMatcher(url)) {
      return false
    }

    return predicate(url, response)
  }, { timeout })
}

export async function gotoAndWaitForInitialData(page, path = '/') {
  const firstResponse = waitForProductsResponse(page, (url) => {
    return url.searchParams.get('offset') === '0'
  }).catch(() => null)

  await page.goto(path)
  await expect(page.getByTestId('results-section')).toBeVisible()
  await expect(page.getByTestId('results-section')).toHaveAttribute('aria-busy', 'false')
  const response = await firstResponse
  return response
}

export function productCards(page) {
  return page.locator('[data-testid^="product-card-"]')
}

export async function waitForCards(page, minimum = 1) {
  await expect.poll(async () => {
    return productCards(page).count()
  }).toBeGreaterThanOrEqual(minimum)
}

export function cardByName(page, name) {
  return page
    .locator('[data-testid^="product-card-"]')
    .filter({ has: page.getByRole('heading', { name }) })
    .first()
}

export async function readProductsCount(page) {
  const text = await page.getByTestId('products-count').innerText()
  const match = text.match(/(\d+)\s+products found/i)
  if (!match) {
    throw new Error(`could not parse products count from: ${text}`)
  }
  return Number.parseInt(match[1], 10)
}

export async function applyFiltersFromPanel(page) {
  const applyTop = page.getByTestId('filters-apply-top')
  const applyBottom = page.getByTestId('filters-apply-bottom')

  if (await applyTop.isVisible()) {
    await applyTop.click()
    return
  }

  await expect(applyBottom).toBeVisible()
  await applyBottom.click()
}

export async function clearDraftFromPanel(page) {
  const resetTop = page.getByTestId('filters-reset-top')
  const resetBottom = page.getByTestId('filters-reset-bottom')

  if (await resetTop.isVisible()) {
    await resetTop.click()
    return
  }

  await expect(resetBottom).toBeVisible()
  await resetBottom.click()
}
