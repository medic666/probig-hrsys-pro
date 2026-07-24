import { post, get } from './client'
import type { User, Permission } from '../types'

export function login(username: string, password: string) {
  return post<{ token: string; user: User }>('/auth/login', { username, password })
}

export function getMe() {
  return get<User>('/auth/me')
}

export function getPermissions() {
  return get<Permission[]>('/auth/permissions')
}
