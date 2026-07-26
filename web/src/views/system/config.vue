<template>
  <div class="page-container">
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="config_name" label="配置名称" width="200" />
      <el-table-column prop="config_key" label="配置键" width="280" />
      <el-table-column label="当前值" min-width="200">
        <template #default="{ row }">
          <template v-if="row.config_key === 'system.encrypt_key'">
            <el-input :model-value="row.config_value ? '******' + row.config_value.slice(-8) : ''" disabled readonly />
          </template>
          <el-input v-else-if="row.value_type === 'number'" v-model="row.config_value" @blur="updateValue(row)" />
          <el-switch v-else-if="row.value_type === 'bool'" v-model="row.config_value" active-value="true" inactive-value="false" @change="updateValue(row)" />
          <el-input v-else v-model="row.config_value" @blur="updateValue(row)" />
        </template>
      </el-table-column>
      <el-table-column prop="config_desc" label="说明" width="200" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getConfigs, updateConfig } from '@/api/audit'

const loading = ref(false)
const list = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const data = await getConfigs()
    list.value = Array.isArray(data) ? data : []
  } catch (e) {} finally { loading.value = false }
}

async function updateValue(row: any) {
  if (row.config_key === 'system.encrypt_key') {
    ElMessage.warning('系统加密密钥不可修改')
    return
  }
  try {
    await updateConfig(row.id, { value: String(row.config_value) })
    ElMessage.success('配置已更新')
  } catch (e) {}
}

onMounted(fetchData)
</script>
