import request from '@/utils/request'

export function getRoles(params: any) {
  return request.get('/roles', { params })
}

export function getAllRoles() {
  return request.get('/roles/all')
}

export function createRole(data: any) {
  return request.post('/roles', data)
}

export function updateRole(id: number, data: any) {
  return request.put(`/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete(`/roles/${id}`)
}

export function restoreRole(id: number) {
  return request.post(`/roles/${id}/restore`)
}

export function getDeletedRoles(params: any) {
  return request.get('/roles/trash', { params })
}

export function assignRolePermissions(id: number, permission_ids: number[]) {
  return request.post(`/roles/${id}/assign-permissions`, { permission_ids })
}

export function getRolePermissions(id: number) {
  return request.get(`/roles/${id}/permissions`)
}

export function getPermissions() {
  return request.get('/permissions')
}
