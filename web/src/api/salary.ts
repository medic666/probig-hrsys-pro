import request from './request'

export function getSalaryEvents(params: any) {
  return request({ url: '/salaries/events', method: 'get', params })
}

export function createSalaryEvent(data: any) {
  return request({ url: '/salaries/events', method: 'post', data })
}

export function updateSalaryEvent(id: number, data: any) {
  return request({ url: `/salaries/events/${id}`, method: 'put', data })
}

export function deleteSalaryEvent(id: number) {
  return request({ url: `/salaries/events/${id}`, method: 'delete' })
}

export function getSalarySummaries(params: any) {
  return request({ url: '/salaries/summaries', method: 'get', params })
}

export function calculateSalarySummaries(data: { person_ids: number[]; belong_month: string }) {
  return request({ url: '/salaries/summaries/cal', method: 'post', data })
}
