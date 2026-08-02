<template>
  <div class="page-container">
    <div class="page-header">
      <h2>月度考勤核算</h2>
      <el-radio-group v-model="viewMode" size="small" style="margin-left:16px">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
    </div>
    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchMonthly" :search-fields="searchFields" :actions="actions" @action="handleAction">
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
      >
        <template #month-list="{ items }">
          <el-descriptions v-if="items.length > 0" :column="3" border size="small" style="max-width:860px">
            <el-descriptions-item label="月份">{{ items[0].belong_month }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <StatusTag :status="items[0].status || 'not_calculated'" />
            </el-descriptions-item>
            <el-descriptions-item label="计薪天数">{{ items[0].salary_days }}</el-descriptions-item>
            <el-descriptions-item label="加权基本工资">{{ items[0].weighted_base_salary }}</el-descriptions-item>
            <el-descriptions-item label="加权餐补">{{ items[0].weighted_meal_allowance }}</el-descriptions-item>
            <el-descriptions-item label="记出勤(天)">{{ hoursToDays(items[0].total_work_hours).toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="工作日加班(天)">{{ hoursToDays(items[0].total_overtime_workday_hours).toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="节假日加班(天)">{{ hoursToDays(items[0].total_overtime_holiday_hours).toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="出勤工资">{{ items[0].attendance_salary }}</el-descriptions-item>
            <el-descriptions-item label="工作日加班工资">{{ items[0].overtime_workday_salary }}</el-descriptions-item>
            <el-descriptions-item label="节假日加班工资">{{ items[0].overtime_holiday_salary }}</el-descriptions-item>
            <el-descriptions-item label="全勤奖">{{ items[0].attendance_bonus }}</el-descriptions-item>
            <el-descriptions-item label="违纪次数">{{ items[0].total_violation_count }}</el-descriptions-item>
            <el-descriptions-item label="有事假">{{ items[0].has_personal_leave_month ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="核算时间">{{ formatDateTime(items[0].last_calc_at) }}</el-descriptions-item>
          </el-descriptions>
        </template>
      </TimeCardPanel>
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

    <el-dialog v-model="detailVisible" title="核算明细" width="700px">
      <el-descriptions v-if="detailRow" :column="2" border>
        <el-descriptions-item label="月份">{{ detailRow.belong_month }}</el-descriptions-item>
        <el-descriptions-item label="计薪天数">{{ detailRow.salary_days }}</el-descriptions-item>
        <el-descriptions-item label="加权基本工资">{{ detailRow.weighted_base_salary }}</el-descriptions-item>
        <el-descriptions-item label="加权餐补">{{ detailRow.weighted_meal_allowance }}</el-descriptions-item>
        <el-descriptions-item label="记出勤(天)">{{ hoursToDays(detailRow.total_work_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班(天)">{{ hoursToDays(detailRow.total_overtime_workday_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班(天)">{{ hoursToDays(detailRow.total_overtime_holiday_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="出勤工资">{{ detailRow.attendance_salary }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班工资">{{ detailRow.overtime_workday_salary }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班工资">{{ detailRow.overtime_holiday_salary }}</el-descriptions-item>
        <el-descriptions-item label="全勤奖">{{ detailRow.attendance_bonus }}</el-descriptions-item>
        <el-descriptions-item label="违纪次数">{{ detailRow.total_violation_count }}</el-descriptions-item>
        <el-descriptions-item label="有事假">{{ detailRow.has_personal_leave_month ? '是' : '否' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import { getMonthlyList, calculateMonthly, exportAttendanceMonthly } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { formatDateTime, hoursToDays } from '@/utils'
import { downloadBlob } from '@/utils/download'

const tableRef=ref(), calcVisible=ref(false), saving=ref(false), calcMonth=ref(''), calcPersonIds=ref<number[]>([])
const personList=ref<{id:number;name:string}[]>([]), detailVisible=ref(false), detailRow=ref<any>(null)
const viewMode=ref<'cards'|'list'>('cards')
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
const actions=[{key:'calc',label:'批量核算',type:'primary' as const},{key:'export',label:'导出',type:'default' as const}]

onMounted(async()=>{personList.value=(await getAllPersons()) as any[]||[]})
async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchMonthly(p:any){
  return (await getMonthlyList(p)) as any
}

function handleAction(k:string){
  if(k==='calc'){calcMonth.value='';calcPersonIds.value=[];calcVisible.value=true}
  else if(k==='export'){handleExport()}
}
async function handleExport(){
  const params: any = {}
  if (calcMonth.value) params.month = calcMonth.value
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
function showDetail(row:any){detailRow.value=row;detailVisible.value=true}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
