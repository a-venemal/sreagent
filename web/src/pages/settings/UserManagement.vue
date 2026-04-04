<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
import { useMessage, NTag, NButton, NSpace, NAvatar } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { userApi } from '@/api'
import type { User } from '@/types'
import { formatTime } from '@/utils/format'
import { AddOutline } from '@vicons/ionicons5'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const usersList = ref<User[]>([])
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)

const form = reactive({
  username: '',
  display_name: '',
  email: '',
  phone: '',
  role: 'member' as User['role'],
  password: '',
  is_active: true,
})

const roleOptions = [
  { label: () => t('settings.admin'), value: 'admin' },
  { label: () => t('settings.teamLead'), value: 'team_lead' },
  { label: () => t('settings.member'), value: 'member' },
  { label: () => t('settings.viewer'), value: 'viewer' },
]

const columns = [
  {
    title: () => t('settings.user'),
    key: 'username',
    width: 200,
    render: (row: User) =>
      h('div', { style: 'display: flex; align-items: center; gap: 8px' }, [
        h(NAvatar, { size: 28, round: true }, { default: () => (row.display_name || row.username).charAt(0).toUpperCase() }),
        h('div', [
          h('div', { style: 'font-weight: 500' }, row.display_name || row.username),
          h('div', { style: 'font-size: 11px; opacity: 0.5' }, row.username),
        ]),
      ]),
  },
  {
    title: () => t('settings.email'),
    key: 'email',
    width: 200,
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('settings.role'),
    key: 'role',
    width: 100,
    render: (row: User) => {
      const typeMap: Record<string, 'info' | 'success' | 'warning' | 'default'> = {
        admin: 'warning',
        team_lead: 'info',
        member: 'success',
        viewer: 'default',
      }
      return h(NTag, { type: typeMap[row.role] || 'default', size: 'small' }, { default: () => row.role })
    },
  },
  {
    title: () => t('common.status'),
    key: 'is_active',
    width: 80,
    render: (row: User) =>
      h(NTag, { type: row.is_active ? 'success' : 'default', size: 'small' }, { default: () => row.is_active ? t('settings.active') : t('settings.inactive') }),
  },
  {
    title: () => t('settings.created'),
    key: 'created_at',
    width: 160,
    render: (row: User) => formatTime(row.created_at),
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 200,
    render: (row: User) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => openEdit(row) }, { default: () => t('common.edit') }),
          h(NButton, {
            size: 'small',
            quaternary: true,
            type: row.is_active ? 'warning' : 'success',
            onClick: () => handleToggleActive(row),
          }, { default: () => row.is_active ? t('settings.deactivate') : t('settings.activate') }),
        ],
      }),
  },
]

async function fetchUsers() {
  loading.value = true
  try {
    const { data } = await userApi.list({ page: 1, page_size: 200 })
    usersList.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('settings.createUser')
  Object.assign(form, {
    username: '',
    display_name: '',
    email: '',
    phone: '',
    role: 'member',
    password: '',
    is_active: true,
  })
  showModal.value = true
}

function openEdit(u: User) {
  editingId.value = u.id
  modalTitle.value = t('settings.editUser')
  Object.assign(form, {
    username: u.username,
    display_name: u.display_name,
    email: u.email,
    phone: u.phone,
    role: u.role,
    password: '',
    is_active: u.is_active,
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.username.trim()) {
    message.warning(t('settings.usernameRequired'))
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      const payload: Partial<User> = {
        username: form.username,
        display_name: form.display_name,
        email: form.email,
        phone: form.phone,
        role: form.role,
      }
      await userApi.update(editingId.value, payload)
      if (form.password.trim()) {
        await userApi.changePassword(editingId.value, { password: form.password })
      }
      message.success(t('settings.userUpdated'))
    } else {
      if (!form.password.trim()) {
        message.warning(t('settings.passwordRequired'))
        saving.value = false
        return
      }
      await userApi.create({
        username: form.username,
        display_name: form.display_name,
        email: form.email,
        phone: form.phone,
        role: form.role,
        password: form.password,
        is_active: form.is_active,
      })
      message.success(t('settings.userCreated'))
    }
    showModal.value = false
    fetchUsers()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function handleToggleActive(user: User) {
  try {
    await userApi.toggleActive(user.id, !user.is_active)
    message.success(user.is_active ? t('settings.userDeactivated') : t('settings.userActivated'))
    fetchUsers()
  } catch (err: any) {
    message.error(err.message)
  }
}

// Expose usersList for other tabs (e.g. team members, biz group members)
defineExpose({ usersList, fetchUsers })

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div>
    <div class="tab-header">
      <n-button type="primary" size="small" @click="openCreate">
        <template #icon><n-icon :component="AddOutline" /></template>
        {{ t('settings.createUser') }}
      </n-button>
    </div>
    <n-data-table
      :loading="loading"
      :columns="columns"
      :data="usersList"
      :row-key="(row: User) => row.id"
      :bordered="false"
      size="small"
    />
    <n-empty v-if="!loading && usersList.length === 0" :description="t('settings.noUsers')" style="padding: 40px 0" />

    <!-- User Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 520px" :bordered="false">
      <n-form label-placement="top">
        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('auth.username')" required>
              <n-input v-model:value="form.username" placeholder="e.g. john.doe" :disabled="!!editingId" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('settings.displayName')">
              <n-input v-model:value="form.display_name" placeholder="e.g. John Doe" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('settings.email')">
              <n-input v-model:value="form.email" placeholder="john@example.com" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('settings.phone')">
              <n-input v-model:value="form.phone" placeholder="+86 ..." />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :x-gap="12" :cols="2">
          <n-gi>
            <n-form-item :label="t('settings.role')">
              <n-select v-model:value="form.role" :options="roleOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="editingId ? t('settings.newPasswordKeep') : t('auth.password')" :required="!editingId">
              <n-input v-model:value="form.password" type="password" :placeholder="t('auth.enterPassword')" show-password-on="click" />
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
.tab-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
