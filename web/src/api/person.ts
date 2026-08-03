import request from '@/utils/request'

// PersonOption 人员选项：基础信息 + 当前快照段域字段（公司/考勤组/在职状态），
// 供人员选择组件（全体/公司/考勤组/在职状态多域筛选）与 NameSelect 共用
export interface PersonOption {
  id: number
  name: string
  company_id?: number
  company_name?: string
  attendance_group?: string
  is_active?: boolean
  entry_date?: string | null
  leave_date?: string | null
}

export function getPersons(params: any) {
  return request.get('/persons', { params })
}

export function getAllPersons(): Promise<PersonOption[]> {
  return request.get('/persons/all') as Promise<PersonOption[]>
}

export function getPersonCards() {
  return request.get('/persons/cards')
}

export function getPerson(id: number) {
  return request.get(`/persons/${id}`)
}

export function upsertPersonProfile(data: any) {
  return request.post('/persons/profile', data)
}

export function deletePerson(id: number) {
  return request.delete(`/persons/${id}`)
}

export function restorePerson(id: number) {
  return request.post(`/persons/${id}/restore`)
}

export function getDeletedPersons(params: any) {
  return request.get('/persons/trash', { params })
}

export function exportPersons(params: any) {
  return request.get('/persons/export', { params, responseType: 'blob' })
}
