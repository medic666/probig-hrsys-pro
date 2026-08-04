<template>
  <div class="page-container">
    <PageHeader title="年假事件流水">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button type="primary" size="small" @click="handleAction('add')">新增</el-button>
      <template #right>
        <el-button size="small" @click="handleAction('trash')">回收站</el-button>
        <el-button size="small" @click="handleAction('export')">导出</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button v-if="row.source_type === 'manual'" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button v-if="row.source_type === 'manual'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          <el-button v-if="row.source_type === 'attendance'" type="primary" link size="small" @click="handleEdit(row)">查看</el-button>
          <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
        </template>
      </ProTable>
    </template>
    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :url-driven="true"
        :fetch-fn="(p: any) => getAnnualLeaveEvents(p)"
        date-field="effective_date"
        :aggregate="'year'"
        :person-dot-map="dotMap"
        badge-position="meta"
      >
        <template #person-badge="{ person }">
          <div class="balance-line" :class="{ 'is-zero': !(balanceMap[person.id] ?? 0) }">
            年假 {{ hoursToDays(balanceMap[person.id] ?? 0).toFixed(2) }} 天
          </div>
        </template>
        <template #day="{ items }">
          <div class="ev-grid">
            <AnnualLeaveEventCard v-for="e in items" :key="e.id" :event="e" @edit="handleEdit" />
          </div>
        </template>
      </TimeCardPanel>
    </template>

    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'annual_leave_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import AnnualLeaveEventCard from '@/components/annual-leave/AnnualLeaveEventCard.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import { getAnnualLeaveEvents, getAnnualLeaveEventBadges, deleteAnnualLeaveEvent, restoreAnnualLeaveEvent, getDeletedAnnualLeaveEvents, exportAnnualLeaveEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'
import { downloadBlob } from '@/utils/download'

import { usePageView } from '@/composables/usePageView'
import { useBadges } from '@/composables/useBadges'

const router = useRouter()
const tableRef = ref()
const { viewMode, isList } = usePageView('cards')
const timePanelRef = ref()
const tv = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)
// 徽章映射：周年月且上月未结转 → orange，否则 green；年假余额映射（meta 位徽章）
const { dotMap, balanceMap, loadDots, loadBalances } = useBadges()

const columns = [
  { prop:'person_name', label:'人员', width:'80' },
  { prop:'event_type', label:'类型', width:'130' },
  { prop:'hours', label:'变动时长(天)', width:'100', formatter:(r:any)=>hoursToDays(r.hours).toFixed(2) },
  { prop:'source_type', label:'来源', width:'110', formatter:(r:any)=>({manual:'人工录入',system_period:'系统周期',attendance:'考勤休假'}[r.source_type]||r.source_type) },
  { prop:'effective_date', label:'生效日期', width:'110' },
  { prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi:fetchPersonOpts },
  { prop:'date', label:'时间范围', type:'date-range' as const, startKey: 'date_start', endKey: 'date_end' },
]
const tc = [{ prop:'event_type', label:'类型' },{ prop:'effective_date', label:'生效日期' }]

async function fetchPersonOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchEvents(p:any){ return (await getAnnualLeaveEvents(p)) as any }
async function fd(p:any){ return (await getDeletedAnnualLeaveEvents(p)) as any }
async function rst(id:number){ return restoreAnnualLeaveEvent(id) }

onMounted(async () => {
  await Promise.all([
    loadDots('annual-leave-badges', getAnnualLeaveEventBadges),
    loadBalances('al-balances', '/annual-leave-balances'),
  ])
})

function handleAction(k:string){
  if(k==='add'){ router.push('/annual-leave-events/create') }
  else if(k==='trash') tv.value=true
  else if(k==='export') handleExport()
}

async function handleExport() {
  const data = await exportAnnualLeaveEvents(tableRef.value?.getSearchParams() || {})
  downloadBlob(data)
}

// 新增=编辑=查看统一跳业务逻辑页：考勤来源 → 该日考勤整日页；manual → 年假事件页
function handleEdit(r: any) {
  if (r.source_type === 'attendance') {
    router.push(`/attendance-events/${r.daily_id}`)
  } else {
    router.push(`/annual-leave-events/${r.id}`)
  }
}

async function handleDelete(r:any){
  try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return}
  try{await deleteAnnualLeaveEvent(r.id);ElMessage.success('已删除');tableRef.value?.refresh();timePanelRef.value?.reload()}catch{ /* */ }
}
function onR(){ tableRef.value?.refresh() }
</script>
<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.page-container{padding:0;background:transparent}

.ev-grid {
  @include card-grid(260px);
}

.balance-line { font-size: 13px; color: #409eff; font-weight: 600; line-height: 20px; }
.balance-line.is-zero { color: #c0c4cc; font-weight: 400; }
</style>
