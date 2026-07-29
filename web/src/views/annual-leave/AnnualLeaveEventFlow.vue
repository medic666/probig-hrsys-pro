<template>
  <div class="page-container"><div class="page-header"><h2>年假事件流水</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button v-if="row.source_type !== 'system_period'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-if="row.source_type !== 'system_period'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" :title="mode==='add'?'新增':'编辑'" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" placeholder="选择" /></el-form-item>
        <el-form-item label="类型" required><el-select v-model="form.event_type" style="width:100%"><el-option v-for="t in types" :key="t" :label="t" :value="t"/></el-select></el-form-item>
        <el-form-item label="时长" required><el-input-number v-model="form.hours" :precision="1" style="width:100%" /></el-form-item>
        <el-form-item label="生效日期" required><el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="s" @click="submit">确定</el-button></template>
    </el-dialog>
    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'annual_leave_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import NameSelect from '@/components/NameSelect.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getAnnualLeaveEvents, createAnnualLeaveEvent, updateAnnualLeaveEvent, deleteAnnualLeaveEvent, restoreAnnualLeaveEvent, getDeletedAnnualLeaveEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'

const tableRef = ref(); const dialogVisible = ref(false); const mode = ref<'add'|'edit'>('add'); const eid = ref(0); const s = ref(false); const tv = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
const types = ref(['grant', 'carryover_deduct', 'adjust'])
const form = reactive({ person_id: null as any, event_type: 'grant', hours: 8, effective_date: '', remark: '' })

const columns = [
  { prop:'id', label:'ID', width:'60' },{ prop:'person_name', label:'人员', width:'80' },
  { prop:'event_type', label:'类型', width:'130' },{ prop:'hours', label:'变动时长', width:'90' },
  { prop:'source_type', label:'来源', width:'90' },{ prop:'effective_date', label:'生效日期', width:'110' },
  { prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi:fetchPersonOpts },
]
const actions = [{ key:'add', label:'新增', type:'primary' as const },{ key:'trash', label:'回收站', type:'default' as const }]
const tc = [{ prop:'id', label:'ID', width:'60' },{ prop:'event_type', label:'类型' }]

async function fetchPersonOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchEvents(p:any){ return (await getAnnualLeaveEvents(p)) as any }
async function fd(p:any){ return (await getDeletedAnnualLeaveEvents(p)) as any }
async function rst(id:number){ return restoreAnnualLeaveEvent(id) }

function handleAction(k:string){
  if(k==='add'){ mode.value='add';eid.value=0;Object.assign(form,{person_id:null,event_type:'grant',hours:8,effective_date:'',remark:''});dialogVisible.value=true }
  else if(k==='trash') tv.value=true
}
async function handleEdit(r:any){ mode.value='edit';eid.value=r.id;Object.assign(form,{person_id:r.person_id,event_type:r.event_type,hours:r.hours,effective_date:r.effective_date,remark:r.remark||''});dialogVisible.value=true }
async function submit(){
  s.value=true
  try { const d:any={person_id:form.person_id,event_type:form.event_type,hours:form.hours,effective_date:form.effective_date,remark:form.remark}
    if(mode.value==='add') await createAnnualLeaveEvent(d); else await updateAnnualLeaveEvent(eid.value,d)
    ElMessage.success('成功');dialogVisible.value=false;tableRef.value?.refresh()
  } catch { /* */ } finally { s.value=false }
}
async function handleDelete(r:any){ try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteAnnualLeaveEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh()}catch{ /* */ } }
function onR(){ tableRef.value?.refresh() }
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
