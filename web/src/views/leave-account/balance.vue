<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import { listBalances } from '@/api/leaveAccount'
import type { BalanceListParams } from '@/api/leaveAccount'
import ProjectionStatus from '@/components/ProjectionStatus.vue'
import { hoursToDays } from '@/utils/unit'

const tableRef = ref()

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'leave_type', label: '假期类型', type: 'select' as const, options: [
    { label: '年假', value: 'annual_leave' }, { label: '调休', value: 'time_off' }
  ]}
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'leave_type', label: '假期类型' },
  { prop: 'balance_hours', label: '可用额度(天)', formatter: (_row: Record<string, unknown>) => hoursToDays(_row.balance_hours as number).toFixed(2) },
  { prop: 'last_calc_at', label: '最后更新时间' },
  { slot: 'status', label: '状态', width: 120 }
]

async function fetchList(params: Record<string, unknown>) {
  return listBalances(params as unknown as BalanceListParams)
}
</script>

<template>
  <div class="page-container">
    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
      :show-actions="false"
    >
      <template #status>
        <ProjectionStatus status="calculated" />
      </template>
    </ProTable>
  </div>
</template>
