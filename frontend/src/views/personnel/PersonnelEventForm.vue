<template>
  <el-dialog
    :model-value="visible"
    :title="editEventId ? '编辑事件' : '新增事件'"
    width="720px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="initForm"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="人员名称" :prop="!props.entityId ? 'person_name' : undefined">
            <el-input v-if="!props.entityId" v-model="personName" placeholder="输入人员姓名" />
            <el-input v-else :model-value="personName" disabled />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="事件名称" prop="event_name">
            <el-input v-model="form.event_name" placeholder="填写事件名称" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="生效日期" prop="effective_date">
            <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">选择要变更的属性</el-divider>
      <el-checkbox-group v-model="selectedFields" style="margin-bottom: 16px">
        <el-checkbox v-for="f in fieldOptions" :key="f.key" :label="f.key" :value="f.key">
          {{ f.label }}
        </el-checkbox>
      </el-checkbox-group>

      <template v-if="selectedFields.length > 0">
        <el-divider content-position="left">填写属性数值</el-divider>
        <el-row :gutter="16">
          <el-col :span="8" v-if="selectedFields.includes('base_salary')">
            <el-form-item label="基本工资">
              <el-input-number v-model="fieldValues.base_salary" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('performance_salary')">
            <el-form-item label="绩效工资">
              <el-input-number v-model="fieldValues.performance_salary" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('pay_days')">
            <el-form-item label="计薪天数">
              <el-input-number v-model="fieldValues.pay_days" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8" v-if="selectedFields.includes('position_allowance')">
            <el-form-item label="职位津贴">
              <el-input-number v-model="fieldValues.position_allowance" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('meal_subsidy')">
            <el-form-item label="餐补">
              <el-input-number v-model="fieldValues.meal_subsidy" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('housing_subsidy')">
            <el-form-item label="房补">
              <el-input-number v-model="fieldValues.housing_subsidy" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('transport_subsidy')">
            <el-form-item label="交通补贴">
              <el-input-number v-model="fieldValues.transport_subsidy" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('heat_subsidy')">
            <el-form-item label="高温补贴">
              <el-input-number v-model="fieldValues.heat_subsidy" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('insurance_compensation')">
            <el-form-item label="保险补偿">
              <el-input-number v-model="fieldValues.insurance_compensation" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('housing_fund_compensation')">
            <el-form-item label="公积金补偿">
              <el-input-number v-model="fieldValues.housing_fund_compensation" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('social_insurance_deduct')">
            <el-form-item label="社保代扣">
              <el-input-number v-model="fieldValues.social_insurance_deduct" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="selectedFields.includes('housing_fund_deduct')">
            <el-form-item label="公积金代扣">
              <el-input-number v-model="fieldValues.housing_fund_deduct" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12" v-if="selectedFields.includes('attendance_group')">
            <el-form-item label="考勤组">
              <el-input v-model="fieldValues.attendance_group" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('hire_date')">
            <el-form-item label="入职日期">
              <el-date-picker v-model="fieldValues.hire_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider v-if="hasExtendedFields" content-position="left">扩展信息</el-divider>
        <el-row :gutter="16" v-if="hasExtendedFields">
          <el-col :span="12" v-if="selectedFields.includes('ext_id_card')">
            <el-form-item label="身份证号">
              <el-input v-model="fieldValues.ext_id_card" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_gender')">
            <el-form-item label="性别">
              <el-select v-model="fieldValues.ext_gender" style="width: 100%">
                <el-option label="男" value="男" />
                <el-option label="女" value="女" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_birthday')">
            <el-form-item label="生日">
              <el-date-picker v-model="fieldValues.ext_birthday" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_phones')">
            <el-form-item label="电话">
              <el-input v-model="fieldValues.ext_phones" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_emails')">
            <el-form-item label="电子邮箱">
              <el-input v-model="fieldValues.ext_emails" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_ethnicity')">
            <el-form-item label="民族">
              <el-input v-model="fieldValues.ext_ethnicity" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_native_place')">
            <el-form-item label="籍贯">
              <el-input v-model="fieldValues.ext_native_place" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_address')">
            <el-form-item label="住址">
              <el-input v-model="fieldValues.ext_address" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_political_status')">
            <el-form-item label="政治面貌">
              <el-input v-model="fieldValues.ext_political_status" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_marital_status')">
            <el-form-item label="婚姻状态">
              <el-select v-model="fieldValues.ext_marital_status" style="width: 100%">
                <el-option label="未婚" value="未婚" />
                <el-option label="已婚" value="已婚" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_bank_accounts')">
            <el-form-item label="银行卡号">
              <el-input v-model="fieldValues.ext_bank_accounts" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="selectedFields.includes('ext_alias')">
            <el-form-item label="别名">
              <el-input v-model="fieldValues.ext_alias" />
            </el-form-item>
          </el-col>
        </el-row>
      </template>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { createPersonnelEvent, updatePersonnelEvent, getPersonnel } from '../../api/personnel'
