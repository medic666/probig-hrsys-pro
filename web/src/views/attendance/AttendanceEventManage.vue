<template>
  <div class="page-container">
    <PageHeader title="考勤事件管理">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="blocks" />
      </template>
    </PageHeader>

    <PageToolbar :right-visible="isList">
      <el-button v-permission="PERM.attendanceEventWrite" type="primary" size="small" @click="handleAction('add')">录入考勤</el-button>
      <el-button v-permission="PERM.attendanceEventWrite" type="success" size="small" @click="handleAction('batch')">批量录入</el-button>
      <el-button v-permission="PERM.attendanceEventWrite" type="warning" size="small" @click="handleAction('import')">钉钉导入</el-button>
      <template #right>
        <el-button v-permission="PERM.attendanceEventWrite" size="small" @click="handleAction('trash')">回收站</el-button>
        <el-button v-permission="PERM.attendanceEventExport" size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="openView(row)">查看</el-button>
          <el-button v-permission="PERM.attendanceEventWrite" type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="PERM.attendanceEventWrite" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button v-permission="PERM.fileRead" type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :url-driven="true"
        :fetch-fn="(p: any) => getAttendanceEvents(p)"
        date-field="event_date"
        status-field="status"
        :pending-values="['pending']"
        :person-dot-map="dotMap"
      >
        <template #day="{ items }">
          <AttendanceDailyDeck
            v-if="items.length > 0"
            :items="items"
            @view="openView"
            @edit="openEdit"
            @confirm="confirmDaily"
            @delete="handleDeleteDaily"
          />
        </template>
      </TimeCardPanel>
    </template>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restore" :columns="trashCols" @restored="onRefresh" />

    <FileAttachDialog v-model:visible="attachVisible" target-type="attendance_event" :target-id="attachFileId" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachDialog from '@/components/FileAttachDialog.vue'
import AttendanceDailyDeck from '@/components/attendance/AttendanceDailyDeck.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import { getAttendanceEvents, getAttendanceEventBadges, deleteAttendanceEvent, restoreAttendanceEvent, getDeletedAttendanceEvents, exportAttendanceEvents, confirmAttendanceDaily } from '@/api/attendance'
import { hoursToDays } from '@/utils'

import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'
import { useBadges } from '@/composables/useBadges'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef = ref()
const trashVisible = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
// 徽章映射：personId → 颜色点（上月无考勤事件 gray / 待确认 orange / 全确认 green）
const { dotMap, loadDots } = useBadges()

const { viewMode, isList } = usePageView('blocks')
const timePanelRef = ref()

function detailSummary(r: any): string {
  const parts = (r.details || []).map((d: any) => {
    let s = d.event_type
    if (d.sub_type) s += `-${d.sub_type}`
    if (d.event_type === '违纪') {
      if (d.minutes) s += ` ${d.minutes}分钟`
    } else {
      s += ` ${hoursToDays(d.hours || 0).toFixed(2)}天`
    }
    return s
  })
  return parts.join('；') || '-'
}

const columns = [
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'event_date', label:'日期', width:'110' },
  { prop:'seq', label:'组号', width:'60' },
  { prop:'status', label:'状态', width:'80', formatter:(r:any)=>({pending:'待确认',confirmed:'已确认'}[r.status]||r.status||'-') },
  { prop:'summary', label:'事件摘要', minWidth:'220', formatter:(r:any)=>detailSummary(r) },
  { prop:'punch_time', label:'打卡时间', width:'110', formatter:(r:any)=>r.punch_time || '-' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const },
  {
    prop:'status', label:'状态', type:'select' as const,
    options: [
      { label:'待确认', value:'pending' },
      { label:'已确认', value:'confirmed' },
    ],
  },
  { prop:'date', label:'日期', type:'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]
const trashCols = [{ prop:'id', label:'ID', width:'60' },{ prop:'event_type', label:'类型' },{ prop:'event_date', label:'日期' }]

async function fetchEvents(p: any) { return (await getAttendanceEvents(p)) as any }
const { run: handleExport } = useExport(exportAttendanceEvents, () => tableRef.value?.getSearchParams() || {})
async function fetchDeleted(p: any) { return (await getDeletedAttendanceEvents(p)) as any }
async function restore(id: number) { return restoreAttendanceEvent(id) }

// 查看=进入详情页查看态；编辑=直达详情页编辑态（?edit=1）
function openView(row: any) {
  router.push(`/attendance-events/${row.id}`)
}

function openEdit(row: any) {
  router.push(`/attendance-events/${row.id}?edit=1`)
}


// 卡片"确认"：提示后直接确认生效（就地转正：目标组提升为当日最新版，其余组降级待确认）
async function confirmDaily(row: any) {
  try {
    await ElMessageBox.confirm(`确认提交 ${row.person_name} ${row.event_date} 第${row.seq}组的整日事件？提交后该组将成为当日最新版并置为已确认，当日其它组将标记为待确认。`, '确认整日', { type: 'warning' })
  } catch { return }
  try {
    await confirmAttendanceDaily(row.id, row.details || [], row.punch_time || '', row.remark || '')
    ElMessage.success('确认成功')
    timePanelRef.value?.reload()
    await loadDots('attendance-events-badges', getAttendanceEventBadges)
  } catch { /* handled */ }
}

// 套卡"删除"：软删除当前显示的考勤组（仅套卡提供）
async function handleDeleteDaily(row: any) {
  try {
    await deleteAttendanceEvent(row.id)
    ElMessage.success('已删除')
    timePanelRef.value?.reload()
    await loadDots('attendance-events-badges', getAttendanceEventBadges)
  } catch { /* handled */ }
}

function handleAction(key: string) {
  if (key==='add') { router.push('/attendance-events/create') }
  else if (key==='batch') { router.push('/attendance-batch/create') }
  else if (key==='import') { router.push('/attendance-import') }
  else if (key==='trash') { trashVisible.value=true }
  else if (key==='export') { handleExport() }
}



async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确认删除？','提示',{type:'warning'}) } catch { return }
  try { await deleteAttendanceEvent(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh() } catch { /* */ }
}
function onRefresh() { tableRef.value?.refresh() }

onMounted(async () => {
  await loadDots('attendance-events-badges', getAttendanceEventBadges)
})
</script>
<style lang="scss" scoped>
.page-container { padding:0; background:transparent; }

.block-group {
  margin-bottom: 16px;
  .group-header {
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
    .group-count { color: #909399; font-size: 12px; font-weight: 400; margin-left: 8px; }
  }
  .block-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
}
.load-more {
  text-align: center;
  margin-top: 12px;
}
.block-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
</style>
