<template>
  <div class="page-container">
    <PageHeader title="调休事件流水">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>

    <PageToolbar>
      <el-button type="primary" size="small" @click="handleAction('add')">新增</el-button>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" />
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :url-driven="true"
        :fetch-fn="(p: any) => getLILEvents(p)"
        date-field="event_date"
      >
        <template #day="{ items }">
          <div class="ev-grid">
            <LILEventCard v-for="e in items" :key="e.id" :event="e" @edit="editDaily" />
          </div>
        </template>
      </TimeCardPanel>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import LILEventCard from '@/components/annual-leave/LILEventCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import { getLILEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'

import { usePageView } from '@/composables/usePageView'

const router = useRouter()
const tableRef = ref()
const { viewMode } = usePageView('cards')
const timePanelRef = ref()

const columns = [
  { prop: 'person_name', label: '人员', width: '80' },
  { prop: 'sub_type', label: '类型', width: '100' },
  { prop: 'event_date', label: '日期', width: '110' },
  { prop: 'hours', label: '时长(天)', width: '90', formatter: (r: any) => hoursToDays(r.hours).toFixed(2) },
  { prop: 'remark', label: '备注' },
]
const searchFields = [
  { prop: 'person_id', label: '人员', type: 'person-select' as const, fetchApi: fetchOpts },
  { prop: 'date', label: '时间范围', type: 'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]

async function fetchOpts(k?: string) { const l = await getAllPersons() as any[]; return k ? l.filter(p => p.name.includes(k)) : l }
async function fetchEvents(p: any) {
  return (await getLILEvents(p)) as any
}

function handleAction(k: string) {
  if (k === 'add') {
    // 调休事件本质是考勤事件（补班出勤/调休），新增统一走考勤事件页
    router.push('/attendance-events/create')
  }
}

// 编辑=查看：调休事件即考勤日事件 → 跳该日考勤整日页
function editDaily(item: any) {
  if (item.daily_id) {
    router.push(`/attendance-events/${item.daily_id}`)
  } else {
    ElMessage.warning('未找到当日考勤记录')
  }
}
</script>
<style scoped>
.page-container { padding: 0; background: transparent; }

.ev-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
</style>
