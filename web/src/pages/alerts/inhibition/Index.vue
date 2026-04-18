<template>
  <div class="inhibition-page">
    <PageHeader :title="t('inhibition.title')" :subtitle="t('inhibition.description')">
      <template #actions>
        <n-button v-if="canManage" type="primary" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          {{ t('inhibition.createRule') }}
        </n-button>
      </template>
    </PageHeader>

    <n-card>
      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: InhibitionRule) => row.id"
        @update:page="handlePageChange"
      />
    </n-card>

    <!-- Create / Edit Modal -->
    <n-modal
      v-model:show="modalVisible"
      :title="editingId ? t('inhibition.editRule') : t('inhibition.createRule')"
      preset="card"
      style="width: 640px"
      :mask-closable="false"
    >
      <n-form ref="formRef" :model="formData" label-placement="top">
        <n-form-item :label="t('inhibition.name')" path="name" :rule="{ required: true, message: t('common.required') }">
          <n-input v-model:value="formData.name" :placeholder="t('inhibition.name')" />
        </n-form-item>

        <n-form-item :label="t('common.description')">
          <n-input v-model:value="formData.description" type="textarea" :rows="2" />
        </n-form-item>

        <!-- Source Match Labels -->
        <n-form-item :label="t('inhibition.sourceMatch')">
          <LabelMatcherEditor v-model:modelValue="formData.source_matchers" :add-label="t('inhibition.addLabel')" />
        </n-form-item>

        <!-- Target Match Labels -->
        <n-form-item :label="t('inhibition.targetMatch')">
          <LabelMatcherEditor v-model:modelValue="formData.target_matchers" :add-label="t('inhibition.addLabel')" />
        </n-form-item>

        <n-form-item :label="t('inhibition.equalLabels')" :feedback="t('inhibition.equalLabelsHint')">
          <n-input v-model:value="formData.equal_labels" placeholder="alertname,namespace" />
        </n-form-item>

        <n-form-item :label="t('inhibition.isEnabled')">
          <n-switch v-model:value="formData.is_enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="modalVisible = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">{{ t('common.save') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, computed, onMounted } from 'vue'
