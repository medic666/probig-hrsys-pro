import { get, post, put, del } from './request'

export interface Company {
  id: number
  name: string
  credit_code: string
  address: string
  contact_phone: string
  bank_name: string
  bank_account: string
  created_at: string
  updated_at: string
}

export interface CompanyCreateRequest {
  name: string
  credit_code: string
  address?: string
  contact_phone?: string
  bank_name?: string
  bank_account?: string
}

export interface CompanyUpdateRequest extends Partial<CompanyCreateRequest> {
  id: number
}

export interface CompanyListParams {
  pageNum: number
  pageSize: number
  name?: string
  credit_code?: string
}

export interface CompanyListResponse {
  list: Company[]
  total: number
}

export function listCompanies(params: CompanyListParams): Promise<CompanyListResponse> {
  return get<CompanyListResponse>('/api/companies', params as unknown as Record<string, unknown>)
}

export function createCompany(data: CompanyCreateRequest): Promise<void> {
  return post<void>('/api/companies', data)
}

export function updateCompany(data: CompanyUpdateRequest): Promise<void> {
  return put<void>('/api/companies', data)
}

export function deleteCompany(id: number): Promise<void> {
  return del<void>('/api/companies', { id })
}

export function restoreCompany(id: number): Promise<void> {
  return post<void>('/api/companies/restore', { id })
}

export function getCompanyDetail(id: number): Promise<Company> {
  return get<Company>('/api/companies/detail', { id })
}

export function listCompanyTrash(params: { pageNum: number; pageSize: number; name?: string }): Promise<CompanyListResponse> {
  return get<CompanyListResponse>('/api/companies/trash', params as unknown as Record<string, unknown>)
}
