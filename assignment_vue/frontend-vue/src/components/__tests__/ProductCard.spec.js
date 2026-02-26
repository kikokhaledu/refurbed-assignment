import { mount } from '@vue/test-utils'
import ProductCard from '../ProductCard.vue'

function product(overrides = {}) {
  return {
    id: 'p1',
    name: 'Phone',
    price: 499.99,
    discount_percent: 10,
    bestseller: true,
    colors: ['blue', 'red', 'green'],
    stock_by_color: {
      blue: 3,
      red: 0,
      green: 1,
    },
    image_url: 'https://example.com/default.jpg',
    image_urls_by_color: {
      blue: 'https://example.com/blue.jpg',
      green: 'https://example.com/green.jpg',
    },
    stock: 4,
    category: 'smartphones',
    brand: 'apple',
    condition: 'refurbished',
    ...overrides,
  }
}

describe('ProductCard', () => {
  it('hides out-of-stock colors from swatches', () => {
    const wrapper = mount(ProductCard, {
      props: {
        product: product(),
      },
    })

    const swatches = wrapper.findAll('button[aria-pressed]')
    expect(swatches).toHaveLength(2)
    expect(wrapper.find('button[title="red"]').exists()).toBe(false)
    expect(wrapper.find('button[title="blue"]').exists()).toBe(true)
    expect(wrapper.find('button[title="green"]').exists()).toBe(true)
  })

  it('switches image and stock label when a visible color is selected', async () => {
    const wrapper = mount(ProductCard, {
      props: {
        product: product(),
      },
    })

    const image = wrapper.find('img')
    expect(image.attributes('src')).toBe('https://example.com/blue.jpg')
    expect(wrapper.text()).toContain('3 in stock (Blue)')

    await wrapper.find('button[title="green"]').trigger('click')

    expect(wrapper.find('img').attributes('src')).toBe('https://example.com/green.jpg')
    expect(wrapper.text()).toContain('1 in stock (Green)')
  })

  it('falls back to base image_url when color-specific image is missing', async () => {
    const wrapper = mount(ProductCard, {
      props: {
        product: product({
          image_urls_by_color: {
            blue: 'https://example.com/blue.jpg',
          },
        }),
      },
    })

    await wrapper.find('button[title="green"]').trigger('click')
    expect(wrapper.find('img').attributes('src')).toBe('https://example.com/default.jpg')
  })

  it('switches to placeholder image when active image fails to load', async () => {
    const wrapper = mount(ProductCard, {
      props: {
        product: product({
          image_urls_by_color: {
            blue: 'https://example.com/broken-blue.jpg',
          },
        }),
      },
    })

    const image = wrapper.find('img')
    await image.trigger('error')

    expect(wrapper.find('img').attributes('src')).toContain('/product-placeholder.svg')
  })

  it('does not render an original-price value when discount is 100 percent', () => {
    const wrapper = mount(ProductCard, {
      props: {
        product: product({
          discount_percent: 100,
          price: 499.99,
        }),
      },
    })

    expect(wrapper.find('p.line-through').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('NaN')
    expect(wrapper.text()).not.toContain('∞')
  })
})
