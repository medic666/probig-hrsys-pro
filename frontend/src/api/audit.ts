import http from './request'

export function getAuditLogList(params: any) {
  return http.get('/audit-logs', { params })
}
