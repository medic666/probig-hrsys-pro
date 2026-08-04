<template>
  <div class="page-container">
    <PageHeader title="月度工资汇总">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button type="primary" size="small" @click="handleAction('calc')">批量核算</el-button>
      <template #right>
        <el-button size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchSummaries" :search-fields="searchFields">
        <template #status="{ row }">
          <StatusTag :status="row.status || 'not_calculated'" />
        </template>
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="openDetail(row, 'summary')">明细</el-button>
          <el-button type="success" link size="small" @click="openDetail(row, 'versions')">版本</el-button>
          <el-button type="warning" link size="small" @click="openDetail(row, 'trace')">追溯</el-button>
        </template>
      </ProTable>
    </template>
    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :url-driven="true"
        :fetch-fn="(p: any) => getSalarySummaries(p)"
        month-field="belong_month"
        status-field="status"
        :pending-values="['data_changed']"
        :has-day-level="false"
        :detail-route="summaryDetailRoute"
      />
    </template>

    <BatchActionDrawer v-model:visible="calcVisible" title="批量核算" :submit-fn="submitFn" @done="onCalcDone" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import BatchActionDrawer from '@/components/BatchActionDrawer.vue'
import { getSalarySummaries, calculateSalaries, exportSalarySummaries } from '@/api/salary'
import { getAllPersons } from '@/api/person'
import { formatDateTime } from '@/utils'
import { downloadBlob } from '@/utils/download'

import { usePageView } from '@/composables/usePageView'

const router = useRouter()
const tableRef=ref()
const calcVisible=ref(false)
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'attendance_salary',label:'出勤工资',width:'100'},{prop:'performance_salary',label:'绩效工资',width:'100'},
  {prop:'attendance_bonus',label:'全勤奖',width:'80'},{prop:'final_salary',label:'实发工资',width:'110'},
  {prop:'status',label:'状态',width:'110',slot:'status'},
  {prop:'last_calc_at',label:'核算时间',width:'160',formatter:(r:any)=>formatDateTime(r.last_calc_at)},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const,fetchApi:fetchPersonOpts},
  {prop:'month',label:'月份',type:'month' as const},
]

async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchSummaries(p:any){
  return (await getSalarySummaries(p)) as any
}

// 月份点击 → 工资汇总聚合详情页（明细/版本/对比/追溯，URL 携带人员+月份）
function summaryDetailRoute(person: { id: number; name: string }, month: string) {
  return `/salary-summaries/${person.id}/${month}?name=${encodeURIComponent(person.name)}`
}

// 列表查看（明细/版本/追溯）→ 聚合详情页对应 tab
function openDetail(row: any, tab: string) {
  router.push(`/salary-summaries/${row.person_id}/${row.belong_month}?tab=${tab}`)
}

function handleAction(k:string){
  if(k==='calc'){ calcVisible.value=true }
  else if(k==='export'){handleExport()}
}
async function handleExport(){
  const data = await exportSalarySummaries(tableRef.value?.getSearchParams() || {})
  downloadBlob(data)
}

const submitFn = (data: any) => calculateSalaries(data)

function onCalcDone(){
  tableRef.value?.refresh()
  timePanelRef.value?.reload()
}
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
