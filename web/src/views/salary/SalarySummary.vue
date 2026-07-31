<template>
  <div class="page-container"><div class="page-header"><h2>月度工资汇总</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchSummaries" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #status="{ row }">
        <StatusTag :status="row.status || 'not_calculated'" />
      </template>
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">明细</el-button>
        <el-button type="success" link size="small" @click="showVersions(row)">版本</el-button>
        <el-button type="warning" link size="small" @click="showTrace(row)">追溯</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="calcVisible" title="批量核算" width="450px">
      <el-form label-width="80px">
        <el-form-item label="月份" required><el-date-picker v-model="calcMonth" type="month" value-format="YYYY-MM" style="width:100%"/></el-form-item>
        <el-form-item label="人员"><el-select v-model="calcPersonIds" multiple placeholder="不选则全部" style="width:100%"><el-option v-for="p in personList" :key="p.id" :label="p.name" :value="p.id"/></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="calcVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="doCalc">开始核算</el-button></template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="薪资构成明细" width="700px">
      <el-descriptions v-if="detailRow" :column="2" border size="small">
        <el-descriptions-item label="月份">{{ detailRow.belong_month }}</el-descriptions-item>
        <el-descriptions-item label="计薪天数">{{ detailRow.salary_days }}</el-descriptions-item>
        <el-descriptions-item label="出勤工资">{{ detailRow.attendance_salary }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班工资">{{ detailRow.overtime_workday_salary }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班工资">{{ detailRow.overtime_holiday_salary }}</el-descriptions-item>
        <el-descriptions-item label="年假结转工资">{{ detailRow.annual_leave_carryover_salary }}</el-descriptions-item>
        <el-descriptions-item label="全勤奖">{{ detailRow.attendance_bonus }}</el-descriptions-item>
        <el-descriptions-item label="绩效工资">{{ detailRow.performance_salary }}</el-descriptions-item>
        <el-descriptions-item label="职位津贴">{{ detailRow.post_allowance }}</el-descriptions-item>
        <el-descriptions-item label="餐补">{{ detailRow.meal_allowance }}</el-descriptions-item>
        <el-descriptions-item label="房补">{{ detailRow.housing_allowance }}</el-descriptions-item>
        <el-descriptions-item label="交通补贴">{{ detailRow.transport_allowance }}</el-descriptions-item>
        <el-descriptions-item label="高温补贴">{{ detailRow.high_temp_allowance }}</el-descriptions-item>
        <el-descriptions-item label="保险补偿">{{ detailRow.insurance_compensation }}</el-descriptions-item>
        <el-descriptions-item label="公积金补偿">{{ detailRow.fund_compensation }}</el-descriptions-item>
        <el-descriptions-item label="提成">{{ detailRow.sales_commission }}</el-descriptions-item>
        <el-descriptions-item label="奖惩">{{ detailRow.reward_punishment }}</el-descriptions-item>
        <el-descriptions-item label="借款还款">{{ detailRow.borrowing_repayment }}</el-descriptions-item>
        <el-descriptions-item label="社保代扣">{{ detailRow.social_security_deduct }}</el-descriptions-item>
        <el-descriptions-item label="公积金代扣">{{ detailRow.housing_fund_deduct }}</el-descriptions-item>
        <el-descriptions-item label="个税代扣">{{ detailRow.tax_deduct }}</el-descriptions-item>
        <el-descriptions-item label="实发工资"><strong style="color:#e6a23c;font-size:16px">{{ detailRow.final_salary }}</strong></el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="versionVisible" title="版本历史" width="750px">
      <el-table :data="versions" border size="small">
        <el-table-column prop="version" label="版本" width="60"/><el-table-column prop="final_salary" label="实发工资" width="100"/>
        <el-table-column prop="operator_name" label="操作人" width="100"/><el-table-column prop="created_at" label="核算时间" width="160"/>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button type="primary" link size="small" @click="showVersionDetail(row)">详情</el-button></template></el-table-column>
      </el-table>
      <template #footer>
        <el-button type="primary" :disabled="versions.length < 2" @click="openCompare">版本对比</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="compareVisible" title="版本对比" width="760px">
      <div class="compare-picker">
        <el-select v-model="compareA" style="width:180px" @change="buildCompare">
          <el-option v-for="v in versions" :key="v.id" :label="'版本 '+v.version" :value="v.id" />
        </el-select>
        <span style="margin:0 8px;color:#909399">对比</span>
        <el-select v-model="compareB" style="width:180px" @change="buildCompare">
          <el-option v-for="v in versions" :key="v.id" :label="'版本 '+v.version" :value="v.id" />
        </el-select>
      </div>
      <el-table :data="compareRows" border size="small" style="margin-top:12px">
        <el-table-column prop="label" label="薪资项" width="150" />
        <el-table-column label="版本A" width="140">
          <template #default="{ row }"><span :class="{ 'cell-changed': row.changed }">{{ row.a }}</span></template>
        </el-table-column>
        <el-table-column label="版本B" width="140">
          <template #default="{ row }"><span :class="{ 'cell-changed': row.changed }">{{ row.b }}</span></template>
        </el-table-column>
        <el-table-column label="差异">
          <template #default="{ row }">
            <span v-if="row.diff === null">-</span>
            <span v-else-if="row.diff === 0" style="color:#909399">不变</span>
            <span v-else :style="{ color: row.diff > 0 ? '#f56c6c' : '#67c23a' }">{{ row.diff > 0 ? '+' : '' }}{{ row.diff }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="verDetailVisible" title="版本详情" width="700px">
      <el-descriptions v-if="verDetail" :column="2" border size="small">
        <el-descriptions-item label="实发工资">{{ verDetail.final_salary }}</el-descriptions-item>
        <el-descriptions-item label="出勤工资">{{ verDetail.attendance_salary }}</el-descriptions-item>
        <el-descriptions-item label="绩效工资">{{ verDetail.performance_salary }}</el-descriptions-item>
        <el-descriptions-item label="全勤奖">{{ verDetail.attendance_bonus }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-drawer v-model="traceVisible" title="工资核算追溯" size="840px">
      <el-tabs v-if="traceData" v-model="traceTab">
        <el-tab-pane label="工资汇总" name="summary">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="计薪天数">{{ traceData.summary.salary_days }}</el-descriptions-item>
            <el-descriptions-item label="加权基本工资">{{ traceData.summary.weighted_base_salary }}</el-descriptions-item>
            <el-descriptions-item label="出勤工资">{{ traceData.summary.attendance_salary }}</el-descriptions-item>
            <el-descriptions-item label="工作日加班工资">{{ traceData.summary.overtime_workday_salary }}</el-descriptions-item>
            <el-descriptions-item label="节假日加班工资">{{ traceData.summary.overtime_holiday_salary }}</el-descriptions-item>
            <el-descriptions-item label="年假结转工资">{{ traceData.summary.annual_leave_carryover_salary }}</el-descriptions-item>
            <el-descriptions-item label="全勤奖">{{ traceData.summary.attendance_bonus }}</el-descriptions-item>
            <el-descriptions-item label="绩效工资">{{ traceData.summary.performance_salary }}</el-descriptions-item>
            <el-descriptions-item label="职位津贴">{{ traceData.summary.post_allowance }}</el-descriptions-item>
            <el-descriptions-item label="餐补">{{ traceData.summary.meal_allowance }}</el-descriptions-item>
            <el-descriptions-item label="房补">{{ traceData.summary.housing_allowance }}</el-descriptions-item>
            <el-descriptions-item label="交通补贴">{{ traceData.summary.transport_allowance }}</el-descriptions-item>
            <el-descriptions-item label="高温补贴">{{ traceData.summary.high_temp_allowance }}</el-descriptions-item>
            <el-descriptions-item label="保险补偿">{{ traceData.summary.insurance_compensation }}</el-descriptions-item>
            <el-descriptions-item label="公积金补偿">{{ traceData.summary.fund_compensation }}</el-descriptions-item>
            <el-descriptions-item label="提成">{{ traceData.summary.sales_commission }}</el-descriptions-item>
            <el-descriptions-item label="奖惩">{{ traceData.summary.reward_punishment }}</el-descriptions-item>
            <el-descriptions-item label="借款还款">{{ traceData.summary.borrowing_repayment }}</el-descriptions-item>
            <el-descriptions-item label="社保代扣">{{ traceData.summary.social_security_deduct }}</el-descriptions-item>
            <el-descriptions-item label="公积金代扣">{{ traceData.summary.housing_fund_deduct }}</el-descriptions-item>
            <el-descriptions-item label="个税代扣">{{ traceData.summary.tax_deduct }}</el-descriptions-item>
            <el-descriptions-item label="实发工资"><strong style="color:#e6a23c">{{ traceData.summary.final_salary }}</strong></el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane label="考勤核算" name="calc">
          <el-descriptions v-if="traceData.attendance_calc && traceData.attendance_calc.id" :column="2" border size="small">
            <el-descriptions-item label="计薪天数">{{ traceData.attendance_calc.salary_days }}</el-descriptions-item>
            <el-descriptions-item label="加权基本工资">{{ traceData.attendance_calc.weighted_base_salary }}</el-descriptions-item>
            <el-descriptions-item label="记出勤工时">{{ traceData.attendance_calc.total_work_hours }}</el-descriptions-item>
            <el-descriptions-item label="工作日加班工时">{{ traceData.attendance_calc.total_overtime_workday_hours }}</el-descriptions-item>
            <el-descriptions-item label="节假日加班工时">{{ traceData.attendance_calc.total_overtime_holiday_hours }}</el-descriptions-item>
            <el-descriptions-item label="全勤奖">{{ traceData.attendance_calc.attendance_bonus }}</el-descriptions-item>
            <el-descriptions-item label="违纪次数">{{ traceData.attendance_calc.total_violation_count }}</el-descriptions-item>
            <el-descriptions-item label="有事假">{{ traceData.attendance_calc.has_personal_leave_month ? '是' : '否' }}</el-descriptions-item>
          </el-descriptions>
          <el-empty v-else description="当月未核算考勤" :image-size="60" />
        </el-tab-pane>

        <el-tab-pane label="每日明细" name="daily">
          <el-table :data="traceData.daily_projections" border size="small" max-height="420">
            <el-table-column prop="work_date" label="日期" width="110" />
            <el-table-column prop="punch_time" label="打卡时间" width="110" />
            <el-table-column label="记出勤(天)" width="100" :formatter="(r:any)=>hoursToDays(r.work_hours).toFixed(2)" />
            <el-table-column label="工作日加班(天)" width="120" :formatter="(r:any)=>hoursToDays(r.overtime_workday_hours).toFixed(2)" />
            <el-table-column label="节假日加班(天)" width="120" :formatter="(r:any)=>hoursToDays(r.overtime_holiday_hours).toFixed(2)" />
            <el-table-column prop="violation_count" label="违纪" width="60" />
            <el-table-column label="原始事件" width="90">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="showDailyEvents(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="职务快照" name="snapshots">
          <el-table :data="traceData.position_snapshots" border size="small" max-height="420">
            <el-table-column prop="effective_start_date" label="起始" width="110" />
            <el-table-column prop="effective_end_date" label="结束" width="110">
              <template #default="{ row: r }">{{ r.effective_end_date === '9999-12-31' ? '至今' : r.effective_end_date }}</template>
            </el-table-column>
            <el-table-column prop="is_active" label="在职" width="60">
              <template #default="{ row: r }">{{ r.is_active ? '是' : '否' }}</template>
            </el-table-column>
            <el-table-column prop="base_salary" label="基本工资" width="100" />
            <el-table-column prop="performance_salary" label="绩效基数" width="100" />
            <el-table-column prop="meal_allowance" label="餐补" width="80" />
            <el-table-column prop="post_allowance" label="职位津贴" width="90" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="工资事件" name="events">
          <el-table :data="traceData.salary_events" border size="small" max-height="420">
            <el-table-column prop="event_type" label="类型" width="100" />
            <el-table-column prop="amount" label="值" width="100" />
            <el-table-column prop="remark" label="备注" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="年假结转" name="carryover">
          <el-table :data="traceData.annual_leave_carryover" border size="small" max-height="420">
            <el-table-column prop="effective_date" label="日期" width="110" />
            <el-table-column label="结转时长(天)" width="110" :formatter="(r:any)=>hoursToDays(r.hours).toFixed(2)" />
            <el-table-column prop="remark" label="备注" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <el-dialog v-model="dailyEventsVisible" title="当日原始考勤事件" width="560px">
      <el-table :data="dailyEvents" border size="small">
        <el-table-column prop="event_type" label="类型" width="90" />
        <el-table-column prop="sub_type" label="子类型" width="110" />
        <el-table-column label="时长(天)" width="90" :formatter="(r:any)=>hoursToDays(r.hours).toFixed(2)" />
        <el-table-column prop="remark" label="备注" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import { getSalarySummaries, calculateSalaries, exportSalarySummaries, getSalaryVersions, getSalaryTrace, getSalaryVersionDetail } from '@/api/salary'
import { getAllPersons } from '@/api/person'
import { formatDateTime, hoursToDays } from '@/utils'
import { downloadBlob } from '@/utils/download'

const tableRef=ref(), calcVisible=ref(false), saving=ref(false), calcMonth=ref(''), calcPersonIds=ref<number[]>([])
const personList=ref<{id:number;name:string}[]>([]), detailVisible=ref(false), detailRow=ref<any>(null)
const versionVisible=ref(false), versions=ref<any[]>([]), verDetailVisible=ref(false), verDetail=ref<any>(null)
const traceVisible=ref(false), traceData=ref<any>(null), traceTab=ref('summary')
const dailyEventsVisible=ref(false), dailyEvents=ref<any[]>([])
const compareVisible=ref(false), compareA=ref(0), compareB=ref(0), compareRows=ref<any[]>([])

const compareFields = [
  { key:'salary_days', label:'计薪天数' },
  { key:'weighted_base_salary', label:'加权基本工资' },
  { key:'total_work_hours', label:'累计记出勤工时' },
  { key:'total_overtime_workday_hours', label:'工作日加班工时' },
  { key:'total_overtime_holiday_hours', label:'节假日加班工时' },
  { key:'attendance_salary', label:'出勤工资' },
  { key:'overtime_workday_salary', label:'工作日加班工资' },
  { key:'overtime_holiday_salary', label:'节假日加班工资' },
  { key:'annual_leave_carryover_deduct', label:'年假结转时长' },
  { key:'annual_leave_carryover_salary', label:'年假结转工资' },
  { key:'attendance_bonus', label:'全勤奖' },
  { key:'performance_salary', label:'绩效工资' },
  { key:'post_allowance', label:'职位津贴' },
  { key:'meal_allowance', label:'餐补' },
  { key:'housing_allowance', label:'房补' },
  { key:'transport_allowance', label:'交通补贴' },
  { key:'high_temp_allowance', label:'高温补贴' },
  { key:'insurance_compensation', label:'保险补偿' },
  { key:'fund_compensation', label:'公积金补偿' },
  { key:'sales_commission', label:'提成' },
  { key:'reward_punishment', label:'奖惩' },
  { key:'borrowing_repayment', label:'借款还款' },
  { key:'social_security_deduct', label:'社保代扣' },
  { key:'housing_fund_deduct', label:'公积金代扣' },
  { key:'tax_deduct', label:'个税代扣' },
  { key:'final_salary', label:'实发工资' },
]

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
const actions=[{key:'calc',label:'批量核算',type:'primary' as const},{key:'export',label:'导出',type:'default' as const}]

onMounted(async()=>{personList.value=(await getAllPersons()) as any[]||[]})
async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchSummaries(p:any){
  return (await getSalarySummaries(p)) as any
}

function handleAction(k:string){
  if(k==='calc'){calcMonth.value='';calcPersonIds.value=[];calcVisible.value=true}
  else if(k==='export'){handleExport()}
}
async function handleExport(){
  const data = await exportSalarySummaries({})
  downloadBlob(data)
}
async function doCalc(){
  if(!calcMonth.value){ElMessage.warning('请选择月份');return}
  saving.value=true
  try{const d=await calculateSalaries({month:calcMonth.value,person_ids:calcPersonIds.value}) as any
    ElMessage.success(`核算完成: 成功${d.success}, 失败${d.fail}, 跳过${d.skip}`)
    calcVisible.value=false;tableRef.value?.refresh()
  }catch{/* */}finally{saving.value=false}
}
function showDetail(r:any){detailRow.value=r;detailVisible.value=true}
async function showVersions(r:any){versions.value=(await getSalaryVersions(r.person_id,r.belong_month)) as any[]||[];versionVisible.value=true}
async function showVersionDetail(r:any){verDetail.value=(await getSalaryVersionDetail(r.id)) as any;verDetailVisible.value=true}
async function showTrace(r:any){
  traceData.value=(await getSalaryTrace(r.person_id,r.belong_month)) as any
  traceTab.value='summary'
  traceVisible.value=true
}
function showDailyEvents(row:any){
  const daily=(traceData.value?.attendance_dailies||[]).find((d:any)=>d.event_date===row.work_date)
  dailyEvents.value=daily?.details||[]
  dailyEventsVisible.value=true
}

function openCompare(){
  compareA.value=versions.value[0].id
  compareB.value=versions.value[1].id
  buildCompare()
  compareVisible.value=true
}
function buildCompare(){
  const va=versions.value.find(v=>v.id===compareA.value)
  const vb=versions.value.find(v=>v.id===compareB.value)
  compareRows.value=compareFields.map(f=>{
    const a=va?va[f.key]:null
    const b=vb?vb[f.key]:null
    const diff=(typeof a==='number'&&typeof b==='number')?Math.round((b-a)*100)/100:null
    return { label:f.label, a, b, diff, changed:a!==b }
  })
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}
.cell-changed{color:#e6a23c;font-weight:600}</style>
