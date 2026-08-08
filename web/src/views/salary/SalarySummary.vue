<template>
  <div class="page-container">
    <PageHeader title="月度工资汇总">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button v-permission="PERM.salarySummaryCalculate" type="primary" size="small" @click="handleAction('calc')">批量核算</el-button>
      <template #right>
        <el-button v-permission="PERM.salaryEventExport" size="small" @click="handleAction('export')">导出</el-button>
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
        :person-dot-map="dotMap"
        :detail-route="summaryDetailRoute"
      />
    </template>

    <BatchActionDrawer v-model:visible="calcVisible" title="批量核算" :submit-fn="submitFn" @done="onCalcDone" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import BatchActionDrawer from '@/components/BatchActionDrawer.vue'
import { getSalarySummaries, getSalarySummariesBadges, calculateSalaries, exportSalarySummaries } from '@/api/salary'
import { formatDateTime } from '@/utils'

import { usePageView } from '@/composables/usePageView'
import { useBadges } from '@/composables/useBadges'
import { useExport } from '@/composables/useExport'
import { SALARY_SUMMARY_FIELDS, fieldsToColumns } from '@/constants/fields'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref()
const calcVisible=ref(false)
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()
// 徽章映射：personId → 颜色点（上月无核算 gray / 汇总过期 orange / 正常 green）
const { dotMap, loadDots } = useBadges()

// 列表字段 = 统一字段表（与明细/版本/对比/追溯/导出同源）+ 状态/核算时间
const columns = fieldsToColumns(SALARY_SUMMARY_FIELDS, [
  { prop: 'status', label: '状态', width: 110, slot: 'status' },
  { prop: 'last_calc_at', label: '核算时间', width: 160, formatter: (r: any) => formatDateTime(r.last_calc_at) },
])
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const},
  {prop:'months',label:'月份',type:'months' as const},
  {prop:'status',label:'状态',type:'select' as const,options:[
    {label:'已核算',value:'calculated'},
    {label:'数据已变动',value:'data_changed'},
  ]},
]

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
const { run: handleExport } = useExport(exportSalarySummaries, () => tableRef.value?.getSearchParams() || {})

const submitFn = (data: any) => calculateSalaries(data)

function onCalcDone(){
  tableRef.value?.refresh()
  timePanelRef.value?.reload()
}

onMounted(async () => {
  await loadDots('salary-summaries-badges', getSalarySummariesBadges)
})
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
