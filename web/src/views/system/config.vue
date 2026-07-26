<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listConfig, updateConfig } from '@/api/system'
import type { SysConfig } from '@/api/system'

const configList = ref<SysConfig[]>([])
const loading = ref(false)

async function fetchConfigs() {
  loading.value = true
  try {
    const res = await listConfig()
    configList.value = res.list
  } finally {
    loading.value = false
  }
}

async function handleUpdate(row: SysConfig) {
  try {
    await updateConfig({ id: row.id, config_value: row.config_value })
    ElMessage.success('配置已更新')
  } catch {
    // error handled by interceptor
  }
}

function getGroupName(key: string): string {
  const parts = key.split('.')
  return parts[0] || '其他'
}

const groupedConfigs = computed(() => {
  const groups: Record<string, SysConfig[]> = {}
  configList.value.forEach((c) => {
    const group = getGroupName(c.config_key)
    if (!groups[group]) groups[group] = []
    groups[group].push(c)
  })
  return groups
})

onMounted(() => {
  fetchConfigs()
})
</script>

<template>
  <div class="page-container">
    <div v-loading="loading">
      <el-card
        v-for="(items, group) in groupedConfigs"
        :key="group"
        class="config-group"
      >
        <template #header>
          <span class="config-group-title">{{ group }} 配置</span>
        </template>
        <el-form label-width="180px">
          <el-form-item
            v-for="item in items"
            :key="item.id"
            :label="item.config_name"
          >
            <div class="config-item">
              <el-input
                v-if="item.value_type === 'string'"
                v-model="item.config_value"
                :disabled="item.config_key === 'system.encrypt_key'"
              />
              <el-input-number
                v-else-if="item.value_type === 'number'"
                v-model="(item.config_value as unknown as number)"
                :disabled="item.config_key === 'system.encrypt_key'"
                style="width: 200px"
              />
              <el-switch
                v-else-if="item.value_type === 'bool'"
                v-model="(item.config_value as unknown as boolean)"
                :disabled="item.config_key === 'system.encrypt_key'"
              />
              <el-select
                v-else-if="item.value_type === 'select'"
                v-model="item.config_value"
                :disabled="item.config_key === 'system.encrypt_key'"
                style="width: 200px"
              >
                <el-option
                  v-for="opt in item.option_values ? JSON.parse(item.option_values) : []"
                  :key="opt"
                  :label="opt"
                  :value="opt"
                />
              </el-select>
              <span class="config-desc">{{ item.config_desc }}</span>
              <el-button
                v-if="item.config_key !== 'system.encrypt_key'"
                type="primary"
                size="small"
                style="margin-left: 12px"
                @click="handleUpdate(item)"
              >保存</el-button>
            </div>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<style scoped lang="scss">
.config-group {
  margin-bottom: 16px;
}
.config-group-title {
  font-weight: bold;
  font-size: 16px;
}
.config-item {
  display: flex;
  align-items: center;
  width: 100%;
}
.config-desc {
  color: #909399;
  font-size: 12px;
  margin-left: 12px;
}
</style>
