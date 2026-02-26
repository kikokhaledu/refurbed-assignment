import { mount } from '@vue/test-utils'
import ActiveFilterChips from '../ActiveFilterChips.vue'

describe('ActiveFilterChips', () => {
  it('renders compact mobile summary and emits chip events', async () => {
    const chips = [
      { id: 'search', label: 'Search: "iphone"' },
      { id: 'sort', label: 'Sort: Popularity' },
    ]

    const wrapper = mount(ActiveFilterChips, {
      props: { chips },
    })

    expect(wrapper.get('[data-testid="active-chip-summary-mobile"]').text()).toContain('2 filters applied')
    expect(wrapper.text()).toContain('Search: "iphone"')
    expect(wrapper.text()).toContain('Sort: Popularity')
    expect(wrapper.find('[data-testid="active-chip-clear-all"]').exists()).toBe(true)

    await wrapper.get('[data-testid="active-chip-search"]').trigger('click')
    expect(wrapper.emitted('remove-chip')).toBeTruthy()
    expect(wrapper.emitted('remove-chip')[0][0]).toEqual(chips[0])

    await wrapper.get('[data-testid="active-chip-clear-all-mobile"]').trigger('click')
    expect(wrapper.emitted('clear-all')).toBeTruthy()
  })

  it('toggles mobile chip list visibility', async () => {
    const wrapper = mount(ActiveFilterChips, {
      props: {
        chips: [{ id: 'category:smartphones', label: 'Category: Smartphones' }],
      },
    })

    const toggle = wrapper.get('[data-testid="active-chip-toggle-mobile"]')
    const list = wrapper.get('[data-testid="active-chip-list"]')

    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(list.classes()).toContain('hidden')

    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(list.classes()).not.toContain('hidden')

    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(list.classes()).toContain('hidden')
  })
})
