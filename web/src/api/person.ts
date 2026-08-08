import request from '@/utils/request'

// 人员模块业务端点（需 person.read/write/export 权限）。
// 结构授权点调用（人员选项/卡片）见 api/reference.ts。

export function getPersons(params: any) {
  return request.get('/persons', { params })
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
