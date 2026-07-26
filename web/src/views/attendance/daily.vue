<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import { listDailyProjections } from '@/api/attendance'
import type { DailyListParams } from '@/api/attendance'
import ProjectionStatus from '@/components/ProjectionStatus.vue'
import { hoursToDays } from '@/utils/unit'

const tableRef = ref()

const searchFields = [
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'work_date_start', label: '日期起', type: 'date' as const },
  { prop: 'work_date_end', label: '日期止', type: 'date' as const }
]

const columns = [
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'work_date', label: '日期' },
  { prop: 'work_hours', label: '记出勤工时(天)', formatter: (_row: Record<string, unknown>) => hoursToDays(_row.work_hours as number).toFixed(2) },
  { prop: 'overtime_workday_hours', label: '工作日加班(小时)' },
  { prop: 'overtime_holiday_hours', label: '节假日加班(小时)' },
  { prop: 'has_personal_leave', label: '事假' },
  { prop: 'violation_count', label: '违纪次数' },
  { prop: 'last_calc_at', label: '状态' },
  { slot: 'status', label: '数据状态', width: 120 }
]

async function fetchList(params: Record<string, unknown>) {
  return listDailyProjections(params as unknown as DailyListParams)
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
