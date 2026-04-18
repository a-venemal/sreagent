<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import type { TreeOption } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { bizGroupApi } from '@/api'
import type { User, BizGroup } from '@/types'
import { kvArrayToRecord } from '@/utils/format'
import { AddOutline } from '@vicons/ionicons5'
import KVEditor from '@/components/common/KVEditor.vue'
import LabelMatcherEditor from '@/components/common/LabelMatcherEditor.vue'
import type { LabelMatcher } from '@/components/common/LabelMatcherEditor.vue'

const props = defineProps<{
  /** All users from the UserManagement tab for member selection */
  allUsers: User[]
}>()

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const list = ref<BizGroup[]>([])
const selectedId = ref<number | null>(null)
const selectedGroup = ref<BizGroup | null>(null)
const members = ref<any[]>([])
const membersLoading = ref(false)
const showModal = ref(false)
const modalTitle = ref('')
const editingId = ref<number | null>(null)
const saving = ref(false)
const showAddMemberModal = ref(false)
const selectedMemberUserId = ref<number | null>(null)
const selectedMemberRole = ref<string>('member')

const form = reactive({
  name: '',
  description: '',
  labels: [] as { key: string; value: string }[],
  match_labels: [] as LabelMatcher[],
})

const memberRoleOptions = [
  { label: 'Admin', value: 'admin' },
  { label: 'Member', value: 'member' },
]

const allUserOptions = computed(() =>
  props.allUsers.map(u => ({ label: u.display_name || u.username, value: u.id }))
)

/**
 * Convert a flat list of BizGroup (with "/" in name) into NTree options.
 * e.g. "DBA/MySQL" -> DBA > MySQL
 */
function buildTree(groups: BizGroup[]): TreeOption[] {
  const root: TreeOption[] = []
  const nodeMap = new Map<string, TreeOption>()

  const sorted = [...groups].sort((a, b) => a.name.localeCompare(b.name))

  for (const g of sorted) {
    const parts = g.name.split('/')
    let currentPath = ''
    let parentChildren = root

    for (let i = 0; i < parts.length; i++) {
      currentPath = currentPath ? `${currentPath}/${parts[i]}` : parts[i]
      const isLeaf = i === parts.length - 1

      if (isLeaf) {
        const node: TreeOption = {
          key: g.id,
          label: parts[i],
          children: undefined,
        }
        const existing = nodeMap.get(currentPath)
        if (existing) {
          existing.key = g.id
        } else {
          parentChildren.push(node)
          nodeMap.set(currentPath, node)
        }
      } else {
        if (!nodeMap.has(currentPath)) {
          const intermediate: TreeOption = {
            key: `__path__${currentPath}`,
            label: parts[i],
            children: [],
          }
          parentChildren.push(intermediate)
          nodeMap.set(currentPath, intermediate)
        }
        const parent = nodeMap.get(currentPath)!
        if (!parent.children) parent.children = []
        parentChildren = parent.children
      }
    }
  }

  return root
}

const treeOptions = computed(() => buildTree(list.value))

