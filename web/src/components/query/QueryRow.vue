<script setup lang="ts">
import { NSelect, NInput, NButton, NIcon, NSpace } from 'naive-ui'
import PromQLEditor from './PromQLEditor.vue'
import type { QueryTarget, QuerySeriesItem } from '@/types/query'
import type { DataSource } from '@/types'

defineProps<{
  target: QueryTarget
  datasources: DataSource[]
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

function onDsUpdate(id: string, value: number) {
  emit('update', id, { datasourceId: value })
}
</script>

<template>
  <div class="query-row" :class="{ disabled: !target.enabled }">
    <div class="query-row-header">
      <span class="query-label">Query {{ String.fromCharCode(65 + index) }}</span>
      <NSpace size="small">
        <NButton
          quaternary
          size="tiny"
          :type="target.enabled ? 'primary' : 'default'"
          @click="emit('toggle', target.id)"
        >
          {{ target.enabled ? 'A' : 'H' }}
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
      <NSelect
        :value="target.datasourceId"
        :options="datasources.map(ds => ({ label: `${ds.name} (${ds.type})`, value: ds.id }))"
        placeholder="Select datasource"
        filterable
        size="small"
        style="width: 240px; flex-shrink: 0"
        @update:value="(v: number) => onDsUpdate(target.id, v)"
      />

      <div class="editor-wrapper">
        <PromQLEditor
          :model-value="target.expression"
          :datasource-id="target.datasourceId"
          placeholder="Enter PromQL expression... (Ctrl+Enter to execute)"
          @update:model-value="(v: string) => onExprUpdate(target.id, v)"
          @execute="emit('execute', target.id)"
        />
      </div>

      <NInput
        :value="target.legendFormat"
        placeholder="Legend: {{instance}}"
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
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  background: #fafafa;
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
  color: #666;
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
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 4px;
  color: #cf1322;
  font-size: 12px;
}
</style>
