<script setup lang="ts">
import { h, ref, reactive, computed, onMounted } from 'vue'
import { useMessage, NTag, NButton, NSpace, NPopconfirm, NAvatar } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { teamApi } from '@/api'
import type { User, Team } from '@/types'
import { kvArrayToRecord } from '@/utils/format'
import { AddOutline } from '@vicons/ionicons5'
import KVEditor from '@/components/common/KVEditor.vue'

const props = defineProps<{
  /** All users from the UserManagement tab for member selection */
  allUsers: User[]
}>()

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const teamsList = ref<Team[]>([])
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)

// Team members management
const showMembersModal = ref(false)
const membersTeamId = ref<number | null>(null)
const membersTeamName = ref('')
const teamMembers = ref<User[]>([])
const selectedMemberUserId = ref<number | null>(null)
const membersLoading = ref(false)

const form = reactive({
  name: '',
  description: '',
  labels: [] as { key: string; value: string }[],
})

const columns = [
  {
    title: () => t('common.name'),
    key: 'name',
    width: 180,
    render: (row: Team) =>
      h('div', { style: 'font-weight: 500' }, row.name),
  },
  {
    title: () => t('common.description'),
    key: 'description',
    ellipsis: { tooltip: true },
  },
  {
    title: () => t('alert.labels'),
    key: 'labels',
    width: 200,
    render: (row: Team) => {
      const entries = Object.entries(row.labels || {})
      if (entries.length === 0) return h('span', { style: 'opacity: 0.3' }, '-')
      return h(NSpace, { size: 4 }, {
        default: () => entries.slice(0, 3).map(([k, v]) =>
          h(NTag, { size: 'small', bordered: false }, { default: () => `${k}=${v}` })
        ).concat(entries.length > 3 ? [h('span', { style: 'font-size: 11px; opacity: 0.35' }, `+${entries.length - 3} more`)] : []),
      })
    },
  },
  {
    title: () => t('settings.members'),
    key: 'members',
    width: 80,
    render: (row: Team) =>
      h('span', {}, row.members?.length ?? '-'),
  },
  {
    title: () => t('common.actions'),
    key: 'actions',
    width: 220,
    render: (row: Team) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'small', quaternary: true, type: 'info', onClick: () => openEdit(row) }, { default: () => t('common.edit') }),
          h(NButton, { size: 'small', quaternary: true, type: 'primary', onClick: () => openMembers(row) }, { default: () => t('settings.members') }),
          h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
            trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => t('common.delete') }),
            default: () => t('settings.deleteTeamConfirm'),
          }),
        ],
      }),
  },
]

const allUserOptions = computed(() =>
  props.allUsers.map(u => ({ label: u.display_name || u.username, value: u.id }))
)

