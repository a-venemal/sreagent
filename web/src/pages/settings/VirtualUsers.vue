<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton, NPopconfirm, NAvatar } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { userApi } from '@/api'
import type { User } from '@/types'
import { AddOutline } from '@vicons/ionicons5'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const list = ref<User[]>([])
const showModal = ref(false)
const saving = ref(false)

const form = reactive({
  display_name: '',
  user_type: 'bot' as 'bot' | 'channel',
  notify_target: '',
})

const columns = [
  {
    title: () => t('settings.displayName'),
    key: 'display_name',
    width: 160,
    render: (row: User) =>
      h('div', { style: 'display: flex; align-items: center; gap: 8px' }, [
        h(NAvatar, { size: 28, round: true }, { default: () => (row.display_name || row.username).charAt(0).toUpperCase() }),
        h('div', { style: 'font-weight: 500' }, row.display_name || row.username),
      ]),
  },
  {
    title: () => t('settings.userType'),
    key: 'user_type',
    width: 120,
    render: (row: User) =>
      h(NTag, {
        type: row.user_type === 'bot' ? 'info' : row.user_type === 'channel' ? 'warning' : 'default',
        size: 'small',
      }, { default: () => row.user_type === 'bot' ? t('settings.botType') : row.user_type === 'channel' ? t('settings.virtualChannelType') : row.user_type || '-' }),
  },
  {
    title: () => t('auth.username'),
    key: 'username',
    width: 160,
    render: (row: User) => h('span', { style: 'font-size: 12px; opacity: 0.6; font-family: monospace' }, row.username),
  },
  {
    title: () => t('settings.notifyTarget'),
    key: 'notify_target',
    ellipsis: { tooltip: true },
    render: (row: User) => h('span', { style: 'font-size: 12px; opacity: 0.55' }, row.notify_target || '-'),
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 100,
    render: (row: User) =>
      h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
        trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => t('common.delete') }),
        default: () => t('settings.deleteVirtualConfirm'),
      }),
  },
]

async function fetchList() {
  loading.value = true
  try {
    const { data } = await userApi.list({ page: 1, page_size: 200 })
    list.value = (data.data.list || []).filter(u => u.user_type === 'bot' || u.user_type === 'channel')
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { display_name: '', user_type: 'bot', notify_target: '' })
  showModal.value = true
}

async function handleSave() {
  if (!form.display_name.trim()) {
    message.warning(t('settings.displayNameRequired'))
    return
  }
  saving.value = true
  try {
    const username = `virtual_${form.display_name.toLowerCase().replace(/\s+/g, '_').replace(/[^a-z0-9_]/g, '')}_${Date.now().toString(36)}`
    await userApi.createVirtual({
      username,
      display_name: form.display_name,
      user_type: form.user_type,
      notify_target: form.notify_target || undefined,
    })
    message.success(t('settings.virtualUserCreated'))
    showModal.value = false
    fetchList()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await userApi.delete(id)
    message.success(t('settings.userDeleted'))
    fetchList()
  } catch (err: any) {
    message.error(err.message)
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div>
    <div class="tab-header">
      <n-button type="primary" size="small" @click="openCreate">
        <template #icon><n-icon :component="AddOutline" /></template>
        {{ t('settings.createVirtual') }}
      </n-button>
    </div>
    <n-data-table
      :loading="loading"
      :columns="columns"
      :data="list"
      :row-key="(row: User) => row.id"
      :bordered="false"
      size="small"
    />
    <n-empty v-if="!loading && list.length === 0" :description="t('settings.noVirtualUsers')" style="padding: 40px 0" />

    <!-- Virtual User Create Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="t('settings.createVirtual')" style="width: 480px" :bordered="false">
      <n-form label-placement="top">
        <n-form-item :label="t('settings.displayName')" required>
          <n-input v-model:value="form.display_name" :placeholder="t('settings.displayNamePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('settings.userType')">
          <n-radio-group v-model:value="form.user_type">
            <n-space>
              <n-radio value="bot">{{ t('settings.botType') }}</n-radio>
              <n-radio value="channel">{{ t('settings.virtualChannelType') }}</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>
        <n-form-item :label="t('settings.notifyTarget')">
          <n-input
            v-model:value="form.notify_target"
            type="textarea"
            :rows="3"
            :placeholder="form.user_type === 'bot' ? t('settings.botNotifyTargetHint') : t('settings.channelNotifyTargetHint')"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">{{ t('common.create') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.tab-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
