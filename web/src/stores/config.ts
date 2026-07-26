import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listConfig } from '@/api/system'

export const useConfigStore = defineStore('config', () => {
  const configMap = ref<Record<string, string>>({})

  async function fetchConfig() {
    const res = await listConfig()
    const map: Record<string, string> = {}
    res.list.forEach((item) => {
      map[item.config_key] = item.config_value
    })
    configMap.value = map
  }

  function getConfig(key: string, defaultValue = ''): string {
    return configMap.value[key] ?? defaultValue
  }

  return {
    configMap,
    fetchConfig,
    getConfig
  }
})
