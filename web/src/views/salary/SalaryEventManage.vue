<template>
  <div class="page-container">
    <div class="page-header">
      <h2>工资事件管理</h2>
      <el-radio-group v-model="viewMode" size="small" style="margin-left:16px">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
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
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getSalaryEvents(p)"
        month-field="belong_month"
        :has-day-level="false"
      >
        <template #month-list="{ items }">
          <el-table :data="items" border size="small" style="max-width:720px">
            <el-table-column prop="belong_month" label="月份" width="90" />
            <el-table-column prop="event_type" label="类型" width="100" />
            <el-table-column prop="amount" label="值" width="110" />
            <el-table-column prop="remark" label="备注" />
            <el-table-column label="操作" width="90">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </TimeCardPanel>
    </template>

    <el-dialog v-model="dialogVisible" :title="mode==='add'?'新增':'编辑'" width="420px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts"/></el-form-item>
        <el-form-item label="归属月份" required><el-date-picker v-model="form.belong_month" type="month" value-format="YYYY-MM" style="width:100%"/></el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type" style="width:100%"><el-option v-for="t in types" :key="t" :label="t" :value="t"/></el-select>
        </el-form-item>
        <el-form-item label="值" required><el-input-number v-model="form.amount" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="s" @click="submit">确定</el-button></template>
    </el-dialog>
    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR"/>

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'salary_event'" :target-id="attachFileId" />
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
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import { getSalaryEvents, createSalaryEvent, updateSalaryEvent, deleteSalaryEvent, restoreSalaryEvent, getDeletedSalaryEvents, exportSalaryEvents } from '@/api/salary'
import { getAllPersons } from '@/api/person'
import { downloadBlob } from '@/utils/download'

const tableRef=ref(), dialogVisible=ref(false), mode=ref<'add'|'edit'>('add'), eid=ref(0), s=ref(false), tv=ref(false), attachVisible=ref(false), attachFileId=ref<number|null>(null)
const viewMode=ref<'cards'|'list'>('cards')
const timePanelRef=ref()
const types=['绩效系数','提成','奖惩','借款还款','个税扣除']
const form=reactive({person_id:null as any,belong_month:'',event_type:'绩效系数',amount:1,remark:''})

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'event_type',label:'类型',width:'100'},  {prop:'amount',label:'值',width:'100'},
  {prop:'remark',label:'备注'},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const, fetchApi:fetchPersonOpts},
  {prop:'event_type',label:'类型',type:'select' as const,options:types.map(t=>({label:t,value:t}))},
]
const actions=[{key:'add',label:'新增',type:'primary' as const},{key:'trash',label:'回收站',type:'default' as const},{key:'export',label:'导出',type:'default' as const}]
const tc=[{prop:'id',label:'ID',width:'60'},{prop:'event_type',label:'类型'},{prop:'belong_month',label:'月份'}]

async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchEvents(p:any){return (await getSalaryEvents(p)) as any}
async function fd(p:any){return (await getDeletedSalaryEvents(p)) as any}
async function rst(id:number){return restoreSalaryEvent(id)}

function handleAction(k:string){
  if(k==='add'){mode.value='add';eid.value=0;Object.assign(form,{person_id:null,belong_month:'',event_type:'绩效系数',amount:1,remark:''});dialogVisible.value=true}
  else if(k==='trash') tv.value=true
  else if(k==='export') handleExport()
}

async function handleExport() {
  const data = await exportSalaryEvents({})
  downloadBlob(data)
}
async function handleEdit(r:any){mode.value='edit';eid.value=r.id;Object.assign(form,{person_id:r.person_id,belong_month:r.belong_month,event_type:r.event_type,amount:r.amount,remark:r.remark||''});dialogVisible.value=true}
async function submit(){
  s.value=true
  try{const d:any={person_id:form.person_id,belong_month:form.belong_month,event_type:form.event_type,amount:form.amount,remark:form.remark}
    if(mode.value==='add') await createSalaryEvent(d); else await updateSalaryEvent(eid.value,d)
    ElMessage.success('成功');dialogVisible.value=false;tableRef.value?.refresh();timePanelRef.value?.reload()
  }catch{/* */}finally{s.value=false}
}
async function handleDelete(r:any){try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteSalaryEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh();timePanelRef.value?.reload()}catch{ /* */ }}
function onR(){tableRef.value?.refresh()}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
