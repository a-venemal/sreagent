<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { scheduleApi, teamApi, userApi } from '@/api'
import type { Schedule, Team, User, OnCallShift } from '@/types'

import ScheduleSidebar from './ScheduleSidebar.vue'
import ScheduleModal from './ScheduleModal.vue'
import ShiftModal from './ShiftModal.vue'
import ParticipantsList from './ParticipantsList.vue'

const message = useMessage()
const { t } = useI18n()

// ===== Color palette for users =====
const userColors = ['#18a058','#4098fc','#f0a020','#d03050','#9c27b0','#00bcd4','#ff5722','#607d8b']
const userColorMap = ref<Map<number, string>>(new Map())

function getUserColor(userId: number): string {
  if (!userColorMap.value.has(userId)) {
    const idx = userColorMap.value.size % userColors.length
    userColorMap.value.set(userId, userColors[idx])
  }
  return userColorMap.value.get(userId)!
}

function getUserName(userId: number): string {
  const u = users.value.find(u => u.id === userId)
  return u ? (u.display_name || u.username) : `#${userId}`
}

// ===== Data =====
const loading = ref(false)
const schedules = ref<Schedule[]>([])
const teams = ref<Team[]>([])
const users = ref<User[]>([])
const onCallMap = ref<Record<number, User | null>>({})
const selectedSchedule = ref<Schedule | null>(null)

// ===== Week navigation =====
function getMonday(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1)
  d.setDate(diff)
  d.setHours(0, 0, 0, 0)
  return d
}

const currentWeekStart = ref(getMonday(new Date()))

function prevWeek() {
  currentWeekStart.value = new Date(currentWeekStart.value.getTime() - 7 * 86400000)
}
function nextWeek() {
  currentWeekStart.value = new Date(currentWeekStart.value.getTime() + 7 * 86400000)
}
function goToday() {
  currentWeekStart.value = getMonday(new Date())
}

const weekDays = computed(() =>
  Array.from({ length: 7 }, (_, i) => {
    const d = new Date(currentWeekStart.value)
    d.setDate(d.getDate() + i)
    return d
  })
)

const weekRangeLabel = computed(() => {
  const start = weekDays.value[0]
  const end = weekDays.value[6]
  const fmt = (d: Date) => `${d.getMonth() + 1}/${d.getDate()}`
  return `${start.getFullYear()} ${fmt(start)} - ${fmt(end)}`
})

// Time axis
const timeLabels = Array.from({ length: 13 }, (_, i) => `${(i * 2).toString().padStart(2, '0')}:00`)

// Current time line
const currentTimePercent = ref(0)
function updateCurrentTime() {
  const now = new Date()
  const minutes = now.getHours() * 60 + now.getMinutes()
  currentTimePercent.value = (minutes / (24 * 60)) * 100
}
let currentTimeInterval: ReturnType<typeof setInterval>
onMounted(() => { updateCurrentTime(); currentTimeInterval = setInterval(updateCurrentTime, 60000) })
onUnmounted(() => clearInterval(currentTimeInterval))

function isToday(d: Date): boolean {
  const now = new Date()
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate()
}

// ===== Shifts =====
const shifts = ref<OnCallShift[]>([])
const shiftsLoading = ref(false)

async function fetchShifts() {
  if (!selectedSchedule.value) return
  shiftsLoading.value = true
  try {
    const start = weekDays.value[0].toISOString()
    const end = new Date(weekDays.value[6].getTime() + 86400000).toISOString()
    const { data } = await scheduleApi.listShifts(selectedSchedule.value.id, { start, end })
    shifts.value = data.data || []
  } catch {
    shifts.value = []
  } finally {
    shiftsLoading.value = false
  }
}

watch([selectedSchedule, currentWeekStart], () => {
  fetchShifts()
}, { immediate: false })

function getShiftsForDay(day: Date): OnCallShift[] {
  const dayStart = new Date(day)
  dayStart.setHours(0, 0, 0, 0)
  const dayEnd = new Date(day)
  dayEnd.setHours(23, 59, 59, 999)
  return shifts.value.filter(s => {
    const start = new Date(s.start_time)
    const end = new Date(s.end_time)
    return start <= dayEnd && end >= dayStart
  })
}

