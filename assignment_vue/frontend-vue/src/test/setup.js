class NoopIntersectionObserver {
  constructor() {}

  observe() {}

  unobserve() {}

  disconnect() {}
}

if (typeof window !== 'undefined' && !window.IntersectionObserver) {
  window.IntersectionObserver = NoopIntersectionObserver
}

if (typeof globalThis !== 'undefined' && !globalThis.IntersectionObserver) {
  globalThis.IntersectionObserver = NoopIntersectionObserver
}
