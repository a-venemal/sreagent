/**
 * Format an ISO date string to a localized display string.
 * Uses the current document language to determine locale.
 */
export function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return '-'
  // Detect locale from <html lang="..."> or fallback to en-US
  const htmlLang = typeof document !== 'undefined' ? document.documentElement.lang : ''
  const locale = htmlLang === 'zh-CN' ? 'zh-CN' : 'en-US'
  return d.toLocaleString(locale, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

/**
 * Format a duration in seconds to a human-readable string.
 * e.g. 270 => "4m 30s", 3661 => "1h 1m 1s"
 */
export function formatDuration(seconds: number): string {
  if (seconds < 0) return '0s'
  if (seconds === 0) return '0s'

  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)

  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`)

  return parts.join(' ')
}

/**
 * Convert an array of { key, value } pairs to a Record<string, string>.
 * Empty keys are silently skipped; keys are trimmed.
 */
export function kvArrayToRecord(arr: { key: string; value: string }[]): Record<string, string> {
  const record: Record<string, string> = {}
  for (const item of arr) {
    if (item.key.trim()) {
      record[item.key.trim()] = item.value
    }
  }
  return record
}

/**
 * Convert a Record<string, string> to an array of { key, value } pairs.
 * Useful for populating KVEditor from API data.
 */
export function recordToKVArray(record: Record<string, string> | null | undefined): { key: string; value: string }[] {
  if (!record) return []
  return Object.entries(record).map(([key, value]) => ({ key, value }))
}
