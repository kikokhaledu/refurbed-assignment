<template>
  <div>
    <label for="product-search" class="sr-only">Search products</label>
    <div class="relative">
      <svg
        class="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400"
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
      >
        <path
          d="M10.5 18a7.5 7.5 0 1 1 5.853-12.19A7.5 7.5 0 0 1 10.5 18Zm0 0L21 21"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
      <input
        id="product-search"
        data-testid="search-input"
        :value="modelValue"
        :disabled="disabled"
        type="search"
        autocomplete="off"
        placeholder="Search products..."
        class="h-12 w-full rounded-xl border border-slate-200 bg-slate-50 pl-12 pr-4 text-base text-slate-900 placeholder:text-slate-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-700 disabled:cursor-not-allowed disabled:opacity-70 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:placeholder:text-slate-500"
        @input="emitUpdate"
      />
    </div>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])

function emitUpdate(event) {
  emit('update:modelValue', event.target.value)
}
</script>
