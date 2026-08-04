<template>
  <div class="page-container">
    <PageHeader title="日记工时查询">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <template #right>
        <el-button size="small" @click="handleExport">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchDaily" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="showEvents(row)">查看原始事件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :url-driven="true"
        :fetch-fn="(p: any) => getDailyProjections(p)"
        date-field="work_date"
        status-field="status"
        :pending-values="['pending']"
        :person-dot-map="dotMap"
      >
        <template #day="{ date, items }">
          <div
            v-if="items.length > 0"
            class="proj-card"
            @click="showEvents({ person_id: items[0].person_id, work_date: date })"
          >
            <div class="pc-header">
              <span class="pc-date">{{ date }}</span>
              <span class="pc-person">{{ items[0].person_name || '' }}</span>
            </div>
            <div class="pc-lines">
              <div class="pc-line">记出勤：{{ hoursToDays(items[0].work_hours || 0).toFixed(2) }} 天</div>
              <div class="pc-line" :class="{ 'is-alert': items[0].overtime_workday_hours > 0 }">工作日加班：{{ hoursToDays(items[0].overtime_workday_hours || 0).toFixed(2) }} 天</div>
              <div class="pc-line">节假日加班：{{ hoursToDays(items[0].overtime_holiday_hours || 0).toFixed(2) }} 天</div>
              <div v-if="items[0].violation_count > 0" class="pc-line is-alert">有违纪</div>
              <div v-if="items[0].has_personal_leave" class="pc-line is-alert">有事假</div>
              <div v-if="items[0].remark" class="pc-line">备注：{{ items[0].remark }}</div>
            </div>
          </div>
        </template>
      </TimeCardPanel>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import { getDailyProjections, getDailyProjectionBadges, getEventsByDate, exportDailyProjections } from '@/api/attendance'
import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'
import { useBadges } from '@/composables/useBadges'
import { hoursToDays } from '@/utils'

const router = useRouter()
const tableRef = ref()
const { viewMode, isList } = usePageView('cards')
// 徽章映射：personId → 颜色点（上月无投影 gray / 同月事假+加班 orange / 正常 green）
const { dotMap, loadDots } = useBadges()

const columns = [
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'work_date', label:'日期', width:'110' },
  { prop:'status', label:'状态', width:'80', formatter:(r:any)=>({pending:'待确认',confirmed:'已确认'}[r.status]||r.status||'-') },
  { prop:'work_hours', label:'记出勤(天)', width:'110', formatter:(r:any)=>hoursToDays(r.work_hours).toFixed(2) },
  { prop:'overtime_workday_hours', label:'工作日加班(天)', width:'120', formatter:(r:any)=>hoursToDays(r.overtime_workday_hours).toFixed(2) },
  { prop:'overtime_holiday_hours', label:'节假日加班(天)', width:'120', formatter:(r:any)=>hoursToDays(r.overtime_holiday_hours).toFixed(2) },
  { prop:'violation_count', label:'违纪次数', width:'80' },
  { prop:'has_personal_leave', label:'有事假', width:'70', formatter:(r:any)=>r.has_personal_leave?'是':'' },
  { prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const },
  { prop:'date', label:'日期范围', type:'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]

const { run: handleExport } = useExport(exportDailyProjections, () => tableRef.value?.getSearchParams() || {})

async function fetchDaily(p: any) {
  return (await getDailyProjections(p)) as any
}

// 查看原始事件 = 进入该日考勤整日页（只读：日记工时模块进入无编辑/确认能力）
async function showEvents(row: any) {
  try {
    const d = (await getEventsByDate(row.person_id, row.work_date)) as any
    if (d?.daily_id) {
      router.push(`/attendance-events/${d.daily_id}?readonly=1`)
    } else {
      ElMessage.warning('未找到当日考勤记录')
    }
  } catch { /* handled */ }
}


onMounted(async () => {
  await loadDots('attendance-daily-badges', getDailyProjectionBadges)
})
</script>
<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.page-container{padding:0;background:transparent}

.proj-card {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
  padding: 10px 12px;
  cursor: pointer;
  transition: box-shadow 0.2s;

  @include hover-capable {
    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }
  }

  .pc-header {
    display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
    .pc-date { font-weight: 600; font-size: 14px; color: #303133; }
    .pc-person { color: #909399; font-size: 12px; }
  }
  .pc-lines {
    .pc-line { font-size: 12px; line-height: 22px; color: #606266; }
    /* 异常项蓝字提醒：工作日加班（有加班时）/有违纪/有事假 */
    .pc-line.is-alert { color: #409eff; font-weight: 600; }
  }
}
</style>
