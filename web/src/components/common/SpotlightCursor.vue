<script setup lang="ts">
import { ref, onMounted, onUnmounted, inject } from 'vue'
import type { Ref } from 'vue'

const isDark = inject<Ref<boolean>>('isDark', ref(true))

const x = ref(-9999)
const y = ref(-9999)
const visible = ref(false)
let rafId = 0

function onMouseMove(e: MouseEvent) {
  cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(() => {
    x.value = e.clientX
    y.value = e.clientY
    visible.value = true
  })
}

function onMouseLeave() {
  visible.value = false
}

const mq = window.matchMedia('(prefers-reduced-motion: reduce)')

onMounted(() => {
  if (mq.matches) return
  window.addEventListener('mousemove', onMouseMove, { passive: true })
  document.addEventListener('mouseleave', onMouseLeave)
})

onUnmounted(() => {
  window.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseleave', onMouseLeave)
  cancelAnimationFrame(rafId)
})
</script>

<template>
  <div
    v-if="isDark && visible"
    class="spotlight"
    :style="{ '--mx': x + 'px', '--my': y + 'px' }"
    aria-hidden="true"
  />
</template>

<style scoped>
.spotlight {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background: radial-gradient(
    circle var(--sre-spotlight-size) at var(--mx) var(--my),
    color-mix(in srgb, var(--sre-brand-accent) calc(var(--sre-spotlight-opacity) * 100%), transparent) 0%,
    transparent 60%
  );
  transition: opacity 0.4s ease;
}
</style>
