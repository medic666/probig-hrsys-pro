import { get, post, put, del } from './request'

export interface AttendanceEvent {
  id: number
  person_id: number
  person_name: string
  event_date: string
  punch_time: string
  event_type: string
  sub_type: string
  hours: number
  late_minutes: number
  is_special_approval: boolean
  remark: string
  created_at: string
  updated_at: string
}

export interface AttendanceEventCreateRequest {
  person_id: number
  event_date: string
  event_date_end?: string
  punch_time?: string
  event_type: string
  sub_type: string
  hours: number
  late_minutes?: number
  is_special_approval?: boolean
  remark?: string
}

export interface AttendanceBatchCreateRequest {
  person_ids: number[]
  attendance_group?: string
  event_date: string
  event_date_end?: string
  event_type: string
  sub_type: string
  hours: number
  late_minutes?: number
  is_special_approval?: boolean
  remark?: string
}

export interface AttendanceEventUpdateRequest extends Partial<AttendanceEventCreateRequest> {
  id: number
}

export interface AttendanceDailyProjection {
  id: number
  person_id: number
  person_name: string
  work_date: string
  punch_time: string
  work_hours: number
  overtime_workday_hours: number
  overtime_holiday_hours: number
  has_personal_leave: boolean
  violation_count: number
  remark: string
  last_calc_at: string
}

export interface AttendanceMonthlySalary {
  id: number
  person_id: number
  person_name: string
  belong_month: string
  salary_days: number
  weighted_base_salary: number
  weighted_meal_allowance: number
  total_work_hours: number
  total_overtime_workday_hours: number
  total_overtime_holiday_hours: number
  attendance_salary: number
  overtime_workday_salary: number
  overtime_holiday_salary: number
  has_personal_leave_month: boolean
  total_violation_count: number
  attendance_bonus: number
  status: string
  last_calc_at: string
}

export interface EventListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  event_date_start?: string
  event_date_end?: string
  event_type?: string
  sub_type?: string
}

export interface EventListResponse {
  list: AttendanceEvent[]
  total: number
}

export interface DailyListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  work_date_start?: string
  work_date_end?: string
}

export interface DailyListResponse {
  list: AttendanceDailyProjection[]
  total: number
}

export interface MonthlySalaryListParams {
  pageNum: number
  pageSize: number
  belong_month?: string
  person_id?: number
  attendance_group?: string
}

export interface MonthlySalaryListResponse {
  list: AttendanceMonthlySalary[]
  total: number
}

export interface CalcMonthlyRequest {
  belong_month: string
  person_ids?: number[]
  attendance_group?: string
}

export function listEvents(params: EventListParams): Promise<EventListResponse> {
  return get<EventListResponse>('/api/attendance-events', params as unknown as Record<string, unknown>)
}

export function createEvent(data: AttendanceEventCreateRequest): Promise<void> {
  return post<void>('/api/attendance-events', data)
}

export function createBatchEvents(data: AttendanceBatchCreateRequest): Promise<{ success_count: number; fail_details: { person_name: string; reason: string }[] }> {
  return post('/api/attendance-events/batch', data)
}

export function updateEvent(data: AttendanceEventUpdateRequest): Promise<void> {
  return put<void>('/api/attendance-events', data)
}

export function deleteEvent(id: number): Promise<void> {
  return del<void>('/api/attendance-events', { id })
}

export function restoreEvent(id: number): Promise<void> {
  return post<void>('/api/attendance-events/restore', { id })
}

export function listDailyProjections(params: DailyListParams): Promise<DailyListResponse> {
  return get<DailyListResponse>('/api/attendance-daily', params as unknown as Record<string, unknown>)
}

export function listMonthlySalary(params: MonthlySalaryListParams): Promise<MonthlySalaryListResponse> {
  return get<MonthlySalaryListResponse>('/api/attendance-monthly', params as unknown as Record<string, unknown>)
}

export function calcMonthlySalary(data: CalcMonthlyRequest): Promise<void> {
  return post<void>('/api/attendance-monthly/calc', data)
}

export function exportMonthlySalary(params: { belong_month: string; person_ids?: number[] }): Promise<Blob> {
  return get<Blob>('/api/attendance-monthly/export', params as unknown as Record<string, unknown>, { responseType: 'blob' })
}