import type { PersonnelSnapshot } from '../../types'

const props = defineProps<{
  visible: boolean
  entityId: number
  editSnapshot: PersonnelSnapshot | null
  eventId?: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'success'): void
}>()

const formRef = ref()
const submitting = ref(false)
const selectedFields = ref<string[]>([])
const personName = ref('')

const form = reactive({
  event_name: '',
  effective_date: '',
})

const fieldValues = reactive<Record<string, any>>({
  attendance_group: '',
  hire_date: null,
  base_salary: 0,
  performance_salary: 0,
  pay_days: 21.75,
  position_allowance: 0,
  meal_subsidy: 0,
  housing_subsidy: 0,
  transport_subsidy: 0,
  heat_subsidy: 0,
  insurance_compensation: 0,
  housing_fund_compensation: 0,
  social_insurance_deduct: 0,
  housing_fund_deduct: 0,
  ext_id_card: '',
  ext_gender: '',
  ext_birthday: '',
  ext_phones: '',
  ext_emails: '',
  ext_ethnicity: '',
  ext_native_place: '',
  ext_address: '',
  ext_political_status: '',
  ext_marital_status: '',
  ext_bank_accounts: '',
  ext_alias: '',
})

const rules = {
  effective_date: [{ required: true, message: '请选择生效日期', trigger: 'change' }],
}

const fieldOptions = [
  { key: 'base_salary', label: '基本工资' },
  { key: 'performance_salary', label: '绩效工资' },
  { key: 'pay_days', label: '计薪天数' },
  { key: 'position_allowance', label: '职位津贴' },
  { key: 'meal_subsidy', label: '餐补' },
  { key: 'housing_subsidy', label: '房补' },
  { key: 'transport_subsidy', label: '交通补贴' },
  { key: 'heat_subsidy', label: '高温补贴' },
  { key: 'insurance_compensation', label: '保险补偿' },
  { key: 'housing_fund_compensation', label: '公积金补偿' },
  { key: 'social_insurance_deduct', label: '社保代扣' },
  { key: 'housing_fund_deduct', label: '公积金代扣' },
  { key: 'attendance_group', label: '考勤组' },
  { key: 'hire_date', label: '入职日期' },
  { key: 'ext_id_card', label: '身份证号' },
  { key: 'ext_gender', label: '性别' },
  { key: 'ext_birthday', label: '生日' },
  { key: 'ext_phones', label: '电话' },
  { key: 'ext_emails', label: '电子邮箱' },
  { key: 'ext_ethnicity', label: '民族' },
  { key: 'ext_native_place', label: '籍贯' },
  { key: 'ext_address', label: '住址' },
  { key: 'ext_political_status', label: '政治面貌' },
  { key: 'ext_marital_status', label: '婚姻状态' },
  { key: 'ext_bank_accounts', label: '银行卡号' },
  { key: 'ext_alias', label: '别名' },
]

const hasExtendedFields = computed(() => {
  return selectedFields.value.some(f => f.startsWith('ext_'))
})

