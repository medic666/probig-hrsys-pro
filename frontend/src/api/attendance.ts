import http from './request'

export function getAttendanceEventList(params: any) {
  return http.get('/attendance-events', { params })
}

export function createAttendanceEvent(data: any) {
  return http.post('/attendance-events', data)
}

export function updateAttendanceEvent(id: number, data: any) {
  return http.put(`/attendance-events/${id}`, data)
}

export function deleteAttendanceEvent(id: number) {
  return http.delete(`/attendance-events/${id}`)
}

export function getAttendanceSummaryList(params: any) {
  return http.get('/attendance-summaries', { params })
}

export function calculateAttendance(data: any) {
  return http.post('/attendance-summaries/calculate', data)
}

export function lockAttendanceSummary(data: any) {
  return http.post('/attendance-summaries/lock', data)
}
