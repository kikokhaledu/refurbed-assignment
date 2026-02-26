<template>
  <div v-if="show" class="mt-10 flex justify-center">
    <button
      type="button"
      data-testid="load-more-button"
      :disabled="disabled || loading"
      class="min-w-[260px] rounded-xl border border-slate-300 bg-white px-6 py-3 text-lg font-semibold text-slate-800 transition hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 disabled:cursor-not-allowed disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800"
      @click="$emit('click')"
    >
      <span v-if="loading">Loading...</span>
      <span v-else>Load more</span>
    </button>
  </div>
</template>

<script setup>
defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['click'])
</script>
