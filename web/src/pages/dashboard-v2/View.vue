<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, NInput, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { dashboardV2Api } from '@/api'
import type { DashboardV2, DashboardConfig, PanelConfig, VariableConfig } from '@/types/dashboard'
import { useTimeRange } from '@/composables/useTimeRange'
import { useQueryEngine, createDefaultTarget } from '@/composables/useQueryEngine'
import { useVariable } from '@/composables/useVariable'
import TimeRangePicker from '@/components/time/TimeRangePicker.vue'
import RefreshPicker from '@/components/time/RefreshPicker.vue'
import QueryPanel from '@/components/query/QueryPanel.vue'
import QueryResultChart from '@/components/query/QueryResultChart.vue'
import type { DataSource } from '@/types'
import { datasourceApi } from '@/api'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const { t } = useI18n()

const isNew = computed(() => route.params.id === 'new')
const dashboard = ref<DashboardV2 | null>(null)
const loading = ref(false)
const saving = ref(false)
const config = ref<DashboardConfig>({
  panels: [],
  layout: { cols: 24, rowHeight: 100 },
  variables: [],
})

const datasources = ref<DataSource[]>([])

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

const variableConfig = ref<VariableConfig[]>(config.value.variables || [])
const { variableList, replaceVariables, setValue, resolveAll } = useVariable(variableConfig, timeRange)

async function fetchDatasources() {
  try {
    const res = await datasourceApi.list({ page: 1, page_size: 100 })
    datasources.value = (res.data.data.list || []).filter((ds: any) => ds.is_enabled)
  } catch { /* ignore */ }
}

async function fetchDashboard() {
  if (isNew.value) return
  loading.value = true
  try {
    const res = await dashboardV2Api.get(Number(route.params.id))
    dashboard.value = res.data.data
    if (dashboard.value.config) {
      try {
        config.value = JSON.parse(dashboard.value.config)
        variableConfig.value = config.value.variables || []
      } catch { /* ignore */ }
    }
  } catch (err: any) {
    message.error(err.message || 'Failed to load dashboard')
    router.back()
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const cfg = { ...config.value, variables: variableConfig.value }
    const data = {
      name: dashboard.value?.name || 'Untitled',
      description: dashboard.value?.description || '',
      tags: dashboard.value?.tags || {},
      config: JSON.stringify(cfg),
      is_public: dashboard.value?.is_public || false,
    }
    if (isNew.value) {
      const res = await dashboardV2Api.create(data)
      message.success('Dashboard created')
      router.replace({ name: 'DashboardV2View', params: { id: res.data.data.id } })
    } else if (dashboard.value) {
      await dashboardV2Api.update(dashboard.value.id, data)
      message.success('Dashboard saved')
    }
  } catch (err: any) {
    message.error(err.message || 'Save failed')
  } finally {
    saving.value = false
  }
}

function handleExecuteSingle(id: string) {
  const target = targets.value.find(t => t.id === id)
  if (target) executeQuery(target)
}

onMounted(() => {
  fetchDatasources()
  fetchDashboard()
})
</script>

<template>
  <div class="dashboard-view">
    <div class="dashboard-header">
      <div class="header-left">
        <NButton quaternary size="small" @click="router.push({ name: 'DashboardV2List' })">
          &larr; Back
        </NButton>
        <NInput
          v-if="dashboard || isNew"
          :value="dashboard?.name || 'Untitled'"
          size="small"
          style="width: 300px"
          @update:value="(v: string) => { if (dashboard) dashboard.name = v }"
        />
      </div>
      <div class="header-right">
        <TimeRangePicker
          :time-range="timeRange"
          :is-relative="isRelative"
          :relative-duration="relativeDuration"
          @set-relative="setRelative"
          @set-absolute="setAbsolute"
        />
        <RefreshPicker
          :value="autoRefreshInterval"
          @update:value="(v) => autoRefreshInterval = v"
        />
        <NButton type="primary" size="small" :loading="saving" @click="handleSave">
          Save
        </NButton>
      </div>
    </div>

    <!-- Variable pickers -->
    <div v-if="variableList.length > 0" class="variable-bar">
      <div v-for="v in variableList" :key="v.config.name" class="var-item">
        <label>{{ v.config.label || v.config.name }}</label>
        <NSelect
          v-if="v.config.type === 'query' || v.config.type === 'custom'"
          :value="v.value"
          :options="v.options.map(o => ({ label: o, value: o }))"
          :loading="v.loading"
          size="small"
          style="width: 160px"
          @update:value="(val: string) => setValue(v.config.name, val)"
        />
        <NInput
          v-else-if="v.config.type === 'textbox'"
          :value="v.value"
          size="small"
          style="width: 160px"
          @update:value="(val: string) => setValue(v.config.name, val)"
        />
        <span v-else class="var-value">{{ v.value }}</span>
      </div>
    </div>

    <!-- Query panel -->
    <QueryPanel
      :targets="targets"
      :datasources="datasources"
      :loading="globalLoading"
      @add="addTarget"
      @remove="removeTarget"
      @toggle="toggleTarget"
      @update="updateTarget"
      @execute="handleExecuteSingle"
      @execute-all="executeAll"
    />

    <!-- Results -->
    <div v-if="targets.some(t => t.series && t.series.length > 0)" class="results-section">
      <QueryResultChart
        :targets="targets"
        :time-range="timeRange"
        :height="400"
      />
    </div>
  </div>
</template>

<style scoped>
.dashboard-view {
  padding: 20px;
  max-width: 1600px;
}
.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.variable-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
}
.var-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.var-item label {
  font-size: 12px;
  color: #666;
  white-space: nowrap;
}
.var-value {
  font-size: 13px;
  padding: 4px 8px;
  background: #f5f5f5;
  border-radius: 4px;
}
.results-section {
  margin-top: 16px;
}
</style>
