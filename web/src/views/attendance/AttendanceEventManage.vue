<template>
  <div class="page-container">
    <div class="page-header">
      <h2>考勤事件管理</h2>
      <el-radio-group v-model="viewMode" size="small" style="margin-left:16px">
        <el-radio-button value="list">列表</el-radio-button>
        <el-radio-button value="blocks">每日方块</el-radio-button>
      </el-radio-group>
    </div>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" :actions="actions" @action="handleAction">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <div class="block-toolbar">
        <el-form inline>
          <el-form-item label="人员">
            <NameSelect v-model="blockPersonId" :fetch-api="fetchPersonOpts" placeholder="全部人员" style="width:180px" />
          </el-form-item>
          <el-form-item label="日期范围">
            <el-date-picker v-model="blockDateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width:260px" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="blockStatus" placeholder="全部状态" clearable style="width:130px">
              <el-option label="待确认" value="pending" />
              <el-option label="已确认" value="confirmed" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="resetBlocks">查询</el-button>
          </el-form-item>
        </el-form>
        <div class="block-actions-bar">
          <el-button type="primary" size="small" @click="handleAction('add')">新增事件</el-button>
          <el-button type="success" size="small" @click="handleAction('batch')">批量新增</el-button>
          <el-button type="warning" size="small" @click="handleAction('import')">钉钉导入</el-button>
          <el-button size="small" @click="handleAction('trash')">回收站</el-button>
          <el-button size="small" @click="handleAction('export')">导出</el-button>
        </div>
      </div>

      <div v-loading="loadingBlocks" style="min-height:120px">
        <div v-for="group in blockGroups" :key="group.person_id" class="block-group">
          <div class="group-header">
            {{ group.person_name }}
            <span class="group-count">{{ group.items.length }} 天</span>
          </div>
          <div class="block-grid">
            <AttendanceDailyBlock
              v-for="d in group.items"
              :key="d.id"
              :daily="d"
              :edited="editedSet.has(d.id)"
              @edit="openEdit"
              @confirm="confirmDaily"
            />
          </div>
        </div>
        <el-empty v-if="!loadingBlocks && blocksData.length === 0" description="暂无数据" :image-size="60" />
        <div v-if="blocksData.length < blocksTotal" class="load-more">
          <el-button :loading="loadingBlocks" @click="loadBlocksMore">
            加载更多（{{ blocksData.length }}/{{ blocksTotal }}）
          </el-button>
        </div>
      </div>

      <DailyEditDialog v-model:visible="editVisible" :daily="editDailyRow" @saved="onBlockSaved" />
    </template>

    <el-dialog v-model="dialogVisible" :title="dialogMode==='add'?'新增考勤事件':'编辑考勤事件'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" placeholder="选择人员" /></el-form-item>
        <el-form-item label="日期" required><el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type" placeholder="选择类型" style="width:100%" @change="onTypeChange">
            <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.event_type !== '违纪'" label="子类型" required>
          <el-select v-model="form.sub_type" placeholder="选择子类型" style="width:100%">
            <el-option v-for="s in currentSubTypes" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.event_type !== '违纪'" label="时长(天)" required>
          <el-input-number v-model="form.hours" :min="0" :precision="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchVisible" title="批量新增" width="500px">
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="人员">
          <el-select v-model="batchForm.person_ids" multiple placeholder="选择人员" style="width:100%">
            <el-option v-for="p in personList" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="起始日期"><el-date-picker v-model="batchForm.start_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="结束日期"><el-date-picker v-model="batchForm.end_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型">
          <el-select v-model="batchForm.event_type" style="width:100%" @change="onBatchTypeChange"><el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" /></el-select>
        </el-form-item>
        <el-form-item v-if="batchForm.event_type !== '违纪'" label="子类型">
          <el-select v-model="batchForm.sub_type" style="width:100%"><el-option v-for="s in batchSubTypes" :key="s" :label="s" :value="s" /></el-select>
        </el-form-item>
        <el-form-item v-if="batchForm.event_type !== '违纪'" label="每日时长(天)" required>
          <el-input-number v-model="batchForm.hours" :min="0" :precision="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="batchForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleBatchSubmit">确定</el-button>
      </template>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restore" :columns="trashCols" @restored="onRefresh" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'attendance_event'" :target-id="attachFileId" />
    </el-dialog>

    <el-dialog v-model="importVisible" title="钉钉考勤导入" width="720px">
      <el-steps :active="importStep" simple style="margin-bottom:16px">
        <el-step title="上传文件" />
        <el-step title="匹配确认" />
        <el-step title="导入执行" />
      </el-steps>

      <div v-if="importStep === 0">
        <el-upload ref="uploadRef" :auto-upload="false" :limit="1" accept=".xlsx" :on-change="onImportFileChange" :on-remove="()=>importFile=null">
          <el-button type="primary">选择钉钉月度汇总文件</el-button>
        </el-upload>
        <el-button style="margin-top:12px" type="primary" :loading="previewing" :disabled="!importFile" @click="doPreview">解析预览</el-button>
      </div>

      <div v-else-if="importStep === 1">
        <el-table :data="importPreview" border size="small" max-height="360">
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
        <el-button style="margin-top:8px" @click="importStep=0">上一步</el-button>
        <el-button style="margin-top:8px" type="primary" :disabled="importPreview.some(r=>!r.person_id)" @click="importStep=2">下一步</el-button>
      </div>

      <div v-else>
        <el-form label-width="90px">
          <el-form-item label="归属月份" required>
            <el-date-picker v-model="importMonth" type="month" value-format="YYYY-MM" style="width:100%" />
          </el-form-item>
        </el-form>
        <el-alert type="info" :closable="false" title="导入后系统自动标记为「待确认」的记录，请到待确认页面核实后再参与核算。" style="margin-bottom:12px" />
        <el-button @click="importStep=1">上一步</el-button>
        <el-button type="primary" :loading="importing" @click="doImport">确认导入</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import NameSelect from '@/components/NameSelect.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import AttendanceDailyBlock from '@/components/attendance/AttendanceDailyBlock.vue'
