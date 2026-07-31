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
    const data = (await request.get('/system-configs')) as {
      key: string
      value: string
    }[]
    const map: Record<string, string> = {}
    for (const item of data || []) {
      map[item.key] = item.value
    }
    setConfigs(map)
  }

  return {
    configMap,
    setConfig,
    setConfigs,
    getConfig,
    fetchConfigs,
  }
})
