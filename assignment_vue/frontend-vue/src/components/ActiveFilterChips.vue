<template>
  <section
    v-if="chips.length > 0"
    data-testid="active-filter-chips"
    class="mb-5"
    aria-label="Active filters"
  >
    <div class="mb-2 flex items-center justify-between gap-2 lg:hidden">
      <p data-testid="active-chip-summary-mobile" class="text-sm font-medium text-slate-600 dark:text-slate-300">
        {{ chipsSummaryLabel }}
      </p>
      <div class="flex items-center gap-2">
        <button
          type="button"
          data-testid="active-chip-toggle-mobile"
          :aria-expanded="isMobileExpanded ? 'true' : 'false'"
          aria-controls="active-chip-list"
          class="inline-flex items-center rounded-full border border-slate-300 bg-white px-3 py-1.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
          @click="isMobileExpanded = !isMobileExpanded"
        >
          {{ isMobileExpanded ? 'Hide' : 'Show' }}
        </button>
        <button
          type="button"
          data-testid="active-chip-clear-all-mobile"
          class="inline-flex items-center rounded-full border border-transparent bg-slate-900 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-slate-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:bg-slate-100 dark:text-slate-900 dark:hover:bg-slate-300"
          @click="$emit('clear-all')"
        >
          Clear all
        </button>
      </div>
    </div>

    <div
      id="active-chip-list"
      data-testid="active-chip-list"
      :class="[isMobileExpanded ? 'flex' : 'hidden', 'flex-wrap items-center gap-2 lg:flex']"
    >
      <button
        v-for="chip in chips"
        :key="chip.id"
        type="button"
        :data-testid="`active-chip-${chip.id}`"
        class="inline-flex items-center gap-2 rounded-full border border-slate-300 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
        @click="$emit('remove-chip', chip)"
      >
        <span>{{ chip.label }}</span>
        <span aria-hidden="true" class="text-xs">x</span>
        <span class="sr-only">Remove {{ chip.label }}</span>
      </button>

      <button
        type="button"
        data-testid="active-chip-clear-all"
        class="hidden items-center rounded-full border border-transparent bg-slate-900 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-slate-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 dark:bg-slate-100 dark:text-slate-900 dark:hover:bg-slate-300 lg:inline-flex"
        @click="$emit('clear-all')"
      >
        Clear all
      </button>
    </div>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  chips: {
    type: Array,
    default: () => [],
  },
})

const isMobileExpanded = ref(false)
const chipsSummaryLabel = computed(() => {
  const count = props.chips.length
  return `${count} filter${count === 1 ? '' : 's'} applied`
})

defineEmits(['remove-chip', 'clear-all'])
</script>
