<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文件管理</span>
          <el-upload :action="`/api/files/upload`" :headers="uploadHeaders" :on-success="onUploadSuccess" :show-file-list="false">
            <el-button type="primary">上传文件</el-button>
          </el-upload>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="query" placeholder="搜索文件名" style="width:240px" clearable @keyup.enter="fetchList" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="name" label="文件名" min-width="200" />
        <el-table-column prop="mime_type" label="类型" width="120" />
        <el-table-column label="大小" width="100">
          <template #default="{ row }">{{ (row.size / 1024).toFixed(1) }} KB</template>
        </el-table-column>
        <el-table-column prop="created_at" label="上传时间" width="170" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleDownload(row.id)">下载</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
              <template #reference><el-button size="small" type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top:16px"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20)
const loading = ref(false); const query = ref('')
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('token')}` }

async function fetchList() {
  loading.value = true
  const res = await request.get('/files', { params: { query: query.value, page: page.value, page_size: pageSize.value } })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function onUploadSuccess() { fetchList(); ElMessage.success('上传成功') }

function handleDownload(id: number) {
  window.open(`/api/files/${id}/download?token=${localStorage.getItem('token')}`, '_blank')
}

async function handleDelete(id: number) {
  await request.delete(`/files/${id}`)
  fetchList(); ElMessage.success('删除成功')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-bar { margin-bottom: 16px; }
</style>
