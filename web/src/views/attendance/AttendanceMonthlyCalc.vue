<template>
  <div class="page-container">
    <PageHeader title="月度考勤核算">
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
        :detail-route="monthDetailRoute"
      />
    </template>

    <el-dialog v-model="calcVisible" title="批量核算" width="450px">
      <el-form label-width="80px">
        <el-form-item label="月份"><el-date-picker v-model="calcMonth" type="month" value-format="YYYY-MM" style="width:100%" /></el-form-item>
        <el-form-item label="人员">
          <el-select v-model="calcPersonIds" multiple placeholder="不选则核算全部" style="width:100%">
            <el-option v-for="p in personList" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="calcVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doCalc">开始核算(不选则全部在职人员)</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import { getMonthlyList, calculateMonthly, exportAttendanceMonthly } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { formatDateTime } from '@/utils'
import { downloadBlob } from '@/utils/download'

import { usePageView } from '@/composables/usePageView'

const router = useRouter()
const tableRef=ref(), calcVisible=ref(false), saving=ref(false), calcMonth=ref(''), calcPersonIds=ref<number[]>([])
const personList=ref<{id:number;name:string}[]>([])
const { viewMode, isList } = usePageView('cards')
const timePanelRef=ref()

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'attendance_salary',label:'出勤工资',width:'100'},{prop:'overtime_workday_salary',label:'工作日加班工资',width:'120'},
  {prop:'overtime_holiday_salary',label:'节假日加班工资',width:'120'},  {prop:'attendance_bonus',label:'全勤奖',width:'80'},
  {prop:'status',label:'状态',width:'110',slot:'status'},
  {prop:'last_calc_at',label:'核算时间',width:'160',formatter:(r:any)=>formatDateTime(r.last_calc_at)},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const,fetchApi:fetchPersonOpts},
  {prop:'month',label:'月份',type:'month' as const},
]

onMounted(async()=>{personList.value=(await getAllPersons()) as any[]||[]})
async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchMonthly(p:any){
  return (await getMonthlyList(p)) as any
}

// 月份点击 → 业务逻辑页（URL 携带人员+月份，返回后可恢复层级）
function monthDetailRoute(person: { id: number; name: string }, month: string) {
  return `/attendance-monthly/${person.id}/${month}?name=${encodeURIComponent(person.name)}`
}

function handleAction(k:string){
  if(k==='calc'){calcMonth.value='';calcPersonIds.value=[];calcVisible.value=true}
  else if(k==='export'){handleExport()}
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
async function doCalc(){
  if(!calcMonth.value){ElMessage.warning('请选择月份');return}
  saving.value=true
  try{const d=await calculateMonthly({month:calcMonth.value,person_ids:calcPersonIds.value}) as any
    ElMessage.success(`核算完成: 成功${d.success}条, 失败${d.fail}条`);calcVisible.value=false;tableRef.value?.refresh();timePanelRef.value?.reload()
  }catch{/* */}finally{saving.value=false}
}
// 查看明细 = 进入月度考勤核算详情页（URL 携带人员+月份）
function showDetail(row: any) {
  router.push(`/attendance-monthly/${row.person_id}/${row.belong_month}`)
}
</script>
<style scoped>.page-container{padding:0;background:transparent}</style>
