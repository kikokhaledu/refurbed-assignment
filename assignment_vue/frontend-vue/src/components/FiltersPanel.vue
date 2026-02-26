<template>
  <section data-testid="filters-panel" class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
    <div ref="topControlsRef" data-testid="filters-top-controls" class="mb-6 flex items-center justify-between">
      <h2 class="text-2xl font-semibold text-slate-800 dark:text-slate-100">Filters</h2>
      <div class="flex items-center gap-2">
        <button
          type="button"
          data-testid="filters-reset-top"
          class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
          @click="$emit('clear')"
        >
          Reset
        </button>
        <button
          type="button"
          data-testid="filters-apply-top"
          class="rounded-lg border border-teal-700 bg-teal-700 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-teal-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-teal-600 disabled:cursor-not-allowed disabled:border-slate-300 disabled:bg-slate-200 disabled:text-slate-500 dark:disabled:border-slate-700 dark:disabled:bg-slate-800 dark:disabled:text-slate-400"
          :disabled="applyDisabled"
          @click="$emit('apply')"
        >
          Apply
        </button>
      </div>
    </div>
    <p
      v-if="hasPendingChanges && !showBottomActionBar"
      data-testid="filter-unapplied-warning-top"
      class="mb-4 text-sm font-medium text-amber-700 dark:text-amber-300"
    >
      You have unapplied changes.
    </p>

    <div class="space-y-6">
      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Category</h3>
        <ul class="mt-3 space-y-2">
          <li v-for="category in categoryOptions" :key="category">
            <label class="inline-flex cursor-pointer items-center gap-3 text-base text-slate-700 dark:text-slate-300">
              <input
                :data-testid="`filter-category-${category}`"
                :checked="selectedCategories.includes(category)"
                :disabled="disabled"
                type="checkbox"
                class="h-4 w-4 rounded border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
                @change="toggleCategory(category)"
              />
              <span class="capitalize">{{ category }}</span>
            </label>
          </li>
        </ul>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Brand</h3>
        <ul class="mt-3 space-y-2">
          <li v-for="brand in brandOptions" :key="brand">
            <label class="inline-flex cursor-pointer items-center gap-3 text-base text-slate-700 dark:text-slate-300">
              <input
                :data-testid="`filter-brand-${brand}`"
                :checked="selectedBrands.includes(brand)"
                :disabled="disabled"
                type="checkbox"
                class="h-4 w-4 rounded border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
                @change="toggleBrand(brand)"
              />
              <span class="capitalize">{{ brand }}</span>
            </label>
          </li>
        </ul>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Colors</h3>
        <ul class="mt-3 grid grid-cols-2 gap-2">
          <li v-for="color in colorOptions" :key="color">
            <button
              type="button"
              :data-testid="`filter-color-${color}`"
              :disabled="disabled"
              :aria-pressed="selectedColors.includes(color)"
              class="flex w-full min-w-0 items-center gap-2 rounded-lg border px-2 py-1.5 text-left text-sm font-medium transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
              :class="
                selectedColors.includes(color)
                  ? 'border-teal-700 bg-teal-50 text-teal-900 ring-2 ring-teal-600/35 dark:border-teal-300 dark:bg-teal-950/50 dark:text-teal-100 dark:ring-teal-300/50'
                  : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700'
              "
              @click="toggleColor(color)"
            >
              <span class="inline-flex min-w-0 items-center gap-2">
                <span
                  class="inline-block h-4 w-4 shrink-0 rounded-full border border-slate-300 dark:border-slate-600"
                  :style="{ backgroundColor: swatchColor(color) }"
                />
                <span class="truncate capitalize">{{ color }}</span>
              </span>
              <span
                v-if="selectedColors.includes(color)"
                class="ml-auto h-2.5 w-2.5 shrink-0 rounded-full bg-teal-700 dark:bg-teal-200"
                aria-hidden="true"
              ></span>
              <span v-if="selectedColors.includes(color)" class="sr-only">
                Selected
              </span>
            </button>
          </li>
        </ul>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Price range</h3>
        <div class="mt-3 space-y-3">
          <div class="slider-shell relative h-8">
            <div
              class="slider-inner absolute inset-y-0"
              @mouseenter="isSliderHovering = true"
              @mouseleave="isSliderHovering = false"
            >
              <span
                data-testid="price-min-badge"
                class="pointer-events-none absolute z-40 whitespace-nowrap rounded-full border border-slate-300 bg-white px-3 py-1 text-xs font-semibold text-slate-700 opacity-0 shadow-sm transition-opacity duration-150 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200"
                :class="showSliderIndicators ? 'opacity-100' : 'opacity-0'"
                :style="minIndicatorStyle"
              >
                Min EUR {{ sliderMin }}
              </span>
              <span
                data-testid="price-max-badge"
                class="pointer-events-none absolute z-40 whitespace-nowrap rounded-full border border-slate-300 bg-white px-3 py-1 text-xs font-semibold text-slate-700 opacity-0 shadow-sm transition-opacity duration-150 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200"
                :class="showSliderIndicators ? 'opacity-100' : 'opacity-0'"
                :style="maxIndicatorStyle"
              >
                Max EUR {{ sliderMax }}
              </span>

              <div
                class="absolute left-0 right-0 top-1/2 h-2 -translate-y-1/2 rounded-full bg-slate-200 dark:bg-slate-700"
              ></div>
              <div
                class="absolute top-1/2 h-2 -translate-y-1/2 rounded-full bg-teal-700 dark:bg-teal-400"
                :style="sliderRangeStyle"
              ></div>

              <input
                ref="minSliderRef"
                data-testid="price-slider-min"
                :value="sliderMin"
                :disabled="disabled"
                type="range"
                :min="normalizedFloor"
                :max="normalizedCeiling"
                step="1"
                aria-label="Minimum price"
                :class="['dual-thumb-slider', activeThumb === 'min' ? 'z-30' : 'z-20']"
                @pointerdown="handleThumbPointerDown('min')"
                @pointerup="handleThumbPointerUp"
                @pointercancel="handleThumbPointerUp"
                @focus="handleThumbFocus('min')"
                @blur="handleThumbBlur"
                @input="handleMinRangeInput"
              />
              <input
                ref="maxSliderRef"
                data-testid="price-slider-max"
                :value="sliderMax"
                :disabled="disabled"
                type="range"
                :min="normalizedFloor"
                :max="normalizedCeiling"
                step="1"
                aria-label="Maximum price"
                :class="['dual-thumb-slider', activeThumb === 'max' ? 'z-30' : 'z-20']"
                @pointerdown="handleThumbPointerDown('max')"
                @pointerup="handleThumbPointerUp"
                @pointercancel="handleThumbPointerUp"
                @focus="handleThumbFocus('max')"
                @blur="handleThumbBlur"
                @input="handleMaxRangeInput"
              />
            </div>
          </div>

          <div class="flex items-center justify-between text-xs font-medium text-slate-500 dark:text-slate-400">
            <span>EUR {{ normalizedFloor }}</span>
            <span>EUR {{ normalizedCeiling }}</span>
          </div>
        </div>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Discount</h3>
        <div class="bestseller-options mt-3">
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-onsale-all"
              type="radio"
              name="on-sale-filter"
              value="all"
              :checked="onSale === 'all'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:onSale', 'all')"
            />
            <span>All</span>
          </label>
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-onsale-true"
              type="radio"
              name="on-sale-filter"
              value="true"
              :checked="onSale === 'true'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:onSale', 'true')"
            />
            <span>On sale only</span>
          </label>
        </div>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Condition</h3>
        <ul class="mt-3 space-y-2">
          <li v-for="condition in conditionOptions" :key="condition">
            <label class="inline-flex cursor-pointer items-center gap-3 text-base text-slate-700 dark:text-slate-300">
              <input
                :data-testid="`filter-condition-${condition}`"
                :checked="selectedConditions.includes(condition)"
                :disabled="disabled"
                type="checkbox"
                class="h-4 w-4 rounded border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
                @change="toggleCondition(condition)"
              />
              <span class="capitalize">{{ condition }}</span>
            </label>
          </li>
        </ul>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Availability</h3>
        <div class="bestseller-options mt-3">
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-instock-all"
              type="radio"
              name="stock-filter"
              value="all"
              :checked="inStock === 'all'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:inStock', 'all')"
            />
            <span>All</span>
          </label>
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-instock-true"
              type="radio"
              name="stock-filter"
              value="true"
              :checked="inStock === 'true'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:inStock', 'true')"
            />
            <span>In stock</span>
          </label>
        </div>
      </section>

      <section>
        <h3 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Bestseller</h3>
        <div class="bestseller-options mt-3">
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-bestseller-all"
              type="radio"
              name="bestseller-filter"
              value="all"
              :checked="bestseller === 'all'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:bestseller', 'all')"
            />
            <span>All</span>
          </label>
          <label class="bestseller-option text-base text-slate-700 dark:text-slate-300">
            <input
              data-testid="filter-bestseller-true"
              type="radio"
              name="bestseller-filter"
              value="true"
              :checked="bestseller === 'true'"
              :disabled="disabled"
              class="h-4 w-4 border-slate-300 bg-white text-teal-600 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-teal-400 dark:focus:ring-teal-400"
              @change="$emit('update:bestseller', 'true')"
            />
            <span>Bestseller only</span>
          </label>
        </div>
      </section>

      <p v-if="validationMessage" data-testid="filter-validation-message" class="mt-1 text-sm text-rose-600 dark:text-rose-300">
        {{ validationMessage }}
      </p>
    </div>

    <div
      v-if="showBottomActionBar"
      data-testid="filters-bottom-actions"
      class="sticky bottom-3 z-10 mt-6 rounded-xl border border-slate-200 bg-white/95 p-3 shadow-lg backdrop-blur dark:border-slate-700 dark:bg-slate-900/95"
    >
      <p
        v-if="hasPendingChanges"
        data-testid="filter-unapplied-warning-bottom"
        class="mb-2 text-sm font-medium text-amber-700 dark:text-amber-300"
      >
        You have unapplied changes.
      </p>
      <div class="flex items-center justify-end gap-2">
        <button
          type="button"
          data-testid="filters-reset-bottom"
          class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
          @click="$emit('clear')"
        >
          Reset
        </button>
        <button
          type="button"
          data-testid="filters-apply-bottom"
          class="rounded-lg border border-teal-700 bg-teal-700 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-teal-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-teal-600 disabled:cursor-not-allowed disabled:border-slate-300 disabled:bg-slate-200 disabled:text-slate-500 dark:disabled:border-slate-700 dark:disabled:bg-slate-800 dark:disabled:text-slate-400"
          :disabled="applyDisabled"
          @click="$emit('apply')"
        >
          Apply
        </button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { swatchColorForToken } from '../constants/colors'

