import { get, post, put, del } from './request'

export interface LeaveAccountEvent {
  id: number
  person_id: number
  person_name: string
  leave_type: string
  event_type: string
  source_type: string
  batch_id: number
  hours: number
  effective_date: string
  remark: string
  created_at: string
}

export interface LeaveAccountEventCreateRequest {
  person_id: number
  leave_type: string
  event_type: string
  hours: number
  effective_date: string
  remark?: string
}

export interface LeaveAccountEventUpdateRequest extends Partial<LeaveAccountEventCreateRequest> {
  id: number
}

export interface LeaveAccountBalance {
  id: number
  person_id: number
  person_name: string
  leave_type: string
  balance_hours: number
  last_calc_at: string
  status: string
}

export interface BalanceDetail {
  total_grant: number
  total_used: number
  total_adjust: number
  total_carryover: number
}

export interface CarryoverBatch {
  id: number
  batch_no: string
  business_type: string
  business_period: string
  operator_id: number
  operator_name: string
  status: number
  total_count: number
  remark: string
  created_at: string
  executed_at: string
  canceled_at: string
}

export interface EventListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  leave_type?: string
  effective_date_start?: string
  effective_date_end?: string
  source_type?: string
}

export interface EventListResponse {
  list: LeaveAccountEvent[]
  total: number
}

export interface BalanceListParams {
  pageNum: number
  pageSize: number
  person_id?: number
  leave_type?: string
}

export interface BalanceListResponse {
  list: LeaveAccountBalance[]
  total: number
}

export interface BatchListParams {
  pageNum: number
  pageSize: number
  batch_no?: string
  business_period?: string
  status?: number
  operator_name?: string
}

export interface BatchListResponse {
  list: CarryoverBatch[]
  total: number
}

export interface CarryoverRequest {
  target_month: string
}

export function listEvents(params: EventListParams): Promise<EventListResponse> {
  return get<EventListResponse>('/api/leave-account-events', params as unknown as Record<string, unknown>)
}

export function createManualEvent(data: LeaveAccountEventCreateRequest): Promise<void> {
  return post<void>('/api/leave-account-events/manual', data)
}

export function updateManualEvent(data: LeaveAccountEventUpdateRequest): Promise<void> {
  return put<void>('/api/leave-account-events/manual', data)
}

export function deleteManualEvent(id: number): Promise<void> {
  return del<void>('/api/leave-account-events/manual', { id })
}

export function listBalances(params: BalanceListParams): Promise<BalanceListResponse> {
  return get<BalanceListResponse>('/api/leave-account-balances', params as unknown as Record<string, unknown>)
}

export function listBalanceDetail(personId: number, leaveType: string): Promise<BalanceDetail> {
  return get<BalanceDetail>('/api/leave-account-balances/detail', { person_id: personId, leave_type: leaveType })
}

export function listBatches(params: BatchListParams): Promise<BatchListResponse> {
  return get<BatchListResponse>('/api/leave-account-batches', params as unknown as Record<string, unknown>)
}

export function executeCarryover(data: CarryoverRequest): Promise<void> {
  return post<void>('/api/leave-account-batches/carryover', data)
}

export function cancelBatch(batchId: number): Promise<void> {
  return post<void>('/api/leave-account-batches/cancel', { batch_id: batchId })
}
