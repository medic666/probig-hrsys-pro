<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listEvents, createEvent, updateEvent, deleteEvent } from '@/api/salary'
import type { SalaryEvent, EventListParams } from '@/api/salary'
import { useRoute } from 'vue-router'

const route = useRoute()
const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增工资事件')
const editingRow = ref<SalaryEvent | null>(null)

const defaultParams = ref<Record<string, unknown>>({})
if (route.query.person_id) {
  defaultParams.value.person_id = Number(route.query.person_id)
}

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'belong_month', label: '归属月份', type: 'month' as const },
  { prop: 'event_type', label: '事件类型', type: 'select' as const, options: [
    { label: '绩效系数', value: '绩效系数' }, { label: '提成', value: '提成' },
    { label: '奖惩', value: '奖惩' }, { label: '借款还款', value: '借款还款' },
    { label: '个税扣除', value: '个税扣除' }, { label: '其他', value: '其他' }
  ]}
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'belong_month', label: '归属月份' },
  { prop: 'event_type', label: '事件类型' },
  { prop: 'amount', label: '金额/系数' },
  { prop: 'event_name', label: '事件名称' },
  { prop: 'remark', label: '备注' },
  { slot: 'actions', label: '操作', width: 160, fixed: 'right' as const }
]

const formFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', required: true },
  { prop: 'belong_month', label: '归属月份', type: 'month' as const, required: true },
  { prop: 'event_type', label: '事件类型', type: 'select' as const, required: true, options: [
    { label: '绩效系数', value: '绩效系数' }, { label: '提成', value: '提成' },
    { label: '奖惩', value: '奖惩' }, { label: '借款还款', value: '借款还款' },
    { label: '个税扣除', value: '个税扣除' }, { label: '其他', value: '其他' }
  ]},
  { prop: 'amount', label: '金额/系数', type: 'number' as const, required: true },
  { prop: 'event_name', label: '事件名称', type: 'input' as const, required: true },
  { prop: 'remark', label: '备注', type: 'textarea' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listEvents(params as unknown as EventListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增工资事件'
  formVisible.value = true
}

function handleEdit(row: SalaryEvent) {
  editingRow.value = row
  formTitle.value = '编辑工资事件'
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

async function handleDelete(row: SalaryEvent) {
  try {
    await ElMessageBox.confirm(`确定要删除该工资事件吗？`, '确认删除', { type: 'warning' })
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
    />
  </div>
</template>
