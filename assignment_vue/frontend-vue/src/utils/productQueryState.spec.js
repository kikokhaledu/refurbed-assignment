import { describe, expect, it } from 'vitest'
import { parseURLState } from './productQueryState'

describe('productQueryState.parseURLState', () => {
  it('drops non-numeric price values from URL state', () => {
    const { filters } = parseURLState('?minPrice=abc&maxPrice=500')
    expect(filters.minPrice).toBe('')
    expect(filters.maxPrice).toBe('500')
  })

  it('drops negative price values from URL state', () => {
    const { filters } = parseURLState('?minPrice=-2&maxPrice=500')
    expect(filters.minPrice).toBe('')
    expect(filters.maxPrice).toBe('500')
  })

  it('drops both price values when min is greater than max', () => {
    const { filters } = parseURLState('?minPrice=900&maxPrice=200')
    expect(filters.minPrice).toBe('')
    expect(filters.maxPrice).toBe('')
  })

  it('keeps valid price bounds from URL state', () => {
    const { filters } = parseURLState('?minPrice=200&maxPrice=900')
    expect(filters.minPrice).toBe('200')
    expect(filters.maxPrice).toBe('900')
  })
})
