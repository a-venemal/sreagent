<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NSelect, NButton, NSpace, NTag, NAlert, NSpin,
  NDataTable, NTabs, NTabPane, NInputNumber, NIcon,
  useMessage,
} from 'naive-ui'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TooltipComponent, LegendComponent, GridComponent, DataZoomComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { datasourceApi } from '@/api'
import type { DataSource, DataSourceType } from '@/types'
import PromQLEditor from '@/components/query/PromQLEditor.vue'

use([CanvasRenderer, LineChart, TooltipComponent, LegendComponent, GridComponent, DataZoomComponent])

const { t } = useI18n()
const message = useMessage()

type ResultMode = 'chart' | 'table'

// --- state ---
const datasources = ref<DataSource[]>([])
const selectedDsId = ref<number | null>(null)
const expression = ref('')
const loading = ref(false)
const errorMsg = ref('')
const logEntries = ref<any[]>([])
const metricData = ref<any>(null)
const logTotal = ref(0)
const logTruncated = ref(false)
const logLimit = ref(200)
const resultMode = ref<ResultMode>('chart')

// --- time ---
const now = ref(Date.now())
const rangeH = ref(1)
const timeStart = computed(() => Math.floor((now.value - rangeH.value * 3600000) / 1000))
const timeEnd = computed(() => Math.floor(now.value / 1000))

const timeOptions = [
  { label: 'Last 1 hour', value: 1 },
  { label: 'Last 6 hours', value: 6 },
  { label: 'Last 24 hours', value: 24 },
  { label: 'Last 3 days', value: 72 },
  { label: 'Last 7 days', value: 168 },
]

// --- computed ---
const selectedDs = computed(() => datasources.value.find(d => d.id === selectedDsId.value))
const isLogs = computed(() => selectedDs.value?.type === 'victorialogs')
const isMetrics = computed(() => !isLogs.value && !!selectedDs.value)

function dsLabel(ds: DataSource): string {
  const typeLabel = typeBadge(ds.type) as string || ds.type
  return `${ds.name} (${typeLabel})`
}

function typeBadge(t: DataSourceType): string {
  const m: Record<string, string> = {
    prometheus: 'Prometheus',
    victoriametrics: 'VM',
    victorialogs: 'VLogs',
    zabbix: 'Zabbix',
  }
  return m[t] || t
}

function typeColor(t: DataSourceType): string {
  const m: Record<string, string> = {
    prometheus: '#e6522c',
    victoriametrics: '#1a7f37',
    victorialogs: '#0550ae',
    zabbix: '#d32f2f',
  }
  return m[t] || '#666'
}

// --- actions ---
async function loadDs() {
  try {
    const res = await datasourceApi.list({ page: 1, page_size: 100 })
    const list = res.data?.data?.list
    datasources.value = (Array.isArray(list) ? list : []).filter((d: any) => d.is_enabled)
    if (datasources.value.length && !selectedDsId.value) selectedDsId.value = datasources.value[0].id
  } catch { /* ignore */ }
}

async function run() {
  if (!selectedDsId.value || !expression.value.trim()) return
  loading.value = true
  errorMsg.value = ''
  metricData.value = null
  logEntries.value = []
  try {
    if (isLogs.value) {
      const res = await datasourceApi.logQuery(selectedDsId.value, {
        expression: expression.value,
        start: timeStart.value,
        end: timeEnd.value,
        limit: logLimit.value,
      })
      const data = res.data.data
      logEntries.value = (data.entries || []).map((e: any, i: number) => ({ ...e, _key: i }))
      logTotal.value = data.total || 0
      logTruncated.value = data.truncated || false
    } else {
      const diff = (timeEnd.value - timeStart.value)
      const step = diff <= 3600 ? '15s' : diff <= 21600 ? '1m' : diff <= 86400 ? '5m' : '15m'
      const res = await datasourceApi.rangeQuery(selectedDsId.value, {
        expression: expression.value,
        start: timeStart.value,
        end: timeEnd.value,
        step,
      })
      metricData.value = res.data.data
    }
  } catch (e: any) {
    errorMsg.value = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Query failed'
  } finally {
    loading.value = false
  }
}

