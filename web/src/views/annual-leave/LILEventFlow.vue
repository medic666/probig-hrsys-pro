<template>
  <div class="page-container">
    <PageHeader title="调休事件流水">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>

    <PageToolbar>
      <el-button v-permission="PERM.attendanceWrite" type="primary" size="small" @click="handleAction('add')">新增</el-button>
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
        :aggregate="'year'"
        badge-position="meta"
      >
        <template #person-badge="{ person }">
          <div class="balance-line" :class="{ 'is-zero': !(balanceMap[person.id] ?? 0) }">
            调休余额 {{ hoursToDays(balanceMap[person.id] ?? 0).toFixed(2) }} 天
          </div>
        </template>
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import LILEventCard from '@/components/annual-leave/LILEventCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import { getLILEvents } from '@/api/annual-leave'
import { hoursToDays } from '@/utils'

import { usePageView } from '@/composables/usePageView'
import { useBadges } from '@/composables/useBadges'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef = ref()
const { viewMode } = usePageView('cards')
const timePanelRef = ref()
// 调休余额映射（meta 位徽章）
const { balanceMap, loadBalances } = useBadges()

const columns = [
  { prop: 'person_name', label: '人员', width: '80' },
  { prop: 'sub_type', label: '类型', width: '100' },
  { prop: 'event_date', label: '日期', width: '110' },
  { prop: 'hours', label: '时长(天)', width: '90', formatter: (r: any) => hoursToDays(r.hours).toFixed(2) },
  { prop: 'remark', label: '备注' },
]
const searchFields = [
  { prop: 'person_id', label: '人员', type: 'person-select' as const },
  { prop: 'date', label: '时间范围', type: 'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]

async function fetchEvents(p: any) {
  return (await getLILEvents(p)) as any
}

onMounted(async () => {
  await loadBalances('lil-balances', '/lil-balances')
})

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
<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.page-container { padding: 0; background: transparent; }

.ev-grid {
  @include card-grid(260px);
}

.balance-line { font-size: 13px; color: #67c23a; font-weight: 600; line-height: 20px; }
.balance-line.is-zero { color: #c0c4cc; font-weight: 400; }
</style>
