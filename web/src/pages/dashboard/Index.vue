<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NTag } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { dashboardApi, alertEventApi, engineApi } from '@/api'
import type { DashboardStats, AlertEvent, EngineStatus } from '@/types'
import { formatTime } from '@/utils/format'
import { getSeverityType, getEventStatusType } from '@/utils/alert'
import PageHeader from '@/components/common/PageHeader.vue'
import {
  AlertCircleOutline,
  ServerOutline,
  CheckmarkCircleOutline,
  ReaderOutline,
  PulseOutline,
} from '@vicons/ionicons5'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const eventsLoading = ref(false)
const engineLoading = ref(false)

const stats = ref<DashboardStats>({
  total_datasources: 0,
  total_rules: 0,
  active_alerts: 0,
  resolved_today: 0,
  total_users: 0,
  total_teams: 0,
})

const engineStatus = ref<EngineStatus>({
  running: false,
  total_rules: 0,
  active_alerts: 0,
  uptime: '',
})

const recentAlerts = ref<AlertEvent[]>([])

const statCards = ref([
  { titleKey: 'dashboard.activeAlerts', key: 'active_alerts' as const, icon: AlertCircleOutline, color: '#e88080' },
  { titleKey: 'dashboard.dataSources', key: 'total_datasources' as const, icon: ServerOutline, color: '#18a058' },
  { titleKey: 'dashboard.resolvedToday', key: 'resolved_today' as const, icon: CheckmarkCircleOutline, color: '#70c0e8' },
  { titleKey: 'dashboard.totalRules', key: 'total_rules' as const, icon: ReaderOutline, color: '#f2c97d' },
])

const alertColumns = [
  {
    title: () => t('alert.severity'),
    key: 'severity',
    width: 100,
    render: (row: AlertEvent) =>
      h(NTag, { type: getSeverityType(row.severity), size: 'small', round: true }, { default: () => row.severity.toUpperCase() }),
  },
  {
    title: () => t('alert.alertName'),
    key: 'alert_name',
    ellipsis: { tooltip: true },
    render: (row: AlertEvent) =>
      h('a', {
        style: 'color: var(--sre-info); cursor: pointer; text-decoration: none',
        onClick: () => router.push(`/alerts/events/${row.id}`),
      }, row.alert_name),
  },
  {
    title: () => t('common.status'),
    key: 'status',
    width: 120,
    render: (row: AlertEvent) =>
      h(NTag, { type: getEventStatusType(row.status), size: 'small' }, { default: () => row.status }),
  },
  {
    title: () => t('alert.source'),
    key: 'source',
    width: 120,
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('alert.firedAt'),
    key: 'fired_at',
    width: 180,
    render: (row: AlertEvent) => formatTime(row.fired_at),
  },
  {
    title: () => t('alert.fireCount'),
    key: 'fire_count',
    width: 70,
  },
]

