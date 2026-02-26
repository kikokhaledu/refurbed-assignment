export const SORT_OPTIONS = [
  { value: 'popularity', label: 'Popularity' },
  { value: 'price_asc', label: 'Price: Low to High' },
  { value: 'price_desc', label: 'Price: High to Low' },
]

export const SORT_LABELS = Object.freeze(
  SORT_OPTIONS.reduce((labels, option) => {
    labels[option.value] = option.label
    return labels
  }, {}),
)

export const SORT_MODES = Object.freeze(SORT_OPTIONS.map((option) => option.value))
