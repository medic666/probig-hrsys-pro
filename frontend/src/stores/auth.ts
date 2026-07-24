import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, Permission } from '../types'
import { login as apiLogin, getMe, getPermissions } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)
  const permissions = ref<Permission[]>([])

  const isLoggedIn = computed(() => !!token.value)

  function hasPermission(module: string, action: string): boolean {
    return permissions.value.some((p) => p.module === module && p.action === action)
  }

  function hasAnyPermission(module: string): boolean {
    return permissions.value.some((p) => p.module === module)
  }

  async function login(username: string, password: string) {
    const res = await apiLogin(username, password)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    await fetchPermissions()
  }

  async function fetchUser() {
    try {
      const res = await getMe()
      user.value = res.data
    } catch {
      logout()
    }
  }

  async function fetchPermissions() {
    try {
      const res = await getPermissions()
      permissions.value = res.data
    } catch {
      permissions.value = []
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    permissions.value = []
    localStorage.removeItem('token')
  }

  return {
    token,
    user,
    permissions,
    isLoggedIn,
    hasPermission,
    hasAnyPermission,
    login,
    fetchUser,
    fetchPermissions,
    logout,
  }
})
