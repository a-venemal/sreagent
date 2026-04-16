// ===== API Response Types =====
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// ===== Auth =====
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_in: number
}

export type UserType = 'human' | 'bot' | 'channel'

export interface User {
  id: number
  username: string
  display_name: string
  email: string
  phone: string
  lark_user_id: string
  avatar: string
  role: 'admin' | 'team_lead' | 'member' | 'viewer' | 'global_viewer'
  is_active: boolean
  created_at: string
  user_type?: UserType
  notify_target?: string
}

// ===== DataSource =====
export type DataSourceType = 'prometheus' | 'victoriametrics' | 'zabbix' | 'victorialogs'
export type DataSourceStatus = 'healthy' | 'unhealthy' | 'unknown'

export interface DataSource {
  id: number
  name: string
  type: DataSourceType
  endpoint: string
  description: string
  labels: Record<string, string>
  status: DataSourceStatus
  auth_type: string
  health_check_interval: number
  is_enabled: boolean
  created_at: string
  updated_at: string
}

// ===== Alert Rule =====
export type AlertSeverity = 'critical' | 'warning' | 'info'
export type AlertRuleStatus = 'enabled' | 'disabled' | 'muted'

export interface AlertRule {
  id: number
  name: string
  display_name: string
  description: string
  datasource_id: number
  datasource?: DataSource
  expression: string
  for_duration: string
  severity: AlertSeverity
  labels: Record<string, string>
  annotations: Record<string, string>
  status: AlertRuleStatus
  group_name: string
  version: number
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
}

// ===== Alert Event =====
export type AlertEventStatus = 'firing' | 'acknowledged' | 'assigned' | 'resolved' | 'closed' | 'silenced'

export interface AlertEvent {
  id: number
  fingerprint: string
  rule_id: number | null
  rule?: AlertRule
  alert_name: string
  severity: AlertSeverity
  status: AlertEventStatus
  labels: Record<string, string>
  annotations: Record<string, string>
  source: string
  generator_url: string
  fired_at: string
  acked_at: string | null
  resolved_at: string | null
  closed_at: string | null
  acked_by: number | null
  acked_by_user?: User
  assigned_to: number | null
  assigned_to_user?: User
  resolution: string
  fire_count: number
  silenced_until?: string
  silence_reason?: string
  oncall_user_id?: number | null
  oncall_user?: User
  is_dispatched?: boolean
  created_at: string
}

export interface AlertEventFilter {
  status?: string[]
  severity?: string[]
  start_time?: string
  end_time?: string
  source?: string
  alert_name?: string
  business_line?: string
  view_mode?: AlertViewMode
  user_id?: number
  page: number
  page_size: number
}

export type AlertViewMode = 'mine' | 'unassigned' | 'all'

export type TimelineAction =
  | 'created'
  | 'acknowledged'
  | 'assigned'
  | 'commented'
  | 'escalated'
  | 'resolved'
  | 'closed'
  | 'reopened'
  | 'notified'

export interface AlertTimeline {
  id: number
  event_id: number
  action: TimelineAction
  operator_id: number | null
  operator?: User
  note: string
  extra: string
  created_at: string
}

// ===== Team =====
export interface Team {
  id: number
  name: string
  description: string
  labels: Record<string, string>
  members?: User[]
}

// ===== Schedule =====
export interface Schedule {
  id: number
  name: string
  team_id: number
  team?: Team
  description: string
  rotation_type: 'daily' | 'weekly' | 'custom'
  timezone: string
  handoff_time: string
  handoff_day: number
  is_enabled: boolean
  severity_filter?: string
  created_at: string
}

export interface OnCallShift {
  id: number
  schedule_id: number
  user_id: number
  user?: User
  start_time: string   // ISO date string
  end_time: string
  severity_filter: string  // "" | "critical" | "critical,warning" etc
  source: 'manual' | 'rotation'
  note: string
  created_at: string
}

export interface ScheduleParticipant {
  id: number
  schedule_id: number
  user_id: number
  user?: User
  position: number
}

