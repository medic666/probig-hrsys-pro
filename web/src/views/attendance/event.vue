<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, DocumentAdd } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listEvents, createEvent, updateEvent, deleteEvent } from '@/api/attendance'
import type { AttendanceEvent, EventListParams } from '@/api/attendance'
import { daysToHours } from '@/utils/unit'
import { useRoute } from 'vue-router'

const route = useRoute()
const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增考勤事件')
const editingRow = ref<AttendanceEvent | null>(null)

const defaultParams = ref<Record<string, unknown>>({})
if (route.query.person_id) {
  defaultParams.value.person_id = Number(route.query.person_id)
}

const eventTypes: Record<string, string[]> = {
  '出勤': ['普通出勤', '补班出勤', '外勤出勤'],
  '休假': ['调休', '事假', '病假', '年假', '法定假', '福利假'],
  '加班': ['工作日加班', '节假日加班'],
  '违纪': ['缺卡', '迟到', '早退']
}

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'event_date_start', label: '日期起', type: 'date' as const },
  { prop: 'event_date_end', label: '日期止', type: 'date' as const },
  { prop: 'event_type', label: '事件类型', type: 'select' as const, options: [
    { label: '出勤', value: '出勤' }, { label: '休假', value: '休假' },
    { label: '加班', value: '加班' }, { label: '违纪', value: '违纪' }
  ]},
  { prop: 'sub_type', label: '子类型', type: 'input' as const }
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'event_date', label: '日期' },
  { prop: 'event_type', label: '事件类型' },
  { prop: 'sub_type', label: '子类型' },
  { prop: 'hours', label: '时长(小时)' },
  { prop: 'is_special_approval', label: '特批' },
  { prop: 'remark', label: '备注' },
  { slot: 'actions', label: '操作', width: 160, fixed: 'right' as const }
]

const subTypeOptions = ref<{ label: string; value: string }[]>([])

function onEventTypeChange(selectedType: string) {
  const subTypes = eventTypes[selectedType] || []
  subTypeOptions.value = subTypes.map((s) => ({ label: s, value: s }))
}

const formFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', required: true },
  { prop: 'event_date', label: '日期', type: 'date' as const, required: true },
  { prop: 'event_type', label: '事件类型', type: 'select' as const, required: true, options: [
    { label: '出勤', value: '出勤' }, { label: '休假', value: '休假' },
    { label: '加班', value: '加班' }, { label: '违纪', value: '违纪' }
  ]},
  { prop: 'sub_type', label: '子类型', type: 'select' as const, required: true, options: [
    { label: '普通出勤', value: '普通出勤' }, { label: '补班出勤', value: '补班出勤' }, { label: '外勤出勤', value: '外勤出勤' },
    { label: '调休', value: '调休' }, { label: '事假', value: '事假' }, { label: '病假', value: '病假' },
    { label: '年假', value: '年假' }, { label: '法定假', value: '法定假' }, { label: '福利假', value: '福利假' },
    { label: '工作日加班', value: '工作日加班' }, { label: '节假日加班', value: '节假日加班' },
    { label: '缺卡', value: '缺卡' }, { label: '迟到', value: '迟到' }, { label: '早退', value: '早退' }
  ]},
  { prop: 'hours', label: '时长(天)', type: 'number' as const, required: true, min: 0 },
  { prop: 'punch_time', label: '打卡时间', type: 'input' as const },
  { prop: 'late_minutes', label: '迟到/早退分钟', type: 'number' as const, min: 0 },
  { prop: 'is_special_approval', label: '是否特批', type: 'switch' as const },
  { prop: 'remark', label: '备注', type: 'textarea' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listEvents(params as unknown as EventListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增考勤事件'
  formVisible.value = true
}

function handleEdit(row: AttendanceEvent) {
  editingRow.value = row
  formTitle.value = '编辑考勤事件'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { ...editingRow.value }
}

async function handleSubmit(data: Record<string, unknown>) {
  const submitData = { ...data }
  if (submitData.hours) {
    submitData.hours = daysToHours(submitData.hours as number)
  }
  if (editingRow.value) {
    await updateEvent(submitData as any)
  } else {
    await createEvent(submitData as any)
  }
}

async function handleDelete(row: AttendanceEvent) {
  try {
    await ElMessageBox.confirm(`确定要删除该考勤事件吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteEvent(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

function handleFormSuccess() {
  tableRef.value?.refresh()
}
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增事件</el-button>
        <el-button type="success" :icon="DocumentAdd">批量新增</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
      :default-params="defaultParams"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link :icon="Delete" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="formVisible"
      :title="formTitle"
      :form-fields="formFields"
      :initial-data="getInitialData()"
      :submit-api="handleSubmit"
      @success="handleFormSuccess"
    />
  </div>
</template>
