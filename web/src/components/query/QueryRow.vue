<script setup lang="ts">
import { NInput, NButton, NSpace } from 'naive-ui'
import PromQLEditor from './PromQLEditor.vue'
import type { QueryTarget } from '@/types/query'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  target: QueryTarget
  index: number
  canRemove: boolean
}>()

const emit = defineEmits<{
  (e: 'update', id: string, patch: Partial<QueryTarget>): void
  (e: 'remove', id: string): void
  (e: 'toggle', id: string): void
  (e: 'execute', id: string): void
}>()

function onExprUpdate(id: string, value: string) {
  emit('update', id, { expression: value })
}

function onLegendUpdate(id: string, value: string) {
  emit('update', id, { legendFormat: value })
}
</script>

<template>
  <div class="query-row" :class="{ disabled: !target.enabled }">
    <div class="query-row-header">
      <span class="query-label">{{ t('explore.queryLabel', { n: String.fromCharCode(65 + index) }) }}</span>
      <NSpace size="small">
        <NButton
          quaternary
          size="tiny"
          :type="target.enabled ? 'primary' : 'default'"
          @click="emit('toggle', target.id)"
        >
          {{ target.enabled ? t('explore.toggleOn') : t('explore.toggleOff') }}
        </NButton>
        <NButton
          v-if="canRemove"
          quaternary
          size="tiny"
          type="error"
          @click="emit('remove', target.id)"
        >
          &times;
        </NButton>
      </NSpace>
    </div>

    <div class="query-row-body">
      <div class="editor-wrapper">
        <PromQLEditor
          :model-value="target.expression"
          :datasource-id="target.datasourceId"
          :placeholder="t('explore.enterExpression')"
          @update:model-value="(v: string) => onExprUpdate(target.id, v)"
          @execute="emit('execute', target.id)"
        />
      </div>

      <NInput
        :value="target.legendFormat"
        :placeholder="t('explore.legendFormat')"
        size="small"
        style="width: 200px; flex-shrink: 0"
        @update:value="(v: string) => onLegendUpdate(target.id, v)"
      />
    </div>

    <div v-if="target.state === 'error' && target.error" class="query-error">
      {{ target.error }}
    </div>
  </div>
</template>

<style scoped>
.query-row {
  border: 1px solid var(--sre-border);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  background: var(--sre-bg-sunken);
  transition: opacity 0.2s;
}
.query-row.disabled {
  opacity: 0.5;
}
.query-row-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.query-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--sre-text-secondary);
  text-transform: uppercase;
}
.query-row-body {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}
.editor-wrapper {
  flex: 1;
  min-width: 0;
}
.query-error {
  margin-top: 8px;
  padding: 8px;
  background: var(--sre-danger-soft, rgba(207, 19, 34, 0.08));
  border: 1px solid var(--sre-danger-ring, rgba(207, 19, 34, 0.2));
  border-radius: 4px;
  color: var(--sre-danger);
  font-size: 12px;
}
</style>
