<template>
  <div>
    <el-table :data="fileList" border size="small">
      <el-table-column label="文件名" :formatter="(r:any)=>r.file?.original_name||'-'" />
      <el-table-column label="大小" width="100" :formatter="(r:any)=>r.file?.size?((r.file.size/1024/1024).toFixed(2)+' MB'):'-'" />
      <el-table-column label="操作" width="140">
        <template #default="{ row: r }">
          <el-button type="success" link size="small" @click="handleDownload(r)">下载</el-button>
          <el-button type="danger" link size="small" @click="disassociate(r.relation.id)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div style="margin-top:8px;display:flex;gap:8px">
      <el-button size="small" @click="uploadInput.click()">上传文件</el-button>
      <el-button size="small" @click="assocVisible=true">关联已有文件</el-button>
    </div>
    <input ref="uploadInput" type="file" style="display:none" @change="handleUploadInput" />

    <el-dialog v-model="assocVisible" title="选择要关联的文件" width="600px">
      <el-table :data="allFiles" border size="small" @selection-change="(rows:any)=>selectedFiles=rows">
        <el-table-column type="selection" width="50" :selectable="(r:any)=>!r._associated" />
        <el-table-column label="文件名" :formatter="(r:any)=>r.original_name||'-'" />
        <el-table-column label="大小" width="100" :formatter="(r:any)=>r.size?((r.size/1024/1024).toFixed(2)+' MB'):'-'" />
        <el-table-column label="状态" width="80">
          <template #default="{row:r}"><el-tag v-if="r._associated" type="info" size="small">已关联</el-tag></template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="assocVisible=false">取消</el-button>
        <el-button type="primary" :loading="associating" @click="handleAssociateFiles">确定关联</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFilesByTarget, getFiles, uploadFile, associateFile, disassociateFile, downloadFile } from '@/api/file'

const props = defineProps<{ targetType: string; targetId: number | null }>()

const fileList = ref<any[]>([])
const uploadInput = ref<HTMLInputElement>()
const assocVisible = ref(false)
const allFiles = ref<any[]>([])
const selectedFiles = ref<any[]>([])
const associating = ref(false)

async function loadFiles() {
  if (!props.targetId) { fileList.value = []; return }
  try { fileList.value = (await getFilesByTarget(props.targetType, props.targetId)) as any[] || [] } catch { fileList.value = [] }
}

watch(() => props.targetId, () => loadFiles())
onMounted(() => loadFiles())

async function handleUploadInput(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.[0] || !props.targetId) return
  const data = (await uploadFile(input.files[0])) as any
  await associateFile(data.id, props.targetType, props.targetId)
  ElMessage.success('上传成功')
  input.value = ''
  loadFiles()
}

async function handleAssociateFiles() {
  if (!props.targetId || selectedFiles.value.length === 0) return
  associating.value = true
  try {
    for (const f of selectedFiles.value) { await associateFile(f.id, props.targetType, props.targetId) }
    ElMessage.success(`已关联 ${selectedFiles.value.length} 个文件`)
    assocVisible.value = false; loadFiles()
  } catch { /* */ } finally { associating.value = false }
}

async function handleDownload(r: any) {
  if (r.file) downloadFile(r.file.id, r.file.original_name)
}

async function disassociate(relationId: number) {
  try { await ElMessageBox.confirm('确认移除该附件？', '提示', { type: 'warning' }) } catch { return }
  await disassociateFile(relationId)
  ElMessage.success('已移除')
  loadFiles()
}

watch(assocVisible, async (v) => {
  if (v) {
    const d = (await getFiles({pageNum:1,pageSize:100})) as any
    const files = (d.list || []).map((f:any)=>({...f,_associated:false}))
    const associated = fileList.value.map((r:any)=>r.file?.id).filter(Boolean)
    for (const f of files) { if (associated.includes(f.id)) { f._associated = true } else { f._associated = false } }
    allFiles.value = files
  }
})
</script>
