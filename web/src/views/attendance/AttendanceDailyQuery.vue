<template>
  <div class="page-container">
    <div class="page-header"><h2>日记工时查询</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchDaily" :search-fields="searchFields" @action="noop">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showEvents(row)">查看原始事件</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="eventsVisible" title="当日考勤事件" width="600px">
      <el-table :data="dailyEvents" border size="small">
        <el-table-column prop="event_type" label="事件类型" width="80" />
        <el-table-column prop="sub_type" label="子类型" width="100" />
        <el-table-column prop="hours" label="时长(天)" width="90">
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
import { getDailyProjections, getEventsByDate } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'

const tableRef = ref()
const eventsVisible = ref(false)
const dailyEvents = ref<any[]>([])

const columns = [
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'work_date', label:'日期', width:'110' },
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
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
