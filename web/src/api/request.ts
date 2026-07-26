import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
})

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0) {
      return res.data
    }
    if (res.code === 401) {
      localStorage.removeItem('token')
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
      return Promise.reject(new Error(res.msg || 'Error'))
    }
    if (res.code === 403) {
      ElMessage.error('无操作权限')
      return Promise.reject(new Error(res.msg || 'Error'))
    }
    ElMessage.error(res.msg || '操作失败')
    return Promise.reject(new Error(res.msg || 'Error'))
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default service