async function initForm() {
  form.event_name = ''
  form.effective_date = ''
  selectedFields.value = []
  resetFieldValues()

  const s = props.editSnapshot
  if (s) {
    personName.value = s.name
    form.effective_date = new Date().toISOString().slice(0, 10)
    fieldValues.base_salary = s.base_salary
    fieldValues.performance_salary = s.performance_salary
    fieldValues.pay_days = s.pay_days
    fieldValues.position_allowance = s.position_allowance
    fieldValues.meal_subsidy = s.meal_subsidy
    fieldValues.housing_subsidy = s.housing_subsidy
    fieldValues.transport_subsidy = s.transport_subsidy
    fieldValues.heat_subsidy = s.heat_subsidy
    fieldValues.insurance_compensation = s.insurance_compensation
    fieldValues.housing_fund_compensation = s.housing_fund_compensation
    fieldValues.social_insurance_deduct = s.social_insurance_deduct
    fieldValues.housing_fund_deduct = s.housing_fund_deduct
    fieldValues.attendance_group = s.attendance_group
    fieldValues.hire_date = s.hire_date
    const ext = s.extended_info || {}
    fieldValues.ext_id_card = ext.id_card || ''
    fieldValues.ext_gender = ext.gender || ''
    fieldValues.ext_birthday = ext.birthday || ''
    fieldValues.ext_phones = ext.phones || ''
    fieldValues.ext_emails = ext.emails || ''
    fieldValues.ext_ethnicity = ext.ethnicity || ''
    fieldValues.ext_native_place = ext.native_place || ''
    fieldValues.ext_address = ext.address || ''
    fieldValues.ext_political_status = ext.political_status || ''
    fieldValues.ext_marital_status = ext.marital_status || ''
    fieldValues.ext_bank_accounts = ext.bank_accounts || ''
    fieldValues.ext_alias = ext.alias || ''
  } else if (props.entityId) {
    loadPersonName()
  } else {
    personName.value = ''
    form.effective_date = new Date().toISOString().slice(0, 10)
    selectedFields.value = fieldOptions.map(f => f.key)
  }
}

async function loadPersonName() {
  try {
    const res = await getPersonnel(props.entityId)
    personName.value = res.data?.name || ''
  } catch {}
}

function resetFieldValues() {
  fieldValues.attendance_group = ''
  fieldValues.hire_date = null
  fieldValues.base_salary = 0
  fieldValues.performance_salary = 0
  fieldValues.pay_days = 21.75
  fieldValues.position_allowance = 0
  fieldValues.meal_subsidy = 0
  fieldValues.housing_subsidy = 0
  fieldValues.transport_subsidy = 0
  fieldValues.heat_subsidy = 0
  fieldValues.insurance_compensation = 0
  fieldValues.housing_fund_compensation = 0
  fieldValues.social_insurance_deduct = 0
  fieldValues.housing_fund_deduct = 0
  fieldValues.ext_id_card = ''
  fieldValues.ext_gender = ''
  fieldValues.ext_birthday = ''
  fieldValues.ext_phones = ''
  fieldValues.ext_emails = ''
  fieldValues.ext_ethnicity = ''
  fieldValues.ext_native_place = ''
  fieldValues.ext_address = ''
  fieldValues.ext_political_status = ''
  fieldValues.ext_marital_status = ''
  fieldValues.ext_bank_accounts = ''
  fieldValues.ext_alias = ''
}

function buildChangedFields(): Record<string, any> {
  const cf: Record<string, any> = {}
  for (const f of selectedFields.value) {
    if (f.startsWith('ext_')) {
      if (!cf.extended_info) cf.extended_info = {}
      const extKey = f.replace('ext_', '')
      cf.extended_info[extKey] = fieldValues[f]
    } else {
      cf[f] = fieldValues[f]
    }
  }
  return cf
}

async function handleSubmit() {
  if (selectedFields.value.length === 0) {
    ElMessage.warning('请至少选择一个要变更的属性')
    return
  }
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = {
      entity_id: props.entityId || 0,
      event_type: props.editSnapshot ? 'update' : 'create',
      event_name: form.event_name,
      effective_date: form.effective_date,
      name: personName.value,
      changed_fields: buildChangedFields(),
    }

    if (props.eventId) {
      await updatePersonnelEvent(props.eventId, payload)
    } else {
      await createPersonnelEvent(payload)
    }

    ElMessage.success(props.editSnapshot ? '事件更新成功' : '事件创建成功')
    emit('update:visible', false)
    emit('success')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}
</script>
