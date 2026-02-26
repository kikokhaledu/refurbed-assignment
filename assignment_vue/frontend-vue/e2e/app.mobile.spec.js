import { expect, test } from '@playwright/test'
import { gotoAndWaitForInitialData, waitForCards, waitForProductsResponse } from './helpers'

test.describe('Mobile Product Discovery', () => {
  test('FE-E2E-055 FE-E2E-056 floating filters button opens modal', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const trigger = page.getByTestId('mobile-filters-trigger')
    await expect(trigger).toBeVisible()
    await trigger.click()

    await expect(page.getByTestId('mobile-filters-modal')).toBeVisible()
    await expect(page.locator('#filters-modal')).toHaveAttribute('role', 'dialog')
  })

  test('FE-E2E-057 clicking modal backdrop closes filters modal', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('mobile-filters-trigger').click()
    await expect(page.getByTestId('mobile-filters-modal')).toBeVisible()
    await page.mouse.click(10, 10)
    await expect(page.getByTestId('mobile-filters-modal')).toHaveCount(0)
  })

  test('FE-E2E-058 FE-E2E-059 FE-E2E-060 FE-E2E-061 escape closes modal, focus is trapped/restored, and body scroll is locked', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    const trigger = page.getByTestId('mobile-filters-trigger')
    await trigger.click()

    const modal = page.getByTestId('mobile-filters-modal')
    await expect(modal).toBeVisible()
    await expect.poll(() => page.evaluate(() => document.body.style.overflow)).toBe('hidden')

    for (let index = 0; index < 8; index += 1) {
      await page.keyboard.press('Tab')
      await expect
        .poll(() =>
          page.evaluate(() => {
            const modalNode = document.querySelector('#filters-modal')
            return Boolean(modalNode && modalNode.contains(document.activeElement))
          }),
        )
        .toBeTruthy()
    }

    await page.keyboard.press('Escape')
    await expect(modal).toHaveCount(0)
    await expect(trigger).toBeFocused()
    await expect.poll(() => page.evaluate(() => document.body.style.overflow)).not.toBe('hidden')
  })

  test('FE-E2E-062 FE-E2E-063 floating top button appears on scroll and returns to top', async ({ page }) => {
    await gotoAndWaitForInitialData(page)
    await waitForCards(page, 1)

    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await expect(page.getByTestId('mobile-scroll-top')).toBeVisible()

    await page.getByTestId('mobile-scroll-top').click()
    await expect.poll(() => page.evaluate(() => window.scrollY)).toBeLessThan(20)
  })

  test('FE-E2E-075 mobile active filters chips are collapsed by default with summary and toggle', async ({ page }) => {
    await gotoAndWaitForInitialData(page)

    await page.getByTestId('mobile-filters-trigger').click()
    const mobileModal = page.getByTestId('mobile-filters-modal')
    await mobileModal.getByTestId('filter-category-smartphones').check()

    const applyResponse = waitForProductsResponse(page, (url) => {
      return url.searchParams.getAll('category').includes('smartphones')
    })
    await mobileModal.getByTestId('filters-apply-top').click()
    await applyResponse

    await expect(page.getByTestId('mobile-filters-modal')).toHaveCount(0)
    await expect(page.getByTestId('active-chip-summary-mobile')).toContainText('1 filter applied')

    const toggle = page.getByTestId('active-chip-toggle-mobile')
    await expect(toggle).toHaveAttribute('aria-expanded', 'false')
    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-expanded', 'true')
    await expect(page.getByTestId('active-chip-category:smartphones')).toBeVisible()

    const clearResponse = waitForProductsResponse(page, (url) => !url.searchParams.has('category'))
    await page.getByTestId('active-chip-clear-all-mobile').click()
    await clearResponse
    await expect(page.getByTestId('active-filter-chips')).toHaveCount(0)
  })
})
