<template>
  <el-dialog
    :model-value="visible"
    :title="title"
    :close-on-click-modal="false"
    width="600px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="100px"
      :disabled="formLoading"
    >
      <el-row :gutter="20">
        <el-col v-for="field in formFields" :key="field.prop" :span="field.span || 24">
          <el-form-item :label="field.label" :prop="field.prop">
            <el-input
              v-if="field.type === 'input'"
              v-model="formData[field.prop]"
              :placeholder="field.placeholder || '请输入'"
            />
            <el-input-number
              v-else-if="field.type === 'number'"
              v-model="formData[field.prop]"
              :placeholder="field.placeholder || '请输入'"
              :precision="field.precision || 2"
              style="width: 100%"
            />
            <el-select
              v-else-if="field.type === 'select'"
              v-model="formData[field.prop]"
              :placeholder="field.placeholder || '请选择'"
              style="width: 100%"
            >
              <el-option
                v-for="opt in field.options"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
            <el-date-picker
              v-else-if="field.type === 'date'"
              v-model="formData[field.prop]"
              type="date"
              :placeholder="field.placeholder || '选择日期'"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
            <el-input
              v-else-if="field.type === 'textarea'"
              v-model="formData[field.prop]"
              type="textarea"
              :rows="3"
              :placeholder="field.placeholder || '请输入'"
            />
            <el-switch
              v-else-if="field.type === 'switch'"
              v-model="formData[field.prop]"
            />
            <NameSelect
              v-else-if="field.type === 'person-select'"
              v-model="formData[field.prop]"
              :fetch-api="field.fetchApi!"
              :placeholder="field.placeholder || '请选择'"
            />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="formLoading" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'

export interface FormField {
  prop: string
  label: string
  type: 'input' | 'number' | 'select' | 'date' | 'textarea' | 'switch' | 'person-select'
  options?: { label: string; value: any }[]
  placeholder?: string
  span?: number
  precision?: number
  defaultValue?: any
  fetchApi?: (keyword?: string) => Promise<{ id: number; name: string }[]>
}

const props = defineProps<{
  visible: boolean
  title: string
  mode: 'add' | 'edit'
  formFields: FormField[]
  rules?: FormRules
  submitApi: (data: any) => Promise<any>
  editData?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const formLoading = ref(false)
const formData = reactive<Record<string, any>>({})

function initForm() {
  for (const key of Object.keys(formData)) {
    delete formData[key]
  }
  for (const field of props.formFields) {
    formData[field.prop] = field.defaultValue ?? ''
  }
}

function fillEditData() {
  if (!props.editData) return
  for (const field of props.formFields) {
    if (props.editData[field.prop] !== undefined) {
      formData[field.prop] = props.editData[field.prop]
    }
  }
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      initForm()
      if (props.mode === 'edit') {
        fillEditData()
      }
      setTimeout(() => {
        formRef.value?.clearValidate()
      }, 0)
    }
  },
)

function handleClose() {
  emit('update:visible', false)
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  formLoading.value = true
  try {
    const data: Record<string, any> = {}
    for (const key of Object.keys(formData)) {
      if (formData[key] !== '' && formData[key] !== null && formData[key] !== undefined) {
        data[key] = formData[key]
      }
    }
    for (const field of props.formFields) {
      if (field.type === 'person-select' && (data[field.prop] === undefined || data[field.prop] === null)) {
        data[field.prop] = 0
      }
    }
    await props.submitApi(data)
    ElMessage.success(props.mode === 'add' ? '添加成功' : '修改成功')
    emit('update:visible', false)
    emit('success')
  } catch {
    // error handled by request interceptor
  } finally {
    formLoading.value = false
  }
}
</script>