import {
  NButton, NCard, NDataTable, NForm, NFormItem, NIcon, NInput, NModal,
  NSpace, NSwitch, NTag, useMessage, useDialog,
} from 'naive-ui'
import type { DataTableColumns, FormInst } from 'naive-ui'
import { AddOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { inhibitionRuleApi } from '@/api'
import type { InhibitionRule } from '@/types'
import PageHeader from '@/components/common/PageHeader.vue'
import LabelMatcherEditor from '@/components/common/LabelMatcherEditor.vue'
import type { LabelMatcher } from '@/components/common/LabelMatcherEditor.vue'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()

const canManage = computed(() => auth.canManage)

// ---- List state ----
const list = ref<InhibitionRule[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20

const pagination = computed(() => ({
  page: currentPage.value,
  pageSize,
  itemCount: total.value,
  showSizePicker: false,
}))

async function fetchList() {
  loading.value = true
  try {
    const res = await inhibitionRuleApi.list({ page: currentPage.value, page_size: pageSize })
    list.value = res.data.data.list || []
    total.value = res.data.data.total || 0
  } catch {
    message.error(t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchList()
}

onMounted(fetchList)

// ---- Helpers for op-encoded label encoding/decoding ----
function recordToMatchers(record: Record<string, string>): LabelMatcher[] {
  return Object.entries(record || {}).map(([key, raw]) => {
    for (const op of ['!=', '=~', '!~'] as const) {
      if (raw.startsWith(op)) return { key, op, value: raw.slice(op.length) }
    }
    return { key, op: '=', value: raw }
  })
}

function matchersToRecord(matchers: LabelMatcher[]): Record<string, string> {
  return Object.fromEntries(matchers.map(m => {
    const v = m.op === '=' ? m.value : `${m.op}${m.value}`
    return [m.key, v]
  }))
}

// ---- Modal state ----
interface InhibitionForm {
  name: string
  description: string
  source_matchers: LabelMatcher[]
  target_matchers: LabelMatcher[]
  equal_labels: string
  is_enabled: boolean
}

const modalVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInst | null>(null)

const defaultForm = (): InhibitionForm => ({
  name: '',
  description: '',
  source_matchers: [],
  target_matchers: [],
  equal_labels: '',
  is_enabled: true,
})

const formData = ref<InhibitionForm>(defaultForm())

function openCreate() {
  editingId.value = null
  formData.value = defaultForm()
  modalVisible.value = true
}

function openEdit(row: InhibitionRule) {
  editingId.value = row.id
  formData.value = {
    name: row.name,
    description: row.description,
    source_matchers: recordToMatchers(row.source_match ?? {}),
    target_matchers: recordToMatchers(row.target_match ?? {}),
    equal_labels: row.equal_labels,
    is_enabled: row.is_enabled,
  }
  modalVisible.value = true
}

async function handleSave() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const payload = {
      name: formData.value.name,
      description: formData.value.description,
      source_match: matchersToRecord(formData.value.source_matchers),
      target_match: matchersToRecord(formData.value.target_matchers),
      equal_labels: formData.value.equal_labels,
      is_enabled: formData.value.is_enabled,
    }
    if (editingId.value) {
      await inhibitionRuleApi.update(editingId.value, payload)
      message.success(t('common.updateSuccess'))
    } else {
      await inhibitionRuleApi.create(payload)
      message.success(t('common.createSuccess'))
    }
    modalVisible.value = false
    fetchList()
  } catch {
    message.error(t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

function handleDelete(row: InhibitionRule) {
  dialog.warning({
    title: t('common.confirmDelete'),
    content: `${t('common.confirmDeleteMsg')} "${row.name}"?`,
    positiveText: t('common.delete'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        await inhibitionRuleApi.delete(row.id)
        message.success(t('common.deleteSuccess'))
        fetchList()
      } catch {
        message.error(t('common.deleteFailed'))
      }
    },
  })
}

// ---- Table columns ----
function renderLabels(labels: Record<string, string>) {
  const entries = Object.entries(labels || {})
  if (!entries.length) return h('span', { style: 'color: #999' }, '-')
  return h('div', { class: 'label-tags' }, entries.map(([k, v]) =>
    h(NTag, { size: 'small', style: 'margin: 2px' }, { default: () => `${k}=${v}` })
  ))
}

const columns = computed<DataTableColumns<InhibitionRule>>(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: t('inhibition.name'), key: 'name', minWidth: 140 },
  {
    title: t('inhibition.sourceMatch'),
    key: 'source_match',
    render: (row) => renderLabels(row.source_match),
    minWidth: 160,
  },
  {
    title: t('inhibition.targetMatch'),
    key: 'target_match',
    render: (row) => renderLabels(row.target_match),
    minWidth: 160,
  },
  { title: t('inhibition.equalLabels'), key: 'equal_labels', render: (row) => row.equal_labels || '-' },
  {
    title: t('inhibition.isEnabled'),
    key: 'is_enabled',
    width: 80,
    render: (row) => h(NTag, { type: row.is_enabled ? 'success' : 'default', size: 'small' }, {
      default: () => row.is_enabled ? t('common.enabled') : t('common.disabled'),
    }),
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 120,
    render: (row) => h(NSpace, { size: 'small' }, {
      default: () => [
        canManage.value && h(NButton, {
          size: 'small', quaternary: true, onClick: () => openEdit(row),
        }, { icon: () => h(NIcon, { component: CreateOutline }) }),
        canManage.value && h(NButton, {
          size: 'small', quaternary: true, type: 'error', onClick: () => handleDelete(row),
        }, { icon: () => h(NIcon, { component: TrashOutline }) }),
      ].filter(Boolean),
    }),
  },
])
</script>

<style scoped>
.inhibition-page { padding: 0; }
.label-tags { display: flex; flex-wrap: wrap; gap: 4px; }
</style>
