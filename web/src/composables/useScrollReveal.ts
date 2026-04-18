/**
 * useScrollReveal — IntersectionObserver-based reveal for elements.
 *
 * Usage:
 *   const { revealRef } = useScrollReveal()
 *   <div ref="revealRef" data-reveal="out"> ... </div>
 *
 * Or observe multiple els via the returned `observe(el)` helper.
 *
 * CSS (in global.css or component):
 *   [data-reveal="out"] { opacity:0; transform:translateY(24px); }
 *   [data-reveal="in"]  { opacity:1; transform:translateY(0);
 *                         transition: opacity 500ms ease-out, transform 500ms cubic-bezier(0.34,1.56,0.64,1); }
 */
import { ref, onMounted, onUnmounted } from 'vue'

export function useScrollReveal(options?: IntersectionObserverInit) {
  const revealRef = ref<HTMLElement | null>(null)
  let observer: IntersectionObserver | null = null

  const defaultOpts: IntersectionObserverInit = {
    threshold: 0.12,
    rootMargin: '0px 0px -40px 0px',
    ...options,
  }

  function callback(entries: IntersectionObserverEntry[]) {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        entry.target.setAttribute('data-reveal', 'in')
        // Don't un-observe — let it stay "in" once visible
      }
    }
  }

  function observe(el: HTMLElement | null) {
    if (!el || !observer) return
    el.setAttribute('data-reveal', 'out')
    observer.observe(el)
  }

  onMounted(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    observer = new IntersectionObserver(callback, defaultOpts)
    if (revealRef.value) observe(revealRef.value)
  })

  onUnmounted(() => observer?.disconnect())

  return { revealRef, observe }
}
