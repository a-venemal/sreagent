import type { AlertSeverity, AlertEventStatus, AlertRuleStatus, DataSourceStatus, TimelineAction } from '@/types'

// ===== Severity Helpers =====

/**
 * Map alert severity to Naive UI NTag `type` prop.
 */
export function getSeverityType(severity: string): 'error' | 'warning' | 'info' | 'default' {
  switch (severity) {
    case 'critical': return 'error'
    case 'warning':  return 'warning'
    case 'info':     return 'info'
    default:         return 'default'
  }
}

/**
 * Map alert severity to a hex color string.
 */
export function getSeverityColor(severity: string): string {
  switch (severity) {
    case 'critical': return '#e88080'
    case 'warning':  return '#f2c97d'
    case 'info':     return '#70c0e8'
    default:         return '#999'
  }
}

/**
 * Return a CSS class name for table row highlighting by severity.
 */
export function severityRowClass(row: { severity?: string }): string {
  if (row.severity === 'critical') return 'row-critical'
  if (row.severity === 'warning') return 'row-warning'
  return ''
}

// ===== Alert Event Status Helpers =====

/**
 * Map alert event status to Naive UI NTag `type` prop.
 */
export function getEventStatusType(status: string): 'error' | 'warning' | 'info' | 'success' | 'default' {
  switch (status) {
    case 'firing':       return 'error'
    case 'acknowledged': return 'warning'
    case 'assigned':     return 'info'
    case 'resolved':     return 'success'
    case 'closed':       return 'default'
    case 'silenced':     return 'default'
    default:             return 'default'
  }
}

/**
 * Map alert event status to a hex color string.
 */
export function getStatusColor(status: string): string {
  switch (status) {
    case 'firing':       return '#e88080'
    case 'acknowledged': return '#f2c97d'
    case 'assigned':     return '#70c0e8'
    case 'resolved':     return '#18a058'
    case 'closed':       return '#666'
    case 'silenced':     return '#a855f7'
    default:             return '#999'
  }
}

/**
 * Build an NTag `color` prop object from a hex color for a subtle transparent look.
 * Returns `{ color: hex + '18', textColor: hex, borderColor: 'transparent' }`
 */
export function statusTagColor(status: string) {
  const hex = getStatusColor(status)
  return { color: hex + '18', textColor: hex, borderColor: 'transparent' }
}

/**
 * i18n key map for alert event statuses.
 */
const statusLabelKeys: Record<string, string> = {
  firing:       'alert.firing',
  acknowledged: 'alert.acknowledged',
  assigned:     'alert.assigned',
  resolved:     'alert.resolved',
  closed:       'alert.closed',
  silenced:     'alert.silenced',
}

/**
 * Return the i18n key for a given alert event status, or the status string itself.
 */
export function getStatusLabelKey(status: string): string {
  return statusLabelKeys[status] || status
}

// ===== Alert Rule Status Helpers =====

/**
 * Map alert rule status to Naive UI NTag `type` prop.
 */
export function getRuleStatusType(status: string): 'success' | 'default' | 'warning' {
  switch (status) {
    case 'enabled':  return 'success'
    case 'disabled': return 'default'
    case 'muted':    return 'warning'
    default:         return 'default'
  }
}

// ===== Datasource Status Helpers =====

/**
 * Map datasource health status to Naive UI NTag `type` prop.
 */
export function getDatasourceStatusType(status: string): 'success' | 'error' | 'warning' {
  switch (status) {
    case 'healthy':   return 'success'
    case 'unhealthy': return 'error'
    default:          return 'warning'
  }
}

// ===== Timeline Helpers =====

/**
 * Map timeline action to Naive UI NTag/Timeline `type` prop.
 */
export function getTimelineType(action: string): 'error' | 'warning' | 'info' | 'success' | 'default' {
  switch (action) {
    case 'created':      return 'error'
    case 'resolved':     return 'success'
    case 'closed':       return 'default'
    case 'acknowledged': return 'warning'
    case 'assigned':     return 'info'
    case 'escalated':    return 'error'
    case 'commented':    return 'info'
    default:             return 'info'
  }
}

// ===== Row highlight CSS (to be imported by pages using :deep or global) =====

/**
 * Shared row highlight styles. Pages should add these to their <style> block or
 * use the global CSS classes defined in global.css.
 */
export const ROW_HIGHLIGHT_CSS = `
.row-critical { background-color: rgba(232, 128, 128, 0.04); }
.row-warning  { background-color: rgba(242, 201, 125, 0.04); }
`
