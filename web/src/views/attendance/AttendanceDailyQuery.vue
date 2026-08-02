<template>
  <div class="page-container">
    <PageHeader title="日记工时查询">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchDaily" :search-fields="searchFields" @action="noop">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="showEvents(row)">查看原始事件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getDailyProjections(p)"
        date-field="work_date"
        status-field="status"
        :pending-values="['pending']"
      >
        <template #day="{ date, items }">
          <div v-if="items.length > 0" class="proj-card" :class="{ 'is-pending': items[0].status === 'pending' }">
            <div class="pc-header">
              <span class="pc-date">{{ date }}</span>
              <span class="pc-person">{{ items[0].person_name || '' }}</span>
              <el-tag v-if="items[0].status === 'pending'" type="warning" size="small">待确认</el-tag>
            </div>
            <div class="pc-lines">
              <div class="pc-line">记出勤：{{ hoursToDays(items[0].work_hours || 0).toFixed(2) }} 天</div>
              <div class="pc-line">工作日加班：{{ hoursToDays(items[0].overtime_workday_hours || 0).toFixed(2) }} 天</div>
              <div class="pc-line">节假日加班：{{ hoursToDays(items[0].overtime_holiday_hours || 0).toFixed(2) }} 天</div>
              <div class="pc-line">违纪次数：{{ items[0].violation_count || 0 }}</div>
              <div v-if="items[0].has_personal_leave" class="pc-line">有事假</div>
              <div v-if="items[0].remark" class="pc-line">备注：{{ items[0].remark }}</div>
            </div>
          </div>
        </template>
      </TimeCardPanel>
    </template>

    <el-dialog v-model="eventsVisible" title="当日考勤事件" width="600px">
      <el-table :data="dailyEvents" border size="small">
        <el-table-column prop="event_type" label="事件类型" width="80" />
        <el-table-column prop="sub_type" label="子类型" width="100" />
        <el-table-column label="时长(天)" width="90">
          <template #default="{ row: r }">{{ hoursToDays(r.hours).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="punch_time" label="打卡时间" width="90" />
        <el-table-column prop="remark" label="备注" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import { getDailyProjections, getEventsByDate } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'

const tableRef = ref()
const viewMode = ref<'cards'|'list'>('cards')
const eventsVisible = ref(false)
const dailyEvents = ref<any[]>([])

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
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOptions },
  { prop:'date', label:'日期范围', type:'date-range' as const },
]

async function fetchPersonOptions(k?: string) { const l=(await getAllPersons()) as any[]||[]; return k?l.filter((p:any)=>p.name.includes(k)):l }
async function fetchDaily(p: any) {
  return (await getDailyProjections(p)) as any
}

async function showEvents(row: any) { dailyEvents.value = (await getEventsByDate(row.person_id, row.work_date)) as any[]||[]; eventsVisible.value=true }
function noop() {}
</script>
<style scoped>
.page-container{padding:0;background:transparent}

.proj-card {
  width: 260px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
  padding: 10px 12px;
  &.is-pending {
    border-color: #eebe77;
  }
  .pc-header {
    display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
    .pc-date { font-weight: 600; font-size: 14px; color: #303133; }
    .pc-person { color: #909399; font-size: 12px; }
  }
  .pc-lines {
    .pc-line { font-size: 12px; line-height: 22px; color: #606266; }
  }
}
</style>