async function fetchStats() {
  loading.value = true
  try {
    const { data } = await dashboardApi.getStats()
    stats.value = data.data
  } catch (err: any) {
    message.error(err.message || t('dashboard.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchEngineStatus() {
  engineLoading.value = true
  try {
    const { data } = await engineApi.getStatus()
    engineStatus.value = data.data
  } catch (err: any) {
    // Engine status is non-critical, fail silently
    engineStatus.value = { running: false, total_rules: 0, active_alerts: 0, uptime: '' }
  } finally {
    engineLoading.value = false
  }
}

async function fetchRecentAlerts() {
  eventsLoading.value = true
  try {
    const { data } = await alertEventApi.list({ page: 1, page_size: 10, status: ['firing'] })
    recentAlerts.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message || t('dashboard.loadAlertsFailed'))
  } finally {
    eventsLoading.value = false
  }
}

onMounted(() => {
  fetchStats()
  fetchEngineStatus()
  fetchRecentAlerts()
})
</script>

<template>
  <div class="dashboard">
    <PageHeader :title="t('dashboard.title')" :subtitle="t('dashboard.subtitle')" />

    <!-- Stats Cards -->
    <n-spin :show="loading">
      <n-grid :x-gap="16" :y-gap="16" :cols="5" responsive="screen" style="margin-bottom: 24px">
        <n-gi v-for="card in statCards" :key="card.key">
          <n-card class="stat-card card-hover" :bordered="false">
            <div class="stat-content">
              <div class="stat-info">
                <div class="stat-label">{{ t(card.titleKey) }}</div>
                <div class="stat-value">{{ stats[card.key] }}</div>
              </div>
              <div class="stat-icon" :style="{ background: card.color + '15', color: card.color }">
                <n-icon :component="card.icon" :size="28" />
              </div>
            </div>
          </n-card>
        </n-gi>

        <!-- Engine Status Card -->
        <n-gi>
          <n-card class="stat-card card-hover" :bordered="false">
            <div class="stat-content">
              <div class="stat-info">
                <div class="stat-label">
                  {{ t('engine.title') }}
                </div>
                <div class="engine-status-row">
                  <span class="engine-dot" :class="engineStatus.running ? 'engine-dot--running' : 'engine-dot--stopped'"></span>
                  <span class="engine-status-text">{{ engineStatus.running ? t('engine.running') : t('engine.stopped') }}</span>
                </div>
                <div class="engine-meta">
                  <span>{{ engineStatus.total_rules }} {{ t('engine.totalRules') }}</span>
                  <span style="margin: 0 6px; opacity: 0.25">|</span>
                  <span>{{ engineStatus.active_alerts }} {{ t('engine.activeAlerts') }}</span>
                </div>
              </div>
              <div class="stat-icon" :style="{ background: (engineStatus.running ? '#18a058' : '#e88080') + '15', color: engineStatus.running ? '#18a058' : '#e88080' }">
                <n-icon :component="PulseOutline" :size="28" />
              </div>
            </div>
          </n-card>
        </n-gi>
      </n-grid>
    </n-spin>

    <!-- Additional mini stats -->
    <n-grid :x-gap="16" :y-gap="16" :cols="2" responsive="screen" style="margin-bottom: 24px">
      <n-gi>
        <n-card class="stat-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">{{ t('dashboard.totalUsers') }}</div>
              <div class="stat-value" style="font-size: 22px">{{ stats.total_users }}</div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">{{ t('dashboard.totalTeams') }}</div>
              <div class="stat-value" style="font-size: 22px">{{ stats.total_teams }}</div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- Recent Alerts Table -->
    <n-card :title="t('dashboard.recentAlerts')" :bordered="false" style="background: var(--sre-bg-card); border-radius: 12px">
      <template #header-extra>
        <n-button text type="primary" @click="router.push('/alerts/events')">
          {{ t('dashboard.viewAll') }}
        </n-button>
      </template>

      <n-data-table
        v-if="recentAlerts.length > 0 || eventsLoading"
        :loading="eventsLoading"
        :columns="alertColumns"
        :data="recentAlerts"
        :row-key="(row: AlertEvent) => row.id"
        :bordered="false"
        size="small"
        :pagination="false"
      />

      <n-empty
        v-if="!eventsLoading && recentAlerts.length === 0"
        :description="t('dashboard.noAlerts')"
        style="padding: 40px 0"
      >
        <template #extra>
          <n-button type="primary" size="small" @click="router.push('/datasources')">
            {{ t('dashboard.configDatasources') }}
          </n-button>
        </template>
      </n-empty>
    </n-card>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.stat-card {
  background: var(--sre-bg-card);
  border-radius: 12px;
}

.stat-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-label {
  font-size: 13px;
  color: var(--sre-text-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--sre-text-primary);
  line-height: 1;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.engine-status-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.engine-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.engine-dot--running {
  background: #18a058;
  box-shadow: 0 0 6px rgba(24, 160, 88, 0.6);
  animation: engine-pulse 2s ease-in-out infinite;
}

.engine-dot--stopped {
  background: #e88080;
  box-shadow: 0 0 6px rgba(232, 128, 128, 0.4);
}

@keyframes engine-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 6px rgba(24, 160, 88, 0.6); }
  50% { opacity: 0.6; box-shadow: 0 0 12px rgba(24, 160, 88, 0.9); }
}

.engine-status-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--sre-text-primary);
  line-height: 1;
}

.engine-meta {
  font-size: 12px;
  color: var(--sre-text-secondary);
  margin-top: 6px;
}
</style>
