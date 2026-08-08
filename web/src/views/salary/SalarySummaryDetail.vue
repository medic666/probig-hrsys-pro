<template>
  <BusinessPage :title="`${personName} · ${month} 月度工资汇总`" back-to="/salary-summaries">
    <el-tabs v-model="activeTab">
      <!-- 工资明细 -->
      <el-tab-pane label="工资明细" name="summary">
        <div v-loading="loading" class="detail-wrap">
          <FieldDescriptions v-if="summaryRow" :fields="SALARY_SUMMARY_FIELDS" :data="summaryRow" :column="2">
            <template #prefix>
              <el-descriptions-item label="状态">
                <StatusTag :status="summaryRow.status || 'not_calculated'" />
              </el-descriptions-item>
              <el-descriptions-item label="核算时间">{{ formatDateTime(summaryRow.last_calc_at) }}</el-descriptions-item>
            </template>
          </FieldDescriptions>
          <el-empty v-else-if="!loading" description="当月未核算工资，请先在工资汇总页执行核算" :image-size="60" />
        </div>
      </el-tab-pane>

      <!-- 版本历史 -->
      <el-tab-pane label="版本历史" name="versions">
        <el-table v-loading="versionsLoading" :data="versions" border size="small">
          <el-table-column type="expand">
            <template #default="{ row }">
              <FieldDescriptions :fields="SALARY_SUMMARY_FIELDS" :data="row" :column="3" class="expand-desc" />
            </template>
          </el-table-column>
          <el-table-column prop="version" label="版本" width="70" />
          <el-table-column prop="final_salary" label="实发工资" width="110" />
          <el-table-column prop="operator_name" label="操作人" width="110" />
          <el-table-column prop="calc_batch_no" label="核算批次" min-width="180" />
          <el-table-column prop="created_at" label="核算时间" width="170" :formatter="(r: any) => formatDateTime(r.created_at)" />
        </el-table>
        <el-empty v-if="!versionsLoading && versions.length === 0" description="暂无版本记录" :image-size="60" />
      </el-tab-pane>

      <!-- 版本对比 -->
      <el-tab-pane label="版本对比" name="compare">
        <el-empty v-if="versions.length < 2" description="版本不足 2 个，无法对比" :image-size="60" />
        <template v-else>
          <div class="compare-picker">
            <el-select v-model="compareA" style="width:200px" @change="buildCompare">
              <el-option v-for="v in versions" :key="v.id" :label="'版本 ' + v.version" :value="v.id" />
            </el-select>
            <span style="margin:0 8px;color:#909399">对比</span>
            <el-select v-model="compareB" style="width:200px" @change="buildCompare">
              <el-option v-for="v in versions" :key="v.id" :label="'版本 ' + v.version" :value="v.id" />
            </el-select>
          </div>
          <el-table :data="compareRows" border size="small" style="margin-top:12px">
            <el-table-column prop="label" label="薪资项" width="150" />
            <el-table-column label="版本A" width="140">
              <template #default="{ row }"><span :class="{ 'cell-changed': row.changed }">{{ row.a }}</span></template>
            </el-table-column>
            <el-table-column label="版本B" width="140">
              <template #default="{ row }"><span :class="{ 'cell-changed': row.changed }">{{ row.b }}</span></template>
            </el-table-column>
            <el-table-column label="差异">
              <template #default="{ row }">
                <span v-if="row.diff === null">-</span>
                <span v-else-if="row.diff === 0" style="color:#909399">不变</span>
                <span v-else :style="{ color: row.diff > 0 ? '#f56c6c' : '#67c23a' }">{{ row.diff > 0 ? '+' : '' }}{{ row.diff }}</span>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </el-tab-pane>

      <!-- 全链路追溯 -->
      <el-tab-pane label="全链路追溯" name="trace">
        <div v-if="!traceData" v-loading="traceLoading" class="detail-wrap" />
        <el-tabs v-else v-model="traceTab">
          <el-tab-pane label="工资汇总" name="summary">
            <FieldDescriptions :fields="SALARY_SUMMARY_FIELDS" :data="traceData.summary" :column="2" />
          </el-tab-pane>

          <el-tab-pane label="考勤核算" name="calc">
            <AttendanceCalcDescriptions :calc="traceData.attendance_calc" empty-text="当月未核算考勤" />
          </el-tab-pane>

          <el-tab-pane label="每日明细" name="daily">
            <el-table :data="traceData.daily_projections" border size="small" max-height="420">
              <el-table-column type="expand">
                <template #default="{ row }">
                  <el-table :data="dailyEventsOf(row)" border size="small">
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

          <el-tab-pane label="职务快照" name="snapshots">
            <el-table :data="traceData.position_snapshots" border size="small" max-height="420">
              <el-table-column prop="effective_start_date" label="起始" width="110" />
              <el-table-column label="结束" width="110">
                <template #default="{ row: r }">{{ r.effective_end_date === '9999-12-31' ? '至今' : r.effective_end_date }}</template>
              </el-table-column>
              <el-table-column label="在职" width="60">
                <template #default="{ row: r }">{{ r.is_active ? '是' : '否' }}</template>
              </el-table-column>
              <el-table-column prop="base_salary" label="基本工资" width="100" />
              <el-table-column prop="performance_salary" label="绩效基数" width="100" />
              <el-table-column prop="meal_allowance" label="餐补" width="80" />
              <el-table-column prop="housing_allowance" label="房补" width="80" />
              <el-table-column prop="post_allowance" label="职位津贴" width="90" />
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="工资事件" name="events">
            <el-table :data="traceData.salary_events" border size="small" max-height="420">
              <el-table-column prop="event_type" label="类型" width="100" />
              <el-table-column prop="amount" label="值" width="100" />
              <el-table-column prop="remark" label="备注" />
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="年假结转" name="carryover">
            <el-table :data="traceData.annual_leave_carryover" border size="small" max-height="420">
              <el-table-column prop="effective_date" label="日期" width="110" />
              <el-table-column label="结转时长(天)" width="110" :formatter="(r: any) => hoursToDays(r.hours).toFixed(2)" />
              <el-table-column prop="remark" label="备注" />
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
import FieldDescriptions from '@/components/FieldDescriptions.vue'
import StatusTag from '@/components/StatusTag.vue'
import AttendanceCalcDescriptions from '@/components/attendance/AttendanceCalcDescriptions.vue'
import { SALARY_SUMMARY_FIELDS } from '@/constants/fields'
import { getSalarySummaries, getSalaryVersions, getSalaryTrace } from '@/api/salary'
import { formatDateTime, hoursToDays } from '@/utils'

