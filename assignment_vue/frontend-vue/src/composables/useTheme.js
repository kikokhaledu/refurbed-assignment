import { ref } from 'vue'

export function useTheme(themeStorageKey = 'refurbed-theme') {
  const isDarkMode = ref(false)

  function applyTheme(nextDarkMode) {
    isDarkMode.value = nextDarkMode
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('dark', nextDarkMode)
    }
    if (typeof window !== 'undefined') {
      try {
        window.localStorage.setItem(themeStorageKey, nextDarkMode ? 'dark' : 'light')
      } catch {
        // Ignore localStorage access errors.
      }
    }
  }

  function initializeTheme() {
    if (typeof window === 'undefined') {
      applyTheme(false)
      return
    }

    let nextDarkMode = false
    try {
      const storedTheme = window.localStorage.getItem(themeStorageKey)
      if (storedTheme === 'dark') {
        nextDarkMode = true
      } else if (storedTheme === 'light') {
        nextDarkMode = false
      } else {
        nextDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
      }
    } catch {
      nextDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    applyTheme(nextDarkMode)
  }

  function toggleDarkMode() {
    applyTheme(!isDarkMode.value)
  }

  return {
    isDarkMode,
    initializeTheme,
    toggleDarkMode,
    applyTheme,
  }
}
