<template>
  <article
    :data-testid="`product-card-${product.id}`"
    class="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-xl dark:border-slate-800 dark:bg-slate-900"
  >
    <div
      v-if="product.bestseller"
      class="absolute left-0 top-4 z-10 rounded-r-md bg-teal-700 px-3 py-1 text-sm font-semibold text-white shadow-lg"
    >
      Bestseller
    </div>

    <div
      v-if="product.discount_percent > 0"
      class="absolute right-4 top-4 z-10 grid h-14 w-14 place-items-center rounded-full bg-indigo-600 text-sm font-semibold text-white shadow-lg"
    >
      -{{ product.discount_percent }}%
    </div>

    <figure class="bg-slate-100 px-4 pb-4 pt-8">
      <img
        :data-testid="`product-image-${product.id}`"
        :key="`${product.id}-${selectedColor}`"
        :src="resolvedImageUrl"
        :alt="selectedColor ? `${product.name} in ${selectedColor}` : `${product.name} product image`"
        loading="lazy"
        decoding="async"
        width="420"
        height="420"
        class="mx-auto h-56 w-full object-contain transition duration-300"
        @error="handleImageError"
      />
    </figure>

    <div class="space-y-4 p-4">
      <h3 class="line-clamp-2 text-3xl font-semibold tracking-tight text-slate-900 dark:text-slate-100">
        {{ product.name }}
      </h3>

      <ul
        v-if="visibleProductColors.length > 0"
        class="flex gap-2"
        :aria-label="`Available colors: ${visibleProductColors.join(', ')}`"
      >
        <li v-for="color in visibleProductColors" :key="`${product.id}-${color}`">
          <button
            type="button"
            :data-testid="`product-swatch-${product.id}-${normalizeToken(color)}`"
            class="block h-5 w-5 rounded-full border border-slate-300 transition hover:scale-110 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-600"
            :class="
              selectedColor === normalizeToken(color)
                ? 'ring-2 ring-teal-600 ring-offset-2 ring-offset-white dark:ring-teal-300 dark:ring-offset-slate-900'
                : ''
            "
            :style="{ backgroundColor: swatchColor(color) }"
            :title="color"
            :aria-label="`Preview ${product.name} in ${color}`"
            :aria-pressed="selectedColor === normalizeToken(color)"
            @click="selectColor(color)"
          ></button>
        </li>
      </ul>

      <div class="flex items-end gap-2">
        <p class="text-4xl font-bold text-teal-700">{{ formattedPrice }}</p>
        <p
          v-if="showOriginalPrice"
          class="pb-1 text-xl font-semibold text-slate-400 line-through dark:text-slate-500"
        >
          {{ formattedOriginalPrice }}
        </p>
      </div>

      <p :data-testid="`product-stock-${product.id}`" class="text-sm font-medium text-slate-500 dark:text-slate-400">
        {{ stockLabel }}
      </p>
    </div>
  </article>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { swatchColorForToken } from '../constants/colors'

const props = defineProps({
  product: {
    type: Object,
    required: true,
  },
})

const currencyFormatter = new Intl.NumberFormat('de-AT', {
  style: 'currency',
  currency: 'EUR',
})

const selectedColor = ref('')
const imageLoadFailed = ref(false)
const fallbackImageURL = '/product-placeholder.svg'

const normalizedProductColors = computed(() => {
  if (!Array.isArray(props.product.colors)) {
    return []
  }
  return props.product.colors
    .map((color) => normalizeToken(color))
    .filter((color) => Boolean(color))
})

const colorImageMap = computed(() => {
  const rawMap = props.product.image_urls_by_color
  if (!rawMap || typeof rawMap !== 'object') {
    return {}
  }
  const map = {}
  for (const [rawColor, rawURL] of Object.entries(rawMap)) {
    const color = normalizeToken(rawColor)
    const imageURL = String(rawURL ?? '').trim()
    if (!color || !imageURL) {
      continue
    }
    map[color] = imageURL
  }
  return map
})

const normalizedStockByColor = computed(() => {
  const rawMap = props.product.stock_by_color
  if (!rawMap || typeof rawMap !== 'object') {
    return {}
  }

  const map = {}
  for (const [rawColor, rawStock] of Object.entries(rawMap)) {
    const color = normalizeToken(rawColor)
    if (!color) {
      continue
    }
    const parsedStock = Number(rawStock)
    if (!Number.isFinite(parsedStock)) {
      continue
    }
    map[color] = Math.max(0, Math.trunc(parsedStock))
  }
  return map
})

