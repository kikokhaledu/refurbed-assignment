<template>
  <section data-testid="product-grid" class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
    <template v-if="loading">
      <article
        v-for="skeleton in skeletonCount"
        :key="`skeleton-${skeleton}`"
        data-testid="product-skeleton"
        class="overflow-hidden rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900"
        aria-hidden="true"
      >
        <div class="h-64 animate-pulse rounded-xl bg-slate-200 dark:bg-slate-700" />
        <div class="mt-4 h-7 w-3/4 animate-pulse rounded bg-slate-200 dark:bg-slate-700" />
        <div class="mt-4 h-5 w-1/2 animate-pulse rounded bg-slate-200 dark:bg-slate-700" />
        <div class="mt-4 h-8 w-2/3 animate-pulse rounded bg-slate-200 dark:bg-slate-700" />
      </article>
    </template>

    <ProductCard
      v-for="product in products"
      v-else
      :key="product.id"
      :product="product"
    />
  </section>
</template>

<script setup>
import ProductCard from './ProductCard.vue'

defineProps({
  products: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
  skeletonCount: {
    type: Number,
    default: 6,
  },
})
</script>
