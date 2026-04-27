<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { datasourceApi } from '@/api'
import type { DataSource, QueryResponse } from '@/types'
import PageHeader from '@/components/common/PageHeader.vue'

const message = useMessage()
const { t } = useI18n()

const datasources = ref<DataSource[]>([])
const selectedDsId = ref<number | null>(null)
const expression = ref('')
const queryTime = ref(0)
const loading = ref(false)
const queryResult = ref<QueryResponse | null>(null)
const queryError = ref('')

async function fetchDatasources() {
  try {
    const { data } = await datasourceApi.list({ page: 1, page_size: 100 })
    datasources.value = (data.data.list || []).filter(ds => ds.is_enabled)
  } catch (err: any) {
    message.error(err.message || t('datasource.fetchFailed'))
  }
}

async function handleQuery() {
  if (!selectedDsId.value) { message.warning(t('datasource.selectDatasource')); return }
  if (!expression.value.trim()) { message.warning(t('datasource.queryExpression')); return }

  loading.value = true
  queryResult.value = null
  queryError.value = ''
  try {
    const { data } = await datasourceApi.query(selectedDsId.value, {
      expression: expression.value,
      time: queryTime.value,
    })
    queryResult.value = data.data
  } catch (err: any) {
    queryError.value = err.message || 'Query failed'
  } finally {
    loading.value = false
  }
}

function formatTimestamp(ts: number) {
  return new Date(ts * 1000).toLocaleString()
}

onMounted(fetchDatasources)
</script>

<template>
  <div class="query-page">
    <PageHeader :title="t('datasource.queryTitle')" :subtitle="t('datasource.querySubtitle')" />

    <n-card :bordered="false" class="content-card">
      <n-space vertical :size="16">
        <div>
          <label>{{ t('datasource.selectDatasource') }}</label>
          <n-select
            v-model:value="selectedDsId"
            :options="datasources.map(ds => ({ label: `${ds.name} (${ds.type})`, value: ds.id }))"
            :placeholder="datasources.length === 0 ? t('datasource.noEnabledDatasource') : t('datasource.selectDatasource')"
            filterable
          />
        </div>
        <div>
          <label>{{ t('datasource.queryTime') }}</label>
          <n-select
            v-model:value="queryTime"
            :options="[
              { label: 'now', value: 0 },
              { label: '5m ago', value: 300 },
              { label: '15m ago', value: 900 },
              { label: '1h ago', value: 3600 },
            ]"
            :placeholder="t('datasource.queryTime')"
          />
        </div>
        <div>
          <label>{{ t('datasource.queryExpression') }}</label>
          <n-input
            v-model:value="expression"
            type="textarea"
            :placeholder="t('datasource.queryPlaceholder')"
            :rows="3"
            @keyup.ctrl.enter="handleQuery"
          />
        </div>
        <n-button
          type="primary"
          :loading="loading"
          @click="handleQuery"
          :disabled="!selectedDsId || !expression.trim()"
        >
          {{ t('datasource.executeQuery') }}
        </n-button>
      </n-space>
    </n-card>

    <n-alert v-if="queryError" type="error" style="margin-top: 16px" closable @close="queryError = ''">
      {{ queryError }}
    </n-alert>

    <n-card v-if="queryResult" :bordered="false" class="content-card" style="margin-top: 16px">
      <template #header>
        <n-space align="center">
          <span>{{ t('datasource.queryResult') }}</span>
          <n-tag size="small" type="info">{{ queryResult.result_type }}</n-tag>
          <n-tag size="small">{{ queryResult.series?.length ?? queryResult.raw_count ?? 0 }} series</n-tag>
        </n-space>
      </template>

      <template v-if="!queryResult.series || queryResult.series.length === 0">
        <n-empty :description="t('datasource.queryNoResult')" />
      </template>

      <template v-else-if="queryResult.result_type === 'vector' || queryResult.result_type === 'matrix'">
        <n-data-table
          :columns="[
            { title: 'Labels', key: 'labels', minWidth: 200, ellipsis: { tooltip: true } },
            { title: 'Values', key: 'values', minWidth: 300 },
          ]"
          :data="queryResult.series.map((s, i) => ({
            key: i,
            labels: Object.entries(s.labels).map(([k, v]) => `${k}=${v}`).join(', '),
            values: s.values.map(v => `${formatTimestamp(v.ts)}: ${v.value}`).join('\n'),
          }))"
          :max-height="400"
          size="small"
        />
      </template>

      <template v-else>
        <n-code :code="JSON.stringify(queryResult.series, null, 2)" language="json" show-line-numbers />
      </template>
    </n-card>
  </div>
</template>

<style scoped>
.query-page { max-width: 1400px; }
.content-card { border-radius: 12px; }
.content-card label { display: block; margin-bottom: 4px; font-size: 13px; color: #666; }
</style>
