import request from '@/utils/request'

export function getAttendanceEvents(params: any) {
  return request.get('/attendance-events', { params })
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
export function calculateMonthly(data: any) {
  return request.post('/attendance-monthly/calculate', data)
}
