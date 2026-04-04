<template>
  <div class="page-header" :class="{ 'page-header--flex': hasActions }">
    <div>
      <h2 class="page-title">{{ title }}</h2>
      <n-text v-if="subtitle" depth="3">{{ subtitle }}</n-text>
    </div>
    <div v-if="hasActions" class="header-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSlots } from 'vue'
import { NText } from 'naive-ui'

defineProps<{
  title: string
  subtitle?: string
}>()

const slots = useSlots()
const hasActions = !!slots.actions
</script>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.page-header--flex {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: var(--sre-text-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
