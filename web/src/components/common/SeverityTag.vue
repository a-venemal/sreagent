<template>
  <n-tag :type="getSeverityType(severity)" size="small" :bordered="bordered" round>
    <slot>{{ severity }}</slot>
  </n-tag>
</template>

<script setup lang="ts">
import { NTag } from 'naive-ui'
import { getSeverityType } from '@/utils/alert'

withDefaults(defineProps<{
  severity: string
  bordered?: boolean
}>(), {
  bordered: true,
})
</script>
