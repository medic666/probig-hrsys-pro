import { get } from './client'
import type { PageData, AuditLog } from '../types'

export function listAuditLogs(params?: { page?: number; page_size?: number; target_type?: string; user_id?: number }) {
  return get<PageData<AuditLog>>('/audit-logs', params)
}
