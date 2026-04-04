<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton, NSpace, NPopconfirm, NTooltip } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { messageTemplateApi } from '@/api'
import type { MessageTemplate } from '@/types'
import { AddOutline } from '@vicons/ionicons5'
import PageHeader from '@/components/common/PageHeader.vue'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const templates = ref<MessageTemplate[]>([])
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)

const showPreviewModal = ref(false)
const previewLoading = ref(false)
const previewResult = ref('')

const showVariablesHint = ref(false)

const form = reactive({
  name: '',
  description: '',
  type: 'text' as 'text' | 'html' | 'markdown' | 'lark_card',
  content: '',
})

const typeOptions = [
  { label: () => t('template.text'), value: 'text' },
  { label: () => t('template.html'), value: 'html' },
  { label: () => t('template.markdown'), value: 'markdown' },
  { label: () => t('template.larkCard'), value: 'lark_card' },
]

function getTypeTagType(type: string): 'default' | 'info' | 'success' | 'warning' | 'error' {
  const map: Record<string, 'default' | 'info' | 'success' | 'warning' | 'error'> = {
    text: 'default',
    html: 'info',
    markdown: 'success',
    lark_card: 'warning',
  }
  return map[type] || 'default'
}

const columns = [
  {
    title: () => t('common.name'),
    key: 'name',
    width: 200,
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('template.type'),
    key: 'type',
    width: 120,
    render: (row: MessageTemplate) =>
      h(NTag, { type: getTypeTagType(row.type), size: 'small', bordered: false, round: true }, { default: () => row.type }),
  },
  {
    title: () => t('template.builtin'),
    key: 'is_builtin',
    width: 80,
    render: (row: MessageTemplate) =>
      row.is_builtin
        ? h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => t('template.builtin') })
        : h(NTag, { size: 'small', bordered: false }, { default: () => t('template.custom') }),
  },
  {
    title: () => t('common.description'),
    key: 'description',
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 220,
    render: (row: MessageTemplate) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => openEdit(row) }, { default: () => t('common.edit') }),
          h(NButton, { size: 'small', quaternary: true, type: 'warning', onClick: () => handlePreview(row) }, { default: () => t('template.preview') }),
          row.is_builtin
            ? h(NTooltip, {}, {
                trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error', disabled: true }, { default: () => t('common.delete') }),
                default: () => t('template.builtinCannotDelete'),
              })
            : h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
                trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => t('common.delete') }),
                default: () => t('template.deleteConfirm'),
              }),
        ],
      }),
  },
]

async function fetchData() {
  loading.value = true
  try {
    const { data } = await messageTemplateApi.list({ page: 1, page_size: 100 })
    templates.value = data.data.list || []
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
    type: 'text',
    content: '',
  })
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('template.create')
  resetForm()
  showModal.value = true
}

function openEdit(row: MessageTemplate) {
  editingId.value = row.id
  modalTitle.value = t('template.edit')
  Object.assign(form, {
    name: row.name,
    description: row.description,
    type: row.type,
    content: row.content || '',
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning(t('template.nameRequired'))
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      type: form.type,
      content: form.content,
    }
    if (editingId.value) {
      await messageTemplateApi.update(editingId.value, payload)
      message.success(t('template.updated'))
    } else {
      await messageTemplateApi.create(payload)
      message.success(t('template.created'))
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
    await messageTemplateApi.delete(id)
    message.success(t('template.deleted'))
    fetchData()
  } catch (err: any) {
    message.error(err.message)
  }
}

async function handlePreview(row: MessageTemplate) {
  previewLoading.value = true
  previewResult.value = ''
  showPreviewModal.value = true
  try {
    const { data } = await messageTemplateApi.preview({ content: row.content, type: row.type })
    previewResult.value = data.data.rendered
  } catch (err: any) {
    previewResult.value = ''
    message.error(t('template.previewFailed') + ': ' + err.message)
  } finally {
    previewLoading.value = false
  }
}

async function handlePreviewFromForm() {
  previewLoading.value = true
  previewResult.value = ''
  showPreviewModal.value = true
  try {
    const { data } = await messageTemplateApi.preview({ content: form.content, type: form.type })
    previewResult.value = data.data.rendered
  } catch (err: any) {
    previewResult.value = ''
    message.error(t('template.previewFailed') + ': ' + err.message)
  } finally {
    previewLoading.value = false
  }
}

const availableVariables = '{{.AlertName}} {{.Severity}} {{.Status}} {{.Labels}} {{.Annotations}} {{.FiredAt}} {{.Value}} {{.Duration}} {{.RuleName}}'

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="page-container">
    <PageHeader :title="t('template.title')" :subtitle="t('template.subtitle')">
      <template #actions>
        <n-button type="primary" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          {{ t('template.create') }}
        </n-button>
      </template>
    </PageHeader>

    <n-card :bordered="false" class="content-card">
      <n-data-table
        :loading="loading"
        :columns="columns"
        :data="templates"
        :row-key="(row: MessageTemplate) => row.id"
        :bordered="false"
        size="small"
      />
      <n-empty v-if="!loading && templates.length === 0" :description="t('template.noData')" style="padding: 40px 0" />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 600px" :bordered="false">
      <n-form label-placement="top">
        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('template.name')" required>
              <n-input v-model:value="form.name" placeholder="e.g. default-alert-template" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('template.type')">
              <n-select v-model:value="form.type" :options="typeOptions" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item :label="t('template.description')">
          <n-input v-model:value="form.description" :placeholder="t('template.description')" />
        </n-form-item>

        <!-- Available variables hint -->
        <n-collapse>
          <n-collapse-item :title="t('template.availableVariables')" name="variables">
            <n-code
              :code="availableVariables"
              language="text"
              style="font-size: 12px"
            />
          </n-collapse-item>
        </n-collapse>

        <n-form-item :label="t('template.content')" style="margin-top: 12px">
          <n-input
            v-model:value="form.content"
            type="textarea"
            :rows="12"
            :placeholder="t('common.enterContent')"
            style="font-family: monospace; font-size: 12px"
          />
        </n-form-item>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="handlePreviewFromForm" :loading="previewLoading">{{ t('template.preview') }}</n-button>
          <n-button @click="showModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">
            {{ editingId ? t('common.update') : t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Preview Modal -->
    <n-modal v-model:show="showPreviewModal" preset="card" :title="t('template.previewResult')" style="width: 600px" :bordered="false">
      <n-spin :show="previewLoading">
        <n-card :bordered="true" size="small" style="min-height: 120px" class="preview-card">
          <pre v-if="previewResult" class="preview-content" style="white-space: pre-wrap; word-break: break-word; margin: 0; font-family: inherit;">{{ previewResult }}</pre>
          <n-empty v-else-if="!previewLoading" :description="t('common.noPreview')" style="padding: 20px 0" />
        </n-card>
      </n-spin>
      <template #action>
        <n-space justify="end">
          <n-button @click="showPreviewModal = false">{{ t('common.close') }}</n-button>
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

.preview-card {
  background: var(--sre-bg-dark);
}

.preview-content {
  font-family: monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--sre-text-primary);
  line-height: 1.6;
}
</style>