const props = defineProps({
  selectedCategories: {
    type: Array,
    default: () => [],
  },
  selectedBrands: {
    type: Array,
    default: () => [],
  },
  selectedColors: {
    type: Array,
    default: () => [],
  },
  selectedConditions: {
    type: Array,
    default: () => [],
  },
  bestseller: {
    type: String,
    default: 'all',
  },
  onSale: {
    type: String,
    default: 'all',
  },
  inStock: {
    type: String,
    default: 'all',
  },
  minPrice: {
    type: String,
    default: '',
  },
  maxPrice: {
    type: String,
    default: '',
  },
  priceFloor: {
    type: Number,
    default: 0,
  },
  priceCeiling: {
    type: Number,
    default: 0,
  },
  categoryOptions: {
    type: Array,
    default: () => [],
  },
  brandOptions: {
    type: Array,
    default: () => [],
  },
  colorOptions: {
    type: Array,
    default: () => [],
  },
  conditionOptions: {
    type: Array,
    default: () => [],
  },
  validationMessage: {
    type: String,
    default: '',
  },
  hasPendingChanges: {
    type: Boolean,
    default: false,
  },
  applyDisabled: {
    type: Boolean,
    default: true,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits([
  'apply',
  'clear',
  'update:selectedCategories',
  'update:selectedBrands',
  'update:selectedColors',
  'update:selectedConditions',
  'update:bestseller',
  'update:onSale',
  'update:inStock',
  'update:minPrice',
  'update:maxPrice',
])

const topControlsRef = ref(null)
const isTopControlsVisible = ref(true)
const showBottomActionBar = computed(() => !isTopControlsVisible.value)
const localSliderMin = ref(0)
const localSliderMax = ref(0)
const activeThumb = ref('max')
const minSliderRef = ref(null)
const maxSliderRef = ref(null)
const sliderMinGap = 1
const isSliderHovering = ref(false)
const isSliderInteracting = ref(false)

let topControlsObserver = null

onMounted(() => {
  if (typeof window === 'undefined' || typeof window.IntersectionObserver !== 'function') {
    return
  }
  if (!topControlsRef.value) {
    return
  }

  topControlsObserver = new window.IntersectionObserver(
    (entries) => {
      const [entry] = entries
      isTopControlsVisible.value = entry?.isIntersecting ?? true
    },
    { threshold: 0.05 },
  )

  topControlsObserver.observe(topControlsRef.value)

  window.addEventListener('pointerup', handleGlobalPointerRelease)
  window.addEventListener('pointercancel', handleGlobalPointerRelease)
})

onBeforeUnmount(() => {
  if (topControlsObserver) {
    topControlsObserver.disconnect()
    topControlsObserver = null
  }

  if (typeof window !== 'undefined') {
    window.removeEventListener('pointerup', handleGlobalPointerRelease)
    window.removeEventListener('pointercancel', handleGlobalPointerRelease)
  }
})

const normalizedFloor = computed(() => {
  const floor = Number.isFinite(props.priceFloor) ? Math.round(props.priceFloor) : 0
  const ceiling = Number.isFinite(props.priceCeiling) ? Math.round(props.priceCeiling) : 0
  return Math.min(floor, ceiling)
})

const normalizedCeiling = computed(() => {
  const floor = Number.isFinite(props.priceFloor) ? Math.round(props.priceFloor) : 0
  const ceiling = Number.isFinite(props.priceCeiling) ? Math.round(props.priceCeiling) : 0
  return Math.max(floor, ceiling)
})

watch(
  [() => props.priceFloor, () => props.priceCeiling, () => props.minPrice, () => props.maxPrice],
  () => {
    const floor = normalizedFloor.value
    const ceiling = normalizedCeiling.value

    const parsedMin = parsePriceValue(props.minPrice)
    const parsedMax = parsePriceValue(props.maxPrice)

    let nextMin = parsedMin === null ? floor : clampToRange(parsedMin, floor, ceiling)
    let nextMax = parsedMax === null ? ceiling : clampToRange(parsedMax, floor, ceiling)

    if (nextMin > nextMax) {
      nextMax = nextMin
    }

    localSliderMin.value = nextMin
    localSliderMax.value = nextMax
    void syncSliderDomValues()
  },
  { immediate: true },
)

const sliderMin = computed(() => localSliderMin.value)
const sliderMax = computed(() => localSliderMax.value)

const sliderRangeStyle = computed(() => {
  const floor = normalizedFloor.value
  const ceiling = normalizedCeiling.value
  if (!Number.isFinite(floor) || !Number.isFinite(ceiling) || ceiling <= floor) {
    return {
      left: '0%',
      width: '0%',
    }
  }

  const minPercent = sliderPercents.value.min
  const maxPercent = sliderPercents.value.max

  return {
    left: `${Math.max(0, Math.min(100, minPercent))}%`,
    width: `${Math.max(0, Math.min(100, maxPercent) - Math.max(0, minPercent))}%`,
  }
})

const sliderPercents = computed(() => {
  const floor = normalizedFloor.value
  const ceiling = normalizedCeiling.value
  if (!Number.isFinite(floor) || !Number.isFinite(ceiling) || ceiling <= floor) {
    return { min: 0, max: 0 }
  }

  const range = ceiling - floor
  const rangeStart = Math.min(sliderMin.value, sliderMax.value)
  const rangeEnd = Math.max(sliderMin.value, sliderMax.value)

  const min = ((rangeStart - floor) / range) * 100
  const max = ((rangeEnd - floor) / range) * 100

  return {
    min: Math.max(0, Math.min(100, min)),
    max: Math.max(0, Math.min(100, max)),
  }
})

const showSliderIndicators = computed(() => isSliderHovering.value || isSliderInteracting.value)

const minIndicatorStyle = computed(() => {
  const leftPercent = sliderPercents.value.min
  return {
    top: '-0.2rem',
    left: `${leftPercent}%`,
    transform: 'translate(-50%, -100%)',
  }
})

const maxIndicatorStyle = computed(() => {
  const leftPercent = sliderPercents.value.max
  return {
    top: '-0.2rem',
    left: `${leftPercent}%`,
    transform: 'translate(-50%, -100%)',
  }
})

function toggleCategory(category) {
  const selected = new Set(props.selectedCategories)
  if (selected.has(category)) {
    selected.delete(category)
  } else {
    selected.add(category)
  }
  emit('update:selectedCategories', [...selected])
}

function toggleBrand(brand) {
  const selected = new Set(props.selectedBrands)
  if (selected.has(brand)) {
    selected.delete(brand)
  } else {
    selected.add(brand)
  }
  emit('update:selectedBrands', [...selected])
}

function toggleColor(color) {
  const selected = new Set(props.selectedColors)
  if (selected.has(color)) {
    selected.delete(color)
  } else {
    selected.add(color)
  }
  emit('update:selectedColors', [...selected])
}

function toggleCondition(condition) {
  const selected = new Set(props.selectedConditions)
  if (selected.has(condition)) {
    selected.delete(condition)
  } else {
    selected.add(condition)
  }
  emit('update:selectedConditions', [...selected])
}

function swatchColor(colorName) {
  return swatchColorForToken(colorName)
}

function handleMinRangeInput(event) {
  const raw = Number.parseInt(event.target.value, 10)
  const floor = normalizedFloor.value
  const ceiling = normalizedCeiling.value
  const enforcedGap = ceiling > floor ? sliderMinGap : 0
  const maxForMin = Math.max(floor, sliderMax.value - enforcedGap)
  const safeMin = Number.isNaN(raw) ? floor : clampToRange(raw, floor, maxForMin)
  localSliderMin.value = safeMin
  emit('update:minPrice', String(safeMin))
}

function handleMaxRangeInput(event) {
  const raw = Number.parseInt(event.target.value, 10)
  const floor = normalizedFloor.value
  const ceiling = normalizedCeiling.value
  const enforcedGap = ceiling > floor ? sliderMinGap : 0
  const minForMax = Math.min(ceiling, sliderMin.value + enforcedGap)
  const safeMax = Number.isNaN(raw) ? ceiling : clampToRange(raw, minForMax, ceiling)
  localSliderMax.value = safeMax
  emit('update:maxPrice', String(safeMax))
}

function setActiveThumb(thumb) {
  activeThumb.value = thumb
}

function handleThumbPointerDown(thumb) {
  setActiveThumb(thumb)
  isSliderInteracting.value = true
}

function handleThumbPointerUp() {
  isSliderInteracting.value = false
}

function handleThumbFocus(thumb) {
  setActiveThumb(thumb)
  isSliderInteracting.value = true
}

function handleThumbBlur() {
  if (!isSliderHovering.value) {
    isSliderInteracting.value = false
  }
}

async function syncSliderDomValues() {
  await nextTick()

  if (minSliderRef.value instanceof HTMLInputElement) {
    minSliderRef.value.value = String(localSliderMin.value)
  }
  if (maxSliderRef.value instanceof HTMLInputElement) {
    maxSliderRef.value.value = String(localSliderMax.value)
  }
}

function parsePriceValue(rawValue) {
  const normalized = String(rawValue ?? '').trim()
  if (!normalized) {
    return null
  }
  const parsed = Number.parseFloat(normalized)
  if (!Number.isFinite(parsed)) {
    return null
  }
  return Math.round(parsed)
}

function clampToRange(value, min, max) {
  return Math.min(max, Math.max(min, value))
}

function handleGlobalPointerRelease() {
  if (!isSliderHovering.value) {
    isSliderInteracting.value = false
  }
}
</script>

<style scoped>
.bestseller-options {
  display: grid;
  row-gap: 0.75rem;
}

.bestseller-option {
  display: flex;
  align-items: center;
  column-gap: 0.75rem;
}

.bestseller-option input[type='radio'] {
  margin: 0;
  flex-shrink: 0;
}

.slider-shell {
  --thumb-size: 1.1rem;
  --thumb-half: calc(var(--thumb-size) / 2);
}

.slider-inner {
  left: var(--thumb-half);
  right: var(--thumb-half);
}

.dual-thumb-slider {
  position: absolute;
  inset: 0;
  width: 100%;
  margin: 0;
  pointer-events: none;
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
}

.dual-thumb-slider::-webkit-slider-runnable-track {
  height: 0;
  background: transparent;
}

.dual-thumb-slider::-moz-range-track {
  height: 0;
  background: transparent;
}

.dual-thumb-slider::-webkit-slider-thumb {
  pointer-events: auto;
  height: 1.1rem;
  width: 1.1rem;
  margin-top: -0.55rem;
  border-radius: 9999px;
  border: 2px solid #0f172a;
  background: #ffffff;
  cursor: pointer;
  -webkit-appearance: none;
}

.dual-thumb-slider::-moz-range-thumb {
  pointer-events: auto;
  height: 1.1rem;
  width: 1.1rem;
  border-radius: 9999px;
  border: 2px solid #0f172a;
  background: #ffffff;
  cursor: pointer;
}

:global(.dark) .dual-thumb-slider::-webkit-slider-thumb {
  border-color: #22d3ee;
  background: #0f172a;
}

:global(.dark) .dual-thumb-slider::-moz-range-thumb {
  border-color: #22d3ee;
  background: #0f172a;
}

.dual-thumb-slider:disabled::-webkit-slider-thumb {
  cursor: not-allowed;
  opacity: 0.6;
}

.dual-thumb-slider:disabled::-moz-range-thumb {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>

