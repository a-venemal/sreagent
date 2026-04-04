<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton, NSpace, NPopconfirm, NSwitch } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { notifyRuleApi } from '@/api'
import type { NotifyRule } from '@/types'
import { AddOutline } from '@vicons/ionicons5'
import { kvArrayToRecord } from '@/utils/format'
import { getSeverityType } from '@/utils/alert'
import KVEditor from '@/components/common/KVEditor.vue'
import PageHeader from '@/components/common/PageHeader.vue'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const rules = ref<NotifyRule[]>([])
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)

const form = reactive({
  name: '',
  description: '',
  severities: [] as string[],
  match_labels: [] as { key: string; value: string }[],
  pipeline: '[]',
  notify_configs: '[]',
  repeat_interval: 3600,
  callback_url: '',
  is_enabled: true,
})

const severityOptions = [
  { label: () => t('alert.critical'), value: 'critical' },
  { label: () => t('alert.warning'), value: 'warning' },
  { label: () => t('alert.info'), value: 'info' },
]

const columns = [
  {
    title: () => t('common.name'),
    key: 'name',
    width: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('notifyRule.severities'),
    key: 'severities',
    width: 200,
    render: (row: NotifyRule) => {
      const sevs = (row.severities || '').split(',').filter(Boolean)
      if (sevs.length === 0) return h('span', { style: 'color: #666' }, '-')
      return h(NSpace, { size: 4 }, {
        default: () => sevs.map(s =>
          h(NTag, { size: 'small', type: getSeverityType(s), round: true, bordered: false }, { default: () => s })
        ),
      })
    },
  },
  {
    title: () => t('notifyRule.matchLabels'),
    key: 'match_labels',
    width: 220,
    render: (row: NotifyRule) => {
      const labels = row.match_labels || {}
      const entries = Object.entries(labels)
      if (entries.length === 0) return h('span', { style: 'color: #666' }, '-')
      return h(NSpace, { size: 4 }, {
        default: () => entries.map(([k, v]) =>
          h(NTag, { size: 'small', bordered: false }, { default: () => `${k}=${v}` })
        ),
      })
    },
  },
  {
    title: () => t('notifyRule.repeatInterval'),
    key: 'repeat_interval',
    width: 120,
    render: (row: NotifyRule) => `${row.repeat_interval}s`,
  },
  {
    title: () => t('common.enabled'),
    key: 'is_enabled',
    width: 80,
    render: (row: NotifyRule) =>
      h(NTag, { type: row.is_enabled ? 'success' : 'default', size: 'small' }, { default: () => row.is_enabled ? t('common.on') : t('common.off') }),
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 160,
    render: (row: NotifyRule) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => openEdit(row) }, { default: () => t('common.edit') }),
          h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => t('common.delete') }),
            default: () => t('notifyRule.deleteConfirm'),
          }),
        ],
      }),
  },
]

async function fetchData() {
  loading.value = true
  try {
    const { data } = await notifyRuleApi.list({ page: 1, page_size: 100 })
    rules.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, {
    name: '',
    description: '',
    severities: [],
    match_labels: [],
    pipeline: '[]',
    notify_configs: '[]',
    repeat_interval: 3600,
    callback_url: '',
    is_enabled: true,
  })
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('notifyRule.create')
  resetForm()
  showModal.value = true
}

function openEdit(row: NotifyRule) {
  editingId.value = row.id
  modalTitle.value = t('notifyRule.edit')
  Object.assign(form, {
    name: row.name,
    description: row.description,
    severities: (row.severities || '').split(',').filter(Boolean),
    match_labels: Object.entries(row.match_labels || {}).map(([key, value]) => ({ key, value })),
    pipeline: row.pipeline || '[]',
    notify_configs: row.notify_configs || '[]',
    repeat_interval: row.repeat_interval,
    callback_url: row.callback_url || '',
    is_enabled: row.is_enabled,
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning(t('notifyRule.nameRequired'))
    return
  }

  // Validate JSON fields
  try {
    JSON.parse(form.pipeline)
  } catch {
    message.warning(t('notifyRule.pipeline') + ': Invalid JSON')
    return
  }
  try {
    JSON.parse(form.notify_configs)
  } catch {
    message.warning(t('notifyRule.notifyConfigs') + ': Invalid JSON')
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      severities: form.severities.join(','),
      match_labels: kvArrayToRecord(form.match_labels),
      pipeline: form.pipeline,
      notify_configs: form.notify_configs,
      repeat_interval: form.repeat_interval,
      callback_url: form.callback_url,
      is_enabled: form.is_enabled,
    }
    if (editingId.value) {
      await notifyRuleApi.update(editingId.value, payload)
      message.success(t('notifyRule.updated'))
    } else {
      await notifyRuleApi.create(payload)
      message.success(t('notifyRule.created'))
    }
    showModal.value = false
    fetchData()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await notifyRuleApi.delete(id)
    message.success(t('notifyRule.deleted'))
    fetchData()
  } catch (err: any) {
    message.error(err.message)
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="page-container">
    <PageHeader :title="t('notifyRule.title')" :subtitle="t('notifyRule.subtitle')">
      <template #actions>
        <n-button type="primary" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          {{ t('notifyRule.create') }}
        </n-button>
      </template>
    </PageHeader>

    <n-card :bordered="false" class="content-card">
      <n-data-table
        :loading="loading"
        :columns="columns"
        :data="rules"
        :row-key="(row: NotifyRule) => row.id"
        :bordered="false"
        size="small"
      />
      <n-empty v-if="!loading && rules.length === 0" :description="t('notifyRule.noData')" style="padding: 40px 0" />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 600px" :bordered="false">
      <n-form label-placement="top">
        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('notifyRule.name')" required>
              <n-input v-model:value="form.name" placeholder="e.g. Critical Alert Notify" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('common.enabled')">
              <n-switch v-model:value="form.is_enabled" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item :label="t('notifyRule.description')">
          <n-input v-model:value="form.description" :placeholder="t('notifyRule.description')" />
        </n-form-item>

        <n-form-item :label="t('notifyRule.severities')">
          <n-select
            v-model:value="form.severities"
            :options="severityOptions"
            multiple
            :placeholder="t('common.selectSeverities')"
          />
        </n-form-item>

        <n-form-item :label="t('notifyRule.matchLabels')">
          <KVEditor v-model:modelValue="form.match_labels" :add-label="t('notifyRule.addLabel')" />
        </n-form-item>

        <n-form-item :label="t('notifyRule.pipeline')">
          <n-input
            v-model:value="form.pipeline"
            type="textarea"
            :rows="4"
            :placeholder="t('notifyRule.pipelineHint')"
            style="font-family: monospace; font-size: 12px"
          />
        </n-form-item>

        <n-form-item :label="t('notifyRule.notifyConfigs')">
          <n-input
            v-model:value="form.notify_configs"
            type="textarea"
            :rows="4"
            :placeholder="t('notifyRule.notifyConfigsHint')"
            style="font-family: monospace; font-size: 12px"
          />
        </n-form-item>

        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('notifyRule.repeatInterval')">
              <n-input-number v-model:value="form.repeat_interval" :min="0" style="width: 100%" />
              <template #feedback>
                <n-text depth="3" style="font-size: 12px">{{ t('notifyRule.repeatIntervalHint') }}</n-text>
              </template>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('notifyRule.callbackUrl')">
              <n-input v-model:value="form.callback_url" placeholder="https://..." />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">
            {{ editingId ? t('common.update') : t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 1400px;
}

.content-card {
  border-radius: 12px;
}
</style>
