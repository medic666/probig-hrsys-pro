<template>
  <div class="page-container">
    <PageHeader title="月度考勤核算">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button v-permission="PERM.attendanceMonthlyCalculate" type="primary" size="small" @click="handleAction('calc')">批量核算</el-button>
      <template #right>
        <el-button v-permission="PERM.attendanceEventExport" size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchMonthly" :search-fields="searchFields">
        <template #status="{ row }">
          <StatusTag :status="row.status || 'not_calculated'" />
        </template>
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="showDetail(row)">查看明细</el-button>
        </template>
      </ProTable>
    </template>
    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getMonthlyList(p)"
        month-field="belong_month"
        status-field="status"
        :pending-values="['data_changed']"
        :has-day-level="false"
        :url-driven="true"
        :person-dot-map="dotMap"
        :detail-route="monthDetailRoute"
      />
    </template>

    <BatchActionDrawer v-model:visible="calcVisible" title="批量核算" :submit-fn="submitFn" @done="onCalcDone" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import BatchActionDrawer from '@/components/BatchActionDrawer.vue'
import { getMonthlyList, getAttendanceMonthlyBadges, calculateMonthly, exportAttendanceMonthly } from '@/api/attendance'
import { formatDateTime } from '@/utils'
import { downloadBlob } from '@/utils/download'

import { usePageView } from '@/composables/usePageView'
import { useBadges } from '@/composables/useBadges'
import { ATTENDANCE_CALC_FIELDS, fieldsToColumns } from '@/constants/fields'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref()
const calcVisible=ref(false)
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()
// 徽章映射：personId → 颜色点（上月核算为空 gray / 核算过期 orange / 已核算 green）
const { dotMap, loadDots } = useBadges()

// 列表字段 = 统一字段表（与详情/追溯/导出同源）+ 状态/核算时间
const columns = fieldsToColumns(ATTENDANCE_CALC_FIELDS, [
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

async function fetchMonthly(p:any){
  return (await getMonthlyList(p)) as any
}

onMounted(async () => {
  await loadDots('attendance-monthly-badges', getAttendanceMonthlyBadges)
})

// 月份点击 → 业务逻辑页（URL 携带人员+月份，返回后可恢复层级）
function monthDetailRoute(person: { id: number; name: string }, month: string) {
  return `/attendance-monthly/${person.id}/${month}?name=${encodeURIComponent(person.name)}`
}

function handleAction(k:string){
  if(k==='calc'){ calcVisible.value=true }
  else if(k==='export'){handleExport()}
}

const submitFn = (data: any) => calculateMonthly(data)

function onCalcDone(){
  tableRef.value?.refresh()
  timePanelRef.value?.reload()
}
async function handleExport(){
  const params = tableRef.value?.getSearchParams() || {}
  const d = (await getMonthlyList(params)) as any
  const changedCount = (d.list || []).filter((r:any)=>r.status==='data_changed').length
  if (changedCount > 0) {
    try {
      await ElMessageBox.confirm(`当前筛选结果中有 ${changedCount} 条「数据已变动」记录，导出结果可能不准确，确认导出？`, '提示', { type: 'warning' })
    } catch { return }
  }
  const data = await exportAttendanceMonthly(params)
  downloadBlob(data)
}
// 查看明细 = 进入月度考勤核算详情页（URL 携带人员+月份）
function showDetail(row: any) {
  router.push(`/attendance-monthly/${row.person_id}/${row.belong_month}`)
}
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
