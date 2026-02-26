import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

export function useMobileFilterDialog() {
  const isMobileFiltersOpen = ref(false)
  const mobileFiltersDialogRef = ref(null)
  const mobileFiltersCloseButtonRef = ref(null)
  const mobileFiltersTriggerRef = ref(null)

  let previouslyFocusedElement = null
  let previousBodyOverflow = ''

  function getFocusableElements(container) {
    const selector = [
      'a[href]',
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
    ].join(',')

    return [...container.querySelectorAll(selector)].filter((element) => {
      return element instanceof HTMLElement && element.offsetParent !== null
    })
  }

  function focusMobileDialog() {
    if (!(mobileFiltersDialogRef.value instanceof HTMLElement)) {
      return
    }

    const focusable = getFocusableElements(mobileFiltersDialogRef.value)
    if (mobileFiltersCloseButtonRef.value instanceof HTMLElement) {
      mobileFiltersCloseButtonRef.value.focus()
      return
    }

    if (focusable.length > 0) {
      focusable[0].focus()
      return
    }

    mobileFiltersDialogRef.value.focus()
  }

  function handleMobileDialogKeydown(event) {
    if (!isMobileFiltersOpen.value) {
      return
    }

    if (event.key === 'Escape') {
      event.preventDefault()
      closeMobileFilters()
      return
    }

    if (event.key !== 'Tab') {
      return
    }

    if (!(mobileFiltersDialogRef.value instanceof HTMLElement)) {
      return
    }

    const focusable = getFocusableElements(mobileFiltersDialogRef.value)
    if (focusable.length === 0) {
      event.preventDefault()
      mobileFiltersDialogRef.value.focus()
      return
    }

    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    const active = document.activeElement

    if (event.shiftKey) {
      if (active === first || !mobileFiltersDialogRef.value.contains(active)) {
        event.preventDefault()
        last.focus()
      }
      return
    }

    if (active === last || !mobileFiltersDialogRef.value.contains(active)) {
      event.preventDefault()
      first.focus()
    }
  }

  watch(
    () => isMobileFiltersOpen.value,
    async (isOpen) => {
      if (typeof window === 'undefined' || typeof document === 'undefined') {
        return
      }

      if (isOpen) {
        previouslyFocusedElement =
          document.activeElement instanceof HTMLElement ? document.activeElement : null
        previousBodyOverflow = document.body.style.overflow
        document.body.style.overflow = 'hidden'
        window.addEventListener('keydown', handleMobileDialogKeydown)
        await nextTick()
        focusMobileDialog()
        return
      }

      document.body.style.overflow = previousBodyOverflow
      window.removeEventListener('keydown', handleMobileDialogKeydown)
      await nextTick()

      if (mobileFiltersTriggerRef.value instanceof HTMLElement) {
        mobileFiltersTriggerRef.value.focus()
      } else if (previouslyFocusedElement instanceof HTMLElement) {
        previouslyFocusedElement.focus()
      }
    },
  )

  function openMobileFilters() {
    isMobileFiltersOpen.value = true
  }

  function closeMobileFilters() {
    isMobileFiltersOpen.value = false
  }

  function toggleMobileFilters() {
    if (isMobileFiltersOpen.value) {
      closeMobileFilters()
      return
    }
    openMobileFilters()
  }

  onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleMobileDialogKeydown)
    }

    if (typeof document !== 'undefined') {
      document.body.style.overflow = previousBodyOverflow
    }
  })

  return {
    isMobileFiltersOpen,
    mobileFiltersDialogRef,
    mobileFiltersCloseButtonRef,
    mobileFiltersTriggerRef,
    openMobileFilters,
    closeMobileFilters,
    toggleMobileFilters,
  }
}
