import { defineStore } from 'pinia'
import { reactive } from 'vue'
import request from '@/utils/request'

export const useConfigStore = defineStore('config', () => {
  const configMap = reactive<Record<string, string>>({})

  function setConfig(key: string, value: string) {
    configMap[key] = value
  }

  function setConfigs(configs: Record<string, string>) {
    Object.assign(configMap, configs)
  }

  function getConfig(key: string, defaultValue = ''): string {
    return configMap[key] ?? defaultValue
  }

  async function fetchConfigs() {
    const data = (await request.get('/system/configs')) as Record<string, string>
    setConfigs(data || {})
  }

  return {
    configMap,
    setConfig,
    setConfigs,
    getConfig,
    fetchConfigs,
  }
})
