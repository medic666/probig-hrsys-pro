import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const permissions = ref<string[]>([])

  async function login(username: string, password: string) {
    const res = await request.post('/login', { username, password })
    token.value = res.data.token
    userInfo.value = res.data
    permissions.value = res.data.permissions || []
    localStorage.setItem('token', res.data.token)
    return res.data
  }

  async function fetchUserInfo() {
    const res = await request.get('/user/info')
    userInfo.value = res.data
    permissions.value = res.data.permissions || []
    return res.data
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    permissions.value = []
    localStorage.removeItem('token')
    window.location.hash = '#/login'
  }

  function hasPermission(perm: string): boolean {
    return permissions.value.includes(perm)
  }

  return { token, userInfo, permissions, login, fetchUserInfo, logout, hasPermission }
})
