import request from '@/utils/request'

export function getAttendanceEvents(params: any) {
  return request.get('/attendance-events', { params })
}
export function getAttendanceEvent(id: number) {
  return request.get(`/attendance-events/${id}`)
}
export function exportAttendanceEvents(params: any) {
  return request.get('/attendance-events/export', { params, responseType: 'blob' })
}
export function createAttendanceEvent(data: any) {
  return request.post('/attendance-events', data)
}
export function deleteAttendanceEvent(id: number) {
  return request.delete(`/attendance-events/${id}`)
}
export function restoreAttendanceEvent(id: number) {
  return request.post(`/attendance-events/${id}/restore`)
}
export function getDeletedAttendanceEvents(params: any) {
  return request.get('/attendance-events/trash', { params })
}
export function createBatchAttendanceEvents(data: any) {
  return request.post('/attendance-events/batch', data)
}
export function getDailyProjections(params: any) {
  return request.get('/attendance-daily', { params })
}
export function exportDailyProjections(params: any) {
  return request.get('/attendance-daily/export', { params, responseType: 'blob' })
}
export function getEventsByDate(personId: number, date: string) {
  return request.get(`/attendance-daily/${personId}/${date}/events`)
}
export function getMonthlyList(params: any) {
  return request.get('/attendance-monthly', { params })
}
export function exportAttendanceMonthly(params: any) {
  return request.get('/attendance-monthly/export', { params, responseType: 'blob' })
}
export function calculateMonthly(data: any) {
  return request.post('/attendance-monthly/calculate', data)
}
export function getPendingDailies(params: any) {
  return request.get('/attendance-events/pending', { params })
}
// getAttendanceEventBadges 考勤事件徽章（默认上月）：person_id → gray/green/orange
export function getAttendanceEventBadges(month?: string) {
  return request.get('/attendance-events/badges', { params: month ? { month } : {} })
}
// getDailyProjectionBadges 日记工时徽章（默认上月）：person_id → gray/green/orange
export function getDailyProjectionBadges(month?: string) {
  return request.get('/attendance-daily/badges', { params: month ? { month } : {} })
}
// getAttendanceMonthlyBadges 月度核算徽章（默认上月）：person_id → gray/green/orange
export function getAttendanceMonthlyBadges(month?: string) {
  return request.get('/attendance-monthly/badges', { params: month ? { month } : {} })
}
// confirmAttendanceDaily 保存整日考勤（统一保存入口：卡片确认/编辑保存/待确认复核共用）。
// status 缺省为已确认——原子卡片"一键确认"不传状态即生效；编辑保存可显式传待确认
export function confirmAttendanceDaily(id: number, details: any[], punchTime?: string, remark?: string, status?: string) {
  return request.post(`/attendance-events/${id}/confirm`, {
    details,
    punch_time: punchTime || '',
    remark: remark || '',
    status: status || '',
  })
}
export function dingTalkPreview(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return request.post('/attendance-events/import-dingtalk/preview', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
}
export function dingTalkExecute(month: string, filePath: string, mappings: any[]) {
  return request.post('/attendance-events/import-dingtalk/execute', { month, file_path: filePath, mappings })
}
