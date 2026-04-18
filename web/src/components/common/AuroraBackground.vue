<!--
  AuroraBackground — fixed full-viewport backdrop with 3-4 slowly drifting
  blurred color orbs + subtle noise grain. Pointer-events: none so it never
  intercepts clicks. Render once at app root.
-->
<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import type { Ref } from 'vue'

const props = withDefaults(defineProps<{
  /** Intensity multiplier: 'subtle' | 'normal' | 'bold' */
  intensity?: 'subtle' | 'normal' | 'bold'
  /** When true, render above page background-glow but below all content */
  absolute?: boolean
}>(), {
  intensity: 'normal',
  absolute: false,
})

const isDark = inject<Ref<boolean>>('isDark', ref(true))

const opacityScale = computed(() => {
  switch (props.intensity) {
    case 'subtle': return 0.6
    case 'bold':   return 1.25
    default:       return 1
  }
})
</script>

<template>
  <div
    class="aurora-bg"
    :class="{ 'aurora-bg--absolute': absolute, 'aurora-bg--light': !isDark }"
    :style="{ '--aurora-scale': opacityScale }"
    aria-hidden="true"
  >
    <div class="aurora-orb aurora-orb--1" />
    <div class="aurora-orb aurora-orb--2" />
    <div class="aurora-orb aurora-orb--3" />
    <div class="aurora-orb aurora-orb--4" />
    <div class="aurora-grain" />
    <div class="aurora-fade" />
  </div>
</template>

<style scoped>
.aurora-bg {
  position: fixed;
  inset: 0;
  z-index: -2;
  pointer-events: none;
  overflow: hidden;
  contain: strict;
}
.aurora-bg--absolute {
  position: absolute;
  z-index: 0;
}

.aurora-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(var(--sre-aurora-blur));
  opacity: calc(var(--sre-aurora-opacity) * var(--aurora-scale, 1));
  mix-blend-mode: screen;
  will-change: transform;
}

.aurora-bg--light .aurora-orb {
  mix-blend-mode: multiply;
}

.aurora-orb--1 {
  width: 560px; height: 560px;
  left: -8%; top: -12%;
  background: radial-gradient(circle, var(--sre-aurora-1), transparent 70%);
  animation: sre-aurora-drift 22s ease-in-out infinite;
}

.aurora-orb--2 {
  width: 640px; height: 640px;
  right: -10%; top: -6%;
  background: radial-gradient(circle, var(--sre-aurora-2), transparent 70%);
  animation: sre-aurora-drift 28s ease-in-out infinite reverse;
  animation-delay: -4s;
}

.aurora-orb--3 {
  width: 520px; height: 520px;
  left: 25%; bottom: -18%;
  background: radial-gradient(circle, var(--sre-aurora-3), transparent 70%);
  animation: sre-aurora-drift 32s ease-in-out infinite;
  animation-delay: -12s;
  opacity: calc(var(--sre-aurora-opacity) * var(--aurora-scale, 1) * 0.85);
}

.aurora-orb--4 {
  width: 420px; height: 420px;
  right: 15%; bottom: -8%;
  background: radial-gradient(circle, var(--sre-aurora-4), transparent 70%);
  animation: sre-aurora-drift 26s ease-in-out infinite reverse;
  animation-delay: -18s;
  opacity: calc(var(--sre-aurora-opacity) * var(--aurora-scale, 1) * 0.6);
}

.aurora-grain {
  position: absolute;
  inset: 0;
  background-image: var(--sre-noise-url);
  opacity: 0.035;
  mix-blend-mode: overlay;
}
.aurora-bg--light .aurora-grain { opacity: 0.02; }

/* Fade bottom half toward page background so content stays readable */
.aurora-fade {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    180deg,
    transparent 0%,
    transparent 45%,
    var(--sre-bg-page) 92%
  );
}

@media (prefers-reduced-motion: reduce) {
  .aurora-orb { animation: none !important; }
}
</style>
