<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-select v-model="searchForm.belong_month" placeholder="选择月份" clearable style="width: 160px;">
          <el-option v-for="m in months" :key="m" :label="m" :value="m" />
        </el-select>
        <el-button type="primary" @click="calculateVisible = true">工资核算</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="person_id" label="人员ID" min-width="80" />
        <el-table-column prop="belong_month" label="月份" min-width="100" />
        <el-table-column prop="attendance_salary" label="出勤工资" min-width="100" />
        <el-table-column prop="overtime_salary" label="加班工资" min-width="100" />
        <el-table-column prop="attendance_bonus" label="全勤奖" min-width="90" />
        <el-table-column prop="performance_salary" label="绩效工资" min-width="100" />
        <el-table-column prop="total_allowance" label="补贴合计" min-width="100" />
        <el-table-column prop="total_adjustment" label="调整项" min-width="90" />
        <el-table-column prop="total_deduction" label="代扣合计" min-width="90" />
        <el-table-column prop="final_salary" label="实发工资" min-width="100">
          <template #default="{ row }"><b style="color: #409EFF;">{{ row.final_salary }}</b></template>
        </el-table-column>
        <el-table-column label="状态" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_locked ? 'warning' : 'success'">{{ row.is_locked ? '已锁定' : '未锁定' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="100">
          <template #default="{ row }">
            <el-button size="small" @click="handleLock(row, !row.is_locked)">{{ row.is_locked ? '解锁' : '锁定' }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>

    <el-dialog v-model="calculateVisible" title="工资核算" width="400px">
      <el-form :model="calcForm" label-width="100px">
        <el-form-item label="人员ID">
          <el-input v-model.number="calcForm.person_id" />
        </el-form-item>
        <el-form-item label="月份">
          <el-input v-model="calcForm.belong_month" placeholder="YYYY-MM" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="calculateVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCalculate">核算</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import * as api from '../../api/salary'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const calculateVisible = ref(false)
const searchForm = reactive({ belong_month: '' })
const calcForm = reactive({ person_id: 0, belong_month: '' })

const months = computed(() => {
  const ms = []
  for (let y = 2024; y <= 2026; y++)
    for (let m = 1; m <= 12; m++)
      ms.push(`${y}-${String(m).padStart(2, '0')}`)
  return ms
})

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getSalarySummaryList({ page: page.value, page_size: pageSize.value, belong_month: searchForm.belong_month })
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleCalculate() {
  submitting.value = true
  try {
    await api.calculateSalary(calcForm)
    ElMessage.success('核算完成')
    calculateVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleLock(row: any, locked: boolean) {
  await api.lockSalarySummary({ person_id: row.person_id, belong_month: row.belong_month, is_locked: locked })
  ElMessage.success(locked ? '已锁定' : '已解锁')
  fetchData()
}

fetchData()
</script>
