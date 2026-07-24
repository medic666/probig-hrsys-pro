<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-input v-model="searchForm.keyword" placeholder="搜索文件名" style="width: 240px;" clearable @keyup.enter="fetchData" />
        <el-upload :http-request="handleUpload" :show-file-list="false">
          <el-button type="primary">上传文件</el-button>
        </el-upload>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="name" label="文件名" min-width="200" />
        <el-table-column prop="size" label="大小(KB)" min-width="90">
          <template #default="{ row }">{{ (row.size / 1024).toFixed(1) }}</template>
        </el-table-column>
        <el-table-column prop="mime_type" label="类型" min-width="120" />
        <el-table-column prop="created_at" label="上传时间" min-width="160" />
        <el-table-column label="操作" min-width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="downloadFile(row)">下载</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/file'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchForm = reactive({ keyword: '' })

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getFileList({ page: page.value, page_size: pageSize.value, keyword: searchForm.keyword })
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleUpload(options: any) {
  const formData = new FormData()
  formData.append('file', options.file)
  try {
    await api.uploadFile(formData)
    ElMessage.success('上传成功')
    fetchData()
  } catch { ElMessage.error('上传失败') }
}

function downloadFile(row: any) {
  const token = localStorage.getItem('token')
  const url = api.getFileDownloadUrl(row.id)
  const a = document.createElement('a')
  a.href = url
  const xhr = new XMLHttpRequest()
  xhr.open('GET', url, true)
  xhr.setRequestHeader('Authorization', 'Bearer ' + token)
  xhr.responseType = 'blob'
  xhr.onload = () => {
    const blob = new Blob([xhr.response])
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = row.name
    link.click()
  }
  xhr.send()
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除文件「${row.name}」？`, '确认删除', { type: 'warning' })
  await api.deleteFile(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

fetchData()
</script>