import DailyEditDialog from '@/components/attendance/DailyEditDialog.vue'
import { getAttendanceEvents, createAttendanceEvent, updateAttendanceEvent, deleteAttendanceEvent, restoreAttendanceEvent, getDeletedAttendanceEvents, createBatchAttendanceEvents, exportAttendanceEvents, confirmPendingDaily, dingTalkPreview, dingTalkExecute } from '@/api/attendance'
import { hoursToDays } from '@/utils'
import { getAllPersons } from '@/api/person'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add'|'edit'>('add')
const editId = ref(0)
const saving = ref(false)
const trashVisible = ref(false)
const batchVisible = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
const personList = ref<{id:number;name:string}[]>([])
const importVisible = ref(false)
const importStep = ref(0)
const importFile = ref<File | null>(null)
const importPreview = ref<any[]>([])
const importFilePath = ref('')
const importMonth = ref('')
const previewing = ref(false)
const importing = ref(false)
const uploadRef = ref()

const viewMode = ref<'list'|'blocks'>('list')
const blockPersonId = ref<number | null>(null)
const blockDateRange = ref<[string, string] | null>(null)
const blockStatus = ref('')
const blocksData = ref<any[]>([])
const blocksTotal = ref(0)
const blocksPage = ref(1)
const loadingBlocks = ref(false)
const editVisible = ref(false)
const editDailyRow = ref<any>(null)
const editedSet = ref<Set<number>>(new Set())

const eventTypes = ['出勤','休假','加班','违纪']
const subTypeMap: Record<string,string[]> = {
  '出勤':['普通出勤','补班出勤','外勤出勤'],
  '休假':['调休','事假','病假','年假','法定假','福利假'],
  '加班':['工作日加班','节假日加班'],
  '违纪':['缺卡','迟到','早退'],
}

const form = reactive({ person_id: null as any, event_date:'', event_type:'', sub_type:'', hours:8, punch_time:'', remark:'' })
const batchForm = reactive({ person_ids:[] as number[], start_date:'', end_date:'', event_type:'', sub_type:'', hours:8, remark:'' })

const currentSubTypes = computed(() => subTypeMap[form.event_type] || [])
const batchSubTypes = computed(() => subTypeMap[batchForm.event_type] || [])

const blockGroups = computed(() => {
  const map = new Map<number, { person_id: number; person_name: string; items: any[] }>()
  const sorted = [...blocksData.value].sort((a, b) => a.event_date.localeCompare(b.event_date))
  for (const d of sorted) {
    if (!map.has(d.person_id)) {
      map.set(d.person_id, { person_id: d.person_id, person_name: d.person_name || '-', items: [] })
    }
    map.get(d.person_id)!.items.push(d)
  }
  return Array.from(map.values())
})

