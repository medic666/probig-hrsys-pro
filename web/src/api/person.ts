import { get, post, put, del } from './request'

export interface Person {
  id: number
  name: string
  id_card: string
  gender: number
  birthday: string
  nation: string
  native_place: string
  address: string
  political_status: string
  marital_status: number
  alias: string
  status: string
  attendance_group: string
  created_at: string
  updated_at: string
}

export interface PersonCreateRequest {
  name: string
  id_card: string
  gender?: number
  birthday?: string
  nation?: string
  native_place?: string
  address?: string
  political_status?: string
  marital_status?: number
  alias?: string
}

export interface PersonUpdateRequest extends Partial<PersonCreateRequest> {
  id: number
}

export interface PersonListParams {
  pageNum: number
  pageSize: number
  name?: string
  id_card?: string
  attendance_group?: string
  status?: string
}

export interface PersonListResponse {
  list: Person[]
  total: number
}

export interface PersonDetail extends Person {
  phones: Phone[]
  emails: Email[]
  bank_cards: BankCard[]
  files: FileInfo[]
}

export interface Phone {
  id: number
  person_id: number
  phone: string
}

export interface Email {
  id: number
  person_id: number
  email: string
}

export interface BankCard {
  id: number
  person_id: number
  bank_name: string
  card_number: string
  account_name: string
}

export interface FileInfo {
  id: number
  file_name: string
  file_type: string
  file_size: number
  created_at: string
}

export function listPersons(params: PersonListParams): Promise<PersonListResponse> {
  return get<PersonListResponse>('/api/persons', params as unknown as Record<string, unknown>)
}

export function createPerson(data: PersonCreateRequest): Promise<void> {
  return post<void>('/api/persons', data)
}

export function updatePerson(data: PersonUpdateRequest): Promise<void> {
  return put<void>('/api/persons', data)
}

export function deletePerson(id: number): Promise<void> {
  return del<void>('/api/persons', { id })
}

export function restorePerson(id: number): Promise<void> {
  return post<void>('/api/persons/restore', { id })
}

export function getPersonDetail(id: number): Promise<PersonDetail> {
  return get<PersonDetail>('/api/persons/detail', { id })
}

export function listPhones(personId: number): Promise<Phone[]> {
  return get<Phone[]>('/api/persons/phones', { person_id: personId })
}

export function addPhone(data: { person_id: number; phone: string }): Promise<void> {
  return post<void>('/api/persons/phones', data)
}

export function updatePhone(data: { id: number; phone: string }): Promise<void> {
  return put<void>('/api/persons/phones', data)
}

export function deletePhone(id: number): Promise<void> {
  return del<void>('/api/persons/phones', { id })
}

export function listEmails(personId: number): Promise<Email[]> {
  return get<Email[]>('/api/persons/emails', { person_id: personId })
}

export function addEmail(data: { person_id: number; email: string }): Promise<void> {
  return post<void>('/api/persons/emails', data)
}

export function updateEmail(data: { id: number; email: string }): Promise<void> {
  return put<void>('/api/persons/emails', data)
}

export function deleteEmail(id: number): Promise<void> {
  return del<void>('/api/persons/emails', { id })
}

export function listBankCards(personId: number): Promise<BankCard[]> {
  return get<BankCard[]>('/api/persons/bank-cards', { person_id: personId })
}

export function addBankCard(data: Omit<BankCard, 'id'>): Promise<void> {
  return post<void>('/api/persons/bank-cards', data)
}

export function updateBankCard(data: BankCard): Promise<void> {
  return put<void>('/api/persons/bank-cards', data)
}

export function deleteBankCard(id: number): Promise<void> {
  return del<void>('/api/persons/bank-cards', { id })
}

export function listTrash(params: { pageNum: number; pageSize: number; name?: string }): Promise<PersonListResponse> {
  return get<PersonListResponse>('/api/persons/trash', params as unknown as Record<string, unknown>)
}
