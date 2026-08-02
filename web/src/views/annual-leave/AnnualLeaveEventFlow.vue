<template>
  <div class="page-container">
    <PageHeader title="年假事件流水">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>
    <PageToolbar>
      <el-button type="primary" size="small" @click="handleAction('add')">新增</el-button>
      <el-button size="small" @click="handleAction('trash')">回收站</el-button>
      <el-button size="small" @click="handleAction('export')">导出</el-button>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button v-if="row.source_type === 'manual'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button v-if="row.source_type === 'manual'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>
    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getAnnualLeaveEvents(p)"
        date-field="effective_date"
      >
        <template #day="{ date, items }">
          <DayCard :date="date" :events="items" @event-click="handleEventClick" />
        </template>
      </TimeCardPanel>
    </template>

    <el-dialog v-model="dialogVisible" :title="mode==='add'?'新增':'编辑'" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" placeholder="选择" /></el-form-item>
        <el-form-item label="类型" required><el-select v-model="form.event_type" style="width:100%"><el-option v-for="t in types" :key="t" :label="t" :value="t"/></el-select></el-form-item>
        <el-form-item label="时长(小时)" required><el-input-number v-model="form.hours" :precision="1" style="width:100%" /></el-form-item>
        <el-form-item label="生效日期" required><el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="s" @click="submit">确定</el-button></template>
    </el-dialog>
    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR" />

    <DailyEditDialog v-model:visible="editVisible" :daily="editDailyRow" @saved="onEdited" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'annual_leave_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import NameSelect from '@/components/NameSelect.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import DayCard from '@/components/cards/DayCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import DailyEditDialog from '@/components/attendance/DailyEditDialog.vue'
import { getAnnualLeaveEvents, createAnnualLeaveEvent, updateAnnualLeaveEvent, deleteAnnualLeaveEvent, restoreAnnualLeaveEvent, getDeletedAnnualLeaveEvents, exportAnnualLeaveEvents } from '@/api/annual-leave'
import { getAttendanceEvents } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const viewMode = ref<'cards'|'list'>('cards')
const timePanelRef = ref(); const dialogVisible = ref(false); const mode = ref<'add'|'edit'>('add'); const eid = ref(0); const s = ref(false); const tv = ref(false)
const editVisible = ref(false)
const editDailyRow = ref<any>(null)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
const types = ref(['grant', 'carryover_deduct', 'adjust'])
const form = reactive({ person_id: null as any, event_type: 'grant', hours: 8, effective_date: '', remark: '' })

const columns = [
  { prop:'id', label:'ID', width:'60' },{ prop:'person_name', label:'人员', width:'80' },
  { prop:'event_type', label:'类型', width:'130' },{ prop:'hours', label:'变动时长(天)', width:'100', formatter:(r:any)=>hoursToDays(r.hours).toFixed(2) },
  { prop:'source_type', label:'来源', width:'110', formatter:(r:any)=>({manual:'人工录入',system_period:'系统周期',attendance:'考勤休假'}[r.source_type]||r.source_type) },
  { prop:'effective_date', label:'生效日期', width:'110' },
  { prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi:fetchPersonOpts },
  { prop:'date', label:'时间范围', type:'date-range' as const },
]
const tc = [{ prop:'id', label:'ID', width:'60' },{ prop:'event_type', label:'类型' }]

async function fetchPersonOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchEvents(p:any){ return (await getAnnualLeaveEvents(p)) as any }
async function fd(p:any){ return (await getDeletedAnnualLeaveEvents(p)) as any }
async function rst(id:number){ return restoreAnnualLeaveEvent(id) }

function handleAction(k:string){
  if(k==='add'){ mode.value='add';eid.value=0;Object.assign(form,{person_id:null,event_type:'grant',hours:8,effective_date:'',remark:''});dialogVisible.value=true }
  else if(k==='trash') tv.value=true
  else if(k==='export') handleExport()
}

async function handleExport() {
  const data = await exportAnnualLeaveEvents({})
  downloadBlob(data)
}
async function handleEdit(r:any){ mode.value='edit';eid.value=r.id;Object.assign(form,{person_id:r.person_id,event_type:r.event_type,hours:r.hours,effective_date:r.effective_date,remark:r.remark||''});dialogVisible.value=true }

// 统一编辑：考勤来源（休假-年假）→ 打开该日考勤整日编辑；manual → 原年假单事件编辑
async function handleEventClick(r: any) {
  if (r.source_type === 'attendance') {
    await editAttendanceDaily(r)
  } else {
    handleEdit(r)
  }
}

async function editAttendanceDaily(item: any) {
  try {
    let row: any = null
    if (item.daily_id) {
      const d = (await getAttendanceEvents({ pageNum: 1, pageSize: 100 })) as any
      row = (d.list || []).find((x: any) => x.id === item.daily_id) || null
    }
    if (!row && item.person_id && item.effective_date) {
      const d = (await getAttendanceEvents({ person_id: item.person_id, date_start: item.effective_date, date_end: item.effective_date, pageNum: 1, pageSize: 100 })) as any
      row = (d.list || [])[0] || null
    }
    if (!row) { ElMessage.warning('未找到当日考勤记录'); return }
    editDailyRow.value = row
    editVisible.value = true
  } catch { /* handled */ }
}

function onEdited() {
  timePanelRef.value?.reload()
  tableRef.value?.refresh()
}
async function submit(){
  s.value=true
  try { const d:any={person_id:form.person_id,event_type:form.event_type,hours:form.hours,effective_date:form.effective_date,remark:form.remark}
    if(mode.value==='add') await createAnnualLeaveEvent(d); else await updateAnnualLeaveEvent(eid.value,d)
    ElMessage.success('成功');dialogVisible.value=false;tableRef.value?.refresh();timePanelRef.value?.reload()
  } catch { /* */ } finally { s.value=false }
}
async function handleDelete(r:any){ try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteAnnualLeaveEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh();timePanelRef.value?.reload()}catch{ /* */ } }
function onR(){ tableRef.value?.refresh() }
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
