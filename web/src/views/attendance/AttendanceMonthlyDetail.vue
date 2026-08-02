<template>
  <BusinessPage :title="`${personName} · ${month} 月度考勤核算`" back-to="/attendance-monthly">
    <div v-loading="loading" class="detail-wrap">
      <el-descriptions v-if="row" :column="2" border size="small">
        <el-descriptions-item label="状态">
          <StatusTag :status="row.status || 'not_calculated'" />
        </el-descriptions-item>
        <el-descriptions-item label="计薪天数">{{ row.salary_days }}</el-descriptions-item>
        <el-descriptions-item label="加权基本工资">{{ row.weighted_base_salary }}</el-descriptions-item>
        <el-descriptions-item label="加权餐补">{{ row.weighted_meal_allowance }}</el-descriptions-item>
        <el-descriptions-item label="记出勤(天)">{{ hoursToDays(row.total_work_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班(天)">{{ hoursToDays(row.total_overtime_workday_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班(天)">{{ hoursToDays(row.total_overtime_holiday_hours).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="出勤工资">{{ row.attendance_salary }}</el-descriptions-item>
        <el-descriptions-item label="工作日加班工资">{{ row.overtime_workday_salary }}</el-descriptions-item>
        <el-descriptions-item label="节假日加班工资">{{ row.overtime_holiday_salary }}</el-descriptions-item>
        <el-descriptions-item label="全勤奖">{{ row.attendance_bonus }}</el-descriptions-item>
        <el-descriptions-item label="违纪次数">{{ row.total_violation_count }}</el-descriptions-item>
        <el-descriptions-item label="有事假">{{ row.has_personal_leave_month ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="核算时间">{{ formatDateTime(row.last_calc_at) }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else-if="!loading" description="当月无核算记录" :image-size="60" />
    </div>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import StatusTag from '@/components/StatusTag.vue'
import { getMonthlyList } from '@/api/attendance'
import { formatDateTime, hoursToDays } from '@/utils'

const route = useRoute()
const personId = Number(route.params.personId)
const month = String(route.params.month)
const personName = String(route.query.name || '')

const loading = ref(false)
const row = ref<any>(null)

onMounted(async () => {
  loading.value = true
  try {
    const d = (await getMonthlyList({ person_id: personId, month, pageNum: 1, pageSize: 1 })) as any
    row.value = d.list?.[0] || null
  } catch {
    row.value = null
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.detail-wrap {
  min-height: 120px;
}
</style>
