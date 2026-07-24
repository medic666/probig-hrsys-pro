<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <span style="font-size: 14px; color: #666">共 {{ total }} 个文件</span>
      <el-upload
        v-if="auth.hasPermission('file', 'write')"
        :show-file-list="false"
        :http-request="handleUpload"
        accept="*"
      >
        <el-button type="primary">上传文件</el-button>
      </el-upload>
    </div>

    <el-table :data="files" border stripe v-loading="loading">
      <el-table-column prop="original_name" label="文件名" min-width="180" />
      <el-table-column prop="mime_type" label="类型" width="100" />
      <el-table-column label="大小" width="100">
        <template #default="{ row }">{{ formatSize(row.size) }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="上传时间" width="170">
        <template #default="{ row }">{{ row.created_at?.slice(0, 19).replace('T', ' ') }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140">
        <template #default="{ row }">
          <el-button v-if="auth.hasPermission('file', 'write')" text size="small" @click="openAssociate(row)">关联</el-button>
          <el-button v-if="auth.hasPermission('file', 'delete')" text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      style="margin-top: 16px; justify-content: flex-end"
      @current-change="fetchData"
    />

    <el-dialog v-model="assocVisible" title="文件关联" width="400px">
      <el-form label-width="80px">
        <el-form-item label="关联类型">
          <el-select v-model="assocForm.target_type" style="width: 100%">
            <el-option label="人员" value="entity" />
            <el-option label="事件" value="event" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联ID">
          <el-input-number v-model="assocForm.target_id" :min="1" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assocVisible = false">取消</el-button>
        <el-button type="primary" :loading="assocLoading" @click="handleAssociate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../../stores/auth'
import { listFiles, upload, deleteFile, associateFile } from '../../api/file'
import type { FileItem } from '../../types'

const auth = useAuthStore()

const loading = ref(false)
const files = ref<FileItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

async function fetchData() {
  loading.value = true
  try {
    const res = await listFiles({ page: page.value, page_size: pageSize })
    files.value = res.data.list
    total.value = res.data.total
  } catch {} finally { loading.value = false }
}

async function handleUpload(options: any) {
  const fd = new FormData()
  fd.append('file', options.file)
  try {
    await upload(fd)
    ElMessage.success('上传成功')
    fetchData()
  } catch {
    ElMessage.error('上传失败')
  }
}

async function handleDelete(row: FileItem) {
  try {
    await ElMessageBox.confirm('确定删除该文件吗？', '确认删除', { type: 'warning' })
    await deleteFile(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch {}
}

const assocVisible = ref(false)
const assocFileId = ref(0)
const assocLoading = ref(false)
const assocForm = reactive({ target_type: 'entity', target_id: 0 })

function openAssociate(row: FileItem) {
  assocFileId.value = row.id
  assocForm.target_type = 'entity'
  assocForm.target_id = 0
  assocVisible.value = true
}

async function handleAssociate() {
  assocLoading.value = true
  try {
    await associateFile(assocFileId.value, assocForm)
    ElMessage.success('关联成功')
    assocVisible.value = false
  } catch { ElMessage.error('关联失败') } finally { assocLoading.value = false }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(1) + ' MB'
}
</script>
