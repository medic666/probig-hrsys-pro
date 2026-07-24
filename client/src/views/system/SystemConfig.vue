<template>
  <div>
    <el-card>
      <template #header><span>系统配置</span></template>
      <el-table :data="configs" border stripe v-loading="loading">
        <el-table-column prop="config_key" label="配置键" width="200" />
        <el-table-column prop="config_name" label="名称" width="150" />
        <el-table-column prop="config_desc" label="说明" min-width="200" />
        <el-table-column prop="config_value" label="当前值" width="100" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="修改配置" width="400px">
      <el-form>
        <el-form-item label="配置键"><el-input :model-value="editRow.config_key" disabled /></el-form-item>
        <el-form-item label="值">
          <el-input v-if="editRow.value_type === 'string' || editRow.value_type === 'number'" v-model="editValue" />
          <el-select v-else-if="editRow.value_type === 'bool'" v-model="editValue" style="width:100%">
            <el-option label="开启" value="true" /><el-option label="关闭" value="false" />
          </el-select>
          <el-input v-else v-model="editValue" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const configs = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const editRow = ref<any>({})
const editValue = ref('')

async function fetchConfigs() {
  loading.value = true
  const res = await request.get('/configs')
  configs.value = res.data
  loading.value = false
}

function showEdit(row: any) {
  editRow.value = row
  editValue.value = row.config_value
  dialogVisible.value = true
}

async function handleSave() {
  await request.put('/configs', { config_key: editRow.value.config_key, config_value: editValue.value })
  dialogVisible.value = false
  fetchConfigs()
  ElMessage.success('配置已更新')
}

onMounted(fetchConfigs)
</script>
