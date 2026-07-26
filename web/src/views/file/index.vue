<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Delete, Download, RefreshLeft } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import { listFiles, deleteFile, restoreFile, downloadFile } from '@/api/file'
import type { FileListParams } from '@/api/file'
import { downloadFile as downloadUtil } from '@/utils/download'

const tableRef = ref()
const trashVisible = ref(false)

const searchFields = [
  { prop: 'file_name', label: '文件名', type: 'input' as const, placeholder: '请输入文件名' },
  { prop: 'file_type', label: '文件类型', type: 'input' as const },
  { prop: 'upload_time_start', label: '上传时间起', type: 'date' as const },
  { prop: 'upload_time_end', label: '上传时间止', type: 'date' as const }
]

const columns = [
  { prop: 'file_name', label: '文件名' },
  { prop: 'file_type', label: '文件类型' },
  { prop: 'file_size', label: '文件大小' },
  { prop: 'created_at', label: '上传时间' },
  { slot: 'actions', label: '操作', width: 200, fixed: 'right' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listFiles(params as unknown as FileListParams)
}

async function fetchTrashList(params: Record<string, unknown>) {
  return listFiles({ ...params, trash: true } as unknown as FileListParams)
}

function handleUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('file', file)
    ElMessage.success('上传成功')
    tableRef.value?.refresh()
  }
  input.click()
}

async function handleDelete(row: { id: number; file_name: string }) {
  try {
    await ElMessageBox.confirm(`确定要删除文件「${row.file_name}」吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteFile(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

async function handleRestore(row: { id: number; file_name: string }) {
  try {
    await ElMessageBox.confirm(`确定要恢复文件「${row.file_name}」吗？`, '确认恢复', { type: 'warning' })
  } catch {
    return
  }
  await restoreFile(row.id)
  ElMessage.success('恢复成功')
  tableRef.value?.refresh()
}

async function handleDownload(row: { id: number; file_name: string }) {
  try {
    const blob = await downloadFile(row.id)
    downloadUtil(blob, row.file_name)
  } catch {
    ElMessage.error('下载失败')
  }
}

const trashColumns = [
  { prop: 'file_name', label: '文件名' },
  { prop: 'file_type', label: '文件类型' },
  { slot: 'trashActions', label: '操作', width: 120, fixed: 'right' as const }
]
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Upload" @click="handleUpload">上传文件</el-button>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Delete" @click="trashVisible = true">回收站</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="Download" @click="handleDownload(row)">下载</el-button>
        <el-button type="primary" link @click="handleDelete({id: row.id, file_name: row.file_name})">修改关联</el-button>
        <el-button type="danger" link :icon="Delete" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <el-drawer v-model="trashVisible" title="回收站" size="800px">
      <ProTable
        :columns="trashColumns"
        :search-fields="searchFields.slice(0, 1)"
        :api="fetchTrashList"
      >
        <template #trashActions="{ row }">
          <el-button type="primary" link :icon="RefreshLeft" @click="handleRestore(row)">恢复</el-button>
        </template>
      </ProTable>
    </el-drawer>
  </div>
</template>
