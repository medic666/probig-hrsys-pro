import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const permissions = ref<string[]>([])

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('token', val)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('token')
    userInfo.value = null
    permissions.value = []
  }

  function hasPermission(key: string): boolean {
    if (!key) return true
    return permissions.value.includes(key)
  }

  return { token, userInfo, permissions, setToken, clearToken, hasPermission }
})
