import http from './request'

export function login(data: { username: string; password: string }) {
  return http.post('/login', data)
}

export function getUserInfo() {
  return http.get('/user/info')
}

export function changePassword(data: { old_password: string; new_password: string }) {
  return http.post('/user/change-password', data)
}

export function resetUserPassword(data: { user_id: number; new_password: string }) {
  return http.post('/user/reset-password', data)
}