const route = useRoute()
const personId = Number(route.params.personId)
const month = String(route.params.month)
const personName = String(route.query.name || '')
const activeTab = ref(String(route.query.tab || 'summary'))

const loading = ref(false)
const summaryRow = ref<any>(null)

const versions = ref<any[]>([])
const versionsLoading = ref(false)
const compareA = ref(0)
const compareB = ref(0)
const compareRows = ref<any[]>([])

const traceData = ref<any>(null)
const traceLoading = ref(false)
const traceTab = ref('summary')

// 版本对比字段 = 统一字段表去掉月份/人员（版本间两者恒同，无对比意义）
const compareFields = SALARY_SUMMARY_FIELDS.filter((f) => f.key !== 'belong_month' && f.key !== 'person_name')

function dailyEventsOf(proj: any) {
  const daily = (traceData.value?.attendance_dailies || []).find((d: any) => d.event_date === proj.work_date)
  return daily?.details || []
}

function buildCompare() {
  const va = versions.value.find((v) => v.id === compareA.value)
  const vb = versions.value.find((v) => v.id === compareB.value)
  compareRows.value = compareFields.map((f) => {
    const a = va ? va[f.key] : null
    const b = vb ? vb[f.key] : null
    const diff = typeof a === 'number' && typeof b === 'number' ? Math.round((b - a) * 100) / 100 : null
    return { label: f.label, a, b, diff, changed: a !== b }
  })
}

onMounted(async () => {
  loading.value = true
  try {
    const d = (await getSalarySummaries({ person_id: personId, month, pageNum: 1, pageSize: 1 })) as any
    summaryRow.value = d.list?.[0] || null
  } catch {
    summaryRow.value = null
  } finally {
    loading.value = false
  }

  versionsLoading.value = true
  try {
    versions.value = (await getSalaryVersions(personId, month)) as any[] || []
    if (versions.value.length >= 2) {
      // 默认次新对比最新：版本A=次新、版本B=最新，差异 = 最新 − 次新（本次核算的变化）
      compareA.value = versions.value[1].id
      compareB.value = versions.value[0].id
      buildCompare()
    }
  } catch {
    versions.value = []
  } finally {
    versionsLoading.value = false
  }

  traceLoading.value = true
  try {
    traceData.value = (await getSalaryTrace(personId, month)) as any
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
.compare-picker {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.cell-changed {
  color: #e6a23c;
  font-weight: 600;
}
.expand-desc {
  padding: 8px 16px;
}
</style>
