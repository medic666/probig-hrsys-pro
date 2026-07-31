<template>
  <div class="page-container">
    <div class="page-header"><h2>职务事件管理</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        <el-button type="warning" link size="small" @click="attachFileId=row.id;attachVisible=true">附件</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? '新增职务事件' : '编辑职务事件'" width="640px" @close="handleDialogClose">
      <el-form :model="eventForm" label-width="110px">
        <el-form-item label="人员" required>
          <NameSelect v-model="eventForm.person_id" :fetch-api="fetchPersonOptions" placeholder="选择人员" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="eventForm.event_type" placeholder="选择事件类型" style="width:100%">
            <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="eventForm.event_type === '入职'" label="入职日期" required>
          <el-date-picker v-model="eventForm.entry_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item v-else-if="eventForm.event_type === '离职'" label="离职日期" required>
          <el-date-picker v-model="eventForm.leave_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item v-else label="生效日期" required>
          <el-date-picker v-model="eventForm.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="eventForm.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>

        <el-collapse>
          <el-collapse-item title="考勤/福利">
            <el-checkbox v-model="fieldFlags.attendance_group">考勤组</el-checkbox>
            <el-input v-if="fieldFlags.attendance_group" v-model="eventForm.attendance_group" size="small" placeholder="考勤组名称" style="margin:4px 0" />
            <el-checkbox v-model="fieldFlags.has_annual_leave" style="margin-left:16px">享有年假</el-checkbox>
            <el-checkbox v-model="fieldFlags.has_attendance_bonus" style="margin-left:16px">享有全勤奖</el-checkbox>
          </el-collapse-item>

          <el-collapse-item title="薪资基数">
            <el-checkbox v-model="fieldFlags.base_salary">基本工资</el-checkbox>
            <el-input-number v-if="fieldFlags.base_salary" v-model="eventForm.base_salary" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.performance_salary" style="margin-left:16px">绩效工资基数</el-checkbox>
            <el-input-number v-if="fieldFlags.performance_salary" v-model="eventForm.performance_salary" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.salary_days" style="margin-left:16px">计薪天数</el-checkbox>
            <el-input-number v-if="fieldFlags.salary_days" v-model="eventForm.salary_days" :min="0" size="small" style="margin:4px 0;width:100%" />
          </el-collapse-item>

          <el-collapse-item title="补贴">
            <el-checkbox v-model="fieldFlags.post_allowance">职位津贴</el-checkbox>
            <el-input-number v-if="fieldFlags.post_allowance" v-model="eventForm.post_allowance" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.meal_allowance" style="margin-left:16px">餐补</el-checkbox>
            <el-input-number v-if="fieldFlags.meal_allowance" v-model="eventForm.meal_allowance" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.housing_allowance" style="margin-left:16px">房补</el-checkbox>
            <el-input-number v-if="fieldFlags.housing_allowance" v-model="eventForm.housing_allowance" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.transport_allowance">交通补贴</el-checkbox>
            <el-input-number v-if="fieldFlags.transport_allowance" v-model="eventForm.transport_allowance" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.high_temp_allowance" style="margin-left:16px">高温补贴</el-checkbox>
            <el-input-number v-if="fieldFlags.high_temp_allowance" v-model="eventForm.high_temp_allowance" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
          </el-collapse-item>

          <el-collapse-item title="补偿/代扣">
            <el-checkbox v-model="fieldFlags.insurance_compensation">保险补偿</el-checkbox>
            <el-input-number v-if="fieldFlags.insurance_compensation" v-model="eventForm.insurance_compensation" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.fund_compensation" style="margin-left:16px">公积金补偿</el-checkbox>
            <el-input-number v-if="fieldFlags.fund_compensation" v-model="eventForm.fund_compensation" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.social_security_deduct">社保代扣</el-checkbox>
            <el-input-number v-if="fieldFlags.social_security_deduct" v-model="eventForm.social_security_deduct" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
            <el-checkbox v-model="fieldFlags.housing_fund_deduct" style="margin-left:16px">公积金代扣</el-checkbox>
            <el-input-number v-if="fieldFlags.housing_fund_deduct" v-model="eventForm.housing_fund_deduct" :min="0" :precision="2" size="small" style="margin:4px 0;width:100%" />
          </el-collapse-item>
        </el-collapse>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restoreEvent" :columns="trashColumns" @restored="onRefresh" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'position_event'" :target-id="attachFileId" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import NameSelect from '@/components/NameSelect.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getPositionEvents, getPositionEvent, createPositionEvent, updatePositionEvent, deletePositionEvent, restorePositionEvent, getDeletedPositionEvents, exportPositionEvents } from '@/api/position-event'
