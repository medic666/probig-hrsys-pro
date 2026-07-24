import http from './request'

export function getSalaryEventList(params: any) {
  return http.get('/salary-events', { params })
}

export function createSalaryEvent(data: any) {
  return http.post('/salary-events', data)
}

export function updateSalaryEvent(id: number, data: any) {
  return http.put(`/salary-events/${id}`, data)
}

export function deleteSalaryEvent(id: number) {
  return http.delete(`/salary-events/${id}`)
}

export function getSalarySummaryList(params: any) {
  return http.get('/salary-summaries', { params })
}

export function calculateSalary(data: any) {
  return http.post('/salary-summaries/calculate', data)
}

export function lockSalarySummary(data: any) {
  return http.post('/salary-summaries/lock', data)
}
