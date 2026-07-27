import request from '@/utils/request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  user: { id: number; username: string; name: string }
  is_first_login: boolean
  permissions: string[]
  menus: any[]
}

export function login(params: LoginParams): Promise<LoginResult> {
  return request.post('/auth/login', params) as Promise<LoginResult>
}

export function logout(): Promise<void> {
  return request.post('/auth/logout')
}

export function changePassword(old_password: string, new_password: string): Promise<void> {
  return request.post('/auth/change-password', { old_password, new_password })
}

export function getMe(): Promise<{ user: any; permissions: string[]; menus: any[] }> {
  return request.get('/auth/me') as Promise<{ user: any; permissions: string[]; menus: any[] }>
}