async function fetchList() {
  loading.value = true
  try {
    const { data } = await bizGroupApi.list({ page: 1, page_size: 500 })
    list.value = data.data.list || []
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

function handleSelect(keys: Array<string | number>) {
  const key = keys[0]
  if (typeof key === 'string' && key.startsWith('__path__')) return
  selectedId.value = key as number
  const group = list.value.find(g => g.id === key)
  selectedGroup.value = group || null
  if (group) fetchMembers(group.id)
}

async function fetchMembers(groupId: number) {
  membersLoading.value = true
  try {
    const { data } = await bizGroupApi.listMembers(groupId)
    members.value = data.data || []
  } catch (err: any) {
    message.error(err.message)
    members.value = []
  } finally {
    membersLoading.value = false
  }
}

function recordToMatchers(record: Record<string, string> | undefined): LabelMatcher[] {
  return Object.entries(record || {}).map(([key, raw]) => {
    for (const op of ['!=', '=~', '!~'] as const) {
      if (raw.startsWith(op)) return { key, op, value: raw.slice(op.length) }
    }
    return { key, op: '=' as const, value: raw }
  })
}

function matchersToRecord(matchers: LabelMatcher[]): Record<string, string> {
  return Object.fromEntries(matchers.map(m => {
    const v = m.op === '=' ? m.value : `${m.op}${m.value}`
    return [m.key, v]
  }))
}

function openCreate() {
  editingId.value = null
  modalTitle.value = t('bizGroup.create')
  Object.assign(form, { name: '', description: '', labels: [], match_labels: [] })
  showModal.value = true
}

function openEdit() {
  if (!selectedGroup.value) return
  const g = selectedGroup.value
  editingId.value = g.id
  modalTitle.value = t('bizGroup.edit')
  Object.assign(form, {
    name: g.name,
    description: g.description,
    labels: Object.entries(g.labels || {}).map(([key, value]) => ({ key, value })),
    match_labels: recordToMatchers(g.match_labels),
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
      match_labels: matchersToRecord(form.match_labels),
    }
    if (editingId.value) {
      await bizGroupApi.update(editingId.value, payload)
      message.success(t('bizGroup.updated'))
    } else {
      await bizGroupApi.create(payload)
      message.success(t('bizGroup.created'))
    }
    showModal.value = false
    fetchList()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!selectedGroup.value) return
  try {
    await bizGroupApi.delete(selectedGroup.value.id)
    message.success(t('bizGroup.deleted'))
    selectedGroup.value = null
    selectedId.value = null
    members.value = []
    fetchList()
  } catch (err: any) {
    message.error(err.message)
  }
}

function openAddMember() {
  selectedMemberUserId.value = null
  selectedMemberRole.value = 'member'
  showAddMemberModal.value = true
}

async function handleAddMember() {
  if (!selectedGroup.value || !selectedMemberUserId.value) return
  try {
    await bizGroupApi.addMember(selectedGroup.value.id, {
      user_id: selectedMemberUserId.value,
      role: selectedMemberRole.value,
    })
    message.success(t('settings.memberAdded'))
    showAddMemberModal.value = false
    fetchMembers(selectedGroup.value.id)
  } catch (err: any) {
    message.error(err.message)
  }
}

async function handleRemoveMember(userId: number) {
  if (!selectedGroup.value) return
  try {
    await bizGroupApi.removeMember(selectedGroup.value.id, userId)
    message.success(t('settings.memberRemoved'))
    fetchMembers(selectedGroup.value.id)
  } catch (err: any) {
    message.error(err.message)
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="biz-group-layout">
    <!-- Left panel: tree -->
    <div class="biz-group-left">
      <div class="tab-header">
        <n-button type="primary" size="small" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          {{ t('bizGroup.create') }}
        </n-button>
      </div>
      <n-spin :show="loading">
        <n-tree
          :data="treeOptions"
          :selected-keys="selectedId ? [selectedId] : []"
          selectable
          block-line
          @update:selected-keys="handleSelect"
          default-expand-all
          style="min-height: 200px"
        />
        <n-empty v-if="!loading && list.length === 0" :description="t('bizGroup.noData')" style="padding: 40px 0" />
      </n-spin>
    </div>

    <!-- Right panel: selected group detail -->
    <div class="biz-group-right">
      <template v-if="selectedGroup">
        <div class="biz-group-detail-header">
          <h3 style="margin: 0; font-size: 18px; font-weight: 600">{{ selectedGroup.name }}</h3>
          <n-space size="small">
            <n-button size="small" type="info" quaternary @click="openEdit">{{ t('common.edit') }}</n-button>
            <n-popconfirm @positive-click="handleDelete">
              <template #trigger>
                <n-button size="small" type="error" quaternary>{{ t('common.delete') }}</n-button>
              </template>
              {{ t('bizGroup.deleteConfirm') }}
            </n-popconfirm>
          </n-space>
        </div>

        <n-descriptions bordered :column="1" label-placement="left" size="small" style="margin-bottom: 16px">
          <n-descriptions-item :label="t('common.name')">{{ selectedGroup.name }}</n-descriptions-item>
          <n-descriptions-item :label="t('common.description')">{{ selectedGroup.description || '-' }}</n-descriptions-item>
          <n-descriptions-item :label="t('alert.labels')">
            <n-space size="small" v-if="Object.keys(selectedGroup.labels || {}).length > 0">
              <n-tag v-for="(v, k) in selectedGroup.labels" :key="k" size="small" :bordered="false">{{ k }}={{ v }}</n-tag>
            </n-space>
            <span v-else style="opacity: 0.3">-</span>
          </n-descriptions-item>
          <n-descriptions-item :label="t('bizGroup.matchLabels')">
            <n-space size="small" v-if="Object.keys(selectedGroup.match_labels || {}).length > 0">
              <n-tag v-for="(v, k) in selectedGroup.match_labels" :key="k" size="small" type="info" :bordered="false">{{ k }}{{ v.startsWith('!=') || v.startsWith('=~') || v.startsWith('!~') ? v : '=' + v }}</n-tag>
            </n-space>
            <span v-else style="opacity: 0.3">-</span>
          </n-descriptions-item>
        </n-descriptions>

        <div class="biz-members-header">
          <h4 style="margin: 0; font-size: 15px; font-weight: 500">{{ t('settings.members') }}</h4>
          <n-button size="small" type="primary" @click="openAddMember">
            <template #icon><n-icon :component="AddOutline" /></template>
            {{ t('bizGroup.addMember') }}
          </n-button>
        </div>

        <n-spin :show="membersLoading">
          <div class="members-list">
            <div v-for="member in members" :key="member.id" class="member-item">
              <n-avatar :size="28" round>{{ (member.display_name || member.username).charAt(0).toUpperCase() }}</n-avatar>
              <div class="member-info">
                <div class="member-name">{{ member.display_name || member.username }}</div>
                <div class="member-meta">{{ member.email || member.username }}</div>
              </div>
              <n-tag :type="member.role === 'admin' ? 'warning' : 'default'" size="small">{{ member.role || 'member' }}</n-tag>
              <n-popconfirm @positive-click="handleRemoveMember(member.id)">
                <template #trigger>
                  <n-button size="tiny" quaternary type="error">{{ t('common.remove') }}</n-button>
                </template>
                {{ t('settings.removeMemberConfirm') }}
              </n-popconfirm>
            </div>
            <n-empty v-if="!membersLoading && members.length === 0" :description="t('settings.noMembers')" style="padding: 20px 0" />
          </div>
        </n-spin>
      </template>

      <n-empty v-else :description="t('bizGroup.selectGroup')" style="padding: 80px 0" />
    </div>

    <!-- Business Group Create/Edit Modal -->
    <n-modal v-model:show="showModal" preset="card" :title="modalTitle" style="width: 520px" :bordered="false">
      <n-form label-placement="top">
        <n-form-item :label="t('common.name')" required>
          <n-input v-model:value="form.name" placeholder="e.g. DBA/MySQL" />
          <template #feedback>
            <span style="font-size: 11px; opacity: 0.45">{{ t('bizGroup.nameHint') }}</span>
          </template>
        </n-form-item>

        <n-form-item :label="t('common.description')">
          <n-input v-model:value="form.description" type="textarea" :placeholder="t('common.description')" :rows="2" />
        </n-form-item>

        <n-form-item :label="t('settings.labels')">
          <KVEditor v-model="form.labels" :add-label="t('settings.addTeamLabel')" />
        </n-form-item>

        <n-form-item :label="t('bizGroup.matchLabels')">
          <template #feedback>
            <span style="font-size: 11px; opacity: 0.45">{{ t('bizGroup.matchLabelsDesc') }}</span>
          </template>
          <LabelMatcherEditor v-model:modelValue="form.match_labels" :add-label="t('bizGroup.addMatchLabel')" />
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

    <!-- Add Member Modal -->
    <n-modal v-model:show="showAddMemberModal" preset="card" :title="t('bizGroup.addMember')" style="width: 420px" :bordered="false">
      <n-form label-placement="top">
        <n-form-item :label="t('settings.user')" required>
          <n-select
            v-model:value="selectedMemberUserId"
            :options="allUserOptions"
            :placeholder="t('settings.selectUser')"
            filterable
          />
        </n-form-item>
        <n-form-item :label="t('settings.role')">
          <n-select
            v-model:value="selectedMemberRole"
            :options="memberRoleOptions"
          />
        </n-form-item>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showAddMemberModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :disabled="!selectedMemberUserId" @click="handleAddMember">
            {{ t('common.add') }}
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

.biz-group-layout {
  display: flex;
  gap: 24px;
  min-height: 400px;
}

.biz-group-left {
  flex: 0 0 33%;
  max-width: 33%;
  border-right: 1px solid rgba(128, 128, 128, 0.12);
  padding-right: 20px;
}

.biz-group-right {
  flex: 1;
  min-width: 0;
}

.biz-group-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.biz-members-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
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
