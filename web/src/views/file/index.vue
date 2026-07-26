<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.file_name" placeholder="文件名" clearable style="width:200px;" />
      <el-input v-model="search.file_type" placeholder="文件类型(.pdf)" clearable style="width:120px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-upload v-permission="'file:write'" :show-file-list="false" :before-upload="handleUpload">
        <el-button type="primary">上传文件</el-button>
      </el-upload>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="file_name" label="文件名" min-width="200" />
      <el-table-column prop="file_type" label="类型" width="80" />
      <el-table-column label="大小" width="100">
        <template #default="{ row }">{{ fmtSize(row.file_size) }}</template>
      </el-table-column>
      <el-table-column prop="upload_user" label="上传人" width="100" />
      <el-table-column prop="created_at" label="上传时间" width="160" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <a :href="`/api/files/${row.id}/download`" target="_blank">
            <el-button size="small">下载</el-button>
          </a>
          <el-button size="small" type="danger" v-permission="'file:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFileList, uploadFile, deleteFile } from '@/api/file'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ file_name: '', file_type: '' })

function fmtSize(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return bytes.toFixed(1) + ' ' + units[i]
}

async function fetchData() {
  loading.value = true
  try {
    const data = await getFileList({ pageNum: pageNum.value, pageSize: pageSize.value, ...search })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.file_name = ''; search.file_type = ''; pageNum.value = 1; fetchData() }

async function handleUpload(file: File) {
  const form = new FormData()
  form.append('file', file)
  try {
    await uploadFile(form)
    ElMessage.success('上传成功')
    fetchData()
  } catch (e) {}
  return false
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteFile(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
