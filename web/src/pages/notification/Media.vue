<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton, NSpace, NPopconfirm, NTooltip } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { notifyMediaApi } from '@/api'
import type { NotifyMedia } from '@/types'
import { AddOutline } from '@vicons/ionicons5'
import KVEditor from '@/components/common/KVEditor.vue'
import PageHeader from '@/components/common/PageHeader.vue'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const mediaList = ref<NotifyMedia[]>([])
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)
const testingId = ref<number | null>(null)

const form = reactive({
  name: '',
  description: '',
  type: 'lark_webhook' as 'lark_webhook' | 'email' | 'http' | 'script',
  is_enabled: true,
  variables: '{}',
  // lark_webhook config
  webhook_url: '',
  // email config
  smtp_host: '',
  smtp_port: 25,
  username: '',
  password: '',
  from: '',
  // http config
  method: 'POST',
  url: '',
  headers: [] as { key: string; value: string }[],
  body: '',
  // script config
  path: '',
  args: '',
})

const typeOptions = [
  { label: () => t('media.larkWebhook'), value: 'lark_webhook' },
  { label: () => t('media.email'), value: 'email' },
  { label: () => t('media.http'), value: 'http' },
  { label: () => t('media.script'), value: 'script' },
]

const methodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'PATCH', value: 'PATCH' },
]

function getTypeColor(type: string): 'success' | 'info' | 'default' | 'warning' | 'error' {
  const map: Record<string, 'success' | 'info' | 'default' | 'warning' | 'error'> = {
    lark_webhook: 'success',
    email: 'info',
    http: 'default',
    script: 'warning',
  }
  return map[type] || 'default'
}

function getTypeLabel(type: string) {
  const map: Record<string, string> = {
    lark_webhook: 'Lark',
    email: 'Email',
    http: 'HTTP',
    script: 'Script',
  }
  return map[type] || type
}

const columns = [
  {
    title: () => t('common.name'),
    key: 'name',
    width: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('media.type'),
    key: 'type',
    width: 120,
    render: (row: NotifyMedia) =>
      h(NTag, { type: getTypeColor(row.type), size: 'small', bordered: false, round: true }, { default: () => getTypeLabel(row.type) }),
  },
  {
    title: () => t('common.description'),
    key: 'description',
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('media.builtin'),
    key: 'is_builtin',
    width: 80,
    render: (row: NotifyMedia) =>
      row.is_builtin
        ? h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => t('media.builtin') })
        : h(NTag, { size: 'small', bordered: false }, { default: () => t('media.custom') }),
  },
  {
    title: () => t('common.enabled'),
    key: 'is_enabled',
    width: 80,
    render: (row: NotifyMedia) =>
      h(NTag, { type: row.is_enabled ? 'success' : 'default', size: 'small' }, { default: () => row.is_enabled ? t('common.on') : t('common.off') }),
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 220,
    render: (row: NotifyMedia) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => openEdit(row) }, { default: () => t('common.edit') }),
          h(NButton, {
            size: 'small',
            quaternary: true,
            type: 'warning',
            loading: testingId.value === row.id,
            onClick: () => handleTest(row.id),
          }, { default: () => t('common.test') }),
          row.is_builtin
            ? h(NTooltip, {}, {
                trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error', disabled: true }, { default: () => t('common.delete') }),
                default: () => t('media.builtinCannotDelete'),
              })
            : h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
                trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => t('common.delete') }),
                default: () => t('media.deleteConfirm'),
              }),
        ],
      }),
  },
]

