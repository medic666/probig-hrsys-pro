<template>
  <div>
    <el-card>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="config_name" label="配置名称" min-width="160" />
        <el-table-column prop="config_key" label="配置键" min-width="200" />
        <el-table-column prop="config_desc" label="说明" min-width="200" />
        <el-table-column label="当前值" min-width="200">
          <template #default="{ row }">
            <template v-if="row.value_type === 'bool'">
              <el-switch :model-value="row.config_value === 'true'" @change="handleChange(row, $event ? 'true' : 'false')" />
            </template>
            <template v-else-if="row.value_type === 'number'">
              <el-input-number :model-value="Number(row.config_value)" :precision="2" @change="(v: number | undefined) => handleChange(row, String(v ?? ''))" />
            </template>
            <template v-else>
              <el-input :model-value="row.config_value" @change="(v: string) => handleChange(row, v)" />
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as api from '../../api/config'

const loading = ref(false)
const list = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getAllConfigs()
    list.value = res.data || []
  } finally { loading.value = false }
}

async function handleChange(row: any, newValue: string) {
  try {
    await api.updateConfig(row.config_key, { ...row, config_value: newValue })
    ElMessage.success('配置已更新')
    row.config_value = newValue
  } catch { ElMessage.error('更新失败') }
}

fetchData()
</script>
