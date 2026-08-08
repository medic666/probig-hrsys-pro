<template>
  <div class="page-container">
    <PageHeader title="职务事件管理">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>

    <PageToolbar :right-visible="isList">
      <el-button v-permission="PERM.positionEventWrite" type="primary" size="small" @click="handleAction('add')">新增事件</el-button>
      <template #right>
        <el-button v-permission="PERM.positionEventWrite" size="small" @click="handleAction('trash')">回收站</el-button>
        <el-button v-permission="PERM.positionEventExport" size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="openView(row)">查看</el-button>
          <el-button v-permission="PERM.positionEventWrite" type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="PERM.positionEventWrite" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button v-permission="PERM.fileRead" type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getPositionEvents(p)"
        date-field="effective_date"
        :aggregate="'year'"
        :person-dot-map="dotMap"
        :url-driven="true"
      >
        <template #day="{ items }">
          <div class="ev-grid">
            <PositionEventCard v-for="e in items" :key="e.id" :event="e" @click="openView" @edit="openEdit" />
          </div>
        </template>
      </TimeCardPanel>
    </template>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restoreEvent" :columns="trashColumns" @restored="onRefresh" />

    <FileAttachDialog v-model:visible="attachVisible" target-type="position_event" :target-id="attachFileId" />
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
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import PositionEventCard from '@/components/position/PositionEventCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import { getPositionEvents, getPositionEventBadges, deletePositionEvent, restorePositionEvent, getDeletedPositionEvents, exportPositionEvents } from '@/api/position-event'
import { useBadges } from '@/composables/useBadges'

import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef = ref()
const { viewMode, isList } = usePageView('cards')
const timePanelRef = ref()
const trashVisible = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
// 徽章映射：personId → 颜色点（超两年无职务变动为 orange，无事件 gray，正常 green）
const { dotMap, loadDots } = useBadges()

const eventTypes = ['入职', '调薪调岗', '离职']

const columns = [
  { prop: 'person_name', label: '人员', width: '100' },
  { prop: 'event_type', label: '事件类型', width: '100' },
  { prop: 'changed_fields', label: '变更字段', minWidth: '240', formatter: (r: any) => (r.changed_fields || []).join('、') || '-' },
  { prop: 'effective_date', label: '生效日期', width: '110' },
  { prop: 'remark', label: '备注', minWidth: '120' },
]

const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const },
  { prop:'event_type', label:'事件类型', type:'select' as const, options: eventTypes.map(t => ({ label: t, value: t })) },
  { prop:'date', label:'生效日期', type:'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'event_type', label: '事件类型' },
  { prop: 'effective_date', label: '生效日期' },
]

async function fetchEvents(params: any) { return (await getPositionEvents(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedPositionEvents(params)) as any }
async function restoreEvent(id: number) { return restorePositionEvent(id) }

onMounted(async () => {
  await loadDots('position-events-badges', getPositionEventBadges)
})

const { run: handleExport } = useExport(exportPositionEvents, () => tableRef.value?.getSearchParams() || {})

function handleAction(key: string) {
  if (key === 'add') {
    // 新增=编辑=查看统一走业务逻辑页
    router.push('/position-events/create')
  } else if (key === 'trash') { trashVisible.value = true }
  else if (key === 'export') { handleExport() }
}



// 查看=详情页查看态；编辑=直达编辑态
function openView(row: any) {
  router.push(`/position-events/${row.id}`)
}

function openEdit(row: any) {
  router.push(`/position-events/${row.id}?edit=1`)
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确认删除该事件？', '提示', { type: 'warning' }) } catch { return }
  try { await deletePositionEvent(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh(); timePanelRef.value?.reload() } catch { /* handled */ }
}

function onRefresh() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.page-container { padding: 0; background: transparent; }

.ev-grid {
  @include card-grid(260px);
}
</style>
