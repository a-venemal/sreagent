<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import { NTabs, NTabPane, NDataTable, NEmpty, NSpin, NTag, NAlert } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { datasourceApi } from '@/api'
import type { DataSource, LogEntry } from '@/types'
import { useTimeRange } from '@/composables/useTimeRange'
import { useQueryEngine } from '@/composables/useQueryEngine'
import QueryPanel from '@/components/query/QueryPanel.vue'
import QueryResultChart from '@/components/query/QueryResultChart.vue'
import QueryResultTable from '@/components/query/QueryResultTable.vue'
import TimeRangePicker from '@/components/time/TimeRangePicker.vue'
import RefreshPicker from '@/components/time/RefreshPicker.vue'

const { t } = useI18n()
const datasources = ref<DataSource[]>([])
const selectedDsId = ref<number | null>(null)

const {
  timeRange,
  isRelative,
  relativeDuration,
  autoRefreshInterval,
  setRelative,
  setAbsolute,
} = useTimeRange('1h')

const {
  targets,
  globalLoading,
  addTarget,
  removeTarget,
  toggleTarget,
  updateTarget,
  executeAll,
  executeQuery,
} = useQueryEngine(timeRange)

const resultTab = ref('chart')

// --- Derived state ---
const selectedDs = computed(() =>
  datasources.value.find(ds => ds.id === selectedDsId.value) || null
)

const isLogsMode = computed(() => selectedDs.value?.type === 'victorialogs')

const hasResults = computed(() =>
  targets.value.some(t => t.series && t.series.length > 0)
)

// --- Log mode state ---
const expression = ref('')
const limit = ref(200)
const logLoading = ref(false)
const logEntries = ref<LogEntry[]>([])
const truncated = ref(false)
const logError = ref('')

const logColumns: DataTableColumns<LogEntry> = [
  {
    title: t('explore.logTime') || 'Time',
    key: 'timestamp',
    width: 200,
    render(row) {
      const ts = row.timestamp
      if (!ts) return '-'
      try { return new Date(ts).toLocaleString() } catch { return ts }
    },
  },
  {
    title: t('explore.logMessage') || 'Message',
    key: 'message',
    ellipsis: { tooltip: true },
    render(row) { return row.message || '-' },
  },
  {
    title: t('explore.logLabels') || 'Labels',
    key: 'labels',
    width: 400,
    render(row) {
      const labels = row.labels
      if (!labels || Object.keys(labels).length === 0) return '-'
      return Object.entries(labels).slice(0, 5).map(([k, v]) =>
        h(NTag, { size: 'small', bordered: false, style: 'margin: 2px' }, () => `${k}=${v}`)
      )
    },
  },
]

// --- Sync top-level datasource to all targets ---
watch(selectedDsId, (newId) => {
  if (newId && !isLogsMode.value) {
    // Update all existing targets to use the selected datasource
    for (const target of targets.value) {
      if (target.datasourceId !== newId) {
        updateTarget(target.id, { datasourceId: newId })
      }
    }
  }
  // Reset log state when switching
  expression.value = ''
  logEntries.value = []
  logError.value = ''
  truncated.value = false
})

// --- Fetch datasources ---
async function fetchDatasources() {
  try {
    const res = await datasourceApi.list({ page: 1, page_size: 100 })
    datasources.value = (res.data.data.list || []).filter((ds: any) => ds.is_enabled)
    if (datasources.value.length > 0 && !selectedDsId.value) {
      selectedDsId.value = datasources.value[0].id
    }
  } catch {
    // ignore
  }
}

// --- Metrics mode ---
function handleExecuteSingle(id: string) {
  const target = targets.value.find(t => t.id === id)
  if (target) executeQuery(target)
}

function handleAddTarget() {
  addTarget()
  // Sync datasource to the new target
  if (selectedDsId.value) {
    const last = targets.value[targets.value.length - 1]
    if (last && !last.datasourceId) {
      updateTarget(last.id, { datasourceId: selectedDsId.value })
    }
  }
}

// --- Logs mode ---
async function executeLogQuery() {
  if (!selectedDsId.value || !expression.value.trim()) return

  logLoading.value = true
  logError.value = ''
  logEntries.value = []
  truncated.value = false

  try {
    const tr = timeRange.value
    const res = await datasourceApi.logQuery(selectedDsId.value, {
      expression: expression.value,
      start: Math.floor(tr.start / 1000),
      end: Math.floor(tr.end / 1000),
      limit: limit.value,
    })
    const data = res.data.data
    logEntries.value = data.entries || []
    truncated.value = data.truncated || false
  } catch (err: any) {
    logError.value = err?.message || t('explore.queryFailed')
  } finally {
    logLoading.value = false
  }
}

function handleLogKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
    executeLogQuery()
  }
}

onMounted(fetchDatasources)
</script>

<template>
  <div class="explore-page">
    <!-- Header -->
    <div class="explore-header">
      <div class="header-left">
        <h2 class="page-title">{{ t('explore.title') }}</h2>
        <span class="page-subtitle">{{ t('explore.subtitle') }}</span>
      </div>
      <div class="header-right">
        <n-select
          v-model:value="selectedDsId"
          :options="datasources.map(ds => ({ label: `${ds.name} (${ds.type})`, value: ds.id }))"
          :placeholder="t('explore.selectDatasource')"
          filterable
          style="width: 280px"
          size="small"
        />
        <TimeRangePicker
          :time-range="timeRange"
          :is-relative="isRelative"
          :relative-duration="relativeDuration"
          @set-relative="setRelative"
          @set-absolute="setAbsolute"
        />
        <RefreshPicker
          :value="autoRefreshInterval"
          @update:value="(v: number | null) => autoRefreshInterval = v"
        />
      </div>
    </div>

    <!-- No datasource selected -->
    <div v-if="!selectedDsId" class="empty-state">
      <n-empty :description="t('explore.selectDatasource')" />
    </div>

    <!-- METRICS MODE -->
    <template v-if="selectedDsId && !isLogsMode">
      <QueryPanel
        :targets="targets"
        :loading="globalLoading"
        @add="handleAddTarget"
        @remove="removeTarget"
        @toggle="toggleTarget"
        @update="updateTarget"
        @execute="handleExecuteSingle"
        @execute-all="executeAll"
      />

      <div v-if="hasResults" class="explore-results">
        <NTabs v-model:value="resultTab" type="line" size="small">
          <NTabPane name="chart" :tab="t('explore.chart')">
            <QueryResultChart
              :targets="targets"
              :time-range="timeRange"
              :height="400"
            />
          </NTabPane>
          <NTabPane name="table" :tab="t('explore.table')">
            <QueryResultTable :targets="targets" />
          </NTabPane>
        </NTabs>
      </div>
    </template>

    <!-- LOGS MODE -->
    <template v-if="selectedDsId && isLogsMode">
      <div class="query-bar">
        <n-input
          v-model:value="expression"
          :placeholder="t('explore.logQueryPlaceholder')"
          size="small"
          style="flex: 1"
          @keydown="handleLogKeydown"
        />
        <n-input-number
          v-model:value="limit"
          :min="10"
          :max="10000"
          size="small"
          style="width: 120px"
          :placeholder="t('explore.limit')"
        />
        <n-button
          type="primary"
          size="small"
          :loading="logLoading"
          :disabled="!expression.trim()"
          @click="executeLogQuery"
        >
          {{ t('explore.runQuery') }}
        </n-button>
      </div>

      <n-alert v-if="logError" type="error" :show-icon="true" closable style="margin: 12px 0" @close="logError = ''">
        {{ logError }}
      </n-alert>

      <div class="results-section">
        <div class="results-header" v-if="logEntries.length > 0">
          <span class="results-count">
            {{ t('explore.showing') }} {{ logEntries.length }} {{ t('explore.entries') }}
            <n-tag v-if="truncated" type="warning" size="small" style="margin-left: 8px">
              {{ t('explore.truncated') }}
            </n-tag>
          </span>
        </div>

        <n-data-table
          v-if="logEntries.length > 0"
          :columns="logColumns"
          :data="logEntries"
          :row-key="(row: LogEntry) => row.timestamp + row.message"
          :max-height="600"
          :scrollbar-props="{ trigger: 'hover' }"
          size="small"
          striped
        />

        <div v-else-if="!logLoading && !logError" class="empty-state">
          <n-empty :description="t('explore.logEmptyDesc')" />
        </div>

        <div v-if="logLoading" class="loading-overlay">
          <n-spin size="medium" />
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.explore-page {
  max-width: 1600px;
  padding: 20px;
}
.explore-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: var(--sre-text-secondary);
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.query-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--sre-bg-card);
  border-radius: 8px;
  border: 1px solid var(--sre-border);
}
.explore-results {
  margin-top: 16px;
  background: var(--sre-bg-card);
  border-radius: 12px;
  padding: 16px;
}
.results-section {
  margin-top: 16px;
  background: var(--sre-bg-card);
  border-radius: 12px;
  padding: 16px;
  min-height: 200px;
  position: relative;
}
.results-header {
  margin-bottom: 12px;
}
.results-count {
  font-size: 13px;
  color: var(--sre-text-secondary);
}
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}
.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--sre-bg-overlay, rgba(0, 0, 0, 0.5));
  border-radius: 12px;
  z-index: 10;
}
</style>
