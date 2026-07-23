export interface User {
  id: number;
  username: string;
  role_id: number;
  entity_id?: number;
  status: string;
  created_at: string;
}

export interface LoginResponse {
  token: string;
  user: User;
  perms: string[];
}

export interface PageData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface Entity {
  id: number;
  type: string;
  name: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface PersonSnapshotData {
  name: string;
  attendance_group: string;
  entry_date: string;
  basic_salary: number;
  performance_salary: number;
  salary_days: number;
  position_allowance: number;
  meal_subsidy: number;
  housing_subsidy: number;
  transport_subsidy: number;
  heat_subsidy: number;
  insurance_compensation: number;
  housing_fund_compensation: number;
  social_insurance_deduct: number;
  housing_fund_deduct: number;
  phones: string[];
  emails: string[];
  id_number: string;
  gender: string;
  birthday: string;
  ethnicity: string;
  native_place: string;
  address: string;
  bank_cards: string[];
  political_status: string;
  marital_status: string;
  alias: string;
}

export interface OrgSnapshotData {
  company_name: string;
  credit_code: string;
  address: string;
  contact_phone: string;
  bank_name: string;
  bank_account: string;
  business_license_file_id?: number;
  seal_file_id?: number;
}

export interface PersonEventRequest {
  person_id: number;
  effective_date: string;
  event_type: string;
  data: PersonSnapshotData;
}

export interface OrgEventRequest {
  org_id: number;
  effective_date: string;
  event_type: string;
  data: OrgSnapshotData;
}

export interface AttendanceEventRequest {
  person_id: number;
  date: string;
  event_type: string;
  duration: number;
  remark: string;
}

export interface SalaryEventRequest {
  person_id: number;
  period: string;
  event_type: string;
  amount: number;
  detail: string;
}

export interface CalculateRequest {
  period: string;
  person_id?: number;
}
