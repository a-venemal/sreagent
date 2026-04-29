<script setup lang="ts">
import { NSelect } from 'naive-ui'
import { autoRefreshOptions } from '@/composables/useTimeRange'

const props = defineProps<{
  value: number | null
}>()

const emit = defineEmits<{
  (e: 'update:value', value: number | null): void
}>()

const OFF_VALUE = -1

const options = autoRefreshOptions.map(o => ({
  label: o.label,
  value: o.value === null ? OFF_VALUE : o.value,
}))

function handleChange(val: number) {
  emit('update:value', val === OFF_VALUE ? null : val)
}
</script>

<template>
  <NSelect
    :value="value === null ? OFF_VALUE : value"
    :options="options"
    size="small"
    style="width: 80px"
    @update:value="handleChange"
  />
</template>