// --- chart ---
const chartOption = computed(() => {
  if (!metricData.value?.series) return {}
  const seriesList: any[] = []
  const allTimestamps = new Set<number>()
  for (const s of metricData.value.series) {
    const name = formatLegend(s.labels)
    const data: [number, number][] = []
    for (const v of s.values || []) {
      const ts = Number(v.ts) * 1000
      const val = v.value != null ? Number(v.value) : 0
      data.push([ts, val])
      allTimestamps.add(ts)
    }
    seriesList.push({
      name,
      type: 'line',
      data,
      smooth: false,
      showSymbol: false,
      connectNulls: true,
    })
  }
  const timestamps = Array.from(allTimestamps).sort((a, b) => a - b)
  return {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', confine: true },
    legend: {
      type: 'scroll', bottom: 0,
      textStyle: { color: 'var(--sre-text-secondary, #666)', fontSize: 12 },
    },
    grid: { left: 80, right: 20, top: 20, bottom: 40 },
    xAxis: {
      type: 'time',
      data: timestamps,
      axisLabel: { color: 'var(--sre-text-tertiary, #999)', fontSize: 11 },
      axisLine: { lineStyle: { color: 'var(--sre-border, #e0e0e0)' } },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: 'var(--sre-text-tertiary, #999)', fontSize: 11 },
      splitLine: { lineStyle: { color: 'var(--sre-border, #e0e0e0)', type: 'dashed' } },
    },
    series: seriesList,
    dataZoom: [
      { type: 'inside', start: 0, end: 100 },
      { type: 'slider', start: 0, end: 100, height: 24, bottom: 32 },
    ],
  }
})

// --- table ---
const metricColumns = [
  { title: t('explore.metricName') || 'Metric', key: 'name', ellipsis: { tooltip: true } },
  { title: t('explore.value') || 'Value', key: 'value', width: 160 },
  { title: t('explore.labelsHeader') || 'Labels', key: 'labels', ellipsis: { tooltip: true } },
]

const metricTableData = computed(() => {
  if (!metricData.value?.series) return []
  const rows: any[] = []
  let idx = 0
  for (const s of metricData.value.series) {
    for (const v of (s.values || [])) {
      rows.push({
        _key: idx++,
        name: s.labels?.__name__ || '-',
        value: typeof v.value === 'number' ? v.value.toFixed(4) : '-',
        labels: formatLabelsStr(s.labels),
      })
    }
  }
  return rows
})

const logColumns = [
  {
    title: t('explore.logTime') || 'Time',
    key: 'timestamp',
    width: 200,
    render: (r: any) => fmtTs(r.timestamp),
  },
  { title: t('explore.logMessage') || 'Message', key: 'message', ellipsis: { tooltip: true } },
]

// --- helpers ---
function formatLegend(lbs: Record<string, string>): string {
  const parts: string[] = []
  for (const k of Object.keys(lbs)) {
    if (k !== '__name__') parts.push(`${k}="${lbs[k]}"`)
  }
  return parts.length ? parts.join(', ') : (lbs.__name__ || 'value')
}

function formatLabelsStr(lbs: any): string {
  if (!lbs) return '-'
  const parts: string[] = []
  for (const k of Object.keys(lbs)) {
    if (k !== '__name__') parts.push(`${k}=${lbs[k]}`)
  }
  return parts.length ? parts.join(' ') : '-'
}

function fmtTs(ts: any): string {
  if (!ts) return '-'
  try { return new Date(ts).toLocaleString() } catch { return String(ts) }
}

// --- watch ---
watch(selectedDsId, () => {
  expression.value = ''
  metricData.value = null
  logEntries.value = []
  errorMsg.value = ''
})

let refreshTimer: any = null
onMounted(() => {
  loadDs()
})

// Cleanup timer on unmount (not used currently, but ready for auto-refresh)
</script>

