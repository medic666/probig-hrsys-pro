<template>
  <BusinessPage :title="`${personName} · ${month} 月度考勤核算`" back-to="/attendance-monthly">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="月度核算" name="calc">
        <div v-loading="loading" class="detail-wrap">
          <AttendanceCalcDescriptions :calc="row" :show-status="true" :status="row?.status" :show-calc-at="true" empty-text="当月无核算记录" />
        </div>
      </el-tab-pane>
      <el-tab-pane label="全链路追溯" name="trace">
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
      </el-tab-pane>
    </el-tabs>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceCalcDescriptions from '@/components/attendance/AttendanceCalcDescriptions.vue'
import { getMonthlyList, getDailyProjections, getAttendanceEvents } from '@/api/attendance'
import { hoursToDays } from '@/utils'

const route = useRoute()
const personId = Number(route.params.personId)
const month = String(route.params.month)
const personName = String(route.query.name || '')
const activeTab = ref('calc')

const loading = ref(false)
const row = ref<any>(null)

const traceLoading = ref(false)
const traceTab = ref('calc')
const traceData = ref<any>(null)

const [monthStart, monthEnd] = monthRange(month)

// monthRange 由 "YYYY-MM" 计算当月首/末日
function monthRange(m: string): [string, string] {
  const [y, mo] = m.split('-').map(Number)
  const last = new Date(y, mo, 0).getDate()
  return [`${m}-01`, `${m}-${String(last).padStart(2, '0')}`]
}

// dailyEventsOf 展开当天考勤事件明细（按日期匹配考勤日记录，同日多版本取列表首条=最新版）
function dailyEventsOf(proj: any) {
  const daily = (traceData.value?.attendance_dailies || []).find((d: any) => d.event_date === proj.work_date)
  return daily?.details || []
}

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

  traceLoading.value = true
  try {
    // 追溯数据全部来自现有接口：日记工时投影 + 考勤日记录（行内带明细）
    const [proj, dailies] = await Promise.all([
      getDailyProjections({ person_id: personId, date_start: monthStart, date_end: monthEnd, pageNum: 1, pageSize: 100 }),
      getAttendanceEvents({ person_id: personId, date_start: monthStart, date_end: monthEnd, pageNum: 1, pageSize: 100 }),
    ])
    traceData.value = {
      calc: row.value || null,
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
