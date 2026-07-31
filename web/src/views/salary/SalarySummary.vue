<template>
  <div class="page-container"><div class="page-header"><h2>月度工资汇总</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchSummaries" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #status="{ row }">
        <StatusTag :status="row.status || 'not_calculated'" />
      </template>
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">明细</el-button>
        <el-button type="success" link size="small" @click="showVersions(row)">版本</el-button>
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
    </el-dialog>

    <el-dialog v-model="verDetailVisible" title="版本详情" width="700px">
      <el-descriptions v-if="verDetail" :column="2" border size="small">
        <el-descriptions-item label="实发工资">{{ verDetail.final_salary }}</el-descriptions-item>
        <el-descriptions-item label="出勤工资">{{ verDetail.attendance_salary }}</el-descriptions-item>
        <el-descriptions-item label="绩效工资">{{ verDetail.performance_salary }}</el-descriptions-item>
        <el-descriptions-item label="全勤奖">{{ verDetail.attendance_bonus }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import StatusTag from '@/components/StatusTag.vue'
import { getSalarySummaries, calculateSalaries, getSalaryVersions, getSalaryVersionDetail } from '@/api/salary'
import { getAllPersons } from '@/api/person'

const tableRef=ref(), calcVisible=ref(false), saving=ref(false), calcMonth=ref(''), calcPersonIds=ref<number[]>([])
const personList=ref<{id:number;name:string}[]>([]), detailVisible=ref(false), detailRow=ref<any>(null)
const versionVisible=ref(false), versions=ref<any[]>([]), verDetailVisible=ref(false), verDetail=ref<any>(null)

const columns=[
  {prop:'person_name',label:'人员',width:'80'},{prop:'belong_month',label:'月份',width:'90'},
  {prop:'attendance_salary',label:'出勤工资',width:'100'},{prop:'performance_salary',label:'绩效工资',width:'100'},
  {prop:'attendance_bonus',label:'全勤奖',width:'80'},{prop:'final_salary',label:'实发工资',width:'110'},
  {prop:'status',label:'状态',width:'110',slot:'status'},
  {prop:'last_calc_at',label:'核算时间',width:'110'},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const,fetchApi:fetchPersonOpts},
  {prop:'month',label:'月份',type:'month' as const},
]
const actions=[{key:'calc',label:'批量核算',type:'primary' as const}]

onMounted(async()=>{personList.value=(await getAllPersons()) as any[]||[]})
async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchSummaries(p:any){
  const d=(await getSalarySummaries(p)) as any
  const persons=(await getAllPersons()) as any[]||[];const nm:Record<number,string>={};persons.forEach((x:any)=>nm[x.id]=x.name)
  return {list:(d.list||[]).map((r:any)=>({...r,person_name:nm[r.person_id]||'-'})),total:d.total||0}
}

function handleAction(k:string){if(k==='calc'){calcMonth.value='';calcPersonIds.value=[];calcVisible.value=true}}
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
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