watch(viewMode, (v) => {
  if (v === 'blocks' && blocksData.value.length === 0) {
    resetBlocks()
  }
})

function onTypeChange() { form.sub_type = ''; form.punch_time = ''; form.hours = form.event_type==='违纪' ? 0 : 8 }
function onBatchTypeChange() { batchForm.sub_type = ''; batchForm.punch_time = ''; batchForm.hours = batchForm.event_type==='违纪' ? 0 : 8 }

function detailSummary(r: any): string {
  const parts = (r.details || []).map((d: any) => {
    let s = d.event_type
    if (d.sub_type) s += `-${d.sub_type}`
    if (d.event_type === '违纪') {
      if (d.minutes) s += ` ${d.minutes}分钟`
    } else {
      s += ` ${hoursToDays(d.hours || 0).toFixed(2)}天`
    }
    return s
  })
  return parts.join('；') || '-'
}

const columns = [
  { prop:'id', label:'ID', width:'60' },
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'event_date', label:'日期', width:'110' },
  { prop:'status', label:'状态', width:'80', formatter:(r:any)=>({pending:'待确认',confirmed:'已确认'}[r.status]||r.status||'-') },
  { prop:'summary', label:'事件摘要', minWidth:'220', formatter:(r:any)=>detailSummary(r) },
  { prop:'punch_time', label:'打卡时间', width:'110', formatter:(r:any)=>r.punch_time || '-' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOpts },
  { prop:'event_type', label:'事件类型', type:'select' as const, options: eventTypes.map(t=>({label:t,value:t})) },
]
const actions = [
  { key:'add', label:'新增事件', type:'primary' as const },
  { key:'batch', label:'批量新增', type:'success' as const },
  { key:'import', label:'钉钉导入', type:'warning' as const },
  { key:'trash', label:'回收站', type:'default' as const },
  { key:'export', label:'导出', type:'default' as const },
]
const trashCols = [{ prop:'id', label:'ID', width:'60' },{ prop:'event_type', label:'类型' },{ prop:'event_date', label:'日期' }]

async function fetchPersonOpts(k?: string) { const l = await getAllPersons() as any[]; return k ? l.filter(p=>p.name.includes(k)) : l }
async function fetchEvents(p: any) { return (await getAttendanceEvents(p)) as any }
async function fetchDeleted(p: any) { return (await getDeletedAttendanceEvents(p)) as any }
async function restore(id: number) { return restoreAttendanceEvent(id) }

async function loadBlocks(page: number, append: boolean) {
  loadingBlocks.value = true
  try {
    const params: any = { pageNum: page, pageSize: 100 }
    if (blockPersonId.value) params.person_id = blockPersonId.value
    if (blockDateRange.value && blockDateRange.value[0]) params.date_start = blockDateRange.value[0]
    if (blockDateRange.value && blockDateRange.value[1]) params.date_end = blockDateRange.value[1]
    if (blockStatus.value) params.status = blockStatus.value
    const d = (await getAttendanceEvents(params)) as any
    blocksTotal.value = d.total || 0
    blocksData.value = append ? [...blocksData.value, ...(d.list || [])] : (d.list || [])
  } catch {
    if (!append) blocksData.value = []
  } finally {
    loadingBlocks.value = false
  }
}

function resetBlocks() {
  blocksPage.value = 1
  loadBlocks(1, false)
}

function loadBlocksMore() {
  blocksPage.value += 1
  loadBlocks(blocksPage.value, true)
}

function openEdit(row: any) {
  editDailyRow.value = row
  editVisible.value = true
}

function onBlockSaved() {
  editedSet.value.add(editDailyRow.value?.id)
  resetBlocks()
}

async function confirmDaily(row: any) {
  try {
    await ElMessageBox.confirm(`确认提交 ${row.person_name} ${row.event_date} 的整日事件？提交后将一次性完成全部改动。`, '确认整日', { type: 'warning' })
  } catch { return }
  try {
    const d = (await getAttendanceEvents({ person_id: row.person_id, date_start: row.event_date, date_end: row.event_date, pageNum: 1, pageSize: 100 })) as any
    const latest = (d.list || []).find((x: any) => x.id === row.id) || row
    await confirmPendingDaily(row.id, latest.details || [])
    ElMessage.success('确认成功')
    editedSet.value.delete(row.id)
    resetBlocks()
  } catch { /* handled */ }
}

