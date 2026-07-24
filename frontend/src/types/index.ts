export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface PageData<T = any> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface User {
  id: number
  username: string
  real_name: string
  status: string
  roles: Role[]
  created_at: string
}

export interface Role {
  id: number
  name: string
  description: string
  permissions: Permission[]
}

export interface Permission {
  id: number
  module: string
  action: string
}

export interface Entity {
  id: number
  type: string
  name: string
  status: string
}

export interface PersonnelSnapshot {
  id: number
  entity_id: number
  event_id: number
  effective_date: string
  is_latest: boolean
  name: string
  attendance_group: string
  hire_date: string | null
  base_salary: number
  performance_salary: number
  pay_days: number
  position_allowance: number
  meal_subsidy: number
  housing_subsidy: number
  transport_subsidy: number
  heat_subsidy: number
  insurance_compensation: number
  housing_fund_compensation: number
  social_insurance_deduct: number
  housing_fund_deduct: number
  extended_info: Record<string, any>
  created_at: string
}

export interface PersonnelEvent {
  id: number
  entity_id: number
  event_type: string
  event_name: string
  effective_date: string
  name: string
  attendance_group: string
  hire_date: string | null
  base_salary: number
  performance_salary: number
  pay_days: number
  position_allowance: number
  meal_subsidy: number
  housing_subsidy: number
  transport_subsidy: number
  heat_subsidy: number
  insurance_compensation: number
  housing_fund_compensation: number
  social_insurance_deduct: number
  housing_fund_deduct: number
  extended_info: Record<string, any>
  changed_fields: Record<string, any>
  created_by: number
  created_at: string
  updated_at: string
}

export interface OrganizationSnapshot {
  id: number
  entity_id: number
  event_id: number
  effective_date: string
  is_latest: boolean
  company_name: string
  credit_code: string
  address: string
  phone: string
  bank_name: string
  bank_account: string
  business_license_file_id: number | null
  official_seal_file_id: number | null
  created_at: string
}

export interface OrganizationEvent {
  id: number
  entity_id: number
  event_type: string
  event_name: string
  effective_date: string
  company_name: string
  credit_code: string
  address: string
  phone: string
  bank_name: string
  bank_account: string
  business_license_file_id: number | null
  official_seal_file_id: number | null
  changed_fields: Record<string, any>
  created_by: number
  created_at: string
  updated_at: string
}

export interface AttendanceEvent {
  id: number
  entity_id: number
  entity_name: string
  event_category: string
  event_subtype: string
  event_date: string
  duration_days: number
  description: string
  created_by: number
  created_at: string
  updated_at: string
}

export interface AttendanceSummary {
  id: number
  entity_id: number
  entity_name: string
  period_start: string
  period_end: string
  normal_days: number
  makeup_days: number
  lieu_days: number
  personal_days: number
  sick_days: number
  annual_days: number
  statutory_days: number
  welfare_days: number
  workday_overtime: number
  holiday_overtime: number
  missing_card_count: number
  late_count: number
  early_count: number
  annual_allocated: number
  annual_carried_over: number
  calculated_at: string
}

export interface SalaryEvent {
  id: number
  entity_id: number
  entity_name: string
  event_type: string
  amount: number
  description: string
  period_start: string
  period_end: string
  created_by: number
  created_at: string
  updated_at: string
}

export interface SalarySummary {
  id: number
  entity_id: number
  entity_name: string
  period_start: string
  period_end: string
  base_salary: number
  daily_salary: number
  attendance_wage: number
  full_attendance_bonus: number
  overtime_wage: number
  performance_salary: number
  position_allowance: number
  meal_subsidy: number
  housing_subsidy: number
  transport_subsidy: number
  heat_subsidy: number
  insurance_compensation: number
  housing_fund_compensation: number
  performance_adjustment: number
  reward_punishment: number
  loan_deduction: number
  social_insurance_deduct: number
  housing_fund_deduct: number
  tax_deduction: number
  gross_pay: number
  net_pay: number
  calculated_at: string
}

export interface FileItem {
  id: number
  filename: string
  original_name: string
  mime_type: string
  size: number
  uploaded_by: number
  created_at: string
}

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  target_type: string
  target_id: number
  target_name: string
  target_summary: string
  payload: Record<string, any>
  created_at: string
}
