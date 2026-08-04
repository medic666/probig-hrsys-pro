<template>
  <el-form :model="eventForm" label-width="110px">
    <el-form-item label="人员" required>
      <NameSelect v-model="eventForm.person_id" :fetch-api="fetchPersonOptions" placeholder="选择人员" :disabled="isEdit" />
    </el-form-item>
    <el-form-item label="事件类型" required>
      <el-select v-model="eventForm.event_type" style="width:100%" :disabled="isEdit">
        <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
      </el-select>
    </el-form-item>

    <el-form-item v-if="eventForm.event_type === '入职'" label="入职日期" required>
      <el-date-picker v-model="eventForm.entry_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
    </el-form-item>
    <el-form-item v-else-if="eventForm.event_type === '调薪调岗'" label="生效日期" required>
      <el-date-picker v-model="eventForm.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
    </el-form-item>
    <el-form-item v-else-if="eventForm.event_type === '离职'" label="离职日期" required>
      <el-date-picker v-model="eventForm.leave_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
    </el-form-item>

    <el-form-item label="备注">
      <el-input v-model="eventForm.remark" type="textarea" :rows="2" placeholder="备注信息" />
    </el-form-item>

    <!-- 入职：全部字段一次展开填写 -->
    <template v-if="eventForm.event_type === '入职'">
      <el-divider content-position="left">岗位信息</el-divider>
      <el-row :gutter="12">
        <el-col :span="8">
          <el-form-item label="公司组">
            <el-select v-model="eventForm.company_id" clearable placeholder="选择公司组" style="width:100%">
              <el-option v-for="c in companyList" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="部门"><el-input v-model="eventForm.department" placeholder="部门名称" /></el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="职位"><el-input v-model="eventForm.position" placeholder="职位名称" /></el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">考勤/福利</el-divider>
      <el-row :gutter="12">
        <el-col :span="8">
          <el-form-item label="考勤组"><el-input v-model="eventForm.attendance_group" placeholder="考勤组名称" /></el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="享有年假"><el-switch v-model="eventForm.has_annual_leave" /></el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="享有全勤奖"><el-switch v-model="eventForm.has_attendance_bonus" /></el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">薪资参数</el-divider>
      <el-row :gutter="12">
        <el-col :span="8">
          <el-form-item label="基本工资"><el-input-number v-model="eventForm.base_salary" :min="0" :precision="2" style="width:100%" /></el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="绩效工资基数"><el-input-number v-model="eventForm.performance_salary" :min="0" :precision="2" style="width:100%" /></el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="计薪天数" required><el-input-number v-model="eventForm.salary_days" :min="0" :precision="1" style="width:100%" /></el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">补贴</el-divider>
      <el-row :gutter="12">
        <el-col :span="8"><el-form-item label="职位津贴"><el-input-number v-model="eventForm.post_allowance" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="餐补"><el-input-number v-model="eventForm.meal_allowance" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="房补"><el-input-number v-model="eventForm.housing_allowance" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="交通补贴"><el-input-number v-model="eventForm.transport_allowance" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="高温补贴"><el-input-number v-model="eventForm.high_temp_allowance" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
      </el-row>

      <el-divider content-position="left">补偿/代扣</el-divider>
      <el-row :gutter="12">
        <el-col :span="8"><el-form-item label="保险补偿"><el-input-number v-model="eventForm.insurance_compensation" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="公积金补偿"><el-input-number v-model="eventForm.fund_compensation" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="社保代扣"><el-input-number v-model="eventForm.social_security_deduct" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item label="公积金代扣"><el-input-number v-model="eventForm.housing_fund_deduct" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
      </el-row>
    </template>

    <!-- 调薪调岗：动态调整项块 -->
    <template v-else-if="eventForm.event_type === '调薪调岗'">
      <el-divider content-position="left">调整项</el-divider>
      <el-table :data="adjustItems" border size="small">
        <el-table-column label="字段" width="220">
          <template #default="{ row }">
            <el-select v-model="row.field" size="small" placeholder="选择调整字段" style="width:100%">
              <el-option-group v-for="g in adjustFieldGroups" :key="g.label" :label="g.label">
                <el-option v-for="f in g.fields" :key="f.key" :label="f.label" :value="f.key" />
              </el-option-group>
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="调整后值">
          <template #default="{ row }">
            <el-select v-if="fieldMeta[row.field]?.type === 'company'" v-model="row.value" size="small" placeholder="选择公司组" style="width:100%">
              <el-option v-for="c in companyList" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
            <el-switch v-else-if="fieldMeta[row.field]?.type === 'bool'" v-model="row.value" />
            <el-input-number
              v-else-if="fieldMeta[row.field]?.type === 'number'"
              v-model="row.value"
              :min="0"
              :precision="fieldMeta[row.field]?.precision || 2"
              size="small"
              style="width:100%"
            />
            <el-input v-else v-model="row.value" size="small" placeholder="调整后的值" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="adjustItems.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:8px" @click="addAdjustItem">+ 添加调整项</el-button>
    </template>

    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { createPositionEvent, updatePositionEvent, getPositionEvent } from '@/api/position-event'
