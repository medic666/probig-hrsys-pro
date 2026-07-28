import request from '@/utils/request'

export function getAuditLogs(params: any) { return request.get('/audit-logs', { params }) }
export function getAuditLogDetail(id: number) { return request.get(`/audit-logs/${id}`) }
