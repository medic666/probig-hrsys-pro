<template>
  <el-dialog
    v-model="visible"
    title="假勤事件"
    width="500px"
    :close-on-click-modal="false"
    @open="initForm"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="人员" prop="entity_id">
        <div style="display: flex; gap: 8px; align-items: center">
          <span v-if="selectedPersonName" style="flex: 1">{{ selectedPersonName }}</span>
          <span v-else style="flex: 1; color: #c0c4cc">请选择人员</span>
          <el-button size="small" @click="personDialogVisible = true">选择人员</el-button>
        </div>
      </el-form-item>
      <el-form-item label="事件类别" prop="event_category">
        <el-select v-model="form.event_category" style="width: 100%" @change="onCategoryChange">
          <el-option v-for="cat in categories" :key="cat.value" :label="cat.label" :value="cat.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="事件子类" prop="event_subtype">
        <el-select v-model="form.event_subtype" style="width: 100%">
          <el-option v-for="sub in currentSubtypes" :key="sub.value" :label="sub.label" :value="sub.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="日期" prop="event_date">
        <el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
      </el-form-item>
      <el-form-item label="天数/次数">
        <el-input-number v-model="form.duration_days" :min="0" :step="0.5" :controls="false" style="width: 100%" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.description" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="personDialogVisible" title="选择人员" width="500px">
    <div style="margin-bottom: 12px">
      <el-input v-model="personSearch" placeholder="搜索人员" clearable @input="fetchPersonnel" />
    </div>
    <el-table :data="personnelList" border highlight-current-row max-height="400" v-loading="personLoading" @row-click="selectPerson">
      <el-table-column prop="name" label="姓名" />
      <el-table-column label="考勤组" width="100">
        <template #default="{ row }">{{ row.attendance_group || '-' }}</template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="personPage"
      :page-size="20"
      :total="personTotal"
      layout="prev, pager, next"
      small
      style="margin-top: 12px; justify-content: flex-end"
      @current-change="fetchPersonnel"
    />
    <template #footer>
      <el-button @click="personDialogVisible = false">取消</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { createAttendanceEvent, updateAttendanceEvent } from '../../api/attendance'
import { listPersonnel } from '../../api/personnel'
import type { AttendanceEvent, PersonnelSnapshot } from '../../types'

const emit = defineEmits<{
  (e: 'success'): void
}>()

const formRef = ref()
const submitting = ref(false)
const visible = ref(false)
const mode = ref<'create' | 'edit'>('create')
const editingEvent = ref<AttendanceEvent | null>(null)

const form = reactive({
  entity_id: 0,
  event_category: 'attendance',
  event_subtype: 'normal',
  event_date: '',
  duration_days: 1,
  description: '',
})

const selectedPersonName = ref('')
const personDialogVisible = ref(false)
const personSearch = ref('')
const personnelList = ref<PersonnelSnapshot[]>([])
const personPage = ref(1)
const personTotal = ref(0)
const personLoading = ref(false)

onMounted(() => fetchPersonnel())

async function fetchPersonnel() {
  personLoading.value = true
  try {
    const res = await listPersonnel({ search: personSearch.value, page: personPage.value, page_size: 20 })
    personnelList.value = res.data.list
    personTotal.value = res.data.total
  } finally {
    personLoading.value = false
  }
}

function selectPerson(row: PersonnelSnapshot) {
  form.entity_id = row.entity_id
  selectedPersonName.value = row.name
  personDialogVisible.value = false
}

const categories = [
  { value: 'attendance', label: '出勤' },
  { value: 'leave', label: '休假' },
  { value: 'overtime', label: '加班' },
  { value: 'discipline', label: '违纪' },
  { value: 'annual_adjustment', label: '年假调整' },
]

const subtypeMap: Record<string, { value: string; label: string }[]> = {
  attendance: [{ value: 'normal', label: '普通出勤' }, { value: 'makeup', label: '补班出勤' }],
  leave: [
    { value: 'lieu', label: '调休' }, { value: 'personal', label: '事假' },
    { value: 'sick', label: '病假' }, { value: 'annual', label: '年假' },
    { value: 'statutory', label: '法定假' }, { value: 'welfare', label: '福利假' },
  ],
  overtime: [{ value: 'workday', label: '工作日加班' }, { value: 'holiday', label: '节假日加班' }],
  discipline: [{ value: 'missing_card', label: '缺卡' }, { value: 'late', label: '迟到' }, { value: 'early', label: '早退' }],
  annual_adjustment: [{ value: 'allocation', label: '年假配发' }, { value: 'carryover', label: '年假结转' }],
}

const currentSubtypes = computed(() => subtypeMap[form.event_category] || [])

const rules = {
  entity_id: [
    { required: true, message: '请选择人员', trigger: 'change' },
    { validator: (_rule: any, value: number, cb: any) => value > 0 ? cb() : cb(new Error('请选择人员')), trigger: 'change' },
  ],
  event_category: [{ required: true }],
  event_subtype: [{ required: true }],
  event_date: [{ required: true, message: '请选择日期' }],
}

function onCategoryChange() {
  form.event_subtype = currentSubtypes.value[0]?.value || ''
}

function initForm() {
  form.entity_id = 0
  form.event_category = 'attendance'
  form.event_subtype = 'normal'
  form.event_date = ''
  form.duration_days = 1
  form.description = ''
  selectedPersonName.value = ''
  personSearch.value = ''
  personPage.value = 1
  formRef.value?.clearValidate()

  if (mode.value === 'edit' && editingEvent.value) {
    const e = editingEvent.value
    form.entity_id = e.entity_id
    form.event_category = e.event_category
    form.event_subtype = e.event_subtype
    form.event_date = e.event_date
    form.duration_days = e.duration_days
    form.description = e.description
    if (e.entity_name) {
      selectedPersonName.value = e.entity_name
    } else {
      loadPersonName(e.entity_id)
    }
  }
}

async function loadPersonName(entityId: number) {
  try {
    const res = await listPersonnel({ search: '', page: 1, page_size: 100 })
    const found = res.data.list.find((p: PersonnelSnapshot) => p.entity_id === entityId)
    if (found) selectedPersonName.value = found.name
  } catch {}
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = { ...form }
    if (mode.value === 'edit' && editingEvent.value) {
      await updateAttendanceEvent(editingEvent.value.id, payload)
    } else {
      await createAttendanceEvent(payload)
    }
    ElMessage.success(mode.value === 'edit' ? '更新成功' : '创建成功')
    visible.value = false
    emit('success')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function open(m: 'create' | 'edit', data?: AttendanceEvent) {
  mode.value = m
  editingEvent.value = data ?? null
  initForm()
  visible.value = true
}

defineExpose({ open })
</script>