import { getAllPersons } from '@/api/person'
import { getAllCompanies } from '@/api/company'

// 新增=编辑=查看统一表单：event 为 null 或 {id} 缺失 → 新增；{id} → 编辑（打开即回显全部原值）
const props = defineProps<{
  event: any
}>()

const emit = defineEmits<{
  (e: 'saved'): void
  (e: 'cancel'): void
}>()

const saving = ref(false)
const companyList = ref<{ id: number; name: string }[]>([])
const eventTypes = ['入职', '调薪调岗', '离职']

const isEdit = computed(() => props.event?.id != null)

const emptyForm = () => ({
  person_id: null,
  event_type: '',
  entry_date: '',
  effective_date: '',
  leave_date: '',
  remark: '',
  company_id: null,
  department: '',
  position: '',
  attendance_group: '',
  has_annual_leave: true,
  has_attendance_bonus: true,
  base_salary: null,
  performance_salary: null,
  salary_days: 26,
  post_allowance: null,
  meal_allowance: null,
  housing_allowance: null,
  transport_allowance: null,
  high_temp_allowance: null,
  insurance_compensation: null,
  fund_compensation: null,
  social_security_deduct: null,
  housing_fund_deduct: null,
})
const eventForm = reactive<any>(emptyForm())
const adjustItems = ref<any[]>([])

const entryFields = [
  'company_id', 'department', 'position', 'attendance_group',
  'has_annual_leave', 'has_attendance_bonus',
  'base_salary', 'performance_salary', 'salary_days',
  'post_allowance', 'meal_allowance', 'housing_allowance',
  'transport_allowance', 'high_temp_allowance',
  'insurance_compensation', 'fund_compensation',
  'social_security_deduct', 'housing_fund_deduct',
]

const adjustFieldGroups = [
  {
    label: '岗位信息',
    fields: [
      { key: 'company_id', label: '公司组', type: 'company' },
      { key: 'department', label: '部门', type: 'text' },
      { key: 'position', label: '职位', type: 'text' },
    ],
  },
  {
    label: '考勤/福利',
    fields: [
      { key: 'attendance_group', label: '考勤组', type: 'text' },
      { key: 'has_annual_leave', label: '享有年假', type: 'bool' },
      { key: 'has_attendance_bonus', label: '享有全勤奖', type: 'bool' },
    ],
  },
  {
    label: '薪资参数',
    fields: [
      { key: 'base_salary', label: '基本工资', type: 'number', precision: 2 },
      { key: 'performance_salary', label: '绩效工资基数', type: 'number', precision: 2 },
      { key: 'salary_days', label: '计薪天数', type: 'number', precision: 1 },
    ],
  },
  {
    label: '补贴',
    fields: [
      { key: 'post_allowance', label: '职位津贴', type: 'number', precision: 2 },
      { key: 'meal_allowance', label: '餐补', type: 'number', precision: 2 },
      { key: 'housing_allowance', label: '房补', type: 'number', precision: 2 },
      { key: 'transport_allowance', label: '交通补贴', type: 'number', precision: 2 },
      { key: 'high_temp_allowance', label: '高温补贴', type: 'number', precision: 2 },
    ],
  },
  {
    label: '补偿/代扣',
    fields: [
      { key: 'insurance_compensation', label: '保险补偿', type: 'number', precision: 2 },
      { key: 'fund_compensation', label: '公积金补偿', type: 'number', precision: 2 },
      { key: 'social_security_deduct', label: '社保代扣', type: 'number', precision: 2 },
      { key: 'housing_fund_deduct', label: '公积金代扣', type: 'number', precision: 2 },
    ],
  },
]

