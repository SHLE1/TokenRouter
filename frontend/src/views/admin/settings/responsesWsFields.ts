import type { ResponsesWSParameters } from '@/api/admin/settings'

// 每个字段说明页面输入类型、单位和是否放入高级参数。
export const responsesWsFields: { key: keyof ResponsesWSParameters; boolean: boolean; scale: number; advanced: boolean; min: number; step: string; zero: boolean }[] = [
  { key: 'max_conns_per_provider', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'min_idle_per_provider', boolean: false, scale: 1, advanced: false, min: 0, step: '1', zero: false },
  { key: 'max_idle_per_provider', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'dynamic_max_conns_by_provider_concurrency_enabled', boolean: true, scale: 1, advanced: true, min: 1, step: '1', zero: false },
  { key: 'oauth_max_conns_factor', boolean: false, scale: 1, advanced: true, min: 1e-06, step: 'any', zero: false },
  { key: 'apikey_max_conns_factor', boolean: false, scale: 1, advanced: true, min: 1e-06, step: 'any', zero: false },
  { key: 'queue_limit_per_conn', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'pool_target_utilization', boolean: false, scale: 1, advanced: true, min: 1e-06, step: 'any', zero: false },
  { key: 'prewarm_cooldown_ms', boolean: false, scale: 1, advanced: true, min: 0, step: '1', zero: false },
  { key: 'client_first_message_timeout_seconds', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'ingress_inter_turn_idle_timeout_seconds', boolean: false, scale: 1, advanced: false, min: 0, step: '1', zero: true },
  { key: 'max_ingress_connections_per_api_key', boolean: false, scale: 1, advanced: false, min: 0, step: '1', zero: true },
  { key: 'client_read_limit_bytes', boolean: false, scale: 1048576, advanced: false, min: 1, step: 'any', zero: false },
  { key: 'http_bridge_threshold_bytes', boolean: false, scale: 1048576, advanced: false, min: 1, step: 'any', zero: false },
  { key: 'dial_timeout_seconds', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'read_timeout_seconds', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'write_timeout_seconds', boolean: false, scale: 1, advanced: false, min: 1, step: '1', zero: false },
  { key: 'ingress_previous_response_recovery_enabled', boolean: true, scale: 1, advanced: true, min: 1, step: '1', zero: false },
  { key: 'sticky_session_ttl_seconds', boolean: false, scale: 1, advanced: true, min: 1, step: '1', zero: false },
  { key: 'sticky_response_id_ttl_seconds', boolean: false, scale: 1, advanced: true, min: 1, step: '1', zero: false },
]