export interface ScheduleOverride {
  id: number
  schedule_id: number
  user_id: number
  user?: User
  start_time: string
  end_time: string
  reason: string
}

// ===== Escalation Policy =====
export interface EscalationPolicy {
  id: number
  name: string
  team_id: number
  team?: Team
  is_enabled: boolean
  created_at: string
}

export interface EscalationStep {
  id: number
  policy_id: number
  step_order: number
  delay_minutes: number
  target_type: string
  target_id: number
  notify_channel_id: number
}

// ===== Notification =====
export type NotifyChannelType = 'lark_webhook' | 'lark_bot' | 'email' | 'sms' | 'custom_webhook'

export interface NotifyChannel {
  id: number
  name: string
  type: NotifyChannelType
  description: string
  labels: Record<string, string>
  config: string
  is_enabled: boolean
  created_at: string
}

export interface NotifyPolicy {
  id: number
  name: string
  description: string
  match_labels: Record<string, string>
  severities: string
  channel_id: number
  channel?: NotifyChannel
  throttle_minutes: number
  template_name: string
  is_enabled: boolean
  priority: number
  created_at: string
}

// ===== Mute Rule =====
export interface MuteRule {
  id: number
  name: string
  description: string
  match_labels: Record<string, string>
  severities: string
  start_time: string | null
  end_time: string | null
  periodic_start: string
  periodic_end: string
  days_of_week: string
  timezone: string
  created_by: number
  is_enabled: boolean
  rule_ids: string
  created_at: string
}

// ===== Notify Rule (v2, replaces NotifyPolicy) =====
export interface NotifyRule {
  id: number
  name: string
  description: string
  is_enabled: boolean
  severities: string
  match_labels: Record<string, string>
  pipeline: string // JSON array of processor configs
  notify_configs: string // JSON array of notification configs
  repeat_interval: number
  callback_url: string
  created_by: number
  created_at: string
}

// ===== Notify Media (replaces NotifyChannel) =====
export interface NotifyMedia {
  id: number
  name: string
  type: 'lark_webhook' | 'email' | 'http' | 'script'
  description: string
  is_enabled: boolean
  config: string
  variables: string
  is_builtin: boolean
  created_at: string
}

// ===== Message Template =====
export interface MessageTemplate {
  id: number
  name: string
  description: string
  content: string
  type: 'text' | 'html' | 'markdown' | 'lark_card'
  is_builtin: boolean
  created_at: string
}

// ===== Subscribe Rule =====
export interface SubscribeRule {
  id: number
  name: string
  description: string
  is_enabled: boolean
  match_labels: Record<string, string>
  severities: string
  notify_rule_id: number
  user_id: number | null
  team_id: number | null
  created_by: number
  created_at: string
}

// ===== Business Group =====
export interface BizGroup {
  id: number
  name: string
  description: string
  parent_id: number | null
  labels: Record<string, string>
  children?: BizGroup[]
  created_at: string
}

// ===== Alert Channel =====
export interface AlertChannel {
  id: number
  name: string
  description: string
  match_labels: Record<string, string>
  severities: string
  media_id: number
  media?: NotifyMedia
  template_id: number | null
  template?: MessageTemplate
  throttle_min: number
  is_enabled: boolean
  created_by: number
  created_at: string
}

// ===== User Notify Config =====
export interface UserNotifyConfig {
  id: number
  user_id: number
  media_type: 'lark_personal' | 'email' | 'webhook'
  config: string
  is_enabled: boolean
}

// ===== Engine Status =====
export interface EngineStatus {
  running: boolean
  total_rules: number
  active_alerts: number
  uptime: string
}

// ===== Dashboard =====
export interface DashboardStats {
  total_datasources: number
  total_rules: number
  active_alerts: number
  resolved_today: number
  total_users: number
  total_teams: number
  severity_breakdown: { critical: number; warning: number; info: number }
}

export interface MTTRStats {
  window_hours: number
  /** Mean time to acknowledge in seconds, -1 if no data */
  mtta_seconds: number
  /** Mean time to resolve in seconds, -1 if no data */
  mttr_seconds: number
  acked_count: number
  resolved_count: number
}
