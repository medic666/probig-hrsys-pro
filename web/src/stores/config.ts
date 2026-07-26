import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getConfigs } from '@/api/audit'

export const useConfigStore = defineStore('config', () => {
  const configs = ref<Record<string, string>>({})

  async function fetchConfigs() {
    try {
      const data = await getConfigs()
      if (Array.isArray(data)) {
        for (const c of data) {
          configs.value[c.config_key] = c.config_value
        }
      }
    } catch (e) {
      // ignore
    }
  }

  function getConfig(key: string, defaultValue = ''): string {
    return configs.value[key] || defaultValue
  }

  return { configs, fetchConfigs, getConfig }
})
