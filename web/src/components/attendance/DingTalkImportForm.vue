<template>
  <div class="dingtalk-import-form">
    <el-steps :active="step" simple style="margin-bottom:16px">
      <el-step title="上传文件" />
      <el-step title="匹配确认" />
      <el-step title="导入执行" />
    </el-steps>

    <div v-if="step === 0">
      <el-upload ref="uploadRef" :auto-upload="false" :limit="1" accept=".xlsx" :on-change="onFileChange" :on-remove="()=>importFile=null">
        <el-button type="primary">选择钉钉月度汇总文件</el-button>
      </el-upload>
      <el-button style="margin-top:12px" type="primary" :loading="previewing" :disabled="!importFile" @click="doPreview">解析预览</el-button>
    </div>

    <div v-else-if="step === 1">
      <el-table :data="preview" border size="small" max-height="360">
        <el-table-column prop="excel_name" label="Excel姓名" width="120" />
        <el-table-column label="匹配状态" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.confidence==='exact'" type="success" size="small">精确匹配</el-tag>
            <el-tag v-else-if="row.confidence==='fuzzy'" type="warning" size="small">模糊匹配</el-tag>
            <el-tag v-else type="danger" size="small">未匹配</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配人员">
          <template #default="{ row }">
            <NameSelect v-model="row.person_id" :fetch-api="fetchPersonOpts" placeholder="选择人员" />
          </template>
        </el-table-column>
        <el-table-column prop="matched_name" label="建议匹配" width="110" />
      </el-table>
      <div class="import-hint">未匹配人员请手动选择，已匹配可改选</div>
      <el-button style="margin-top:8px" @click="step=0">上一步</el-button>
      <el-button style="margin-top:8px" type="primary" :disabled="preview.some(r=>!r.person_id)" @click="step=2">下一步</el-button>
    </div>

    <div v-else>
      <el-form label-width="90px">
        <el-form-item label="归属月份" required>
          <el-date-picker v-model="month" type="month" value-format="YYYY-MM" style="width:100%" />
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" title="导入将覆盖当月已有记录（明细/打卡时间/状态）；标记为「待确认」的记录请到待确认页面核实后再参与核算。" style="margin-bottom:12px" />
      <el-button @click="step=1">上一步</el-button>
      <el-button type="primary" :loading="importing" @click="doImport">确认导入</el-button>
    </div>

    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { dingTalkPreview, dingTalkExecute } from '@/api/attendance'
import { getAllPersons } from '@/api/person'

// 钉钉月度汇总导入三步向导：上传解析 → 人员匹配确认 → 月份与执行。
// 文件解析预览与幂等执行均由后端承担，页面只编排流程。
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const step = ref(0)
const importFile = ref<File | null>(null)
const preview = ref<any[]>([])
const filePath = ref('')
const month = ref('')
const previewing = ref(false)
const importing = ref(false)
const uploadRef = ref()

function onFileChange(file: any) { importFile.value = file.raw || null }

async function fetchPersonOpts(k?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[]
  return k ? list.filter(p => p.name.includes(k)) : list
}

async function doPreview() {
  if (!importFile.value) return
  previewing.value = true
  try {
    const d = await dingTalkPreview(importFile.value) as any
    preview.value = (d.preview || []).map((p: any) => ({ ...p, person_id: p.matched_id || null }))
    filePath.value = d.file_path
    step.value = 1
  } catch { /* handled */ } finally { previewing.value = false }
}

async function doImport() {
  if (!month.value) { ElMessage.warning('请选择归属月份'); return }
  importing.value = true
  try {
    const mappings = preview.value.map((p: any) => ({ excel_name: p.excel_name, person_id: p.person_id }))
    const d = await dingTalkExecute(month.value, filePath.value, mappings) as any
    ElMessage.success(`导入完成: 创建${d.created}条, 待确认${d.pending}条`)
    emit('saved')
  } catch { /* handled */ } finally { importing.value = false }
}
</script>

<style scoped>
.import-hint { color: #909399; font-size: 12px; margin-top: 8px; }
.form-footer { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
