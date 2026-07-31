import request from '@/utils/request'

export function getAttendanceEvents(params: any) {
  return request.get('/attendance-events', { params })
}
export function exportAttendanceEvents(params: any) {
  return request.get('/attendance-events/export', { params, responseType: 'blob' })
}
export function createAttendanceEvent(data: any) {
  return request.post('/attendance-events', data)
}
export function updateAttendanceEvent(id: number, data: any) {
  return request.put(`/attendance-events/${id}`, data)
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
export function confirmPendingDaily(id: number, details: any[]) {
  return request.post(`/attendance-events/pending/${id}/confirm`, { details })
}
export function dingTalkPreview(file: File) {
  const fd = new FormData()
  fd.append('file', file)
  return request.post('/attendance-events/import-dingtalk/preview', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
}
export function dingTalkExecute(month: string, filePath: string, mappings: any[]) {
  return request.post('/attendance-events/import-dingtalk/execute', { month, file_path: filePath, mappings })
}
