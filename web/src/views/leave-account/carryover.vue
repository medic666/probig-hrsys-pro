<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { List, RefreshLeft } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listBatches, executeCarryover, cancelBatch } from '@/api/leaveAccount'
import type { BatchListParams } from '@/api/leaveAccount'

const tableRef = ref()
const carryoverVisible = ref(false)

const searchFields = [
  { prop: 'batch_no', label: '批次号', type: 'input' as const },
  { prop: 'business_period', label: '业务周期', type: 'input' as const },
  { prop: 'status', label: '状态', type: 'select' as const, options: [
    { label: '待执行', value: 1 }, { label: '已生效', value: 2 }, { label: '已冲销', value: 3 }, { label: '执行失败', value: 4 }
  ]}
]

const columns = [
  { prop: 'batch_no', label: '批次号' },
  { prop: 'business_period', label: '业务周期' },
  { prop: 'operator_name', label: '操作人' },
  { prop: 'total_count', label: '处理人数' },
  { prop: 'status', label: '状态' },
  { prop: 'executed_at', label: '执行时间' },
  { slot: 'actions', label: '操作', width: 160, fixed: 'right' as const }
]

const carryoverFields = [
  { prop: 'target_month', label: '目标月份', type: 'month' as const, required: true }
]

const statusMap: Record<number, string> = {
  1: '待执行',
  2: '已生效',
  3: '已冲销',
  4: '执行失败'
}

async function fetchList(params: Record<string, unknown>) {
  return listBatches(params as unknown as BatchListParams)
}

async function handleCarryover() {
  carryoverVisible.value = true
}

async function handleCarryoverSubmit(data: Record<string, unknown>) {
  await executeCarryover({ target_month: data.target_month as string })
  ElMessage.success('结转操作已执行')
  tableRef.value?.refresh()
}

async function handleCancelBatch(row: { id: number; batch_no: string }) {
  try {
    await ElMessageBox.confirm(
      `确定要反结账批次「${row.batch_no}」吗？此操作将冲销该批次所有系统事件。`,
      '确认反结账',
      { type: 'warning' }
    )
  } catch {
    return
  }
  await cancelBatch(row.id)
  ElMessage.success('反结账成功')
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
        <el-button type="primary" :icon="List" @click="handleCarryover">年假周年批量结转</el-button>
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
          v-if="row.status === 2"
          type="danger" link :icon="RefreshLeft"
          @click="handleCancelBatch(row)"
        >反结账</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="carryoverVisible"
      title="年假周年批量结转"
      :form-fields="carryoverFields"
      :submit-api="handleCarryoverSubmit"
      @success="handleFormSuccess"
      width="450px"
    />
  </div>
</template>