onMounted(async () => { personList.value = (await getAllPersons()) as any[] || [] })

function handleAction(key: string) {
  if (key==='add') { dialogMode.value='add'; editId.value=0; Object.assign(form,{person_id:null,event_date:'',event_type:'',sub_type:'',hours:8,punch_time:'',remark:''}); dialogVisible.value=true }
  else if (key==='batch') { Object.assign(batchForm,{person_ids:[],start_date:'',end_date:'',event_type:'',sub_type:'',hours:8,remark:''}); batchVisible.value=true }
  else if (key==='import') { importVisible.value=true; importStep.value=0; importFile.value=null; importPreview.value=[]; importFilePath.value=''; importMonth.value='' }
  else if (key==='trash') { trashVisible.value=true }
  else if (key==='export') { handleExport() }
}

function onImportFileChange(file: any) { importFile.value = file.raw || null }

async function doPreview() {
  if (!importFile.value) return
  previewing.value = true
  try {
    const d = await dingTalkPreview(importFile.value) as any
    importPreview.value = (d.preview || []).map((p: any) => ({ ...p, person_id: p.matched_id || null }))
    importFilePath.value = d.file_path
    importStep.value = 1
  } catch { /* handled */ } finally { previewing.value = false }
}

async function doImport() {
  if (!importMonth.value) { ElMessage.warning('请选择归属月份'); return }
  importing.value = true
  try {
    const mappings = importPreview.value.map((p: any) => ({ excel_name: p.excel_name, person_id: p.person_id }))
    const d = await dingTalkExecute(importMonth.value, importFilePath.value, mappings) as any
    ElMessage.success(`导入完成: 创建${d.created}条, 待确认${d.pending}条`)
    importVisible.value = false
    tableRef.value?.refresh()
  } catch { /* handled */ } finally { importing.value = false }
}

async function handleExport() {
  const data = await exportAttendanceEvents({})
  downloadBlob(data)
}

async function handleEdit(row: any) {
  dialogMode.value='edit'; editId.value=row.id
  const d = (await getAttendanceEvents({person_id:row.person_id})) as any
  const e = (d.list||[]).find((x:any)=>x.id===row.id) || row
  Object.assign(form, {person_id:e.person_id,event_date:e.event_date,event_type:e.event_type,sub_type:e.sub_type,hours:e.hours||8,punch_time:e.punch_time||'',remark:e.remark||''})
  dialogVisible.value=true
}

async function handleSubmit() {
  saving.value=true
  try {
    const d: any = { person_id:form.person_id, event_date:form.event_date, event_type:form.event_type, sub_type:form.sub_type }
    if (form.event_type !== '违纪') {
      d.hours = form.hours
    }
    d.remark = form.remark || ''
    if (dialogMode.value==='add') await createAttendanceEvent(d)
    else await updateAttendanceEvent(editId.value, d)
    ElMessage.success(dialogMode.value==='add'?'创建成功':'更新成功')
    dialogVisible.value=false; tableRef.value?.refresh()
  } catch { /* */ } finally { saving.value=false }
}

async function handleBatchSubmit() {
  saving.value=true
  try {
    const d = await createBatchAttendanceEvents(batchForm) as any
    ElMessage.success(`批量创建完成: 成功${d.success}条, 失败${d.fail}条`)
    batchVisible.value=false; tableRef.value?.refresh()
  } catch { /* */ } finally { saving.value=false }
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确认删除？','提示',{type:'warning'}) } catch { return }
  try { await deleteAttendanceEvent(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh() } catch { /* */ }
}
function onRefresh() { tableRef.value?.refresh() }
</script>
<style lang="scss" scoped>
.page-container { padding:0; background:transparent; }
.page-header { margin-bottom:16px; display:flex; align-items:center; h2 { font-size:18px; font-weight:600; color:#303133; } }
.import-hint { color:#909399; font-size:12px; margin-top:8px; }
.block-toolbar {
  background: #fff;
  border-radius: 4px;
  padding: 12px 16px 0;
  margin-bottom: 12px;
  .block-actions-bar {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }
}
.block-group {
  margin-bottom: 16px;
  .group-header {
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
    .group-count { color: #909399; font-size: 12px; font-weight: 400; margin-left: 8px; }
  }
  .block-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
}
.load-more {
  text-align: center;
  margin-top: 12px;
}
</style>