<template>
  <div class="explore-page">
    <!-- Header -->
    <div class="explore-header">
      <div>
        <h2 class="explore-title">{{ t('explore.title') }}</h2>
        <p class="explore-subtitle">{{ t('explore.subtitle') }}</p>
      </div>
      <NSpace align="center" size="small">
        <span style="font-size:12px;color:var(--sre-text-tertiary)">Timezone: {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}</span>
        <NSelect
          v-model:value="rangeH"
          :options="timeOptions"
          size="small"
          style="width:150px"
        />
      </NSpace>
    </div>

    <!-- Data Source Selector -->
    <div style="margin-bottom:12px">
      <NSelect
        v-model:value="selectedDsId"
        :options="datasources.map(d => ({
          label: dsLabel(d),
          value: d.id,
        }))"
        :placeholder="t('explore.selectDatasource') || 'Select datasource'"
        filterable
        clearable
        style="max-width:480px"
      />
      <div v-if="selectedDs" style="display:flex;align-items:center;gap:8px;margin-top:8px">
        <NTag :color="{ color: typeColor(selectedDs.type), textColor: '#fff' }" size="small" :bordered="false">
          {{ typeBadge(selectedDs.type) }}
        </NTag>
        <span style="font-size:12px;color:var(--sre-text-tertiary)">{{ selectedDs.endpoint }}</span>
        <span v-if="selectedDs.version" style="font-size:12px;color:var(--sre-text-tertiary)">{{ selectedDs.version }}</span>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="!selectedDsId" class="explore-empty">
      {{ t('explore.selectDatasource') || 'Select a data source to query' }}
    </div>

    <!-- Query Area -->
    <template v-if="selectedDsId">
      <div class="query-bar">
        <div class="query-editor-wrap">
          <PromQLEditor
            v-if="isMetrics"
            v-model="expression"
            :datasource-id="selectedDsId"
            :placeholder="t('explore.promqlPlaceholder') || 'Enter PromQL expression...'"
            @execute="run"
          />
          <textarea
            v-else
            v-model="expression"
            class="logsql-input"
            rows="3"
            :placeholder="String(t('explore.logQueryPlaceholder') || 'Enter LogsQL expression...')"
            @keyup.ctrl.enter="run"
            @keyup.meta.enter="run"
          ></textarea>
        </div>
        <div class="query-actions">
          <NInputNumber
            v-if="isLogs"
            v-model:value="logLimit"
            :min="10"
            :max="10000"
            size="small"
            style="width:120px"
            :placeholder="String(t('explore.limit') || 'Limit')"
          />
          <NButton
            type="primary"
            :loading="loading"
            :disabled="!expression.trim()"
            @click="run"
          >
            {{ t('explore.runQuery') || 'Run Query' }}
          </NButton>
          <span class="shortcut-hint">Ctrl+Enter</span>
        </div>
      </div>

      <!-- Error -->
      <NAlert
        v-if="errorMsg"
        type="error"
        :title="errorMsg"
        closable
        style="margin-bottom:12px"
        @close="errorMsg = ''"
      />

      <!-- Metrics Results -->
      <template v-if="isMetrics && metricData?.series">
        <div class="results-panel">
          <div class="results-header">
            <span class="results-count">
              {{ metricTableData.length }} data points
              <template v-if="metricData.resultType"> · {{ metricData.resultType }}</template>
            </span>
            <NSpace size="small">
              <NButton
                size="small"
                :type="resultMode === 'chart' ? 'primary' : 'default'"
                @click="resultMode = 'chart'"
              >
                {{ t('explore.chart') || 'Chart' }}
              </NButton>
              <NButton
                size="small"
                :type="resultMode === 'table' ? 'primary' : 'default'"
                @click="resultMode = 'table'"
              >
                {{ t('explore.table') || 'Table' }}
              </NButton>
            </NSpace>
          </div>
          <div v-if="resultMode === 'chart'" class="chart-container">
            <VChart
              :option="chartOption"
              :autoresize="true"
              style="width:100%;height:400px"
            />
          </div>
          <NDataTable
            v-else
            :columns="metricColumns"
            :data="metricTableData"
            :row-key="(r: any) => r._key"
            size="small"
            :single-line="false"
            striped
            max-height="500"
            virtual-scroll
          />
        </div>
      </template>

      <!-- Logs Results -->
      <template v-if="isLogs && logEntries.length">
        <div class="results-panel">
          <div class="results-header">
            <span class="results-count">
              {{ t('explore.showing') || 'Showing' }} {{ logEntries.length }}
              <template v-if="logTotal > 0"> / {{ logTotal }}</template>
              {{ t('explore.entries') || 'entries' }}
            </span>
            <NTag v-if="logTruncated" type="warning" size="small" :bordered="false">
              {{ t('explore.truncated') || 'truncated' }}
            </NTag>
          </div>
          <NDataTable
            :columns="logColumns"
            :data="logEntries"
            :row-key="(r: any) => r._key"
            size="small"
            max-height="600"
            virtual-scroll
          />
        </div>
      </template>

      <!-- No results -->
      <div
        v-if="!loading && !errorMsg && expression.trim() && !metricData?.series && !logEntries.length"
        class="explore-empty"
      >
        {{ t('explore.logEmptyDesc') || 'No results' }}
      </div>
    </template>
  </div>
</template>

<style scoped>
.explore-page {
  max-width: 1600px;
  padding: 24px;
}

.explore-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.explore-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: var(--sre-text-primary);
}

.explore-subtitle {
  font-size: 13px;
  color: var(--sre-text-secondary);
  margin: 0;
}

.explore-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: var(--sre-text-tertiary);
  font-size: 14px;
}

.query-bar {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.query-editor-wrap {
  flex: 1;
  min-width: 0;
}

.logsql-input {
  width: 100%;
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid var(--sre-border, #e0e0e0);
  background: var(--sre-bg-sunken, #f5f5f5);
  color: var(--sre-text-primary);
  font-size: 13px;
  font-family: 'Cascadia Code', 'Fira Code', 'JetBrains Mono', monospace;
  resize: vertical;
  min-height: 60px;
  box-sizing: border-box;
}
.logsql-input:focus {
  outline: none;
  border-color: var(--sre-primary, #18a058);
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.12);
}

.query-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.shortcut-hint {
  font-size: 11px;
  color: var(--sre-text-tertiary);
  white-space: nowrap;
}

.results-panel {
  background: var(--sre-bg-card, #fff);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid var(--sre-border, #e0e0e0);
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.results-count {
  font-size: 13px;
  color: var(--sre-text-secondary);
}

.chart-container {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
