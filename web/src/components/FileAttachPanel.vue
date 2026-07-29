<template>
  <div>
    <el-table :data="fileList" border size="small">
      <el-table-column label="文件名" :formatter="(r:any)=>r.file?.original_name||'-'" />
      <el-table-column label="大小" width="100" :formatter="(r:any)=>r.file?.size?((r.file.size/1024/1024).toFixed(2)+' MB'):'-'" />
      <el-table-column label="操作" width="100">
        <template #default="{ row: r }">
          <el-button type="danger" link size="small" @click="disassociate(r.relation.id)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-upload :show-file-list="false" :http-request="handleUpload" style="margin-top:8px">
      <el-button size="small">上传文件</el-button>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFilesByTarget, uploadFile, associateFile, disassociateFile } from '@/api/file'

const props = defineProps<{ targetType: string; targetId: number | null }>()

const fileList = ref<any[]>([])

async function loadFiles() {
  if (!props.targetId) { fileList.value = []; return }
  try { fileList.value = (await getFilesByTarget(props.targetType, props.targetId)) as any[] || [] } catch { fileList.value = [] }
}

watch(() => props.targetId, () => loadFiles())
onMounted(() => loadFiles())

async function handleUpload(req: any) {
  if (!props.targetId) return
  const data = (await uploadFile(req.file)) as any
  await associateFile(data.id, props.targetType, props.targetId)
  ElMessage.success('上传成功')
  loadFiles()
}

async function disassociate(relationId: number) {
  try { await ElMessageBox.confirm('确认移除该附件？', '提示', { type: 'warning' }) } catch { return }
  await disassociateFile(relationId)
  ElMessage.success('已移除')
  loadFiles()
}
</script>
