<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listEvents, createManualEvent, updateManualEvent, deleteManualEvent } from '@/api/leaveAccount'
import type { LeaveAccountEvent, EventListParams } from '@/api/leaveAccount'

const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增额度事件')
const editingRow = ref<LeaveAccountEvent | null>(null)

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'leave_type', label: '假期类型', type: 'select' as const, options: [
    { label: '年假', value: 'annual_leave' }, { label: '调休', value: 'time_off' }
  ]},
  { prop: 'effective_date_start', label: '日期起', type: 'date' as const },
  { prop: 'effective_date_end', label: '日期止', type: 'date' as const }
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'leave_type', label: '假期类型' },
  { prop: 'event_type', label: '事件类型' },
  { prop: 'hours', label: '变动时长(小时)' },
  { prop: 'effective_date', label: '生效日期' },
  { prop: 'source_type', label: '事件来源' },
  { prop: 'remark', label: '备注' },
  { slot: 'actions', label: '操作', width: 160, fixed: 'right' as const }
]

const formFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', required: true },
  { prop: 'leave_type', label: '假期类型', type: 'select' as const, required: true, options: [
    { label: '年假', value: 'annual_leave' }, { label: '调休', value: 'time_off' }
  ]},
  { prop: 'event_type', label: '事件类型', type: 'select' as const, required: true, options: [
    { label: 'grant(配发)', value: 'grant' }, { label: 'adjust(人工调整)', value: 'adjust' }
  ]},
  { prop: 'hours', label: '额度(小时)', type: 'number' as const, required: true },
  { prop: 'effective_date', label: '生效日期', type: 'date' as const, required: true },
  { prop: 'remark', label: '备注', type: 'textarea' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listEvents(params as unknown as EventListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增额度事件'
  formVisible.value = true
}

function handleEdit(row: LeaveAccountEvent) {
  editingRow.value = row
  formTitle.value = '编辑额度事件'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { ...editingRow.value }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updateManualEvent(data as any)
  } else {
    await createManualEvent(data as any)
  }
}

async function handleDelete(row: LeaveAccountEvent) {
  if (row.source_type === 'system_period') {
    ElMessage.warning('系统结转事件不可单独删除')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除该额度事件吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteManualEvent(row.id)
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
    >
      <template #actions="{ row }">
        <el-button
          v-if="row.source_type !== 'system_period'"
          type="primary" link :icon="Edit"
          @click="handleEdit(row)"
        >编辑</el-button>
        <el-button
          v-if="row.source_type !== 'system_period'"
          type="danger" link :icon="Delete"
          @click="handleDelete(row)"
        >删除</el-button>
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
