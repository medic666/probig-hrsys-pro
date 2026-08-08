<template>
  <div v-if="!traceData" v-loading="traceLoading" class="detail-wrap" />
  <el-tabs v-else v-model="traceTab">
    <el-tab-pane label="考勤核算" name="calc">
      <AttendanceCalcDescriptions :calc="traceData.calc" empty-text="当月未核算考勤" />
    </el-tab-pane>
    <el-tab-pane label="每日明细" name="daily">
      <el-table :data="traceData.daily_projections" border size="small" max-height="420">
        <el-table-column type="expand">
          <template #default="{ row: proj }">
            <el-table :data="dailyEventsOf(proj)" border size="small">
              <el-table-column prop="event_type" label="类型" width="90" />
              <el-table-column prop="sub_type" label="子类型" width="110" />
              <el-table-column label="时长(天)" width="90" :formatter="(r: any) => hoursToDays(r.hours).toFixed(2)" />
              <el-table-column prop="remark" label="备注" />
            </el-table>
          </template>
        </el-table-column>
        <el-table-column prop="work_date" label="日期" width="110" />
        <el-table-column prop="punch_time" label="打卡时间" width="110" />
        <el-table-column label="记出勤(天)" width="100" :formatter="(r: any) => hoursToDays(r.work_hours).toFixed(2)" />
        <el-table-column label="工作日加班(天)" width="120" :formatter="(r: any) => hoursToDays(r.overtime_workday_hours).toFixed(2)" />
        <el-table-column label="节假日加班(天)" width="120" :formatter="(r: any) => hoursToDays(r.overtime_holiday_hours).toFixed(2)" />
        <el-table-column prop="violation_count" label="违纪" width="60" />
      </el-table>
    </el-tab-pane>
  </el-tabs>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AttendanceCalcDescriptions from '@/components/attendance/AttendanceCalcDescriptions.vue'
import { getMonthlyList, getDailyProjections, getAttendanceEvents } from '@/api/attendance'
import { hoursToDays } from '@/utils'

// 月度考勤核算「全链路追溯」子视图：独立组件、数据自加载。
// 由父页面以 v-if（权限门控）+ lazy（激活才渲染）挂载——无权限或未激活时绝不发起请求。
const props = defineProps<{ personId: number; month: string }>()

const [monthStart, monthEnd] = monthRange(props.month)

// monthRange 由 "YYYY-MM" 计算当月首/末日
function monthRange(m: string): [string, string] {
  const [y, mo] = m.split('-').map(Number)
  const last = new Date(y, mo, 0).getDate()
  return [`${m}-01`, `${m}-${String(last).padStart(2, '0')}`]
}

const traceLoading = ref(false)
const traceTab = ref('calc')
const traceData = ref<any>(null)

// dailyEventsOf 展开当天考勤事件明细（按日期匹配考勤日记录，同日多版本取列表首条=最新版）
function dailyEventsOf(proj: any) {
  const daily = (traceData.value?.attendance_dailies || []).find((d: any) => d.event_date === proj.work_date)
  return daily?.details || []
}

onMounted(async () => {
  traceLoading.value = true
  try {
    // 追溯数据：日记工时投影 + 考勤日记录（行内带明细），加上当月核算
    const [calc, proj, dailies] = await Promise.all([
      getMonthlyList({ person_id: props.personId, month: props.month, pageNum: 1, pageSize: 1 }),
      getDailyProjections({ person_id: props.personId, date_start: monthStart, date_end: monthEnd, pageNum: 1, pageSize: 100 }),
      getAttendanceEvents({ person_id: props.personId, date_start: monthStart, date_end: monthEnd, pageNum: 1, pageSize: 100 }),
    ])
    traceData.value = {
      calc: (calc as any)?.list?.[0] || null,
      daily_projections: (proj as any)?.list || [],
      attendance_dailies: (dailies as any)?.list || [],
    }
  } catch {
    traceData.value = null
  } finally {
    traceLoading.value = false
  }
})
</script>

<style scoped>
.detail-wrap {
  min-height: 120px;
}
</style>
