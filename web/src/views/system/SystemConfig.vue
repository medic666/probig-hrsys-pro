<template>
  <div class="page-container">
    <div class="page-header"><h2>系统配置</h2></div>
    <el-card v-for="group in groups" :key="group.name" class="group-card">
      <template #header><span class="group-title">{{ group.name }}</span></template>
      <el-table :data="group.items" border>
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="key" label="配置键" width="240" />
        <el-table-column label="值">
          <template #default="{ row }">
            <el-input v-if="row.value_type === 'string'" v-model="row.draft" size="small" />
            <el-input-number v-else-if="row.value_type === 'number'" v-model="row.draft" size="small" :precision="2" />
            <el-switch v-else-if="row.value_type === 'bool'" v-model="row.draft" />
            <el-checkbox-group v-else-if="row.value_type === 'select'" v-model="row.draftArr" size="small">
              <el-checkbox v-for="o in row.options" :key="o" :label="o" :value="o" />
            </el-checkbox-group>
          </template>
        </el-table-column>
        <el-table-column prop="desc" label="说明" />
      </el-table>
    </el-card>
    <div style="margin-top:16px;display:flex;justify-content:flex-end">
      <el-button type="primary" @click="saveAll">保存配置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

interface ConfigItem {
  key: string
  name: string
  desc: string
  value_type: string
  options?: string[]
  value: string
  group: string
  draft: any
  draftArr?: string[]
}

const configList = ref<ConfigItem[]>([])

const groups = computed(() => {
  const map = new Map<string, ConfigItem[]>()
  for (const item of configList.value) {
    if (!map.has(item.group)) map.set(item.group, [])
    map.get(item.group)!.push(item)
  }
  return Array.from(map.entries()).map(([name, items]) => ({ name, items }))
})

onMounted(async () => {
  const raw = (await request.get('/system-configs')) as ConfigItem[]
  configList.value = (raw || []).map((item) => {
    let draft: any = item.value
    if (item.value_type === 'number') {
      draft = Number(item.value)
    } else if (item.value_type === 'bool') {
      draft = item.value === 'true'
    } else if (item.value_type === 'select') {
      try { draft = JSON.parse(item.value) } catch { draft = [] }
    }
    return { ...item, draft, draftArr: Array.isArray(draft) ? draft : undefined }
  })
})

async function saveAll() {
  let count = 0
  for (const item of configList.value) {
    let newVal: string
    if (item.value_type === 'select') newVal = JSON.stringify(item.draftArr || [])
    else if (item.value_type === 'bool') newVal = String(item.draft)
    else newVal = String(item.draft)
    if (newVal !== item.value) {
      try {
        await request.put(`/system-configs/${item.key}`, { value: newVal })
        item.value = newVal
        count++
      } catch { /* error handled by interceptor */ }
    }
  }
  ElMessage.success(count > 0 ? `已保存 ${count} 项配置` : '无变更')
}
</script>

<style scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
.group-card { margin-bottom: 16px; }
.group-title { font-weight: 600; }
</style>
