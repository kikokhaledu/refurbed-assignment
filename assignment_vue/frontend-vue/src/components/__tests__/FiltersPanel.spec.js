import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import FiltersPanel from '../FiltersPanel.vue'

function baseProps(overrides = {}) {
  return {
    selectedCategories: [],
    selectedBrands: [],
    selectedColors: [],
    selectedConditions: [],
    bestseller: 'all',
    onSale: 'all',
    inStock: 'all',
    minPrice: '',
    maxPrice: '',
    priceFloor: 0,
    priceCeiling: 100,
    categoryOptions: ['smartphones'],
    brandOptions: ['apple', 'samsung'],
    colorOptions: ['blue', 'red'],
    conditionOptions: ['refurbished'],
    validationMessage: '',
    hasPendingChanges: true,
    applyDisabled: false,
    disabled: false,
    ...overrides,
  }
}

describe('FiltersPanel', () => {
  it('clamps minimum slider value to stay at least 1 EUR below current maximum', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '10',
        maxPrice: '50',
      }),
    })

    const sliders = wrapper.findAll('input[type="range"]')
    expect(sliders).toHaveLength(2)

    await sliders[0].setValue('90')

    const emitted = wrapper.emitted('update:minPrice')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toBe('49')
  })

  it('clamps maximum slider value to stay at least 1 EUR above current minimum', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '40',
        maxPrice: '60',
      }),
    })

    const sliders = wrapper.findAll('input[type="range"]')
    expect(sliders).toHaveLength(2)

    await sliders[1].setValue('20')

    const emitted = wrapper.emitted('update:maxPrice')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toBe('41')
  })

  it('keeps both sliders on the same floor/ceiling scale', () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '40',
        maxPrice: '60',
      }),
    })

    const minSlider = wrapper.get('[data-testid="price-slider-min"]')
    const maxSlider = wrapper.get('[data-testid="price-slider-max"]')

    expect(minSlider.attributes('min')).toBe('0')
    expect(minSlider.attributes('max')).toBe('100')
    expect(maxSlider.attributes('min')).toBe('0')
    expect(maxSlider.attributes('max')).toBe('100')
  })

  it('initializes slider thumbs to floor and ceiling when no min/max filters are set', () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '',
        maxPrice: '',
        priceFloor: 100,
        priceCeiling: 900,
      }),
    })

    const minSlider = wrapper.get('[data-testid="price-slider-min"]')
    const maxSlider = wrapper.get('[data-testid="price-slider-max"]')

    expect(minSlider.element.value).toBe('100')
    expect(maxSlider.element.value).toBe('900')
  })

  it('re-syncs slider thumbs to updated floor and ceiling bounds', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        minPrice: '',
        maxPrice: '',
        priceFloor: 0,
        priceCeiling: 0,
      }),
    })

    await wrapper.setProps({
      priceFloor: 99,
      priceCeiling: 1425,
    })
    await nextTick()

    const minSlider = wrapper.get('[data-testid="price-slider-min"]')
    const maxSlider = wrapper.get('[data-testid="price-slider-max"]')
    expect(minSlider.element.value).toBe('99')
    expect(maxSlider.element.value).toBe('1425')
  })

  it('shows bottom sticky actions when top controls scroll out of view', async () => {
    let observerCallback = null
    const observe = vi.fn()
    const disconnect = vi.fn()

    globalThis.IntersectionObserver = vi.fn((cb) => {
      observerCallback = cb
      return { observe, disconnect }
    })

    const wrapper = mount(FiltersPanel, {
      props: baseProps(),
    })

    expect(observe).toHaveBeenCalled()
    expect(wrapper.find('div.sticky.bottom-3').exists()).toBe(false)
    expect(wrapper.text()).toContain('You have unapplied changes.')

    observerCallback?.([{ isIntersecting: false }])
    await nextTick()

    expect(wrapper.find('div.sticky.bottom-3').exists()).toBe(true)
    expect(wrapper.text()).toContain('You have unapplied changes.')

    wrapper.unmount()
    expect(disconnect).toHaveBeenCalled()
  })

  it('emits selected brands updates when a brand is toggled', async () => {
    const wrapper = mount(FiltersPanel, {
      props: baseProps({
        selectedBrands: ['apple'],
      }),
    })

    const samsungCheckbox = wrapper.get('[data-testid="filter-brand-samsung"]')
    await samsungCheckbox.setValue(true)

    const emitted = wrapper.emitted('update:selectedBrands')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toContain('apple')
    expect(emitted[0][0]).toContain('samsung')
  })
})
