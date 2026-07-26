import { get } from './request'

export interface AuditLog {
  id: number
  operator_id: number
  operator_name: string
  target_type: string
  target_id: number
  target_name: string
  action: string
  before_snapshot: string
  after_snapshot: string
  ip: string
  created_at: string
}

export interface AuditListParams {
  pageNum: number
  pageSize: number
  operator_name?: string
  action?: string
  target_type?: string
  created_at_start?: string
  created_at_end?: string
}

export interface AuditListResponse {
  list: AuditLog[]
  total: number
}

export function listAuditLogs(params: AuditListParams): Promise<AuditListResponse> {
  return get<AuditListResponse>('/api/audit-logs', params as unknown as Record<string, unknown>)
}

export function getAuditDetail(id: number): Promise<AuditLog> {
  return get<AuditLog>('/api/audit-logs/detail', { id })
}

export function exportAudit(params: AuditListParams): Promise<Blob> {
  return get<Blob>('/api/audit-logs/export', params as unknown as Record<string, unknown>, { responseType: 'blob' })
}
