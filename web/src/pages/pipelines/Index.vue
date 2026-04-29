<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDataTable, NInput, NSpace, NPopconfirm, NTag, useMessage } from 'naive-ui'
import { pipelineApi } from '@/api'
import type { EventPipeline } from '@/types/pipeline'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const search = ref('')
const list = ref<EventPipeline[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const columns = [
  { title: 'Name', key: 'name', ellipsis: { tooltip: true } },
  { title: 'Description', key: 'description', ellipsis: { tooltip: true },
    render(row: EventPipeline) { return row.description || '-' }
  },
  {
    title: 'Nodes',
    key: 'nodes',
    width: 80,
    render(row: EventPipeline) { return row.nodes?.length || 0 }
  },
  {
    title: 'Status',
    key: 'disabled',
    width: 100,
    render(row: EventPipeline) {
      return h(NTag, {
        type: row.disabled ? 'warning' : 'success',
        size: 'small',
      }, { default: () => row.disabled ? 'Disabled' : 'Active' })
    }
  },
  {
    title: 'Actions',
    key: 'actions',
    width: 200,
    render(row: EventPipeline) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleEdit(row.id) }, { default: () => 'Edit' }),
          h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
            trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => 'Delete' }),
            default: () => 'Delete this pipeline?'
          }),
        ]
      })
    },
  },
]

async function fetchList() {
  loading.value = true
  try {
    const res = await pipelineApi.list({ page: page.value, page_size: pageSize.value, search: search.value || undefined })
    list.value = res.data.data.list || []
    total.value = res.data.data.total || 0
  } catch (err: any) {
    message.error(err.message || 'Failed to load pipelines')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await pipelineApi.delete(id)
    message.success('Deleted')
    fetchList()
  } catch (err: any) {
    message.error(err.message || 'Delete failed')
  }
}

function handleEdit(id: number) {
  router.push({ name: 'PipelineEditor', params: { id } })
}

function handleCreate() {
  router.push({ name: 'PipelineEditor', params: { id: 'new' } })
}

onMounted(fetchList)
</script>

<template>
  <div style="padding: 20px">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <h2 style="margin: 0; font-size: 22px; font-weight: 600">Event Pipelines</h2>
      <NSpace>
        <NInput v-model:value="search" placeholder="Search..." clearable style="width: 200px" @update:value="fetchList" />
        <NButton type="primary" @click="handleCreate">
          + New Pipeline
        </NButton>
      </NSpace>
    </div>

    <NDataTable
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="{ page, pageSize, itemCount: total, onChange: (p: number) => { page = p; fetchList() } }"
      :row-key="(row: EventPipeline) => row.id"
    >
      <template #empty>
        <div style="padding: 40px; text-align: center; color: #999">
          No pipelines yet. Create one to customize alert processing.
        </div>
      </template>
    </NDataTable>
  </div>
</template>
