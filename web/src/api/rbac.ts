import { get, post, put, del } from './request'

export interface User {
  id: number
  username: string
  person_id: number
  person_name: string
  is_active: boolean
  is_first_login: boolean
  roles: Role[]
  created_at: string
}

export interface UserCreateRequest {
  username: string
  password: string
  person_id: number
}

export interface UserUpdateRequest {
  id: number
  username?: string
  person_id?: number
  is_active?: boolean
}

export interface AssignRolesRequest {
  user_id: number
  role_ids: number[]
}

export interface Role {
  id: number
  name: string
  remark: string
  is_system: boolean
  permissions: Permission[]
  created_at: string
}

export interface RoleCreateRequest {
  name: string
  remark?: string
}

export interface RoleUpdateRequest {
  id: number
  name?: string
  remark?: string
}

export interface AssignPermissionsRequest {
  role_id: number
  permission_ids: number[]
}

export interface Permission {
  id: number
  name: string
  key: string
  module: string
  action: string
}

export interface UserListParams {
  pageNum: number
  pageSize: number
  username?: string
  is_active?: boolean
}

export interface UserListResponse {
  list: User[]
  total: number
}

export interface RoleListParams {
  pageNum: number
  pageSize: number
  name?: string
}

export interface RoleListResponse {
  list: Role[]
  total: number
}

export interface PermissionTreeItem {
  module: string
  module_name: string
  permissions: Permission[]
}

export function listUsers(params: UserListParams): Promise<UserListResponse> {
  return get<UserListResponse>('/api/users', params as unknown as Record<string, unknown>)
}

export function createUser(data: UserCreateRequest): Promise<void> {
  return post<void>('/api/users', data)
}

export function updateUser(data: UserUpdateRequest): Promise<void> {
  return put<void>('/api/users', data)
}

export function deleteUser(id: number): Promise<void> {
  return del<void>('/api/users', { id })
}

export function resetPassword(id: number): Promise<void> {
  return post<void>('/api/users/reset-password', { id })
}

export function assignRoles(data: AssignRolesRequest): Promise<void> {
  return post<void>('/api/users/assign-roles', data)
}

export function listRoles(params: RoleListParams): Promise<RoleListResponse> {
  return get<RoleListResponse>('/api/roles', params as unknown as Record<string, unknown>)
}

export function createRole(data: RoleCreateRequest): Promise<void> {
  return post<void>('/api/roles', data)
}

export function updateRole(data: RoleUpdateRequest): Promise<void> {
  return put<void>('/api/roles', data)
}

export function deleteRole(id: number): Promise<void> {
  return del<void>('/api/roles', { id })
}

export function assignPermissions(data: AssignPermissionsRequest): Promise<void> {
  return post<void>('/api/roles/assign-permissions', data)
}

export function listAllPermissions(): Promise<PermissionTreeItem[]> {
  return get<PermissionTreeItem[]>('/api/permissions')
}
