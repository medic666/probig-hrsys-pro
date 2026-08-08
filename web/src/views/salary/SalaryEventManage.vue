<template>
  <div class="page-container">
    <PageHeader title="工资事件管理">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button v-permission="PERM.salaryEventWrite" type="primary" size="small" @click="handleAction('add')">新增</el-button>
      <template #right>
        <el-button v-permission="PERM.salaryEventWrite" size="small" @click="handleAction('trash')">回收站</el-button>
        <el-button v-permission="PERM.salaryEventExport" size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button v-permission="PERM.salaryEventWrite" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button v-permission="PERM.salaryEventWrite" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button v-permission="PERM.fileRead" type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
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
        badge-position="meta"
      >
        <template #person-badge="{ person }">
          <div class="advance-line" :class="{ 'is-zero': !(advanceMap[person.id] ?? 0) }">
            预支待还 {{ formatMoney(advanceMap[person.id] ?? 0) }} 元
          </div>
        </template>
        <template #period-list="{ items }">
          <div class="ev-grid">
            <SalaryEventCard v-for="e in items" :key="e.id" :event="e" @edit="handleEdit" />
          </div>
        </template>
      </TimeCardPanel>
    </template>

    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR"/>

    <FileAttachDialog v-model:visible="attachVisible" target-type="salary_event" :target-id="attachFileId" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachDialog from '@/components/FileAttachDialog.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import SalaryEventCard from '@/components/salary/SalaryEventCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import { getSalaryEvents, getSalaryAdvanceBalances, deleteSalaryEvent, restoreSalaryEvent, getDeletedSalaryEvents, exportSalaryEvents } from '@/api/salary'
import { formatMoney } from '@/utils'

import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'
import { useBadges, fetchBadges } from '@/composables/useBadges'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref(), tv=ref(false), attachVisible=ref(false), attachFileId=ref<number|null>(null)
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()
// 工资预支余额映射（meta 位徽章）
const { balanceMap: advanceMap } = useBadges()

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'event_type',label:'类型',width:'100'},  {prop:'amount',label:'值',width:'100'},
  {prop:'remark',label:'备注'},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const},
  {prop:'months',label:'月份',type:'months' as const},
  {prop:'event_type',label:'类型',type:'select' as const,options:['绩效系数','提成','奖惩','工资预支','预支还款','个税扣除'].map(t=>({label:t,value:t}))},
]
const tc=[{prop:'event_type',label:'类型'},{prop:'belong_month',label:'月份'}]

async function fetchEvents(p:any){return (await getSalaryEvents(p)) as any}
const { run: handleExport } = useExport(exportSalaryEvents, () => tableRef.value?.getSearchParams() || {})
async function fd(p:any){return (await getDeletedSalaryEvents(p)) as any}
async function rst(id:number){return restoreSalaryEvent(id)}

onMounted(async () => {
  try {
    const list = (await fetchBadges('salary-advance-balances', getSalaryAdvanceBalances)) || []
    advanceMap.value = Object.fromEntries(list.map((b: any) => [b.person_id, b.balance ?? 0]))
  } catch {
    advanceMap.value = {}
  }
})

function handleAction(k:string){
  if(k==='add'){ router.push('/salary-events/create') }
  else if(k==='trash') tv.value=true
  else if(k==='export') handleExport()
}


function handleEdit(r:any){ router.push(`/salary-events/${r.id}`) }
async function handleDelete(r:any){try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteSalaryEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh();timePanelRef.value?.reload()}catch{ /* */ }}
function onR(){tableRef.value?.refresh()}
</script>
<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.page-container{padding:0;background:transparent}

.ev-grid {
  @include card-grid(260px);
}

.advance-line { font-size: 13px; color: #e6a23c; font-weight: 600; line-height: 20px; }
.advance-line.is-zero { color: #c0c4cc; font-weight: 400; }
</style>
