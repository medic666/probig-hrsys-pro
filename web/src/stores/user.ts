import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

interface UserInfo {
  id: number
  username: string
  name: string
}

const TOKEN_KEY = 'probig-token'

function loadToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

function saveToken(t: string) {
  if (t) {
    localStorage.setItem(TOKEN_KEY, t)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(loadToken())
  const userInfo = ref<UserInfo | null>(null)
  const isFirstLogin = ref(false)

  function setToken(t: string) {
    token.value = t
    saveToken(t)
  }

  function setUserInfo(info: UserInfo) {
    userInfo.value = info
  }

  function clearUser() {
    token.value = ''
    userInfo.value = null
    isFirstLogin.value = false
    saveToken('')
  }

  async function fetchUserInfo() {
    const data = (await request.get('/user/info')) as UserInfo
    userInfo.value = data
  }

  return {
    token,
    userInfo,
    isFirstLogin,
    setToken,
    setUserInfo,
    clearUser,
    fetchUserInfo,
  }
})
