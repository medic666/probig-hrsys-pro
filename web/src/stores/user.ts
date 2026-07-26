import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, changePassword as changePasswordApi, getUserInfo as getUserInfoApi } from '@/api/auth'
import type { UserInfo } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)
  const isFirstLogin = ref<boolean>(false)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(val: string) {
    token.value = val
    if (val) {
      localStorage.setItem('token', val)
    } else {
      localStorage.removeItem('token')
    }
  }

  async function login(username: string, password: string) {
    const res = await loginApi({ username, password })
    setToken(res.token)
    isFirstLogin.value = res.is_first_login
    return res
  }

  async function logout() {
    setToken('')
    userInfo.value = null
    isFirstLogin.value = false
  }

  async function fetchUserInfo() {
    const info = await getUserInfoApi()
    userInfo.value = info
    return info
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await changePasswordApi({ old_password: oldPassword, new_password: newPassword })
    isFirstLogin.value = false
  }

  return {
    token,
    userInfo,
    isFirstLogin,
    isLoggedIn,
    setToken,
    login,
    logout,
    fetchUserInfo,
    changePassword
  }
})
