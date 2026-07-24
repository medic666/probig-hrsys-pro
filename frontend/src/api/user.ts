import http from './request'

export function getUserList(params: any) {
  return http.get('/users', { params })
}

export function createUser(data: any) {
  return http.post('/users', data)
}

export function updateUser(id: number, data: any) {
  return http.put(`/users/${id}`, data)
}

export function deleteUser(id: number) {
  return http.delete(`/users/${id}`)
}

export function getRoleList() {
  return http.get('/roles')
}

export function createRole(data: any) {
  return http.post('/roles', data)
}

export function updateRole(id: number, data: any) {
  return http.put(`/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return http.delete(`/roles/${id}`)
}

export function getAllPermissions() {
  return http.get('/permissions')
}
