import { get, post, put, del } from './client'
import type { PageData, AttendanceEvent, AttendanceSummary } from '../types'

export function listAttendanceEvents(params?: { entity_id?: number; start_date?: string; end_date?: string; page?: number; page_size?: number }) {
  return get<PageData<AttendanceEvent>>('/attendance/events', params)
}

export function createAttendanceEvent(data: any) {
  return post<AttendanceEvent>('/attendance/events', data)
}

export function updateAttendanceEvent(id: number, data: any) {
  return put<AttendanceEvent>(`/attendance/events/${id}`, data)
}

export function deleteAttendanceEvent(id: number) {
  return del(`/attendance/events/${id}`)
}

export function listAttendanceSummaries(params?: { entity_id?: number; start_date?: string; end_date?: string; page?: number; page_size?: number }) {
  return get<PageData<AttendanceSummary>>('/attendance/summaries', params)
}

export function calculateAttendance(data: { period_start: string; period_end: string }) {
  return post<AttendanceSummary[]>('/attendance/calculate', data)
}
