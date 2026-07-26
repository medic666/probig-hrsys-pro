import { get, post, put, del } from './request'

export interface PositionEvent {
  id: number
  person_id: number
  person_name: string
  event_name: string
  effective_date: string
  attendance_group: string
  has_annual_leave: boolean
  has_attendance_bonus: boolean
  base_salary: number
  performance_salary: number
  salary_days: number
  post_allowance: number
  meal_allowance: number
  housing_allowance: number
  transport_allowance: number
  high_temp_allowance: number
  insurance_compensation: number
  fund_compensation: number
  social_security_deduct: number
  housing_fund_deduct: number
  changed_fields: string
  created_at: string
  updated_at: string
}

export interface PositionEventCreateRequest {
  person_id: number
  event_name: string
  effective_date: string
  attendance_group?: string
  has_annual_leave?: boolean
  has_attendance_bonus?: boolean
  base_salary?: number
  performance_salary?: number
  salary_days?: number
  post_allowance?: number
  meal_allowance?: number
  housing_allowance?: number
  transport_allowance?: number
  high_temp_allowance?: number
  insurance_compensation?: number
  fund_compensation?: number
  social_security_deduct?: number
  housing_fund_deduct?: number
}

export interface PositionEventUpdateRequest extends Partial<PositionEventCreateRequest> {
  id: number
}

export interface PositionSnapshot {
  id: number
  person_id: number
  person_name: string
  effective_start_date: string
  effective_end_date: string
  entry_date: string
  leave_date: string
  attendance_group: string
  has_annual_leave: boolean
  has_attendance_bonus: boolean
  base_salary: number
  performance_salary: number
  salary_days: number
  post_allowance: number
  meal_allowance: number
  housing_allowance: number
  transport_allowance: number
  high_temp_allowance: number
  insurance_compensation: number
  fund_compensation: number
  social_security_deduct: number
  housing_fund_deduct: number
  last_calc_at: string
}

export interface PositionEventListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  effective_date_start?: string
  effective_date_end?: string
  event_name?: string
}

export interface PositionEventListResponse {
  list: PositionEvent[]
  total: number
}

export interface PositionSnapshotListParams {
  pageNum: number
  pageSize: number
  person_id?: number
}

export interface PositionSnapshotListResponse {
  list: PositionSnapshot[]
  total: number
}

export function listEvents(params: PositionEventListParams): Promise<PositionEventListResponse> {
  return get<PositionEventListResponse>('/api/position-events', params as unknown as Record<string, unknown>)
}

export function createEvent(data: PositionEventCreateRequest): Promise<void> {
  return post<void>('/api/position-events', data)
}

export function updateEvent(data: PositionEventUpdateRequest): Promise<void> {
  return put<void>('/api/position-events', data)
}

export function deleteEvent(id: number): Promise<void> {
  return del<void>('/api/position-events', { id })
}

export function restoreEvent(id: number): Promise<void> {
  return post<void>('/api/position-events/restore', { id })
}

export function listSnapshots(params: PositionSnapshotListParams): Promise<PositionSnapshotListResponse> {
  return get<PositionSnapshotListResponse>('/api/position-snapshots', params as unknown as Record<string, unknown>)
}
