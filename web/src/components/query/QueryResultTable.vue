<script setup lang="ts">
import { computed } from 'vue'
import { NDataTable } from 'naive-ui'
import type { QueryTarget } from '@/types/query'

const props = defineProps<{
  targets: QueryTarget[]
}>()

interface TableRow {
  key: string
  labels: string
  value: string
  timestamp: string
}

function formatTimestamp(ts: number): string {
  return new Date(ts).toLocaleString()
}

const columns = [
  { title: 'Labels', key: 'labels', ellipsis: { tooltip: true } },
  { title: 'Value', key: 'value', width: 200 },
  { title: 'Timestamp', key: 'timestamp', width: 200 },
]

const tableData = computed<TableRow[]>(() => {
  const rows: TableRow[] = []
  for (const t of props.targets) {
    if (!t.enabled || !t.series) continue
    for (let i = 0; i < t.series.length; i++) {
      const s = t.series[i]
      const labelStr = Object.entries(s.labels)
        .map(([k, v]) => `${k}="${v}"`)
        .join(', ')
      if (t.resultType === 'vector' && s.values.length > 0) {
        rows.push({
          key: `${t.id}-${i}`,
          labels: labelStr,
          value: String(s.values[0].value),
          timestamp: formatTimestamp(s.values[0].ts),
        })
      } else {
        for (let j = 0; j < s.values.length; j++) {
          const v = s.values[j]
          rows.push({
            key: `${t.id}-${i}-${j}`,
            labels: labelStr,
            value: String(v.value),
            timestamp: formatTimestamp(v.ts),
          })
        }
      }
    }
  }
  return rows.slice(0, 1000) // Limit rows for performance
})
</script>

<template>
  <NDataTable
    :columns="columns"
    :data="tableData"
    :max-height="400"
    :scroll-x="800"
    size="small"
    striped
    :pagination="{ pageSize: 50 }"
  />
</template>
