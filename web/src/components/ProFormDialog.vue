<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

interface FormField {
  prop: string
  label: string
  type: 'input' | 'number' | 'textarea' | 'select' | 'date' | 'month' | 'switch' | 'name-select'
  placeholder?: string
  required?: boolean
  options?: { label: string; value: string | number }[]
  nameType?: string
  min?: number
  max?: number
  disabled?: boolean
}

interface Props {
  visible: boolean
  title: string
  formFields: FormField[]
  initialData?: Record<string, unknown>
  submitApi: (data: Record<string, unknown>) => Promise<void>
  width?: string
}

const props = withDefaults(defineProps<Props>(), {
  initialData: () => ({}),
  width: '600px'
})

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const formRef = ref()
const formData = ref<Record<string, unknown>>({})
const submitting = ref(false)
const isEdit = ref(false)

function initForm() {
  formData.value = {}
  isEdit.value = !!props.initialData && Object.keys(props.initialData).length > 0
  for (const field of props.formFields) {
    if (isEdit.value && props.initialData[field.prop] !== undefined) {
      formData.value[field.prop] = props.initialData[field.prop]
    } else {
      switch (field.type) {
        case 'switch':
          formData.value[field.prop] = false
          break
        case 'number':
          formData.value[field.prop] = undefined
          break
        default:
          formData.value[field.prop] = ''
      }
    }
  }
}

function handleClose() {
  emit('update:visible', false)
  formRef.value?.resetFields()
  formData.value = {}
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const data: Record<string, unknown> = {}
    for (const field of props.formFields) {
      const val = formData.value[field.prop]
      if (val !== '' && val !== undefined && val !== null) {
        data[field.prop] = val
      }
    }
    if (isEdit.value && props.initialData.id) {
      data.id = props.initialData.id
    }
    await props.submitApi(data)
    ElMessage.success(isEdit.value ? '修改成功' : '新增成功')
    emit('success')
    handleClose()
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

watch(() => props.visible, (val) => {
  if (val) {
    initForm()
  }
})
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="title"
    :width="width"
    :close-on-click-modal="false"
    @update:model-value="(val: boolean) => emit('update:visible', val)"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="formData" label-width="100px">
      <el-form-item
        v-for="field in formFields"
        :key="field.prop"
        :label="field.label"
        :prop="field.prop"
        :rules="field.required ? [{ required: true, message: `请输入${field.label}`, trigger: 'blur' }] : []"
      >
        <el-input
          v-if="field.type === 'input'"
          v-model="formData[field.prop]"
          :placeholder="field.placeholder || '请输入'"
          :disabled="field.disabled"
        />
        <el-input-number
          v-else-if="field.type === 'number'"
          v-model="formData[field.prop]"
          :placeholder="field.placeholder || '请输入'"
          :min="field.min"
          :max="field.max"
          :disabled="field.disabled"
          style="width: 100%"
        />
        <el-input
          v-else-if="field.type === 'textarea'"
          v-model="formData[field.prop]"
          type="textarea"
          :rows="3"
          :placeholder="field.placeholder || '请输入'"
          :disabled="field.disabled"
        />
        <el-select
          v-else-if="field.type === 'select'"
          v-model="formData[field.prop]"
          :placeholder="field.placeholder || '请选择'"
          :disabled="field.disabled"
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
          :disabled="field.disabled"
          style="width: 100%"
        />
        <el-date-picker
          v-else-if="field.type === 'month'"
          v-model="formData[field.prop]"
          type="month"
          :placeholder="field.placeholder || '选择月份'"
          value-format="YYYY-MM"
          :disabled="field.disabled"
          style="width: 100%"
        />
        <el-switch
          v-else-if="field.type === 'switch'"
          v-model="formData[field.prop]"
          :disabled="field.disabled"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>
