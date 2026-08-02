import request from '@/utils/request'

export function getUsers(params: any) {
  return request.get('/users', { params })
}

export function getUser(id: number) {
  return request.get(`/users/${id}`)
}

export function createUser(data: any) {
  return request.post('/users', data)
}

export function updateUser(id: number, data: any) {
  return request.put(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete(`/users/${id}`)
}

export function restoreUser(id: number) {
  return request.post(`/users/${id}/restore`)
}

export function resetPassword(id: number) {
  return request.post(`/users/${id}/reset-password`)
}

export function assignUserRoles(id: number, role_ids: number[]) {
  return request.post(`/users/${id}/assign-roles`, { role_ids })
}

export function getDeletedUsers(params: any) {
  return request.get('/users/trash', { params })
}


