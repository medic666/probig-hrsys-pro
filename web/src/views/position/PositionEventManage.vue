<template>
  <div class="page-container">
    <PageHeader title="职务事件管理">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>

    <PageToolbar>
      <el-button type="primary" size="small" @click="handleAction('add')">新增事件</el-button>
      <el-button size="small" @click="handleAction('trash')">回收站</el-button>
      <el-button size="small" @click="handleAction('export')">导出</el-button>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getPositionEvents(p)"
        date-field="effective_date"
      >
        <template #day="{ date, items }">
          <DayCard :date="date" :events="items" @event-click="handleEdit" />
        </template>
      </TimeCardPanel>
    </template>

    <PositionEventEditDialog v-model:visible="dialogVisible" :event="editEvent" @saved="onEventSaved" />

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restoreEvent" :columns="trashColumns" @restored="onRefresh" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'position_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import DayCard from '@/components/cards/DayCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import PositionEventEditDialog from '@/components/position/PositionEventEditDialog.vue'
import { getPositionEvents, deletePositionEvent, restorePositionEvent, getDeletedPositionEvents, exportPositionEvents } from '@/api/position-event'
import { getAllPersons } from '@/api/person'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const viewMode = ref<'cards'|'list'>('cards')
const timePanelRef = ref()
const dialogVisible = ref(false)
const editEvent = ref<any>(null)
const trashVisible = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)

const eventTypes = ['入职', '调薪调岗', '离职']

const columns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'person_name', label: '人员', width: '100' },
  { prop: 'event_type', label: '事件类型', width: '100' },
  { prop: 'company_name', label: '公司组', width: '100', formatter: (r: any) => r.company_name || '-' },
  { prop: 'department', label: '部门', width: '110', formatter: (r: any) => r.department || '-' },
  { prop: 'position', label: '职位', width: '110', formatter: (r: any) => r.position || '-' },
  { prop: 'effective_date', label: '生效日期', width: '110' },
  { prop: 'remark', label: '备注', minWidth: '120' },
  { prop: 'changed_fields', label: '变更字段', formatter: (r: any) => (r.changed_fields || []).join('、') || '-' },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (r: any) => new Date(r.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOptions },
  { prop:'event_type', label:'事件类型', type:'select' as const, options: eventTypes.map(t => ({ label: t, value: t })) },
]

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'event_type', label: '事件类型' },
  { prop: 'effective_date', label: '生效日期' },
]

async function fetchPersonOptions(keyword?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[]
  if (!keyword) return list
  return list.filter(p => p.name.includes(keyword))
}

async function fetchEvents(params: any) { return (await getPositionEvents(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedPositionEvents(params)) as any }
async function restoreEvent(id: number) { return restorePositionEvent(id) }

function handleAction(key: string) {
  if (key === 'add') {
    editEvent.value = null
    dialogVisible.value = true
  } else if (key === 'trash') { trashVisible.value = true }
  else if (key === 'export') { handleExport() }
}

async function handleExport() {
  const data = await exportPositionEvents({})
  downloadBlob(data)
}

async function handleEdit(row: any) {
  editEvent.value = row
  dialogVisible.value = true
}

function onEventSaved() {
  tableRef.value?.refresh()
  timePanelRef.value?.reload()
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确认删除该事件？', '提示', { type: 'warning' }) } catch { return }
  try { await deletePositionEvent(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh(); timePanelRef.value?.reload() } catch { /* handled */ }
}

function onRefresh() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }

</style>
