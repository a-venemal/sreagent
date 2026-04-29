<script setup lang="ts">
import { NButton, NSpace, NIcon } from 'naive-ui'
import QueryRow from './QueryRow.vue'
import type { QueryTarget } from '@/types/query'
import type { DataSource } from '@/types'

defineProps<{
  targets: QueryTarget[]
  datasources: DataSource[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'add'): void
  (e: 'remove', id: string): void
  (e: 'toggle', id: string): void
  (e: 'update', id: string, patch: Partial<QueryTarget>): void
  (e: 'execute', id: string): void
  (e: 'executeAll'): void
}>()
</script>

<template>
  <div class="query-panel">
    <QueryRow
      v-for="(target, i) in targets"
      :key="target.id"
      :target="target"
      :datasources="datasources"
      :index="i"
      :can-remove="targets.length > 1"
      @remove="(id) => emit('remove', id)"
      @toggle="(id) => emit('toggle', id)"
      @update="(id, patch) => emit('update', id, patch)"
      @execute="(id) => emit('execute', id)"
    />

    <div class="query-panel-actions">
      <NSpace>
        <NButton dashed size="small" @click="emit('add')">
          + Add Query
        </NButton>
        <NButton
          type="primary"
          size="small"
          :loading="loading"
          :disabled="!targets.some(t => t.enabled && t.datasourceId && t.expression.trim())"
          @click="emit('executeAll')"
        >
          Run Queries
        </NButton>
      </NSpace>
    </div>
  </div>
</template>

<style scoped>
.query-panel {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
}
.query-panel-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}
</style>
