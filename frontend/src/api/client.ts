import axios from 'axios'
import type { ApiResponse } from '../types'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export function get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
  return client.get(url, { params }).then((r) => r.data)
}

export function post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
  return client.post(url, data).then((r) => r.data)
}

export function put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
  return client.put(url, data).then((r) => r.data)
}

export function del<T>(url: string): Promise<ApiResponse<T>> {
  return client.delete(url).then((r) => r.data)
}

export function upload<T>(url: string, formData: FormData): Promise<ApiResponse<T>> {
  return client.post(url, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }).then((r) => r.data)
}
