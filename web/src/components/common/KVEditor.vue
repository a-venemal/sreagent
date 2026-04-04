<template>
  <div class="kv-editor">
    <div v-for="(item, idx) in modelValue" :key="idx" class="kv-row">
      <n-input
        v-model:value="item.key"
        :placeholder="keyPlaceholder"
        size="small"
        @update:value="emitUpdate"
      />
      <n-input
        v-model:value="item.value"
        :placeholder="valuePlaceholder"
        size="small"
        @update:value="emitUpdate"
      />
      <n-button size="small" quaternary type="error" @click="removeItem(idx)">
        <template #icon><n-icon :component="CloseOutline" /></template>
      </n-button>
    </div>
    <n-button dashed size="small" @click="addItem">
      <template #icon><n-icon :component="AddOutline" /></template>
      {{ addLabel }}
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { NInput, NButton, NIcon } from 'naive-ui'
import { AddOutline, CloseOutline } from '@vicons/ionicons5'

export interface KVItem {
  key: string
  value: string
}

const props = withDefaults(defineProps<{
  modelValue: KVItem[]
  keyPlaceholder?: string
  valuePlaceholder?: string
  addLabel?: string
}>(), {
  keyPlaceholder: 'Key',
  valuePlaceholder: 'Value',
  addLabel: 'Add',
})

const emit = defineEmits<{
  'update:modelValue': [value: KVItem[]]
}>()

function addItem() {
  const updated = [...props.modelValue, { key: '', value: '' }]
  emit('update:modelValue', updated)
}

function removeItem(index: number) {
  const updated = props.modelValue.filter((_, i) => i !== index)
  emit('update:modelValue', updated)
}

function emitUpdate() {
  emit('update:modelValue', [...props.modelValue])
}
</script>

<style scoped>
.kv-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kv-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.kv-row .n-input {
  flex: 1;
}
</style>
