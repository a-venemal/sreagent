<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NTabs, NTabPane, NSpace } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { datasourceApi } from '@/api'
import type { DataSource } from '@/types'
import { useTimeRange } from '@/composables/useTimeRange'
import { useQueryEngine } from '@/composables/useQueryEngine'
import QueryPanel from '@/components/query/QueryPanel.vue'
import QueryResultChart from '@/components/query/QueryResultChart.vue'
import QueryResultTable from '@/components/query/QueryResultTable.vue'
import TimeRangePicker from '@/components/time/TimeRangePicker.vue'
import RefreshPicker from '@/components/time/RefreshPicker.vue'

const { t } = useI18n()
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

const resultTab = ref('chart')

const hasResults = computed(() =>
  targets.value.some(t => t.series && t.series.length > 0)
)

async function fetchDatasources() {
  try {
    const res = await datasourceApi.list({ page: 1, page_size: 100 })
    datasources.value = (res.data.data.list || []).filter((ds: any) => ds.is_enabled)
  } catch {
    // ignore
  }
}

function handleExecuteSingle(id: string) {
  const target = targets.value.find(t => t.id === id)
  if (target) executeQuery(target)
}

onMounted(fetchDatasources)
</script>

<template>
  <div class="explore-page">
    <div class="explore-header">
      <div class="header-left">
        <h2 class="page-title">{{ t('explore.title') }}</h2>
        <span class="page-subtitle">{{ t('explore.subtitle') }}</span>
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
      </div>
    </div>

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

    <div v-if="hasResults" class="explore-results">
      <NTabs v-model:value="resultTab" type="line" size="small">
        <NTabPane name="chart" tab="Chart">
          <QueryResultChart
            :targets="targets"
            :time-range="timeRange"
            :height="400"
          />
        </NTabPane>
        <NTabPane name="table" tab="Table">
          <QueryResultTable :targets="targets" />
        </NTabPane>
      </NTabs>
    </div>
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
  color: #666;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.explore-results {
  margin-top: 16px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
}
</style>
