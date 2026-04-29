<script setup lang="ts">
import { NSelect } from 'naive-ui'
import { autoRefreshOptions } from '@/composables/useTimeRange'

const props = defineProps<{
  value: number | null
}>()

const emit = defineEmits<{
  (e: 'update:value', value: number | null): void
}>()

const options = autoRefreshOptions.map(o => ({
  label: o.label,
  value: o.value,
}))
</script>

<template>
  <NSelect
    :value="value"
    :options="options"
    size="small"
    style="width: 80px"
    @update:value="(v: number | null) => emit('update:value', v)"
  />
</template>
