<template>
  <el-dialog
    v-model="visible"
    title="工资调整事件"
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
      <el-form-item label="事件类型" prop="event_type">
        <el-select v-model="form.event_type" style="width: 100%">
          <el-option v-for="t in types" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="金额" prop="amount">
        <el-input-number v-model="form.amount" :controls="false" style="width: 100%" />
      </el-form-item>
      <el-form-item label="月份">
        <el-date-picker v-model="selectedMonth" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%" />
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
    <el-table :data="personnelList" border highlight-current-row max-height="400" @row-click="selectPerson">
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { createSalaryEvent, updateSalaryEvent } from '../../api/salary'
import { listPersonnel } from '../../api/personnel'
import type { SalaryEvent, PersonnelSnapshot } from '../../types'

const emit = defineEmits<{
  (e: 'success'): void
}>()

const formRef = ref()
const submitting = ref(false)
const visible = ref(false)
const mode = ref<'create' | 'edit'>('create')
const editingEvent = ref<SalaryEvent | null>(null)
const selectedMonth = ref<string>('')

const form = reactive({
  entity_id: 0,
  event_type: 'performance',
  amount: 0,
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

const types = [
  { value: 'performance', label: '业绩' },
  { value: 'reward_punishment', label: '奖惩' },
  { value: 'loan_deduction', label: '借款扣除' },
  { value: 'tax_deduction', label: '个税扣除' },
  { value: 'other', label: '其他' },
]

const rules = {
  entity_id: [
    { required: true, message: '请选择人员', trigger: 'change' },
    { validator: (_rule: any, value: number, cb: any) => value > 0 ? cb() : cb(new Error('请选择人员')), trigger: 'change' },
  ],
  event_type: [{ required: true }],
  amount: [{ required: true }],
}

function initForm() {
  form.entity_id = 0
  form.event_type = 'performance'
  form.amount = 0
  form.description = ''
  selectedMonth.value = ''
  selectedPersonName.value = ''
  personSearch.value = ''
  personPage.value = 1
  formRef.value?.clearValidate()

  if (mode.value === 'edit' && editingEvent.value) {
    const e = editingEvent.value
    form.entity_id = e.entity_id
    form.event_type = e.event_type
    form.amount = e.amount
    form.description = e.description
    if (e.period_start && e.period_start.length >= 7) {
      selectedMonth.value = e.period_start.slice(0, 7)
    }
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
    const periodStart = selectedMonth.value ? selectedMonth.value + '-01' : ''
    const periodEnd = selectedMonth.value ? getMonthEnd(selectedMonth.value) : ''

    const payload = {
      ...form,
      period_start: periodStart,
      period_end: periodEnd,
    }
    if (mode.value === 'edit' && editingEvent.value) {
      await updateSalaryEvent(editingEvent.value.id, payload)
    } else {
      await createSalaryEvent(payload)
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

function getMonthEnd(monthStr: string): string {
  const [y, m] = monthStr.split('-').map(Number)
  const d = new Date(y, m, 0)
  const day = d.getDate()
  return `${monthStr}-${String(day).padStart(2, '0')}`
}

function open(m: 'create' | 'edit', data?: SalaryEvent) {
  mode.value = m
  editingEvent.value = data ?? null
  initForm()
  visible.value = true
}

defineExpose({ open })
</script>
