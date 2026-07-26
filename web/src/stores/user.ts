import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getCurrentUser, changePassword as apiChangePassword } from '@/api/rbac'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const isFirstLogin = ref(false)

  function setToken(val: string) {
    token.value = val
    localStorage.setItem('token', val)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('token')
  }

  async function fetchUserInfo() {
    try {
      const data: any = await getCurrentUser()
      userInfo.value = data
      isFirstLogin.value = data.is_first_login
      return data
    } catch (e) {
      clearToken()
      router.push('/login')
      return null
    }
  }

  async function changePassword(oldPW: string, newPW: string) {
    await apiChangePassword({ old_password: oldPW, new_password: newPW })
    isFirstLogin.value = false
  }

  function logout() {
    clearToken()
    userInfo.value = null
    router.push('/login')
  }

  return { token, userInfo, isFirstLogin, setToken, clearToken, fetchUserInfo, changePassword, logout }
})
