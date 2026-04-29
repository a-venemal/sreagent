export interface PipelineNode {
  id: string
  name: string
  type: 'if' | 'relabel' | 'event_drop' | 'callback' | 'ai_summary'
  config: Record<string, any>
  disabled?: boolean
  continue_on_fail?: boolean
  retry_on_fail?: boolean
  max_retries?: number
  position?: { x: number; y: number }
}

export interface Connections {
  [sourceNodeId: string]: {
    [outputIndex: number]: string[]
  }
}

export interface LabelFilter {
  key: string
  op: '==' | '!=' | '=~' | '!~' | 'in' | 'not_in'
  value: string
}

export interface EventPipeline {
  id: number
  name: string
  description: string
  disabled: boolean
  filter_enable: boolean
  label_filters: LabelFilter[]
  nodes: PipelineNode[]
  connections: Connections
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
}

export interface PipelineExecution {
  id: string
  pipeline_id: number
  event_id: number
  status: 'success' | 'failed' | 'terminated'
  node_results: string
  error_message: string
  duration_ms: number
  started_at: string
  finished_at: string
}

export interface NodeExecutionResult {
  node_id: string
  node_name: string
  node_type: string
  status: 'success' | 'failed' | 'skipped' | 'terminated'
  message?: string
  error?: string
  branch_idx?: number
  duration_ms: number
}

export const PROCESSOR_TYPES = [
  { value: 'if', label: 'If (Condition)', color: '#18a058', icon: 'branch' },
  { value: 'relabel', label: 'Relabel', color: '#2080f0', icon: 'tag' },
  { value: 'event_drop', label: 'Event Drop', color: '#d03050', icon: 'delete' },
  { value: 'callback', label: 'Callback', color: '#f0a020', icon: 'link' },
  { value: 'ai_summary', label: 'AI Summary', color: '#8b5cf6', icon: 'sparkles' },
] as const

export type ProcessorType = typeof PROCESSOR_TYPES[number]['value']
