<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NTag } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { auditLogApi } from '@/api'
import type { AuditLog } from '@/types'
import { formatTime } from '@/utils/format'

const { t } = useI18n()
const loading = ref(false)
const logs = ref<AuditLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

// Filters
const filterAction = ref<string | null>(null)
const filterResourceType = ref<string | null>(null)
const filterDateRange = ref<[number, number] | null>(null)

const actionOptions = [
  { label: 'create', value: 'create' },
  { label: 'update', value: 'update' },
  { label: 'delete', value: 'delete' },
  { label: 'toggle', value: 'toggle' },
  { label: 'acknowledge', value: 'acknowledge' },
  { label: 'assign', value: 'assign' },
  { label: 'resolve', value: 'resolve' },
  { label: 'close', value: 'close' },
  { label: 'silence', value: 'silence' },
]

const resourceTypeOptions = [
  { label: 'alert_rule', value: 'alert_rule' },
  { label: 'alert_event', value: 'alert_event' },
  { label: 'user', value: 'user' },
  { label: 'team', value: 'team' },
  { label: 'datasource', value: 'datasource' },
]

const columns = [
  {
    title: () => t('settings.auditTime'),
    key: 'created_at',
    width: 180,
    render: (row: AuditLog) => h('span', { style: 'font-size: 12px' }, formatTime(row.created_at)),
  },
  {
    title: () => t('settings.auditUser'),
    key: 'username',
    width: 120,
  },
  {
    title: () => t('settings.auditAction'),
    key: 'action',
    width: 110,
    render: (row: AuditLog) =>
      h(NTag, { size: 'small', type: row.action === 'delete' ? 'error' : row.action === 'create' ? 'success' : 'info' }, { default: () => row.action }),
  },
  {
    title: () => t('settings.auditResource'),
    key: 'resource_type',
    width: 130,
    render: (row: AuditLog) =>
      h('div', [
        h('div', { style: 'font-weight: 500' }, row.resource_type),
        row.resource_name ? h('div', { style: 'font-size: 11px; color: var(--sre-text-secondary)' }, row.resource_name) : null,
      ]),
  },
  {
    title: () => t('settings.auditIP'),
    key: 'ip',
    width: 130,
  },
  {
    title: () => t('common.status'),
    key: 'status',
    width: 90,
    render: (row: AuditLog) =>
      h(NTag, { size: 'small', type: row.status === 'success' ? 'success' : 'error' }, { default: () => row.status }),
  },
  {
    title: () => t('settings.auditDetail'),
    key: 'detail',
    ellipsis: { tooltip: true },
  },
]

async function fetchLogs() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (filterAction.value) params.action = filterAction.value
    if (filterResourceType.value) params.resource_type = filterResourceType.value
    if (filterDateRange.value) {
      params.start_time = new Date(filterDateRange.value[0]).toISOString()
      params.end_time = new Date(filterDateRange.value[1]).toISOString()
    }
    const { data } = await auditLogApi.list(params)
    logs.value = data.data.list || []
    total.value = data.data.total
  } catch {
    // silently fail
  } finally {
    loading.value = false
  }
}

function handleFilterChange() {
  page.value = 1
  fetchLogs()
}

onMounted(() => {
  fetchLogs()
})
</script>

<template>
  <div class="audit-log">
    <!-- Filters -->
    <div class="audit-filters">
      <n-select
        v-model:value="filterAction"
        :options="actionOptions"
        :placeholder="t('settings.auditAction')"
        clearable
        style="width: 160px"
        @update:value="handleFilterChange"
      />
      <n-select
        v-model:value="filterResourceType"
        :options="resourceTypeOptions"
        :placeholder="t('settings.auditResource')"
        clearable
        style="width: 160px"
        @update:value="handleFilterChange"
      />
      <n-date-picker
        v-model:value="filterDateRange"
        type="daterange"
        clearable
        style="width: 280px"
        @update:value="handleFilterChange"
      />
    </div>

    <!-- Table -->
    <n-data-table
      :loading="loading"
      :columns="columns"
      :data="logs"
      :row-key="(row: AuditLog) => row.id"
      :bordered="false"
      size="small"
      :pagination="{
        page: page,
        pageSize: pageSize,
        itemCount: total,
        onChange: (p: number) => { page = p; fetchLogs() },
        onUpdatePageSize: (s: number) => { pageSize = s; page = 1; fetchLogs() },
        showSizePicker: true,
        pageSizes: [20, 50, 100],
      }"
    />

    <n-empty
      v-if="!loading && logs.length === 0"
      :description="t('settings.auditNoData')"
      style="padding: 60px 0"
    />
  </div>
</template>

<style scoped>
.audit-filters {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
</style>
