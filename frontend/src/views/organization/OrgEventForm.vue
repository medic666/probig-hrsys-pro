<template>
  <el-dialog
    :model-value="visible"
    :title="editEventId ? '编辑事件' : '新增事件'"
    width="600px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @open="initForm"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="130px">
      <el-form-item label="组织名称">
        <el-input v-if="!props.entityId" v-model="orgName" placeholder="输入组织名称" />
        <el-input v-else :model-value="orgName" disabled />
      </el-form-item>
      <el-form-item label="事件名称" prop="event_name">
        <el-input v-model="form.event_name" placeholder="填写事件名称" />
      </el-form-item>
      <el-form-item label="生效日期" prop="effective_date">
        <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
      </el-form-item>

      <el-divider content-position="left">选择要变更的属性</el-divider>
      <el-checkbox-group v-model="selectedFields" style="margin-bottom: 16px">
        <el-checkbox v-for="f in fieldOptions" :key="f.key" :label="f.key" :value="f.key">
          {{ f.label }}
        </el-checkbox>
      </el-checkbox-group>

      <template v-if="selectedFields.length > 0">
        <el-divider content-position="left">填写属性数值</el-divider>
        <el-form-item v-if="selectedFields.includes('credit_code')" label="统一社会信用代码">
          <el-input v-model="fieldValues.credit_code" />
        </el-form-item>
        <el-form-item v-if="selectedFields.includes('address')" label="地址">
          <el-input v-model="fieldValues.address" />
        </el-form-item>
        <el-form-item v-if="selectedFields.includes('phone')" label="联系电话">
          <el-input v-model="fieldValues.phone" />
        </el-form-item>
        <el-form-item v-if="selectedFields.includes('bank_name')" label="开户行">
          <el-input v-model="fieldValues.bank_name" />
        </el-form-item>
        <el-form-item v-if="selectedFields.includes('bank_account')" label="银行账号">
          <el-input v-model="fieldValues.bank_account" />
        </el-form-item>
      </template>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { createOrganizationEvent, updateOrganizationEvent, getOrganization } from '../../api/organization'
import type { OrganizationSnapshot } from '../../types'

const props = defineProps<{
  visible: boolean
  entityId: number
  editSnapshot: OrganizationSnapshot | null
  eventId?: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'success'): void
}>()

const formRef = ref()
const submitting = ref(false)
const selectedFields = ref<string[]>([])
const orgName = ref('')

const form = reactive({
  event_name: '',
  effective_date: '',
})

const fieldValues = reactive({
  credit_code: '',
  address: '',
  phone: '',
  bank_name: '',
  bank_account: '',
})

const rules = {
  effective_date: [{ required: true, message: '请选择生效日期', trigger: 'change' }],
}

const fieldOptions = [
  { key: 'credit_code', label: '信用代码' },
  { key: 'address', label: '地址' },
  { key: 'phone', label: '联系电话' },
  { key: 'bank_name', label: '开户行' },
  { key: 'bank_account', label: '银行账号' },
]

async function initForm() {
  form.event_name = ''
  form.effective_date = ''
  selectedFields.value = []
  fieldValues.credit_code = ''
  fieldValues.address = ''
  fieldValues.phone = ''
  fieldValues.bank_name = ''
  fieldValues.bank_account = ''

  const s = props.editSnapshot
  if (s) {
    orgName.value = s.company_name
    form.effective_date = new Date().toISOString().slice(0, 10)
    fieldValues.credit_code = s.credit_code
    fieldValues.address = s.address
    fieldValues.phone = s.phone
    fieldValues.bank_name = s.bank_name
    fieldValues.bank_account = s.bank_account
  } else if (props.entityId) {
    loadOrgName()
  } else {
    orgName.value = ''
    form.effective_date = new Date().toISOString().slice(0, 10)
    selectedFields.value = fieldOptions.map(f => f.key)
  }
}

async function loadOrgName() {
  try {
    const res = await getOrganization(props.entityId)
    orgName.value = res.data?.company_name || ''
  } catch {}
}

function buildChangedFields(): Record<string, any> {
  const cf: Record<string, any> = {}
  for (const f of selectedFields.value) {
    cf[f] = fieldValues[f as keyof typeof fieldValues]
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
      company_name: orgName.value,
      changed_fields: buildChangedFields(),
    }

    if (props.eventId) {
      await updateOrganizationEvent(props.eventId, payload)
    } else {
      await createOrganizationEvent(payload)
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
