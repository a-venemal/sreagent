<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = withDefaults(defineProps<{
  value: number
  duration?: number
  decimals?: number
  prefix?: string
  suffix?: string
}>(), {
  duration: 900,
  decimals: 0,
})

const displayed = ref(0)
let startVal = 0
let startTime = 0
let rafId = 0

function easeOut(t: number): number {
  return 1 - Math.pow(1 - t, 3)
}

function animate(timestamp: number) {
  if (!startTime) startTime = timestamp
  const elapsed = timestamp - startTime
  const progress = Math.min(elapsed / props.duration, 1)
  displayed.value = startVal + (props.value - startVal) * easeOut(progress)
  if (progress < 1) {
    rafId = requestAnimationFrame(animate)
  } else {
    displayed.value = props.value
  }
}

function runAnimation(from: number) {
  cancelAnimationFrame(rafId)
  startVal = from
  startTime = 0
  rafId = requestAnimationFrame(animate)
}

onMounted(() => {
  runAnimation(0)
})

watch(() => props.value, (newVal, oldVal) => {
  runAnimation(oldVal ?? 0)
})

const formatted = () => {
  const n = displayed.value.toFixed(props.decimals)
  return (props.prefix ?? '') + n + (props.suffix ?? '')
}
</script>

<template>
  <span class="number-display animated-number">{{ formatted() }}</span>
</template>

<style scoped>
.animated-number {
  display: inline-block;
  font-variant-numeric: tabular-nums;
  transition: none;
}
</style>
