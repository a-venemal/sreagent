<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  variant?: 'default' | 'critical' | 'success' | 'accent'
  interactive?: boolean
  tilt?: boolean
  glow?: boolean
  conic?: boolean | 'critical' | 'strong'
  padding?: string
}>(), {
  variant: 'default',
  interactive: false,
  tilt: false,
  glow: false,
  conic: false,
  padding: 'var(--sre-space-6)',
})

const cardRef = ref<HTMLElement | null>(null)
let rafId = 0

const classes = computed(() => [
  'glow-card',
  'surface-glass-strong',
  'noise-overlay',
  props.interactive && 'glow-card--interactive',
  props.tilt && 'tilt',
  props.glow && `glow-${props.variant}`,
  props.conic === true && 'conic-border',
  props.conic === 'critical' && 'conic-border conic-border--critical',
  props.conic === 'strong' && 'conic-border conic-border--strong',
])

function onMouseMove(e: MouseEvent) {
  if (!props.tilt || !cardRef.value) return
  cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(() => {
    const el = cardRef.value!
    const rect = el.getBoundingClientRect()
    const cx = rect.left + rect.width / 2
    const cy = rect.top + rect.height / 2
    const dx = (e.clientX - cx) / (rect.width / 2)
    const dy = (e.clientY - cy) / (rect.height / 2)
    // max 6 degrees
    el.style.setProperty('--tilt-x', String(dx * 6))
    el.style.setProperty('--tilt-y', String(-dy * 6))
  })
}

function onMouseLeave() {
  if (!props.tilt || !cardRef.value) return
  cancelAnimationFrame(rafId)
  const el = cardRef.value
  el.style.setProperty('--tilt-x', '0')
  el.style.setProperty('--tilt-y', '0')
}

onUnmounted(() => cancelAnimationFrame(rafId))
</script>

<template>
  <div
    ref="cardRef"
    :class="classes"
    :style="{ padding }"
    @mousemove.passive="onMouseMove"
    @mouseleave="onMouseLeave"
  >
    <slot />
  </div>
</template>

<style scoped>
.glow-card {
  border-radius: var(--sre-radius-lg);
  position: relative;
  transition:
    box-shadow var(--sre-duration-base) var(--sre-ease-out),
    border-color var(--sre-duration-base) var(--sre-ease-out),
    transform 220ms var(--sre-ease-out);
  overflow: visible;
}

.glow-card--interactive {
  cursor: pointer;
}

.glow-card--interactive:hover {
  border-color: var(--sre-border-strong);
}

/* Reset tilt on leave with spring feel */
.glow-card:not(:hover).tilt {
  transition:
    box-shadow var(--sre-duration-base) var(--sre-ease-out),
    border-color var(--sre-duration-base) var(--sre-ease-out),
    transform 400ms cubic-bezier(0.34, 1.56, 0.64, 1);
}
</style>
