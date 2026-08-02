<template>
  <div class="page-container">
    <div class="page-header">
      <h2>调休事件流水</h2>
      <el-radio-group v-model="viewMode" size="small" style="margin-left:16px">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
    </div>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" :actions="actions" @action="handleAction" />
    </template>

    <template v-else>
      <TimeCardPanel
        ref="timePanelRef"
        :fetch-fn="(p: any) => getLILEvents(p)"
        date-field="event_date"
      >
        <template #day="{ date, items }">
          <DayCard :date="date" :events="items" @event-click="editDaily" />
        </template>
      </TimeCardPanel>
    </template>

    <el-dialog v-model="dialogVisible" title="新增调休事件" width="440px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="人员" required><NameSelect v-model="form.person_id" :fetch-api="fetchOpts" placeholder="选择人员" /></el-form-item>
        <el-form-item label="日期" required><el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.sub_type" style="width:100%">
            <el-option label="补班出勤" value="补班出勤" />
            <el-option label="调休" value="调休" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(小时)" required>
          <el-input-number v-model="form.hours" :min="0" :precision="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <DailyEditDialog v-model:visible="editVisible" :daily="editDailyRow" @saved="onEdited" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import NameSelect from '@/components/NameSelect.vue'
import TimeCardPanel from '@/components/cards/TimeCardPanel.vue'
import DayCard from '@/components/cards/DayCard.vue'
import DailyEditDialog from '@/components/attendance/DailyEditDialog.vue'
import { getLILEvents } from '@/api/annual-leave'
import { getAttendanceEvents } from '@/api/attendance'
import { createAttendanceEvent } from '@/api/attendance'
import { getAllPersons } from '@/api/person'
import { hoursToDays } from '@/utils'

const tableRef = ref()
const viewMode = ref<'cards' | 'list'>('cards')
const timePanelRef = ref()
const dialogVisible = ref(false)
const saving = ref(false)
const editVisible = ref(false)
const editDailyRow = ref<any>(null)
const form = reactive({ person_id: null as any, event_date: '', sub_type: '补班出勤', hours: 8, remark: '' })

const columns = [
  { prop: 'id', label: 'ID', width: '60' }, { prop: 'person_name', label: '人员', width: '80' },
  { prop: 'sub_type', label: '类型', width: '100' }, { prop: 'event_date', label: '日期', width: '110' },
  { prop: 'hours', label: '时长(天)', width: '90', formatter: (r: any) => hoursToDays(r.hours).toFixed(2) }, { prop: 'remark', label: '备注' },
]
const searchFields = [
  { prop: 'person_id', label: '人员', type: 'person-select' as const, fetchApi: fetchOpts },
  { prop: 'date', label: '时间范围', type: 'date-range' as const },
]
const actions = [{ key: 'add', label: '新增', type: 'primary' as const }]

async function fetchOpts(k?: string) { const l = await getAllPersons() as any[]; return k ? l.filter(p => p.name.includes(k)) : l }
async function fetchEvents(p: any) {
  return (await getLILEvents(p)) as any
}

function handleAction(k: string) {
  if (k === 'add') { Object.assign(form, { person_id: null, event_date: '', sub_type: '补班出勤', hours: 8, remark: '' }); dialogVisible.value = true }
}

async function handleSubmit() {
  if (!form.person_id || !form.event_date) { ElMessage.warning('请选择人员和日期'); return }
  saving.value = true
  try {
    await createAttendanceEvent({
      person_id: form.person_id,
      event_date: form.event_date,
      event_type: form.sub_type === '补班出勤' ? '出勤' : '休假',
      sub_type: form.sub_type,
      hours: form.hours,
      remark: form.remark || '',
    })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    tableRef.value?.refresh()
    timePanelRef.value?.reload()
  } catch { /* */ } finally { saving.value = false }
}

// 统一编辑：点击调休事件 → 打开该日考勤整日编辑（DailyEditDialog + 事务确认）
async function editDaily(item: any) {
  try {
    const dailyId = item.daily_id
    let row: any = null
    if (dailyId) {
      const d = (await getAttendanceEvents({ pageNum: 1, pageSize: 100 })) as any
      row = (d.list || []).find((x: any) => x.id === dailyId) || null
    }
    if (!row && item.person_id && item.event_date) {
      const d = (await getAttendanceEvents({ person_id: item.person_id, date_start: item.event_date, date_end: item.event_date, pageNum: 1, pageSize: 100 })) as any
      row = (d.list || [])[0] || null
    }
    if (!row) { ElMessage.warning('未找到当日考勤记录'); return }
    editDailyRow.value = row
    editVisible.value = true
  } catch { /* handled */ }
}

function onEdited() {
  timePanelRef.value?.reload()
  tableRef.value?.refresh()
}
</script>
<style scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; display: flex; align-items: center; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
</style>
