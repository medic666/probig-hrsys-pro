import request from './request'

export function getAuditList(params: any) {
  return request({ url: '/audits', method: 'get', params })
}

export function getConfigs() {
  return request({ url: '/configs', method: 'get' })
}

export function updateConfig(id: number, data: { value: string }) {
  return request({ url: `/configs/${id}`, method: 'put', data })
}
