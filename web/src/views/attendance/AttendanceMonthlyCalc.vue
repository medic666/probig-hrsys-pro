<template>
  <div class="page-container">
    <div class="page-header"><h2>月度考勤核算</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchMonthly" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">查看明细</el-button>
      </template>
    </ProTable>

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
        <el-descriptions-item label="记出勤工时">{{ detailRow.total_work_hours }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班工时">{{ detailRow.total_overtime_workday_hours }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班工时">{{ detailRow.total_overtime_holiday_hours }}</el-descriptions-item>
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
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import { getMonthlyList, calculateMonthly } from '@/api/attendance'
import { getAllPersons } from '@/api/person'

const tableRef = ref()
const calcVisible = ref(false)
const saving = ref(false)
const calcMonth = ref('')
const calcPersonIds = ref<number[]>([])
const personList = ref<{id:number;name:string}[]>([])
const detailVisible = ref(false)
const detailRow = ref<any>(null)

const columns = [
  { prop:'id', label:'ID', width:'60' },
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'belong_month', label:'月份', width:'90' },
  { prop:'attendance_salary', label:'出勤工资', width:'100' },
  { prop:'overtime_workday_salary', label:'工作日加班工资', width:'120' },
  { prop:'overtime_holiday_salary', label:'节假日加班工资', width:'120' },
  { prop:'attendance_bonus', label:'全勤奖', width:'80' },
  { prop:'last_calc_at', label:'核算时间', width:'110' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOpts },
  { prop:'month', label:'月份', type:'month' as const },
]
const actions = [
  { key:'calc', label:'批量核算', type:'primary' as const },
]

onMounted(async () => { personList.value = (await getAllPersons()) as any[] || [] })

async function fetchPersonOpts(k?: string) {
  const list = (await getAllPersons()) as any[] || []
  return k ? list.filter((p:any)=> p.name.includes(k)) : list
}

async function fetchMonthly(p: any) {
  const d = (await getMonthlyList(p)) as any
  const list = Array.isArray(d) ? d : (d.list||[])
  const persons = (await getAllPersons()) as any[] || []
  const nameMap: Record<number,string> = {}
  persons.forEach((x:any)=> nameMap[x.id]=x.name)
  return { list: list.map((r:any)=>({...r,person_name:nameMap[r.person_id]||'-'})), total: list.length }
}

function handleAction(k: string) {
  if (k==='calc') { calcMonth.value=''; calcPersonIds.value=[]; calcVisible.value=true }
}

async function doCalc() {
  if (!calcMonth.value) { ElMessage.warning('请选择月份'); return }
  saving.value=true
  try {
    const d = await calculateMonthly({ month:calcMonth.value, person_ids:calcPersonIds.value }) as any
    ElMessage.success(`核算完成: 成功${d.success}条, 失败${d.fail}条`)
    calcVisible.value=false; tableRef.value?.refresh()
  } catch { /* */ } finally { saving.value=false }
}

function showDetail(row: any) { detailRow.value=row; detailVisible.value=true }
</script>
<style lang="scss" scoped>
.page-container { padding:0; background:transparent; }
.page-header { margin-bottom:16px; h2 { font-size:18px; font-weight:600; color:#303133; } }
</style>
