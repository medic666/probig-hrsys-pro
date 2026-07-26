import { get, put } from './request'

export interface SysConfig {
  id: number
  config_key: string
  config_value: string
  config_name: string
  config_desc: string
  value_type: string
  option_values: string
  updated_at: string
}

export interface ConfigListResponse {
  list: SysConfig[]
}

export interface UpdateConfigRequest {
  id: number
  config_value: string
}

export function listConfig(): Promise<ConfigListResponse> {
  return get<ConfigListResponse>('/api/sys-config')
}

export function updateConfig(data: UpdateConfigRequest): Promise<void> {
  return put<void>('/api/sys-config', data)
}
