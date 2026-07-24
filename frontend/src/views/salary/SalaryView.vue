<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="工资事件" name="events">
        <div style="display: flex; justify-content: space-between; margin-bottom: 12px">
          <el-input v-model="salaryEventEntityId" placeholder="人员ID" clearable style="width: 120px" @change="fetchEvents" />
          <el-button v-if="auth.hasPermission('salary', 'write')" type="primary" size="small" @click="openCreateEventForm()">新增事件</el-button>
        </div>

        <el-table :data="salaryEvents" border size="small" v-loading="eventLoading">
          <el-table-column prop="entity_name" label="姓名" width="100" />
          <el-table-column prop="event_type" label="事件类型" width="120">
            <template #default="{ row }">{{ typeLabel(row.event_type) }}</template>
          </el-table-column>
          <el-table-column prop="amount" label="金额" width="100" />
          <el-table-column label="期间" width="200">
            <template #default="{ row }">{{ row.period_start }} ~ {{ row.period_end }}</template>
          </el-table-column>
          <el-table-column prop="description" label="备注" min-width="120" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button v-if="auth.hasPermission('salary', 'write')" text size="small" @click="openEditEventForm(row)">编辑</el-button>
              <el-button v-if="auth.hasPermission('salary', 'delete')" text size="small" type="danger" @click="handleDeleteEvent(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination v-model:current-page="eventPage" :page-size="eventPageSize" :total="eventTotal" layout="prev, pager, next" small style="margin-top: 12px; justify-content: flex-end" @current-change="fetchEvents" />
      </el-tab-pane>

      <el-tab-pane label="工资汇总" name="summary">
        <div style="display: flex; justify-content: space-between; margin-bottom: 12px">
          <div style="display: flex; gap: 8px">
            <el-date-picker v-model="salarySummaryDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" @change="fetchSummaries" />
          </div>
          <el-button v-if="auth.hasPermission('salary', 'write')" type="success" size="small" @click="openCalcDialog">工资核算</el-button>
        </div>

        <el-table :data="summaries" border size="small" v-loading="summaryLoading" max-height="500">
          <el-table-column prop="entity_name" label="姓名" width="100" fixed />
          <el-table-column label="期间" width="200" fixed>
            <template #default="{ row }">{{ row.period_start }} ~ {{ row.period_end }}</template>
          </el-table-column>
          <el-table-column label="基本工资" width="100"><template #default="{ r }">{{ r.base_salary.toFixed(2) }}</template></el-table-column>
          <el-table-column label="出勤工资" width="100"><template #default="{ r }">{{ r.attendance_wage.toFixed(2) }}</template></el-table-column>
          <el-table-column label="全勤奖" width="90"><template #default="{ r }">{{ r.full_attendance_bonus.toFixed(2) }}</template></el-table-column>
          <el-table-column label="加班工资" width="100"><template #default="{ r }">{{ r.overtime_wage.toFixed(2) }}</template></el-table-column>
          <el-table-column label="绩效工资" width="100"><template #default="{ r }">{{ r.performance_salary.toFixed(2) }}</template></el-table-column>
          <el-table-column label="津贴合计" width="100"><template #default="{ r }">{{ (r.position_allowance + r.meal_subsidy + r.housing_subsidy + r.transport_subsidy + r.heat_subsidy + r.insurance_compensation + r.housing_fund_compensation).toFixed(2) }}</template></el-table-column>
          <el-table-column label="奖惩" width="90"><template #default="{ r }">{{ r.reward_punishment.toFixed(2) }}</template></el-table-column>
          <el-table-column label="应发合计" width="110"><template #default="{ r }"><strong>{{ r.gross_pay.toFixed(2) }}</strong></template></el-table-column>
          <el-table-column label="代扣合计" width="100"><template #default="{ r }">{{ (r.social_insurance_deduct + r.housing_fund_deduct + r.tax_deduction + r.loan_deduction).toFixed(2) }}</template></el-table-column>
          <el-table-column label="实发合计" width="110"><template #default="{ r }"><strong style="color: #67C23A">{{ r.net_pay.toFixed(2) }}</strong></template></el-table-column>
        </el-table>

        <el-pagination v-model:current-page="summaryPage" :page-size="summaryPageSize" :total="summaryTotal" layout="prev, pager, next" small style="margin-top: 12px; justify-content: flex-end" @current-change="fetchSummaries" />
      </el-tab-pane>
    </el-tabs>

    <SalaryEventForm ref="formRef" @success="fetchEvents" />

    <el-dialog v-model="calcDialogVisible" title="工资核算" width="400px">
      <el-form>
        <el-form-item label="核算期间">
          <el-date-picker v-model="calcRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="calcDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="calcLoading" @click="handleCalculate">开始核算</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../../stores/auth'
