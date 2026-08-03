<template>
  <el-descriptions v-if="calc && calc.id" :column="2" border size="small">
    <el-descriptions-item label="计薪天数">{{ calc.salary_days }}</el-descriptions-item>
    <el-descriptions-item label="加权基本工资">{{ calc.weighted_base_salary }}</el-descriptions-item>
    <el-descriptions-item label="记出勤(天)">{{ hoursToDays(calc.total_work_hours).toFixed(2) }}</el-descriptions-item>
    <el-descriptions-item label="工作日加班(天)">{{ hoursToDays(calc.total_overtime_workday_hours).toFixed(2) }}</el-descriptions-item>
    <el-descriptions-item label="节假日加班(天)">{{ hoursToDays(calc.total_overtime_holiday_hours).toFixed(2) }}</el-descriptions-item>
    <el-descriptions-item label="全勤奖">{{ calc.attendance_bonus }}</el-descriptions-item>
    <el-descriptions-item label="违纪次数">{{ calc.total_violation_count }}</el-descriptions-item>
    <el-descriptions-item label="有事假">{{ calc.has_personal_leave_month ? '是' : '否' }}</el-descriptions-item>
  </el-descriptions>
  <el-empty v-else :description="emptyText" :image-size="60" />
</template>

<script setup lang="ts">
import { hoursToDays } from '@/utils'

// 月度考勤核算展示组件（核心核算结果项）：月度考勤核算详情页与工资追溯-考勤核算共用。
// 仅展示核算核心字段（计薪/加权/工时/全勤/违纪/事假），金额类细项由工资层承接。
withDefaults(
  defineProps<{
    calc?: any
    emptyText?: string
  }>(),
  { calc: undefined, emptyText: '暂无核算记录' },
)
</script>