async function fetchTeams() {
  loading.value = true
  try {
    const { data } = await teamApi.list({ page: 1, page_size: 100 })
    teamsList.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('settings.createTeam')
  Object.assign(form, { name: '', description: '', labels: [] })
  showModal.value = true
}

function openEdit(tm: Team) {
  editingId.value = tm.id
  modalTitle.value = t('settings.editTeam')
  Object.assign(form, {
    name: tm.name,
    description: tm.description,
    labels: Object.entries(tm.labels || {}).map(([key, value]) => ({ key, value })),
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning(t('settings.nameRequired'))
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      labels: kvArrayToRecord(form.labels),
    }
    if (editingId.value) {
      await teamApi.update(editingId.value, payload)
      message.success(t('settings.teamUpdated'))
    } else {
      await teamApi.create(payload)
      message.success(t('settings.teamCreated'))
    }
    showModal.value = false
    fetchTeams()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await teamApi.delete(id)
    message.success(t('settings.teamDeleted'))
    fetchTeams()
  } catch (err: any) {
    message.error(err.message)
  }
}

// Team members management
async function openMembers(tm: Team) {
  membersTeamId.value = tm.id
  membersTeamName.value = tm.name
  selectedMemberUserId.value = null
  showMembersModal.value = true
  await fetchTeamMembers(tm.id)
}

async function fetchTeamMembers(teamId: number) {
  membersLoading.value = true
  try {
    const { data } = await teamApi.listMembers(teamId)
    teamMembers.value = data.data || []
  } catch (err: any) {
    message.error(err.message)
    teamMembers.value = []
  } finally {
    membersLoading.value = false
  }
}

async function handleAddMember() {
  if (!membersTeamId.value || !selectedMemberUserId.value) return

  const existing = teamMembers.value.find(m => m.id === selectedMemberUserId.value)
  if (existing) {
    message.warning(t('settings.memberExists'))
    return
  }

  try {
    await teamApi.addMember(membersTeamId.value, selectedMemberUserId.value)
    message.success(t('settings.memberAdded'))
    selectedMemberUserId.value = null
    await fetchTeamMembers(membersTeamId.value)
    fetchTeams()
  } catch (err: any) {
    message.error(err.message)
  }
}

async function handleRemoveMember(userId: number) {
  if (!membersTeamId.value) return
  try {
    await teamApi.removeMember(membersTeamId.value, userId)
    message.success(t('settings.memberRemoved'))
    await fetchTeamMembers(membersTeamId.value)
    fetchTeams()
  } catch (err: any) {
    message.error(err.message)
  }
}

onMounted(() => {
  fetchTeams()
})
</script>

<template>
  <div>
    <div class="tab-header">
      <n-button type="primary" size="small" @click="openCreate">
        <template #icon><n-icon :component="AddOutline" /></template>
        {{ t('settings.createTeam') }}
      </n-button>
    </div>
    <n-data-table
      :loading="loading"
      :columns="columns"
      :data="teamsList"
      :row-key="(row: Team) => row.id"
      :bordered="false"
      size="small"
    />
    <n-empty v-if="!loading && teamsList.length === 0" :description="t('settings.noTeams')" style="padding: 40px 0" />

    <!-- Team Create/Edit Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 520px" :bordered="false">
      <n-form label-placement="top">
        <n-form-item :label="t('common.name')" required>
          <n-input v-model:value="form.name" placeholder="e.g. Platform Engineering" />
        </n-form-item>

        <n-form-item :label="t('common.description')">
          <n-input v-model:value="form.description" type="textarea" :placeholder="t('common.description')" :rows="2" />
        </n-form-item>

        <n-form-item :label="t('settings.labels')">
          <KVEditor v-model="form.labels" :add-label="t('settings.addTeamLabel')" />
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

    <!-- Team Members Modal -->
    <n-modal v-model:show="showMembersModal" preset="card" :title="t('settings.members') + ' - ' + membersTeamName" style="width: 520px" :bordered="false">
      <n-spin :show="membersLoading">
        <div class="members-list">
          <div v-for="member in teamMembers" :key="member.id" class="member-item">
            <n-avatar :size="28" round>{{ (member.display_name || member.username).charAt(0).toUpperCase() }}</n-avatar>
            <div class="member-info">
              <div class="member-name">{{ member.display_name || member.username }}</div>
              <div class="member-meta">{{ member.email || member.username }} &middot; {{ member.role }}</div>
            </div>
            <n-popconfirm @positive-click="handleRemoveMember(member.id)">
              <template #trigger>
                <n-button size="tiny" quaternary type="error">{{ t('common.remove') }}</n-button>
              </template>
              {{ t('settings.removeMemberConfirm') }}
            </n-popconfirm>
          </div>

          <n-empty v-if="teamMembers.length === 0" :description="t('settings.noMembers')" style="padding: 20px 0" />
        </div>
      </n-spin>

      <n-divider />

      <div style="display: flex; gap: 8px; align-items: center">
        <n-select
          v-model:value="selectedMemberUserId"
          :options="allUserOptions"
          :placeholder="t('settings.selectUserToAdd')"
          filterable
          style="flex: 1"
        />
        <n-button type="primary" @click="handleAddMember" :disabled="!selectedMemberUserId">
          {{ t('common.add') }}
        </n-button>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.tab-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: rgba(128, 128, 128, 0.06);
  border-radius: 8px;
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-name {
  font-size: 14px;
  font-weight: 500;
}

.member-meta {
  font-size: 11px;
  opacity: 0.4;
}
</style>
