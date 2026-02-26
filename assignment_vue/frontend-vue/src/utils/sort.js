import { SORT_LABELS, SORT_MODES } from '../constants/sort'

const allowedSortModes = new Set(SORT_MODES)

function toRawEntries(rawValue) {
  if (Array.isArray(rawValue)) {
    return rawValue
  }
  return [rawValue]
}

export function parseSortModes(rawValue, options = {}) {
  const strict = options.strict === true
  const seen = new Set()
  const modes = []

  for (const rawEntry of toRawEntries(rawValue)) {
    for (const part of String(rawEntry ?? '').split(',')) {
      const mode = part.trim().toLowerCase()
      if (!mode) {
        continue
      }
      if (!allowedSortModes.has(mode)) {
        if (strict) {
          return []
        }
        continue
      }
      if (seen.has(mode)) {
        continue
      }
      seen.add(mode)
      modes.push(mode)
    }
  }

  return modes
}

export function canonicalizeSortModes(modes) {
  if (!Array.isArray(modes) || modes.length === 0) {
    return []
  }

  const canonical = []
  let selectedPriceMode = ''
  for (const mode of modes) {
    if (mode === 'price_asc' || mode === 'price_desc') {
      if (selectedPriceMode) {
        continue
      }
      selectedPriceMode = mode
      canonical.push(mode)
      continue
    }
    canonical.push(mode)
  }

  return canonical
}

export function normalizeSortValue(rawValue, options = {}) {
  const modes = parseSortModes(rawValue, options)
  return canonicalizeSortModes(modes).join(',')
}

export function sortValueLabel(rawValue, emptyLabel = 'Default') {
  const canonical = canonicalizeSortModes(parseSortModes(rawValue))
  if (canonical.length === 0) {
    return emptyLabel
  }

  return canonical.map((mode) => SORT_LABELS[mode] ?? mode).join(' then ')
}
