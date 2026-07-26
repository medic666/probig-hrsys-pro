import request from './request'

export function login(data: { username: string; password: string }) {
  return request({ url: '/login', method: 'post', data })
}

export function getCurrentUser() {
  return request({ url: '/current-user', method: 'get' })
}

export function changePassword(data: { old_password: string; new_password: string }) {
  return request({ url: '/change-password', method: 'post', data })
}

export function getPermissions() {
  return request({ url: '/permissions', method: 'get' })
}

export function getUserList(params: any) {
  return request({ url: '/users', method: 'get', params })
}

export function createUser(data: any) {
  return request({ url: '/users', method: 'post', data })
}

export function updateUser(id: number, data: any) {
  return request({ url: `/users/${id}`, method: 'put', data })
}

export function deleteUser(id: number) {
  return request({ url: `/users/${id}`, method: 'delete' })
}

export function toggleUserStatus(id: number) {
  return request({ url: `/users/${id}/status`, method: 'put' })
}

export function resetUserPassword(id: number) {
  return request({ url: `/users/${id}/reset-password`, method: 'post' })
}

export function getUserRoles(userId: number) {
  return request({ url: `/roles?user_id=${userId}`, method: 'get' })
}

export function getRoles() {
  return request({ url: '/roles', method: 'get' })
}

export function createRole(data: any) {
  return request({ url: '/roles', method: 'post', data })
}

export function updateRole(id: number, data: any) {
  return request({ url: `/roles/${id}`, method: 'put', data })
}

export function deleteRole(id: number) {
  return request({ url: `/roles/${id}`, method: 'delete' })
}

export function getRolePermissions(roleId: number) {
  return request({ url: `/roles/${roleId}/permissions`, method: 'get' })
}

export function setRolePermissions(roleId: number, data: { perm_ids: number[] }) {
  return request({ url: `/roles/${roleId}/permissions`, method: 'put', data })
}

export function getAllPermissions() {
  return request({ url: '/roles/permissions', method: 'get' })
}
