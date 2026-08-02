import axios, { type AxiosInstance } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const instance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

instance.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

instance.interceptors.response.use(
  (response) => {
    if (response.config.responseType === 'blob') {
      const disposition = response.headers['content-disposition'] || ''
      let filename = ''
      // RFC 5987：优先 filename*=UTF-8''<urlencoded>，回退 filename=
      const star = disposition.match(/filename\*=UTF-8''([^;]+)/i)
      if (star) {
        filename = decodeURIComponent(star[1].trim())
      } else {
        const match = disposition.match(/filename=([^;]+)/)
        if (match) filename = decodeURIComponent(match[1].trim())
      }
      return { blob: response.data, filename }
    }
    const { data } = response
    if (data.code !== 0) {
      ElMessage.error(data.msg || '请求失败')
      return Promise.reject(new Error(data.msg || '请求失败'))
    }
    return data.data
  },
  (error) => {
    if (error.response) {
      const { status } = error.response
      if (status === 401) {
        const userStore = useUserStore()
        userStore.clearUser()
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
      } else if (status === 403) {
        ElMessage.error('无操作权限')
      } else {
        ElMessage.error(error.response.data?.msg || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请稍后重试')
    }
    return Promise.reject(error)
  },
)

export default instance
