/**
 * v-magnetic — makes a button subtly follow the cursor (Framer-style).
 * The element shifts toward the cursor by ~18% of the offset from its center.
 * Resets with a spring easing on mouseleave.
 */

interface MagneticEl extends HTMLElement {
  __magnetic__?: { onMove: (e: MouseEvent) => void; onLeave: () => void }
}

export default {
  mounted(el: MagneticEl) {
    let rafId = 0

    const onMove = (e: MouseEvent) => {
      cancelAnimationFrame(rafId)
      rafId = requestAnimationFrame(() => {
        const r = el.getBoundingClientRect()
        const x = ((e.clientX - r.left) - r.width  / 2) * 0.18
        const y = ((e.clientY - r.top)  - r.height / 2) * 0.18
        el.style.transform = `translate(${x}px, ${y}px)`
        el.style.transition = 'transform 60ms linear'
      })
    }

    const onLeave = () => {
      cancelAnimationFrame(rafId)
      el.style.transition = 'transform 400ms cubic-bezier(0.34, 1.56, 0.64, 1)'
      el.style.transform = ''
    }

    el.addEventListener('mousemove', onMove)
    el.addEventListener('mouseleave', onLeave)
    el.__magnetic__ = { onMove, onLeave }
  },

  unmounted(el: MagneticEl) {
    const h = el.__magnetic__
    if (h) {
      el.removeEventListener('mousemove', h.onMove)
      el.removeEventListener('mouseleave', h.onLeave)
    }
  },
}
