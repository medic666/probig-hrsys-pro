<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listEvents, createEvent, updateEvent, deleteEvent } from '@/api/position'
import type { PositionEvent, PositionEventListParams } from '@/api/position'
import { useRoute } from 'vue-router'

const route = useRoute()
const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增职务事件')
const editingRow = ref<PositionEvent | null>(null)

const defaultParams = ref<Record<string, unknown>>({})
if (route.query.person_id) {
  defaultParams.value.person_id = Number(route.query.person_id)
}

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'effective_date_start', label: '生效日期起', type: 'date' as const },
  { prop: 'effective_date_end', label: '生效日期止', type: 'date' as const },
  { prop: 'event_name', label: '事件名称', type: 'input' as const }
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'event_name', label: '事件名称' },
  { prop: 'effective_date', label: '生效日期' },
  { prop: 'changed_fields', label: '变更字段' },
  { prop: 'created_at', label: '创建时间' },
  { slot: 'actions', label: '操作', width: 160, fixed: 'right' as const }
]

const formFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', required: true },
  { prop: 'event_name', label: '事件名称', type: 'select' as const, required: true, options: [
    { label: '入职', value: '入职' }, { label: '转正', value: '转正' }, { label: '调薪', value: '调薪' },
    { label: '调岗', value: '调岗' }, { label: '离职', value: '离职' }
  ]},
  { prop: 'effective_date', label: '生效日期', type: 'date' as const, required: true },
  { prop: 'attendance_group', label: '考勤组', type: 'input' as const },
  { prop: 'base_salary', label: '基本工资', type: 'number' as const, min: 0 },
  { prop: 'performance_salary', label: '绩效工资基数', type: 'number' as const, min: 0 },
  { prop: 'salary_days', label: '计薪天数', type: 'number' as const, min: 0 },
  { prop: 'post_allowance', label: '职位津贴', type: 'number' as const, min: 0 },
  { prop: 'meal_allowance', label: '餐补', type: 'number' as const, min: 0 },
  { prop: 'housing_allowance', label: '房补', type: 'number' as const, min: 0 },
  { prop: 'transport_allowance', label: '交通补贴', type: 'number' as const, min: 0 },
  { prop: 'high_temp_allowance', label: '高温补贴月标准', type: 'number' as const, min: 0 },
  { prop: 'insurance_compensation', label: '保险补偿', type: 'number' as const, min: 0 },
  { prop: 'fund_compensation', label: '公积金补偿', type: 'number' as const, min: 0 },
  { prop: 'social_security_deduct', label: '社保代扣', type: 'number' as const, min: 0 },
  { prop: 'housing_fund_deduct', label: '公积金代扣', type: 'number' as const, min: 0 },
  { prop: 'has_annual_leave', label: '享有年假', type: 'switch' as const },
  { prop: 'has_attendance_bonus', label: '享有全勤奖', type: 'switch' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listEvents(params as unknown as PositionEventListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增职务事件'
  formVisible.value = true
}

function handleEdit(row: PositionEvent) {
  editingRow.value = row
  formTitle.value = '编辑职务事件'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { ...editingRow.value }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updateEvent(data as any)
  } else {
    await createEvent(data as any)
  }
}

async function handleDelete(row: PositionEvent) {
  try {
    await ElMessageBox.confirm(`确定要删除该职务事件吗？`, '确认删除', { type: 'warning' })
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
      width="700px"
    />
  </div>
</template>
