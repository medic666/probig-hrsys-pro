<template>
  <FieldDescriptions v-if="calc && calc.id" :fields="ATTENDANCE_CALC_FIELDS" :data="calc" :column="2">
    <template #prefix>
      <el-descriptions-item v-if="showStatus" label="状态">
        <StatusTag :status="status || 'not_calculated'" />
      </el-descriptions-item>
      <el-descriptions-item v-if="showCalcAt" label="核算时间">{{ formatDateTime(calc.last_calc_at) }}</el-descriptions-item>
    </template>
  </FieldDescriptions>
  <el-empty v-else :description="emptyText" :image-size="60" />
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils'
import FieldDescriptions from '@/components/FieldDescriptions.vue'
import StatusTag from '@/components/StatusTag.vue'
import { ATTENDANCE_CALC_FIELDS } from '@/constants/fields'

// 月度考勤核算展示标准组件：字段由统一字段表驱动（与列表/导出/追溯同口径），
// 人员/月份等展示字段由后端各读取端点统一提供（数据自足），组件纯展示。
// 月度考勤核算详情页传 showStatus/showCalcAt（状态/核算时间置于最前）；
// 追溯页数据源无 status 字段且不需要时间，不传即不渲染。
withDefaults(
  defineProps<{
    calc?: any
    emptyText?: string
    showStatus?: boolean
    status?: string
    showCalcAt?: boolean
  }>(),
  {
    calc: undefined,
    emptyText: '暂无核算记录',
    showStatus: false,
    status: undefined,
    showCalcAt: false,
  },
)
</script>