const visibleProductColors = computed(() => {
  return normalizedProductColors.value.filter((color) => isColorInStock(color))
})

watch(
  () => props.product.id,
  () => {
    selectedColor.value = defaultColorForProduct()
    imageLoadFailed.value = false
  },
)

watch(
  () => visibleProductColors.value.join('|'),
  () => {
    if (!visibleProductColors.value.includes(selectedColor.value)) {
      selectedColor.value = visibleProductColors.value[0] ?? ''
    }
  },
  { immediate: true },
)

watch(
  () => selectedColor.value,
  () => {
    imageLoadFailed.value = false
  },
)

const selectedColorStock = computed(() => {
  const color = selectedColor.value
  if (!color) {
    return null
  }
  if (Object.prototype.hasOwnProperty.call(normalizedStockByColor.value, color)) {
    return normalizedStockByColor.value[color]
  }
  return null
})

const activeImageUrl = computed(() => {
  const mapped = colorImageMap.value[selectedColor.value]
  if (mapped) {
    return mapped
  }
  const fallback = String(props.product.image_url ?? '').trim()
  if (fallback) {
    return fallback
  }
  return fallbackImageURL
})

const resolvedImageUrl = computed(() => {
  if (imageLoadFailed.value) {
    return fallbackImageURL
  }
  return activeImageUrl.value
})

const formattedPrice = computed(() => currencyFormatter.format(props.product.price))

const normalizedDiscountPercent = computed(() => {
  const parsed = Number(props.product.discount_percent)
  if (!Number.isFinite(parsed)) {
    return 0
  }

  return Math.max(0, Math.min(100, Math.trunc(parsed)))
})

const showOriginalPrice = computed(() => {
  return normalizedDiscountPercent.value > 0 && normalizedDiscountPercent.value < 100
})

const formattedOriginalPrice = computed(() => {
  if (!showOriginalPrice.value) {
    return ''
  }

  const discountRatio = normalizedDiscountPercent.value / 100
  const divisor = 1 - discountRatio
  if (!Number.isFinite(divisor) || divisor <= 0) {
    return ''
  }

  const originalPrice = props.product.price / divisor
  if (!Number.isFinite(originalPrice)) {
    return ''
  }

  return currencyFormatter.format(originalPrice)
})

const stockLabel = computed(() => {
  if (selectedColorStock.value !== null) {
    const colorLabel = formatColorLabel(selectedColor.value)
    if (selectedColorStock.value <= 0) {
      return `Out of stock in ${colorLabel}`
    }
    return `${selectedColorStock.value} in stock (${colorLabel})`
  }

  const stock = Number.isFinite(props.product.stock) ? Math.max(0, props.product.stock) : 0
  if (stock <= 0) {
    return 'Out of stock'
  }
  return `${stock} in stock`
})

function swatchColor(colorName) {
  return swatchColorForToken(normalizeToken(colorName))
}

function normalizeToken(value) {
  return String(value ?? '').trim().toLowerCase()
}

function defaultColorForProduct() {
  if (visibleProductColors.value.length === 0) {
    return ''
  }
  return visibleProductColors.value[0]
}

function formatColorLabel(colorName) {
  if (!colorName) {
    return ''
  }
  return colorName.charAt(0).toUpperCase() + colorName.slice(1)
}

function selectColor(colorName) {
  const normalized = normalizeToken(colorName)
  if (!normalized) {
    return
  }
  if (!visibleProductColors.value.includes(normalized)) {
    return
  }
  selectedColor.value = normalized
}

function isColorInStock(color) {
  const normalizedColor = normalizeToken(color)
  if (!normalizedColor) {
    return false
  }

  if (Object.prototype.hasOwnProperty.call(normalizedStockByColor.value, normalizedColor)) {
    return normalizedStockByColor.value[normalizedColor] > 0
  }

  if (Object.keys(normalizedStockByColor.value).length > 0) {
    return false
  }

  const stock = Number.isFinite(props.product.stock) ? Math.max(0, Math.trunc(props.product.stock)) : 0
  return stock > 0
}

function handleImageError(event) {
  const image = event?.target
  if (!(image instanceof HTMLImageElement)) {
    return
  }

  if (image.src.endsWith(fallbackImageURL)) {
    return
  }

  imageLoadFailed.value = true
}
</script>
