export const COLOR_SWATCH_PALETTE = Object.freeze({
  black: '#111827',
  blue: '#3f5bd3',
  gray: '#4b5563',
  green: '#16a34a',
  orange: '#ea580c',
  pink: '#ec4899',
  red: '#e11d48',
  silver: '#94a3b8',
  white: '#f8fafc',
})

export const DEFAULT_SWATCH_COLOR = '#cbd5e1'

export function swatchColorForToken(token) {
  return COLOR_SWATCH_PALETTE[token] ?? DEFAULT_SWATCH_COLOR
}
