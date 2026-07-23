import api from './client'

export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password })

export const getMe = () => api.get('/auth/me')

export const getMenus = () => api.get('/auth/menus')

export const getPermissions = () => api.get('/auth/permissions')

export const getDashboardStats = () => api.get('/dashboard/stats')
