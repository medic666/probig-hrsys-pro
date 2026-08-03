<template>
  <div class="page-container">
    <PageHeader title="工资事件管理">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button type="primary" size="small" @click="handleAction('add')">新增</el-button>
      <template #right>
        <el-button size="small" @click="handleAction('trash')">回收站</el-button>
        <el-button size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
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
        :url-driven="true"
        :fetch-fn="(p: any) => getSalaryEvents(p)"
        month-field="belong_month"
        :has-day-level="false"
      >
        <template #period-list="{ items }">
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

    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR"/>

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'salary_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import { getSalaryEvents, deleteSalaryEvent, restoreSalaryEvent, getDeletedSalaryEvents, exportSalaryEvents } from '@/api/salary'
import { getAllPersons } from '@/api/person'
import { downloadBlob } from '@/utils/download'

import { usePageView } from '@/composables/usePageView'

const router = useRouter()
const tableRef=ref(), tv=ref(false), attachVisible=ref(false), attachFileId=ref<number|null>(null)
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'event_type',label:'类型',width:'100'},  {prop:'amount',label:'值',width:'100'},
  {prop:'remark',label:'备注'},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const, fetchApi:fetchPersonOpts},
  {prop:'event_type',label:'类型',type:'select' as const,options:['绩效系数','提成','奖惩','借款还款','个税扣除'].map(t=>({label:t,value:t}))},
]
const tc=[{prop:'event_type',label:'类型'},{prop:'belong_month',label:'月份'}]

async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchEvents(p:any){return (await getSalaryEvents(p)) as any}
async function fd(p:any){return (await getDeletedSalaryEvents(p)) as any}
async function rst(id:number){return restoreSalaryEvent(id)}

function handleAction(k:string){
  if(k==='add'){ router.push('/salary-events/create') }
  else if(k==='trash') tv.value=true
  else if(k==='export') handleExport()
}

async function handleExport() {
  const data = await exportSalaryEvents(tableRef.value?.getSearchParams() || {})
  downloadBlob(data)
}
function handleEdit(r:any){ router.push(`/salary-events/${r.id}`) }
async function handleDelete(r:any){try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteSalaryEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh();timePanelRef.value?.reload()}catch{ /* */ }}
function onR(){tableRef.value?.refresh()}
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
