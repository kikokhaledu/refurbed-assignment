<template>
  <div ref="rootRef" class="relative flex items-center gap-3">
    <label
      for="product-sort"
      class="text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400"
    >
      Sort
    </label>
    <button
      id="product-sort"
      data-testid="sort-select"
      type="button"
      :disabled="disabled"
      :data-sort-value="modelValue || ''"
      :aria-expanded="isOpen ? 'true' : 'false'"
      aria-haspopup="listbox"
      aria-controls="sort-dropdown-menu"
      class="flex h-10 min-w-[220px] items-center justify-between rounded-lg border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 disabled:cursor-not-allowed disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
      @click="toggleDropdown"
    >
      <span class="truncate">{{ triggerLabel }}</span>
      <svg
        aria-hidden="true"
        class="ml-3 h-4 w-4 shrink-0 text-slate-500 transition dark:text-slate-400"
        :class="isOpen ? 'rotate-180' : ''"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M5.23 7.21a.75.75 0 0 1 1.06.02L10 11.168l3.71-3.938a.75.75 0 1 1 1.08 1.04l-4.25 4.512a.75.75 0 0 1-1.08 0L5.21 8.27a.75.75 0 0 1 .02-1.06z"
          clip-rule="evenodd"
        />
      </svg>
    </button>
    <div
      v-if="isOpen"
      id="sort-dropdown-menu"
      data-testid="sort-dropdown"
      role="listbox"
      aria-multiselectable="true"
      class="absolute right-0 top-12 z-30 w-72 rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl dark:border-slate-700 dark:bg-slate-900"
    >
      <button
        v-for="option in sortOptions"
        :key="option.value"
        type="button"
        :data-testid="`sort-option-${option.value}`"
        :data-selected="isSelected(option.value) ? 'true' : 'false'"
        role="option"
        :aria-selected="isSelected(option.value) ? 'true' : 'false'"
        class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm font-medium transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700"
        :class="
          isSelected(option.value)
            ? 'bg-emerald-50 text-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-100'
            : 'text-slate-700 hover:bg-slate-50 dark:text-slate-200 dark:hover:bg-slate-800'
        "
        @click="toggleMode(option.value)"
      >
        <span>{{ option.label }}</span>
        <span
          class="ml-3 inline-flex h-5 w-5 items-center justify-center rounded-full border"
          :class="
            isSelected(option.value)
              ? 'border-emerald-600 bg-emerald-600 text-white dark:border-emerald-400 dark:bg-emerald-400 dark:text-slate-900'
              : 'border-slate-300 text-transparent dark:border-slate-600'
          "
          aria-hidden="true"
        >
          <svg
            v-if="isSelected(option.value)"
            viewBox="0 0 20 20"
            fill="none"
            class="h-3.5 w-3.5"
            aria-hidden="true"
          >
            <path
              d="M5 10.5L8.5 14L15 7.5"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { SORT_LABELS, SORT_OPTIONS } from '../constants/sort'
import { parseSortModes } from '../utils/sort'

const emit = defineEmits(['update:modelValue'])

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const sortOptions = SORT_OPTIONS

const rootRef = ref(null)
const isOpen = ref(false)
const draftModes = ref([])

const selectedModes = computed(() => parseSortModes(props.modelValue))

const activeModes = computed(() => {
  if (isOpen.value) {
    return [...draftModes.value]
  }
  return selectedModes.value
})

const activeModeSet = computed(() => new Set(activeModes.value))

const triggerLabel = computed(() => {
  const selected = activeModes.value
  if (selected.length === 0) {
    return 'Default'
  }
  return selected.map((mode) => SORT_LABELS[mode]).join(' + ')
})

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) {
      closeDropdown(false)
    }
  },
)

function toggleDropdown() {
  if (props.disabled) {
    return
  }

  if (isOpen.value) {
    closeDropdown(true)
    return
  }

  openDropdown()
}

function openDropdown() {
  draftModes.value = [...selectedModes.value]
  isOpen.value = true
}

function closeDropdown(commit = true) {
  if (!isOpen.value) {
    return
  }

  if (commit) {
    commitDraft()
  }
  isOpen.value = false
}

function isSelected(mode) {
  return activeModeSet.value.has(mode)
}

function commitDraft() {
  const normalized = draftModes.value.join(',')
  if (normalized === String(props.modelValue ?? '')) {
    return
  }
  emit('update:modelValue', normalized)
}

function toggleMode(mode) {
  const next = [...draftModes.value]
  const existingIndex = next.indexOf(mode)

  if (existingIndex >= 0) {
    next.splice(existingIndex, 1)
    draftModes.value = next
    return
  }

  if (mode === 'price_asc') {
    const descIndex = next.indexOf('price_desc')
    if (descIndex >= 0) {
      next.splice(descIndex, 1)
    }
  }

  if (mode === 'price_desc') {
    const ascIndex = next.indexOf('price_asc')
    if (ascIndex >= 0) {
      next.splice(ascIndex, 1)
    }
  }

  next.push(mode)
  draftModes.value = next
}

function handleDocumentPointerDown(event) {
  if (!isOpen.value || !rootRef.value) {
    return
  }
  if (!(event.target instanceof Node)) {
    return
  }
  if (!rootRef.value.contains(event.target)) {
    closeDropdown(true)
  }
}

function handleDocumentKeydown(event) {
  if (!isOpen.value) {
    return
  }
  if (event.key === 'Escape') {
    event.preventDefault()
    closeDropdown(true)
  }
}

onMounted(() => {
  if (typeof document === 'undefined') {
    return
  }
  document.addEventListener('mousedown', handleDocumentPointerDown)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  if (typeof document === 'undefined') {
    return
  }
  document.removeEventListener('mousedown', handleDocumentPointerDown)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>
