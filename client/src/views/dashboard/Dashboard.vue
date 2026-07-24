<template>
  <div>
    <h3>欢迎使用企业人事与行政管理系统</h3>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6" v-for="item in stats" :key="item.label">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 28px; color: #409EFF">{{ item.value }}</div>
            <div style="margin-top: 8px; color: #909399">{{ item.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'

const stats = ref([
  { label: '人员总数', value: 0 },
  { label: '公司总数', value: 0 },
  { label: '文件总数', value: 0 },
  { label: '用户总数', value: 0 },
])

onMounted(async () => {
  try {
    const [p, c, f, u] = await Promise.all([
      request.get('/persons', { params: { page_size: 1 } }),
      request.get('/companies', { params: { page_size: 1 } }),
      request.get('/files', { params: { page_size: 1 } }),
      request.get('/users', { params: { page_size: 1 } }),
    ])
    stats.value[0].value = p.data.total || 0
    stats.value[1].value = c.data.total || 0
    stats.value[2].value = f.data.total || 0
    stats.value[3].value = u.data.total || 0
  } catch {}
})
</script>
