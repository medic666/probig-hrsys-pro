import { get, post, put, del } from './request'

export interface SalaryEvent {
  id: number
  person_id: number
  person_name: string
  belong_month: string
  event_type: string
  amount: number
  event_name: string
  remark: string
  created_at: string
  updated_at: string
}

export interface SalaryEventCreateRequest {
  person_id: number
  belong_month: string
  event_type: string
  amount: number
  event_name: string
  remark?: string
}

export interface SalaryEventUpdateRequest extends Partial<SalaryEventCreateRequest> {
  id: number
}

export interface SalarySummary {
  id: number
  person_id: number
  person_name: string
  belong_month: string
  salary_days: number
  weighted_base_salary: number
  total_work_hours: number
  total_overtime_workday_hours: number
  total_overtime_holiday_hours: number
  attendance_salary: number
  overtime_workday_salary: number
  overtime_holiday_salary: number
  annual_leave_carryover_salary: number
  attendance_bonus: number
  performance_salary: number
  post_allowance: number
  meal_allowance: number
  housing_allowance: number
  transport_allowance: number
  high_temp_allowance: number
  insurance_compensation: number
  fund_compensation: number
  total_adjustment: number
  social_security_deduct: number
  housing_fund_deduct: number
  tax_deduct: number
  final_salary: number
  status: string
  last_calc_at: string
}

export interface EventListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  belong_month?: string
  event_type?: string
}

export interface EventListResponse {
  list: SalaryEvent[]
  total: number
}

export interface SummaryListParams {
  pageNum: number
  pageSize: number
  belong_month_start?: string
  belong_month_end?: string
  person_id?: number
  attendance_group?: string
}

export interface SummaryListResponse {
  list: SalarySummary[]
  total: number
}

export interface CalcSummaryRequest {
  belong_month: string
  person_ids?: number[]
  attendance_group?: string
}

export function listEvents(params: EventListParams): Promise<EventListResponse> {
  return get<EventListResponse>('/api/salary-events', params as unknown as Record<string, unknown>)
}

export function createEvent(data: SalaryEventCreateRequest): Promise<void> {
  return post<void>('/api/salary-events', data)
}

export function updateEvent(data: SalaryEventUpdateRequest): Promise<void> {
  return put<void>('/api/salary-events', data)
}

export function deleteEvent(id: number): Promise<void> {
  return del<void>('/api/salary-events', { id })
}

export function restoreEvent(id: number): Promise<void> {
  return post<void>('/api/salary-events/restore', { id })
}

export function listSummaries(params: SummaryListParams): Promise<SummaryListResponse> {
  return get<SummaryListResponse>('/api/salary-summaries', params as unknown as Record<string, unknown>)
}

export function calcSummary(data: CalcSummaryRequest): Promise<void> {
  return post<void>('/api/salary-summaries/calc', data)
}

export function exportSummary(params: { belong_month: string; person_ids?: number[] }): Promise<Blob> {
  return get<Blob>('/api/salary-summaries/export', params as unknown as Record<string, unknown>, { responseType: 'blob' })
}
