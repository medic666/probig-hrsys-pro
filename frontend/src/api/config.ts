import http from './request'

export function getAllConfigs() {
  return http.get('/configs')
}

export function updateConfig(key: string, data: any) {
  return http.put(`/configs/${key}`, data)
}
