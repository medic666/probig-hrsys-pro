<template>
  <div class="page-container">
    <div class="page-header"><h2>考勤事件管理</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" :title="dialogMode==='add'?'新增考勤事件':'编辑考勤事件'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" placeholder="选择人员" /></el-form-item>
        <el-form-item label="日期" required><el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type" placeholder="选择类型" style="width:100%" @change="onTypeChange">
            <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.event_type !== '打卡时间戳'" label="子类型" required>
          <el-select v-model="form.sub_type" placeholder="选择子类型" style="width:100%">
            <el-option v-for="s in currentSubTypes" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.event_type !== '打卡时间戳' && form.event_type !== '违纪'" label="时长(小时)" required>
          <el-input-number v-model="form.hours" :min="0" :precision="1" style="width:100%" />
        </el-form-item>
        <el-form-item v-if="form.event_type === '打卡时间戳'" label="打卡时间" required>
          <el-input v-model="form.punch_time" placeholder="如 08:30" />
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
        <el-form-item v-if="batchForm.event_type !== '打卡时间戳'" label="子类型">
          <el-select v-model="batchForm.sub_type" style="width:100%"><el-option v-for="s in batchSubTypes" :key="s" :label="s" :value="s" /></el-select>
        </el-form-item>
        <el-form-item v-if="batchForm.event_type !== '打卡时间戳' && batchForm.event_type !== '违纪'" label="每日时长" required>
          <el-input-number v-model="batchForm.hours" :min="0" :precision="1" style="width:100%" />
        </el-form-item>
        <el-form-item v-if="batchForm.event_type === '打卡时间戳'" label="打卡时间" required>
          <el-input v-model="batchForm.punch_time" placeholder="如 08:30" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="batchForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleBatchSubmit">确定</el-button>
      </template>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restore" :columns="trashCols" @restored="onRefresh" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import NameSelect from '@/components/NameSelect.vue'
import { getAttendanceEvents, createAttendanceEvent, updateAttendanceEvent, deleteAttendanceEvent, restoreAttendanceEvent, getDeletedAttendanceEvents, createBatchAttendanceEvents } from '@/api/attendance'
import { getAllPersons } from '@/api/person'

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add'|'edit'>('add')
const editId = ref(0)
const saving = ref(false)
const trashVisible = ref(false)
const batchVisible = ref(false)
const personList = ref<{id:number;name:string}[]>([])

const eventTypes = ['出勤','休假','加班','违纪','打卡时间戳']
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

function onTypeChange() { form.sub_type = '' }
function onBatchTypeChange() { batchForm.sub_type = '' }

const columns = [
  { prop:'id', label:'ID', width:'60' },
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'event_date', label:'日期', width:'110' },
  { prop:'event_type', label:'事件类型', width:'80' },
  { prop:'sub_type', label:'子类型', width:'100' },
  { prop:'hours', label:'时长(小时)', width:'90' },
  { prop:'punch_time', label:'打卡时间', width:'90' },
  { prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOpts },
  { prop:'event_type', label:'事件类型', type:'select' as const, options: eventTypes.map(t=>({label:t,value:t})) },
]
const actions = [
  { key:'add', label:'新增事件', type:'primary' as const },
  { key:'batch', label:'批量新增', type:'success' as const },
  { key:'trash', label:'回收站', type:'default' as const },
]
const trashCols = [{ prop:'id', label:'ID', width:'60' },{ prop:'event_type', label:'类型' },{ prop:'event_date', label:'日期' }]

async function fetchPersonOpts(k?: string) { const l = await getAllPersons() as any[]; return k ? l.filter(p=>p.name.includes(k)) : l }
async function fetchEvents(p: any) { return (await getAttendanceEvents(p)) as any }
async function fetchDeleted(p: any) { return (await getDeletedAttendanceEvents(p)) as any }
async function restore(id: number) { return restoreAttendanceEvent(id) }

onMounted(async () => { personList.value = (await getAllPersons()) as any[] || [] })

function handleAction(key: string) {
  if (key==='add') { dialogMode.value='add'; editId.value=0; Object.assign(form,{person_id:null,event_date:'',event_type:'',sub_type:'',hours:8,punch_time:'',remark:''}); dialogVisible.value=true }
  else if (key==='batch') { Object.assign(batchForm,{person_ids:[],start_date:'',end_date:'',event_type:'',sub_type:'',hours:8,remark:''}); batchVisible.value=true }
  else if (key==='trash') { trashVisible.value=true }
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
    const d: any = { person_id:form.person_id, event_date:form.event_date, event_type:form.event_type, sub_type:form.sub_type, hours:form.hours, punch_time:form.punch_time, remark:form.remark }
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
.page-header { margin-bottom:16px; h2 { font-size:18px; font-weight:600; color:#303133; } }
</style>
