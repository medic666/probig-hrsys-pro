<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="假勤事件" name="events">
        <div style="display: flex; justify-content: space-between; margin-bottom: 12px">
          <div style="display: flex; gap: 8px">
            <el-input v-model="eventEntityId" placeholder="人员ID" clearable style="width: 120px" @change="fetchEvents" />
            <el-date-picker v-model="eventDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" @change="fetchEvents" />
          </div>
          <el-button v-if="auth.hasPermission('attendance', 'write')" type="primary" size="small" @click="openCreateEventForm()">新增事件</el-button>
        </div>

        <el-table :data="events" border size="small" v-loading="eventLoading">
          <el-table-column prop="event_date" label="日期" width="110" />
          <el-table-column prop="entity_name" label="姓名" width="100" />
          <el-table-column prop="event_category" label="事件类别" width="100">
            <template #default="{ row }">{{ categoryLabel(row.event_category) }}</template>
          </el-table-column>
          <el-table-column prop="event_subtype" label="事件子类" width="100">
            <template #default="{ row }">{{ subtypeLabel(row.event_category, row.event_subtype) }}</template>
          </el-table-column>
          <el-table-column prop="duration_days" label="天数/次数" width="90" />
          <el-table-column prop="description" label="备注" min-width="120" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button v-if="auth.hasPermission('attendance', 'write')" text size="small" @click="openEditEventForm(row)">编辑</el-button>
              <el-button v-if="auth.hasPermission('attendance', 'delete')" text size="small" type="danger" @click="handleDeleteEvent(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination v-model:current-page="eventPage" :page-size="eventPageSize" :total="eventTotal" layout="prev, pager, next" small style="margin-top: 12px; justify-content: flex-end" @current-change="fetchEvents" />
      </el-tab-pane>

      <el-tab-pane label="假勤汇总" name="summary">
        <div style="display: flex; justify-content: space-between; margin-bottom: 12px">
          <div style="display: flex; gap: 8px">
            <el-date-picker v-model="summaryDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" @change="fetchSummaries" />
          </div>
          <el-button v-if="auth.hasPermission('attendance', 'write')" type="success" size="small" @click="openCalcDialog">考勤核算</el-button>
        </div>

        <el-table :data="summaries" border size="small" v-loading="summaryLoading">
          <el-table-column prop="entity_name" label="姓名" width="100" />
          <el-table-column prop="period_start" label="期间" width="200">
            <template #default="{ row }">{{ row.period_start }} ~ {{ row.period_end }}</template>
          </el-table-column>
          <el-table-column label="出勤" width="80"><template #default="{ r }">{{ r.normal_days + r.makeup_days }}</template></el-table-column>
          <el-table-column label="病假" width="60"><template #default="{ r }">{{ r.sick_days }}</template></el-table-column>
          <el-table-column label="事假" width="60"><template #default="{ r }">{{ r.personal_days }}</template></el-table-column>
          <el-table-column label="年假" width="60"><template #default="{ r }">{{ r.annual_days }}</template></el-table-column>
          <el-table-column label="缺勤" width="60"><template #default="{ r }">{{ r.missing_card_count + r.late_count + r.early_count }}</template></el-table-column>
          <el-table-column label="加班" width="60"><template #default="{ r }">{{ r.workday_overtime + r.holiday_overtime }}</template></el-table-column>
        </el-table>

        <el-pagination v-model:current-page="summaryPage" :page-size="summaryPageSize" :total="summaryTotal" layout="prev, pager, next" small style="margin-top: 12px; justify-content: flex-end" @current-change="fetchSummaries" />
      </el-tab-pane>
    </el-tabs>

    <AttendanceEventForm ref="formRef" @success="fetchEvents" />

    <el-dialog v-model="calcDialogVisible" title="考勤核算" width="400px">
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
import { listAttendanceEvents, deleteAttendanceEvent, listAttendanceSummaries, calculateAttendance } from '../../api/attendance'
import type { AttendanceEvent, AttendanceSummary } from '../../types'
import AttendanceEventForm from './AttendanceEventForm.vue'

const auth = useAuthStore()
const activeTab = ref('events')

const events = ref<AttendanceEvent[]>([])
const eventTotal = ref(0)
const eventPage = ref(1)
const eventPageSize = 20
const eventLoading = ref(false)
const eventEntityId = ref('')
const eventDateRange = ref<[string, string] | null>(null)

const summaries = ref<AttendanceSummary[]>([])
const summaryTotal = ref(0)
const summaryPage = ref(1)
const summaryPageSize = 20
const summaryLoading = ref(false)
const summaryDateRange = ref<[string, string] | null>(null)

const formRef = ref<InstanceType<typeof AttendanceEventForm>>()

const calcDialogVisible = ref(false)
const calcRange = ref<[string, string] | null>(null)
const calcLoading = ref(false)

function categoryLabel(c: string) {
  return { attendance: '出勤', leave: '休假', overtime: '加班', discipline: '违纪', annual_adjustment: '年假调整' }[c] || c
}

function subtypeLabel(cat: string, sub: string) {
  const m: Record<string, Record<string, string>> = {
    attendance: { normal: '普通出勤', makeup: '补班出勤' },
    leave: { lieu: '调休', personal: '事假', sick: '病假', annual: '年假', statutory: '法定假', welfare: '福利假' },
    overtime: { workday: '工作日加班', holiday: '节假日加班' },
    discipline: { missing_card: '缺卡', late: '迟到', early: '早退' },
    annual_adjustment: { allocation: '年假配发', carryover: '年假结转' },
  }
  return m[cat]?.[sub] || sub
}

async function fetchEvents() {
  eventLoading.value = true
  try {
    const res = await listAttendanceEvents({
      entity_id: eventEntityId.value ? Number(eventEntityId.value) : undefined,
      start_date: eventDateRange.value?.[0],
      end_date: eventDateRange.value?.[1],
      page: eventPage.value,
      page_size: eventPageSize,
    })
    events.value = res.data.list
    eventTotal.value = res.data.total
  } catch { ElMessage.error('加载失败') } finally { eventLoading.value = false }
}

async function fetchSummaries() {
  summaryLoading.value = true
  try {
    const res = await listAttendanceSummaries({
      start_date: summaryDateRange.value?.[0],
      end_date: summaryDateRange.value?.[1],
      page: summaryPage.value,
      page_size: summaryPageSize,
    })
    summaries.value = res.data.list
    summaryTotal.value = res.data.total
  } catch { ElMessage.error('加载失败') } finally { summaryLoading.value = false }
}

function openCreateEventForm() {
  formRef.value!.open('create')
}

function openEditEventForm(event: AttendanceEvent) {
  formRef.value!.open('edit', event)
}

async function handleDeleteEvent(row: AttendanceEvent) {
  try {
    await ElMessageBox.confirm('确定删除该事件吗？', '确认删除', { type: 'warning' })
    await deleteAttendanceEvent(row.id)
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
    await calculateAttendance({ period_start: calcRange.value[0], period_end: calcRange.value[1] })
    ElMessage.success('考勤核算完成')
    calcDialogVisible.value = false
    fetchSummaries()
  } catch { ElMessage.error('核算失败') } finally { calcLoading.value = false }
}
</script>
