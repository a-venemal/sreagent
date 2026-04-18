<!--
  SkeletonCard — full-card loading placeholder with real shimmer sweep.
  Usage: <SkeletonCard :rows="3" :show-avatar="true" />
-->
<script setup lang="ts">
withDefaults(defineProps<{
  rows?: number
  showAvatar?: boolean
  height?: string
}>(), {
  rows: 3,
  showAvatar: false,
  height: 'auto',
})
</script>

<template>
  <div class="sk-card" :style="{ minHeight: height }">
    <div class="sk-shimmer-overlay" />
    <div class="sk-body">
      <div v-if="showAvatar" class="sk-avatar" />
      <div class="sk-content">
        <div class="sk-row sk-row--title" />
        <div v-for="i in rows" :key="i" class="sk-row" :style="{ width: `${75 - i * 8}%` }" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.sk-card {
  background: var(--sre-bg-card);
  border: 1px solid var(--sre-border);
  border-radius: var(--sre-radius-lg);
  padding: var(--sre-space-5);
  position: relative;
  overflow: hidden;
}

/* Shimmer sweep overlay */
.sk-shimmer-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255,255,255,0.045) 40%,
    rgba(255,255,255,0.08) 50%,
    rgba(255,255,255,0.045) 60%,
    transparent 100%
  );
  background-size: 200% 100%;
  animation: sk-sweep 1.8s linear infinite;
}
body.light-theme .sk-shimmer-overlay {
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(0,0,0,0.03) 40%,
    rgba(0,0,0,0.06) 50%,
    rgba(0,0,0,0.03) 60%,
    transparent 100%
  );
  background-size: 200% 100%;
}

@keyframes sk-sweep {
  0%   { background-position: -100% 0; }
  100% { background-position:  200% 0; }
}

.sk-body {
  display: flex;
  gap: var(--sre-space-4);
  position: relative;
  z-index: 1;
}

.sk-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--sre-bg-elevated);
  flex-shrink: 0;
}

.sk-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--sre-space-3);
}

.sk-row {
  height: 12px;
  border-radius: var(--sre-radius-sm);
  background: var(--sre-bg-elevated);
  width: 100%;
}
.sk-row--title {
  height: 16px;
  width: 60%;
  margin-bottom: 4px;
}

@media (prefers-reduced-motion: reduce) {
  .sk-shimmer-overlay { animation: none; }
}
</style>