import { listSalaryEvents, deleteSalaryEvent, listSalarySummaries, calculateSalary } from '../../api/salary'
import type { SalaryEvent, SalarySummary } from '../../types'
import SalaryEventForm from './SalaryEventForm.vue'

const auth = useAuthStore()
const activeTab = ref('events')

const salaryEvents = ref<SalaryEvent[]>([])
const eventTotal = ref(0)
const eventPage = ref(1)
const eventPageSize = 20
const eventLoading = ref(false)
const salaryEventEntityId = ref('')

const summaries = ref<SalarySummary[]>([])
const summaryTotal = ref(0)
const summaryPage = ref(1)
const summaryPageSize = 20
const summaryLoading = ref(false)
const salarySummaryDateRange = ref<[string, string] | null>(null)

const formRef = ref<InstanceType<typeof SalaryEventForm>>()

const calcDialogVisible = ref(false)
const calcRange = ref<[string, string] | null>(null)
const calcLoading = ref(false)

function typeLabel(t: string) {
  return { performance: '业绩', reward_punishment: '奖惩', loan_deduction: '借款扣除', tax_deduction: '个税扣除', other: '其他' }[t] || t
}

async function fetchEvents() {
  eventLoading.value = true
  try {
    const res = await listSalaryEvents({
      entity_id: salaryEventEntityId.value ? Number(salaryEventEntityId.value) : undefined,
      page: eventPage.value,
      page_size: eventPageSize,
    })
    salaryEvents.value = res.data.list
    eventTotal.value = res.data.total
  } catch {} finally { eventLoading.value = false }
}

async function fetchSummaries() {
  summaryLoading.value = true
  try {
    const res = await listSalarySummaries({
      start_date: salarySummaryDateRange.value?.[0],
      end_date: salarySummaryDateRange.value?.[1],
      page: summaryPage.value,
      page_size: summaryPageSize,
    })
    summaries.value = res.data.list
    summaryTotal.value = res.data.total
  } catch {} finally { summaryLoading.value = false }
}

function openCreateEventForm() {
  formRef.value!.open('create')
}

function openEditEventForm(event: SalaryEvent) {
  formRef.value!.open('edit', event)
}

async function handleDeleteEvent(row: SalaryEvent) {
  try {
    await ElMessageBox.confirm('确定删除该事件吗？', '确认删除', { type: 'warning' })
    await deleteSalaryEvent(row.id)
    ElMessage.success('删除成功')
    fetchEvents()
  } catch {}
}

function openCalcDialog() {
  calcRange.value = null
  calcDialogVisible.value = true
}

async function handleCalculate() {
  if (!calcRange.value) { ElMessage.warning('请选择核算期间'); return }
  calcLoading.value = true
  try {
    await calculateSalary({ period_start: calcRange.value[0], period_end: calcRange.value[1] })
    ElMessage.success('工资核算完成')
    calcDialogVisible.value = false
    fetchSummaries()
  } catch { ElMessage.error('核算失败') } finally { calcLoading.value = false }
}
</script>