async function fetchData() {
  loading.value = true
  try {
    const { data } = await notifyMediaApi.list({ page: 1, page_size: 100 })
    mediaList.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function parseConfig(configStr: string): Record<string, any> {
  try {
    return JSON.parse(configStr || '{}')
  } catch {
    return {}
  }
}

function buildConfigString(): string {
  switch (form.type) {
    case 'lark_webhook':
      return JSON.stringify({ webhook_url: form.webhook_url }, null, 2)
    case 'email':
      return JSON.stringify({
        smtp_host: form.smtp_host,
        smtp_port: form.smtp_port,
        username: form.username,
        password: form.password,
        from: form.from,
      }, null, 2)
    case 'http': {
      const hdrs: Record<string, string> = {}
      for (const h of form.headers) {
        if (h.key.trim()) hdrs[h.key.trim()] = h.value
      }
      return JSON.stringify({
        method: form.method,
        url: form.url,
        headers: hdrs,
        body: form.body,
      }, null, 2)
    }
    case 'script':
      return JSON.stringify({ path: form.path, args: form.args }, null, 2)
    default:
      return '{}'
  }
}

function resetForm() {
  Object.assign(form, {
    name: '',
    description: '',
    type: 'lark_webhook',
    is_enabled: true,
    variables: '{}',
    webhook_url: '',
    smtp_host: '',
    smtp_port: 25,
    username: '',
    password: '',
    from: '',
    method: 'POST',
    url: '',
    headers: [],
    body: '',
    path: '',
    args: '',
  })
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('media.create')
  resetForm()
  showModal.value = true
}

function openEdit(row: NotifyMedia) {
  editingId.value = row.id
  modalTitle.value = t('media.edit')
  const cfg = parseConfig(row.config)

  Object.assign(form, {
    name: row.name,
    description: row.description,
    type: row.type,
    is_enabled: row.is_enabled,
    variables: row.variables || '{}',
    webhook_url: cfg.webhook_url || '',
    smtp_host: cfg.smtp_host || '',
    smtp_port: cfg.smtp_port || 25,
    username: cfg.username || '',
    password: cfg.password || '',
    from: cfg.from || '',
    method: cfg.method || 'POST',
    url: cfg.url || '',
    headers: Object.entries(cfg.headers || {}).map(([key, value]) => ({ key, value: String(value) })),
    body: cfg.body || '',
    path: cfg.path || '',
    args: cfg.args || '',
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning(t('media.nameRequired'))
    return
  }

  // Validate variables JSON
  try {
    JSON.parse(form.variables)
  } catch {
    message.warning(t('media.variables') + ': Invalid JSON')
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      type: form.type,
      is_enabled: form.is_enabled,
      config: buildConfigString(),
      variables: form.variables,
    }
    if (editingId.value) {
      await notifyMediaApi.update(editingId.value, payload)
      message.success(t('media.updated'))
    } else {
      await notifyMediaApi.create(payload)
      message.success(t('media.created'))
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
    await notifyMediaApi.delete(id)
    message.success(t('media.deleted'))
    fetchData()
  } catch (err: any) {
    message.error(err.message)
  }
}

async function handleTest(id: number) {
  testingId.value = id
  try {
    const { data } = await notifyMediaApi.test(id)
    if (data.data.success) {
      message.success(t('media.testSuccess'))
    } else {
      message.warning(`${t('media.testFailed')}: ${data.data.message}`)
    }
  } catch (err: any) {
    message.error(err.message)
  } finally {
    testingId.value = null
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="page-container">
    <PageHeader :title="t('media.title')" :subtitle="t('media.subtitle')">
      <template #actions>
        <n-button type="primary" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          {{ t('media.create') }}
        </n-button>
      </template>
    </PageHeader>

    <n-card :bordered="false" class="content-card">
      <n-data-table
        :loading="loading"
        :columns="columns"
        :data="mediaList"
        :row-key="(row: NotifyMedia) => row.id"
        :bordered="false"
        size="small"
      />
      <n-empty v-if="!loading && mediaList.length === 0" :description="t('media.noData')" style="padding: 40px 0" />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 600px" :bordered="false">
      <n-form label-placement="top">
        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('media.name')" required>
              <n-input v-model:value="form.name" placeholder="e.g. SRE Lark Group" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('media.type')">
              <n-select v-model:value="form.type" :options="typeOptions" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item :label="t('media.description')">
          <n-input v-model:value="form.description" :placeholder="t('media.description')" />
        </n-form-item>

        <!-- Dynamic config based on type -->
        <n-divider style="margin: 12px 0">{{ t('media.config') }}</n-divider>

        <!-- Lark Webhook -->
        <template v-if="form.type === 'lark_webhook'">
          <n-form-item :label="t('media.webhookUrl')" required>
            <n-input v-model:value="form.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." />
          </n-form-item>
        </template>

        <!-- Email -->
        <template v-if="form.type === 'email'">
          <n-grid :x-gap="12" :cols="2">
            <n-gi>
              <n-form-item :label="t('media.smtpHost')">
                <n-input v-model:value="form.smtp_host" placeholder="smtp.example.com" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('media.smtpPort')">
                <n-input-number v-model:value="form.smtp_port" :min="1" :max="65535" style="width: 100%" />
              </n-form-item>
            </n-gi>
          </n-grid>
          <n-grid :x-gap="12" :cols="2">
            <n-gi>
              <n-form-item :label="t('media.username')">
                <n-input v-model:value="form.username" placeholder="user@example.com" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item :label="t('media.password')">
                <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="Password" />
              </n-form-item>
            </n-gi>
          </n-grid>
          <n-form-item :label="t('media.from')">
            <n-input v-model:value="form.from" placeholder="noreply@example.com" />
          </n-form-item>
        </template>

        <!-- HTTP -->
        <template v-if="form.type === 'http'">
          <n-grid :x-gap="12" :cols="4">
            <n-gi>
              <n-form-item :label="t('media.method')">
                <n-select v-model:value="form.method" :options="methodOptions" />
              </n-form-item>
            </n-gi>
            <n-gi :span="3">
              <n-form-item :label="t('media.url')">
                <n-input v-model:value="form.url" placeholder="https://api.example.com/webhook" />
              </n-form-item>
            </n-gi>
          </n-grid>
          <n-form-item :label="t('media.headers')">
            <KVEditor v-model:modelValue="form.headers" key-placeholder="Header Name" value-placeholder="Header Value" :add-label="t('media.addHeader')" />
          </n-form-item>
          <n-form-item :label="t('media.body')">
            <n-input
              v-model:value="form.body"
              type="textarea"
              :rows="4"
              placeholder='{"text": "{{.AlertName}} is {{.Status}}"}'
              style="font-family: monospace; font-size: 12px"
            />
          </n-form-item>
        </template>

        <!-- Script -->
        <template v-if="form.type === 'script'">
          <n-form-item :label="t('media.path')">
            <n-input v-model:value="form.path" placeholder="/usr/local/bin/notify.sh" />
          </n-form-item>
          <n-form-item :label="t('media.args')">
            <n-input v-model:value="form.args" placeholder="--severity {{.Severity}} --name {{.AlertName}}" />
          </n-form-item>
        </template>

        <n-divider style="margin: 12px 0" />

        <n-form-item :label="t('media.variables')">
          <n-input
            v-model:value="form.variables"
            type="textarea"
            :rows="3"
            :placeholder="t('media.variablesHint')"
            style="font-family: monospace; font-size: 12px"
          />
        </n-form-item>

        <n-form-item :label="t('common.enabled')">
          <n-switch v-model:value="form.is_enabled" />
        </n-form-item>
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
