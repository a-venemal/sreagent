<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDataTable, NInput, NSpace, NPopconfirm, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { dashboardV2Api } from '@/api'
import type { DashboardV2 } from '@/types/dashboard'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const search = ref('')
const list = ref<DashboardV2[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const columns = [
  { title: 'Name', key: 'name', ellipsis: { tooltip: true } },
  { title: 'Description', key: 'description', ellipsis: { tooltip: true } },
  {
    title: 'Actions',
    key: 'actions',
    width: 200,
    render(row: DashboardV2) {
      return row.id
    },
  },
]

async function fetchList() {
  loading.value = true
  try {
    const res = await dashboardV2Api.list({ page: page.value, page_size: pageSize.value, search: search.value || undefined })
    list.value = res.data.data.list || []
    total.value = res.data.data.total || 0
  } catch (err: any) {
    message.error(err.message || 'Failed to load dashboards')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await dashboardV2Api.delete(id)
    message.success('Deleted')
    fetchList()
  } catch (err: any) {
    message.error(err.message || 'Delete failed')
  }
}

function handleView(id: number) {
  router.push({ name: 'DashboardV2View', params: { id } })
}

onMounted(fetchList)
</script>

<template>
  <div style="padding: 20px">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <h2 style="margin: 0; font-size: 22px; font-weight: 600">Dashboards</h2>
      <NSpace>
        <NInput v-model:value="search" placeholder="Search..." clearable style="width: 200px" @update:value="fetchList" />
        <NButton type="primary" @click="router.push({ name: 'DashboardV2View', params: { id: 'new' } })">
          + New Dashboard
        </NButton>
      </NSpace>
    </div>

    <NDataTable
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="{ page, pageSize, itemCount: total, onChange: (p: number) => { page = p; fetchList() } }"
      :row-key="(row: DashboardV2) => row.id"
    >
      <template #empty>
        <div style="padding: 40px; text-align: center; color: #999">
          No dashboards yet. Create one to get started.
        </div>
      </template>
    </NDataTable>
  </div>
</template>