function shiftStyle(shift: OnCallShift, day: Date): Record<string, string> {
  const dayStart = new Date(day)
  dayStart.setHours(0, 0, 0, 0)
  const dayEnd = new Date(day)
  dayEnd.setHours(24, 0, 0, 0)

  const shiftStart = new Date(shift.start_time)
  const shiftEnd = new Date(shift.end_time)

  const effectiveStart = shiftStart < dayStart ? dayStart : shiftStart
  const effectiveEnd = shiftEnd > dayEnd ? dayEnd : shiftEnd

  const minutesInDay = 24 * 60
  const startMin = (effectiveStart.getTime() - dayStart.getTime()) / 60000
  const endMin = (effectiveEnd.getTime() - dayStart.getTime()) / 60000

  const top = (startMin / minutesInDay) * 100
  const height = Math.max(((endMin - startMin) / minutesInDay) * 100, 2)

  const color = getUserColor(shift.user_id)
  return {
    top: `${top}%`,
    height: `${height}%`,
    position: 'absolute',
    width: '94%',
    left: '3%',
    backgroundColor: color + '22',
    borderLeft: `3px solid ${color}`,
    borderRadius: '4px',
    padding: '2px 4px',
    overflow: 'hidden',
    cursor: 'pointer',
    boxSizing: 'border-box',
    boxShadow: `0 1px 4px ${color}33`,
    zIndex: '1',
  }
}

