import { ref, watch, type Ref } from 'vue'
import { datasourceApi } from '@/api'

interface CompletionCache {
  keys: string[]
  metrics: string[]
}

const cache = new Map<number, CompletionCache>()

export function usePromQLCompletion(datasourceId: Ref<number | null>) {
  const labelKeys = ref<string[]>([])
  const metricNames = ref<string[]>([])
  const loading = ref(false)

  async function load() {
    if (!datasourceId.value) {
      labelKeys.value = []
      metricNames.value = []
      return
    }

    const id = datasourceId.value
    if (cache.has(id)) {
      const cached = cache.get(id)!
      labelKeys.value = cached.keys
      metricNames.value = cached.metrics
      return
    }

    loading.value = true
    try {
      const [keysRes, metricsRes] = await Promise.all([
        datasourceApi.labelKeys(id).catch(() => ({ data: { data: [] as string[] } })),
        datasourceApi.metricNames(id).catch(() => ({ data: { data: [] as string[] } })),
      ])
      const keys = keysRes.data.data || []
      const metrics = metricsRes.data.data || []
      labelKeys.value = keys
      metricNames.value = metrics
      cache.set(id, { keys, metrics })
    } finally {
      loading.value = false
    }
  }

  async function loadLabelValues(key: string): Promise<string[]> {
    if (!datasourceId.value) return []
    try {
      const res = await datasourceApi.labelValues(datasourceId.value, key)
      return res.data.data || []
    } catch {
      return []
    }
  }

  watch(datasourceId, load, { immediate: true })

  return { labelKeys, metricNames, loading, loadLabelValues, refresh: load }
}
