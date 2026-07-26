<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import { listSummaries, calcSummary } from '@/api/salary'
import type { SummaryListParams } from '@/api/salary'
import ProjectionStatus from '@/components/ProjectionStatus.vue'
import { formatDate } from '@/utils/date'

const tableRef = ref()
const calcLoading = ref(false)

const searchFields = [
  { prop: 'belong_month', label: '归属月份', type: 'month' as const },
  { prop: 'person_id', label: '人员', type: 'name-select' as const, nameType: 'person', placeholder: '请选择人员' },
  { prop: 'attendance_group', label: '考勤组', type: 'input' as const }
]

const columns = [
  { prop: 'belong_month', label: '月份' },
  { prop: 'person_name', label: '人员姓名' },
  { prop: 'attendance_salary', label: '出勤工资' },
  { prop: 'attendance_bonus', label: '全勤奖' },
  { prop: 'performance_salary', label: '绩效工资' },
  { prop: 'final_salary', label: '实发工资' },
  { slot: 'status', label: '状态', width: 120 }
]

async function fetchList(params: Record<string, unknown>) {
  return listSummaries(params as unknown as SummaryListParams)
}

async function handleCalc() {
  calcLoading.value = true
  try {
    await calcSummary({ belong_month: formatDate(new Date()).slice(0, 7) })
    ElMessage.success('核算完成')
    tableRef.value?.refresh()
  } catch {
    // error handled by interceptor
  } finally {
    calcLoading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :loading="calcLoading" @click="handleCalc">核算当前月</el-button>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Download">导出</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
      :show-actions="false"
    >
      <template #status="{ row }">
        <ProjectionStatus :status="row.status" />
      </template>
    </ProTable>
  </div>
</template>