function formatShiftTime(shift: OnCallShift): string {
  const s = new Date(shift.start_time)
  const e = new Date(shift.end_time)
  const fmt = (d: Date) => `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  return `${fmt(s)}-${fmt(e)}`
}

// ===== Generate Shifts =====
const showGenerateModal = ref(false)
const generateWeeks = ref(4)
const generating = ref(false)

async function handleGenerateShifts() {
  if (!selectedSchedule.value) return
  generating.value = true
  try {
    await scheduleApi.generateShifts(selectedSchedule.value.id, { weeks: generateWeeks.value })
    message.success(t('schedule.shiftsGenerated'))
    showGenerateModal.value = false
    fetchShifts()
  } catch (err: any) {
    message.error(err.message)
  } finally {
    generating.value = false
  }
}

// ===== Component refs =====
const scheduleModalRef = ref<InstanceType<typeof ScheduleModal> | null>(null)
const shiftModalRef = ref<InstanceType<typeof ShiftModal> | null>(null)
const participantsRef = ref<InstanceType<typeof ParticipantsList> | null>(null)
const activeConfigTab = ref('config')

// ===== Data Fetching =====
async function fetchSchedules() {
  loading.value = true
  try {
    const { data } = await scheduleApi.list({ page: 1, page_size: 100 })
    schedules.value = data.data.list || []
    for (const s of schedules.value) {
      fetchOnCall(s.id)
    }
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

async function fetchOnCall(scheduleId: number) {
  try {
    const { data } = await scheduleApi.getCurrentOnCall(scheduleId)
    onCallMap.value[scheduleId] = data.data
  } catch {
    onCallMap.value[scheduleId] = null
  }
}

async function fetchTeams() {
  try {
    const { data } = await teamApi.list({ page: 1, page_size: 100 })
    teams.value = data.data.list || []
  } catch { /* silent */ }
}

async function fetchUsers() {
  try {
    const { data } = await userApi.list({ page: 1, page_size: 200 })
    users.value = data.data.list || []
  } catch { /* silent */ }
}

function selectSchedule(s: Schedule) {
  selectedSchedule.value = s
  activeConfigTab.value = 'config'
  participantsRef.value?.fetchParticipants()
  fetchShifts()
}

async function handleDeleteSchedule(id: number) {
  try {
    await scheduleApi.delete(id)
    message.success(t('schedule.scheduleDeleted'))
    if (selectedSchedule.value?.id === id) {
      selectedSchedule.value = null
    }
    fetchSchedules()
  } catch (err: any) {
    message.error(err.message)
  }
}

function handleCalendarDayClick(day: Date, event: MouseEvent) {
  if (!selectedSchedule.value) return
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const relY = event.clientY - rect.top
  const fraction = relY / rect.height
  const totalMinutes = fraction * 24 * 60
  const hour = Math.floor(totalMinutes / 60)
  shiftModalRef.value?.openCreate(day, hour)
}

function handleScheduleSaved() {
  fetchSchedules()
}

function handleShiftSaved() {
  fetchShifts()
}

onMounted(() => {
  fetchSchedules()
  fetchTeams()
  fetchUsers()
})
</script>

<template>
  <div class="schedule-page">
    <div class="schedule-layout">
      <!-- Left Sidebar -->
      <ScheduleSidebar
        :schedules="schedules"
        :loading="loading"
        :selected-id="selectedSchedule?.id ?? null"
        :on-call-map="onCallMap"
        @select="selectSchedule"
        @create="scheduleModalRef?.openCreate()"
      />

      <!-- Right Detail Panel -->
      <div class="schedule-detail">
        <template v-if="selectedSchedule">
          <!-- Top bar -->
          <div class="detail-topbar">
            <div class="detail-title">
              <h2>{{ selectedSchedule.name }}</h2>
              <n-tag v-if="selectedSchedule.team" size="small" :bordered="false" type="info">{{ selectedSchedule.team?.name || '' }}</n-tag>
            </div>
            <n-space size="small">
              <n-button size="small" @click="scheduleModalRef?.openEdit(selectedSchedule!)">{{ t('common.edit') }}</n-button>
              <n-popconfirm @positive-click="handleDeleteSchedule(selectedSchedule.id)">
                <template #trigger>
                  <n-button size="small" type="error" quaternary>{{ t('common.delete') }}</n-button>
                </template>
                {{ t('schedule.deleteConfirm') }}
              </n-popconfirm>
              <n-button size="small" type="primary" @click="showGenerateModal = true">
                {{ t('schedule.generateShifts') }}
              </n-button>
            </n-space>
          </div>

          <!-- Week navigation -->
          <div class="week-nav">
            <n-button size="small" quaternary @click="prevWeek">&#x2039;</n-button>
            <span class="week-label">{{ weekRangeLabel }}</span>
            <n-button size="small" quaternary @click="nextWeek">&#x203A;</n-button>
            <n-button size="tiny" @click="goToday" style="margin-left: 8px">{{ t('schedule.today') }}</n-button>
            <n-button size="tiny" type="primary" @click="shiftModalRef?.openCreate()" style="margin-left: 8px">
              + {{ t('schedule.newShift') }}
            </n-button>
          </div>

          <!-- Calendar grid -->
          <div class="calendar-container">
            <n-spin :show="shiftsLoading">
              <div class="calendar-grid">
                <!-- Header row -->
                <div class="cal-header-row">
                  <div class="cal-time-gutter" />
                  <div
                    v-for="(day, i) in weekDays"
                    :key="i"
                    class="cal-day-header"
                    :class="{ today: isToday(day) }"
                  >
                    <span class="cal-day-name">{{ ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'][i] }}</span>
                    <span class="cal-day-num" :class="{ today: isToday(day) }">{{ day.getDate() }}</span>
                  </div>
                </div>

                <!-- Body -->
                <div class="cal-body">
                  <div class="cal-time-gutter-body">
                    <div
                      v-for="label in timeLabels"
                      :key="label"
                      class="cal-time-label"
                    >{{ label }}</div>
                  </div>

                  <div
                    v-for="(day, dayIdx) in weekDays"
                    :key="dayIdx"
                    class="cal-day-col"
                    @click.self="handleCalendarDayClick(day, $event)"
                  >
                    <div
                      v-for="h in 24"
                      :key="h"
                      class="cal-hour-line"
                      :style="{ top: `${((h - 1) / 24) * 100}%` }"
                    />

                    <div
                      v-if="isToday(day)"
                      class="current-time-line"
                      :style="{ top: `${currentTimePercent}%` }"
                    />

                    <div
                      v-for="shift in getShiftsForDay(day)"
                      :key="shift.id"
                      class="shift-block"
                      :style="shiftStyle(shift, day)"
                      @click.stop="shiftModalRef?.openEdit(shift)"
                    >
                      <div class="shift-user" :style="{ color: getUserColor(shift.user_id) }">
                        {{ getUserName(shift.user_id) }}
                      </div>
                      <div class="shift-time">{{ formatShiftTime(shift) }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </n-spin>
          </div>

          <!-- Bottom Config Tabs -->
          <div class="config-tabs">
            <n-tabs v-model:value="activeConfigTab" type="line" size="small" animated>
              <n-tab-pane name="config" :tab="t('schedule.tabConfig')">
                <div class="config-form">
                  <n-descriptions bordered :column="2" label-placement="left" size="small">
                    <n-descriptions-item :label="t('schedule.rotationType')">
                      <n-tag size="small" type="info">{{ selectedSchedule.rotation_type }}</n-tag>
                    </n-descriptions-item>
                    <n-descriptions-item :label="t('schedule.timezone')">{{ selectedSchedule.timezone }}</n-descriptions-item>
                    <n-descriptions-item :label="t('schedule.handoffTime')">{{ selectedSchedule.handoff_time }}</n-descriptions-item>
                    <n-descriptions-item :label="t('schedule.team')">{{ selectedSchedule.team?.name || '-' }}</n-descriptions-item>
                    <n-descriptions-item :label="t('schedule.severityFilter')">
                      <span v-if="selectedSchedule.severity_filter">{{ selectedSchedule.severity_filter }}</span>
                      <span v-else style="opacity: 0.4">{{ t('schedule.allSeverities') }}</span>
                    </n-descriptions-item>
                    <n-descriptions-item :label="t('common.status')">
                      <n-tag :type="selectedSchedule.is_enabled ? 'success' : 'default'" size="small">
                        {{ selectedSchedule.is_enabled ? t('common.active') : t('common.disabled') }}
                      </n-tag>
                    </n-descriptions-item>
                  </n-descriptions>
                </div>
              </n-tab-pane>

              <n-tab-pane name="members" :tab="t('schedule.tabMembers')">
                <ParticipantsList
                  ref="participantsRef"
                  :schedule-id="selectedSchedule.id"
                  :users="users"
                  :get-user-color="getUserColor"
                  :get-user-name="getUserName"
                />
              </n-tab-pane>
            </n-tabs>
          </div>
        </template>

        <!-- No schedule selected -->
        <n-empty v-else :description="t('schedule.selectSchedule')" style="padding: 120px 0">
          <template #extra>
            <n-button type="primary" @click="scheduleModalRef?.openCreate()">+ {{ t('schedule.newSchedule') }}</n-button>
          </template>
        </n-empty>
      </div>
    </div>

    <!-- Modals -->
    <ScheduleModal
      ref="scheduleModalRef"
      :teams="teams"
      @saved="handleScheduleSaved"
    />

    <ShiftModal
      ref="shiftModalRef"
      :schedule-id="selectedSchedule?.id ?? null"
      :users="users"
      @saved="handleShiftSaved"
    />

    <!-- Generate Shifts Modal -->
    <n-modal v-model:show="showGenerateModal" preset="card" :title="t('schedule.generateShifts')" style="width: 420px" :bordered="false">
      <n-form label-placement="top">
        <n-form-item :label="t('schedule.weeksCount')">
          <n-input-number v-model:value="generateWeeks" :min="1" :max="12" style="width: 100%" />
        </n-form-item>
        <n-text depth="3" style="font-size: 12px">{{ t('schedule.generateHint') }}</n-text>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showGenerateModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="generating" @click="handleGenerateShifts">
            {{ t('schedule.confirmGenerate') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.schedule-page {
  height: calc(100vh - 100px);
  display: flex;
  flex-direction: column;
}

.schedule-layout {
  display: flex;
  gap: 0;
  flex: 1;
  min-height: 0;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--sre-border);
  background: var(--sre-bg-card);
}

/* Right Detail Panel */
.schedule-detail {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid var(--sre-border);
  flex-shrink: 0;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.detail-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--sre-text-primary);
}

.week-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 20px;
  border-bottom: 1px solid var(--sre-border);
  flex-shrink: 0;
}

.week-label {
  font-size: 13px;
  font-weight: 500;
  min-width: 160px;
  text-align: center;
  color: var(--sre-text-primary);
}

/* Calendar */
.calendar-container {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
}

.calendar-grid {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.cal-header-row {
  display: grid;
  grid-template-columns: 52px repeat(7, 1fr);
  border-bottom: 1px solid var(--sre-border);
  flex-shrink: 0;
}

.cal-time-gutter {
  border-right: 1px solid var(--sre-border);
}

.cal-day-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 0;
  font-size: 11px;
  color: var(--sre-text-secondary);
  border-right: 1px solid rgba(128, 128, 128, 0.1);
}

.cal-day-header.today {
  color: var(--sre-primary);
}

.cal-day-name {
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 10px;
}

.cal-day-num {
  font-size: 16px;
  font-weight: 500;
  line-height: 1.4;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cal-day-num.today {
  background: var(--sre-primary);
  color: #fff;
}

.cal-body {
  display: grid;
  grid-template-columns: 52px repeat(7, 1fr);
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  position: relative;
}

.cal-time-gutter-body {
  border-right: 1px solid var(--sre-border);
  position: relative;
  height: 1200px;
}

.cal-time-label {
  position: absolute;
  right: 6px;
  font-size: 10px;
  color: var(--sre-text-secondary);
  transform: translateY(-50%);
}

.cal-day-col {
  position: relative;
  height: 1200px;
  border-right: 1px solid rgba(128, 128, 128, 0.08);
  cursor: pointer;
}

.cal-day-col:hover {
  background: rgba(128, 128, 128, 0.02);
}

.cal-hour-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(128, 128, 128, 0.08);
}

.current-time-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: #d03050;
  z-index: 10;
  pointer-events: none;
}

.current-time-line::before {
  content: '';
  position: absolute;
  left: -4px;
  top: -3px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #d03050;
}

.shift-block {
  font-size: 11px;
  line-height: 1.3;
  user-select: none;
  transition: opacity 0.15s;
}

.shift-block:hover {
  opacity: 0.8;
}

.shift-user {
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.shift-time {
  font-size: 10px;
  opacity: 0.75;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Bottom Config Tabs */
.config-tabs {
  flex-shrink: 0;
  border-top: 1px solid var(--sre-border);
  max-height: 260px;
  overflow-y: auto;
  padding: 0 16px 12px;
}

.config-form {
  padding: 12px 0;
}

/* Time label positions (every 2 hours) */
.cal-time-gutter-body .cal-time-label:nth-child(1)  { top: 0px }
.cal-time-gutter-body .cal-time-label:nth-child(2)  { top: 100px }
.cal-time-gutter-body .cal-time-label:nth-child(3)  { top: 200px }
.cal-time-gutter-body .cal-time-label:nth-child(4)  { top: 300px }
.cal-time-gutter-body .cal-time-label:nth-child(5)  { top: 400px }
.cal-time-gutter-body .cal-time-label:nth-child(6)  { top: 500px }
.cal-time-gutter-body .cal-time-label:nth-child(7)  { top: 600px }
.cal-time-gutter-body .cal-time-label:nth-child(8)  { top: 700px }
.cal-time-gutter-body .cal-time-label:nth-child(9)  { top: 800px }
.cal-time-gutter-body .cal-time-label:nth-child(10) { top: 900px }
.cal-time-gutter-body .cal-time-label:nth-child(11) { top: 1000px }
.cal-time-gutter-body .cal-time-label:nth-child(12) { top: 1100px }
.cal-time-gutter-body .cal-time-label:nth-child(13) { top: 1200px }
</style>
