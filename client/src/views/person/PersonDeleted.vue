<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header"><span>人员回收站</span></div>
      </template>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="alias" label="别名" min-width="100" />
        <el-table-column prop="deleted_at" label="删除时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="handleRestore(row.id)">恢复</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

async function fetchList() {
  loading.value = true
  const res = await request.get('/persons/deleted', { params: { page: page.value, page_size: pageSize.value } })
  list.value = res.data.list
  total.value = res.data.total
  loading.value = false
}

async function handleRestore(id: number) {
  await request.put(`/persons/${id}/restore`)
  ElMessage.success('恢复成功')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