import { getAllPersons } from '@/api/person'

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editId = ref(0)
const saving = ref(false)
const trashVisible = ref(false)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)

const eventTypes = ['入职', '调薪调岗', '离职']
const eventForm = reactive<Record<string, any>>({ person_id: null, event_type: '', effective_date: '', remark: '' })
const fieldFlags = reactive<Record<string, boolean>>({})

const allFields = [
  'entry_date','leave_date','attendance_group','has_annual_leave','has_attendance_bonus',
  'base_salary','performance_salary','salary_days',
  'post_allowance','meal_allowance','housing_allowance','transport_allowance','high_temp_allowance',
  'insurance_compensation','fund_compensation','social_security_deduct','housing_fund_deduct',
]

const columns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'person_name', label: '人员', width: '100' },
  { prop: 'event_type', label: '事件类型', width: '100' },
  { prop: 'effective_date', label: '生效日期', width: '110' },
  { prop: 'remark', label: '备注', minWidth: '120' },
  { prop: 'changed_fields', label: '变更字段', formatter: (r: any) => (r.changed_fields || []).join('、') || '-' },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (r: any) => new Date(r.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOptions },
  { prop:'event_type', label:'事件类型', type:'select' as const, options: eventTypes.map(t => ({ label: t, value: t })) },
]

const actions = [
  { key: 'add', label: '新增事件', type: 'primary' as const },
  { key: 'trash', label: '回收站', type: 'default' as const },
  { key: 'export', label: '导出', type: 'default' as const },
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

function resetFlags() { for (const f of allFields) fieldFlags[f] = false }

function handleAction(key: string) {
  if (key === 'add') {
    dialogMode.value = 'add'; editId.value = 0; resetFlags()
    for (const k of allFields.concat(['person_id','event_type','effective_date','remark','entry_date','leave_date'])) eventForm[k] = ''
    eventForm.person_id = null
    dialogVisible.value = true
  } else if (key === 'trash') { trashVisible.value = true }
  else if (key === 'export') { handleExport() }
}

async function handleExport() {
  const data = await exportPositionEvents({}) as any
  const url = URL.createObjectURL(data as Blob)
  const a = document.createElement('a')
  a.href = url; a.download = 'position_events.xlsx'; a.click()
  URL.revokeObjectURL(url)
}

async function handleEdit(row: any) {
  dialogMode.value = 'edit'; editId.value = row.id; resetFlags()
  const detail = (await getPositionEvent(row.id)) as any
  eventForm.person_id = detail.person_id
  eventForm.event_type = detail.event_type || ''
  eventForm.remark = detail.remark || ''
  if (detail.event_type === '入职') {
    eventForm.entry_date = detail.entry_date || detail.effective_date || ''
  } else if (detail.event_type === '离职') {
    eventForm.leave_date = detail.leave_date || detail.effective_date || ''
  } else {
    eventForm.effective_date = detail.effective_date || ''
  }
  for (const f of allFields) {
    if (detail[f] !== null && detail[f] !== undefined && detail[f] !== '') {
      eventForm[f] = detail[f]
      fieldFlags[f] = true
    } else { eventForm[f] = '' }
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  saving.value = true
  try {
    const data: any = { person_id: eventForm.person_id, event_type: eventForm.event_type, remark: eventForm.remark }
    if (eventForm.event_type === '入职') {
      data.entry_date = eventForm.entry_date || undefined
      data.effective_date = eventForm.entry_date
    } else if (eventForm.event_type === '离职') {
      data.leave_date = eventForm.leave_date || undefined
      data.effective_date = eventForm.leave_date
    } else {
      data.effective_date = eventForm.effective_date
    }
    for (const f of allFields) {
      if (fieldFlags[f]) {
        if (f === 'has_annual_leave' || f === 'has_attendance_bonus') {
          data[f] = true
        } else {
          data[f] = eventForm[f] === '' ? null : eventForm[f]
        }
      }
    }
    if (dialogMode.value === 'add') { await createPositionEvent(data) }
    else { await updatePositionEvent(editId.value, data) }
    ElMessage.success(dialogMode.value === 'add' ? '创建成功' : '更新成功')
    dialogVisible.value = false
    tableRef.value?.refresh()
  } catch { /* handled */ } finally { saving.value = false }
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确认删除该事件？', '提示', { type: 'warning' }) } catch { return }
  try { await deletePositionEvent(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh() } catch { /* handled */ }
}

function handleDialogClose() { resetFlags() }
function onRefresh() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
</style>
