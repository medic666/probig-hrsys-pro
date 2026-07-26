import { post, get } from './request'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  is_first_login: boolean
}

export interface UserInfo {
  id: number
  username: string
  person_id: number
  person_name: string
  is_active: boolean
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export function login(data: LoginRequest): Promise<LoginResponse> {
  return post<LoginResponse>('/api/auth/login', data)
}

export function changePassword(data: ChangePasswordRequest): Promise<void> {
  return post<void>('/api/auth/change-password', data)
}

export function getUserInfo(): Promise<UserInfo> {
  return get<UserInfo>('/api/auth/user-info')
}

export function getPermissions(): Promise<string[]> {
  return get<string[]>('/api/auth/permissions')
}
