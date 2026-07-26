<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑' : '新增'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <slot :form="form" />
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { FormInstance } from 'element-plus'

const props = defineProps<{
  modelValue: boolean
  isEdit?: boolean
  rules?: Record<string, any>
  submit: (form: any) => Promise<any>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  success: []
}>()

const visible = ref(props.modelValue)
const form = ref<Record<string, any>>({})
const formRef = ref<FormInstance>()
const submitting = ref(false)

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    nextTick(() => {
      form.value = {}
      formRef.value?.resetFields()
    })
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

function handleClose() {
  visible.value = false
  form.value = {}
  formRef.value?.resetFields()
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitting.value = true
    await props.submit(form.value)
    visible.value = false
    emit('success')
  } catch (e) {
    // validation failed or submit error
  } finally {
    submitting.value = false
  }
}

function setForm(data: Record<string, any>) {
  form.value = { ...data }
}

defineExpose({ setForm, form })
</script>
