<template>
  <n-tag
    size="small"
    :bordered="false"
    :color="statusTagColor(status)"
    round
  >
    <slot>{{ label }}</slot>
  </n-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'
import { statusTagColor, getStatusLabelKey } from '@/utils/alert'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  status: string
}>()

const { t, te } = useI18n()

const label = computed(() => {
  const key = getStatusLabelKey(props.status)
  return te(key) ? t(key) : props.status
})
</script>
