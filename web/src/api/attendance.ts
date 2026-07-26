import request from './request'

export function getAttendanceEvents(params: any) {
  return request({ url: '/attendances/events', method: 'get', params })
}

export function createAttendanceEvent(data: any) {
  return request({ url: '/attendances/events', method: 'post', data })
}

export function createCrossDayEvent(data: any) {
  return request({ url: '/attendances/events/cross-day', method: 'post', data })
}

export function createBatchEvents(data: any) {
  return request({ url: '/attendances/events/batch', method: 'post', data })
}

export function updateAttendanceEvent(id: number, data: any) {
  return request({ url: `/attendances/events/${id}`, method: 'put', data })
}

export function deleteAttendanceEvent(id: number) {
  return request({ url: `/attendances/events/${id}`, method: 'delete' })
}

export function getDailyList(params: any) {
  return request({ url: '/attendances/daily', method: 'get', params })
}

export function getDailyEvents(personId: number, date: string) {
  return request({ url: `/attendances/daily/events/${personId}/${date}`, method: 'get' })
}

export function getAttendanceSalaryList(params: any) {
  return request({ url: '/attendances/salary', method: 'get', params })
}

export function calculateAttendanceSalary(data: any) {
  return request({ url: '/attendances/salary/cal', method: 'post', data })
}