const fieldMeta = computed(() => {
  const map: Record<string, any> = {}
  for (const g of adjustFieldGroups) {
    for (const f of g.fields) map[f.key] = f
  }
  return map
})

onMounted(async () => {
  companyList.value = (await getAllCompanies()) as { id: number; name: string }[] || []
  if (isEdit.value) {
    try {
      // 编辑=查看：取完整事件记录回显全部原值（列表行不含薪资字段）
      const detail = (await getPositionEvent(props.event.id)) as any
      eventForm.person_id = detail.person_id
      eventForm.event_type = detail.event_type || ''
      eventForm.remark = detail.remark || ''
      if (detail.event_type === '入职') {
        eventForm.entry_date = detail.entry_date || detail.effective_date || ''
        for (const key of entryFields) {
          eventForm[key] = detail[key] ?? (key.startsWith('has_') ? true : null)
        }
      } else if (detail.event_type === '离职') {
        eventForm.leave_date = detail.leave_date || detail.effective_date || ''
      } else {
        eventForm.effective_date = detail.effective_date || ''
        for (const g of adjustFieldGroups) {
          for (const f of g.fields) {
            const val = detail[f.key]
            if (val !== null && val !== undefined && val !== '') {
              adjustItems.value.push({ field: f.key, value: val })
            }
          }
        }
      }
    } catch { /* handled */ }
  }
})

function addAdjustItem() {
  adjustItems.value.push({ field: '', value: null })
}

async function fetchPersonOptions(keyword?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[]
  if (!keyword) return list
  return list.filter(p => p.name.includes(keyword))
}

async function handleSubmit() {
  if (!eventForm.person_id) { ElMessage.warning('请选择人员'); return }
  if (!eventForm.event_type) { ElMessage.warning('请选择事件类型'); return }

  const data: any = { person_id: eventForm.person_id, event_type: eventForm.event_type, remark: eventForm.remark || '' }

  if (eventForm.event_type === '入职') {
    if (!eventForm.entry_date) { ElMessage.warning('请选择入职日期'); return }
    if (eventForm.salary_days == null || eventForm.salary_days <= 0) { ElMessage.warning('请填写计薪天数，且必须大于 0'); return }
    data.entry_date = eventForm.entry_date
    data.effective_date = eventForm.entry_date
    for (const key of entryFields) {
      const val = eventForm[key]
      if (val === null || val === undefined || val === '') continue
      data[key] = val
    }
  } else if (eventForm.event_type === '离职') {
    if (!eventForm.leave_date) { ElMessage.warning('请选择离职日期'); return }
    data.leave_date = eventForm.leave_date
    data.effective_date = eventForm.leave_date
  } else {
    if (!eventForm.effective_date) { ElMessage.warning('请选择生效日期'); return }
    data.effective_date = eventForm.effective_date
    for (const item of adjustItems.value) {
      if (!item.field) continue
      const meta = fieldMeta.value[item.field]
      if (!meta) continue
      if (meta.type === 'text' && !item.value) continue
      if (item.field === 'salary_days' && (item.value == null || item.value <= 0)) { ElMessage.warning('计薪天数必须大于 0'); return }
      data[item.field] = item.value
    }
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await updatePositionEvent(props.event.id, data)
    } else {
      await createPositionEvent(data)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    emit('saved')
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
