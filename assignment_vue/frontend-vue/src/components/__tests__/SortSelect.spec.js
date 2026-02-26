import { mount } from '@vue/test-utils'
import SortSelect from '../SortSelect.vue'

describe('SortSelect', () => {
  it('emits selected single sort value on close', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: '',
      },
    })

    await wrapper.get('[data-testid="sort-select"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-price_desc"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    await wrapper.get('[data-testid="sort-select"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')[0][0]).toBe('price_desc')
  })

  it('emits combined non-contradicting sorts', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: '',
      },
    })

    await wrapper.get('[data-testid="sort-select"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-popularity"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-price_asc"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    await wrapper.get('[data-testid="sort-select"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')[0][0]).toBe('popularity,price_asc')
  })

  it('preserves user pick order for non-contradicting sorts', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: '',
      },
    })

    await wrapper.get('[data-testid="sort-select"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-price_asc"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-popularity"]').trigger('click')
    await wrapper.get('[data-testid="sort-select"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')[0][0]).toBe('price_asc,popularity')
  })

  it('auto-unchecks contradicting price sort', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: 'price_asc',
      },
    })

    await wrapper.get('[data-testid="sort-select"]').trigger('click')
    await wrapper.get('[data-testid="sort-option-price_desc"]').trigger('click')
    await wrapper.get('[data-testid="sort-select"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')[0][0]).toBe('price_desc')
  })

  it('renders only base sort options', async () => {
    const wrapper = mount(SortSelect, {
      props: {
        modelValue: '',
      },
    })

    await wrapper.get('[data-testid="sort-select"]').trigger('click')
    const options = wrapper.findAll('[data-testid^="sort-option-"]').map((option) => option.attributes('data-testid').replace('sort-option-', ''))
    expect(options).toEqual([
      'popularity',
      'price_asc',
      'price_desc',
    ])
  })
})
