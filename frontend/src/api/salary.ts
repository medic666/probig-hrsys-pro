import { get, post, put, del } from './client'
import type { PageData, SalaryEvent, SalarySummary } from '../types'

export function listSalaryEvents(params?: { entity_id?: number; page?: number; page_size?: number }) {
  return get<PageData<SalaryEvent>>('/salary/events', params)
}

export function createSalaryEvent(data: any) {
  return post<SalaryEvent>('/salary/events', data)
}

export function updateSalaryEvent(id: number, data: any) {
  return put<SalaryEvent>(`/salary/events/${id}`, data)
}

export function deleteSalaryEvent(id: number) {
  return del(`/salary/events/${id}`)
}

export function listSalarySummaries(params?: { entity_id?: number; start_date?: string; end_date?: string; page?: number; page_size?: number }) {
  return get<PageData<SalarySummary>>('/salary/summaries', params)
}

export function calculateSalary(data: { period_start: string; period_end: string }) {
  return post<SalarySummary[]>('/salary/calculate', data)
}
