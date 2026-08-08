<template>
  <el-dialog
    :model-value="visible"
    title="修改密码"
    width="420px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="!force"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item v-if="!force" label="旧密码" prop="oldPassword">
        <el-input v-model="form.oldPassword" type="password" show-password />
      </el-form-item>
      <el-form-item label="新密码" prop="newPassword">
        <el-input v-model="form.newPassword" type="password" show-password />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input v-model="form.confirmPassword" type="password" show-password />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="force" @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { changePassword } from '@/api/auth'

const props = withDefaults(
  defineProps<{
    visible: boolean
    force?: boolean
  }>(),
  { force: false },
)

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== form.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const validateNotDefault = (_rule: any, value: string, callback: any) => {
  if (value === 'admin123' || value === '123456') {
    callback(new Error('新密码不能与默认密码相同'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  oldPassword: props.force ? [] : [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 4, message: '密码长度不能小于4位', trigger: 'blur' },
    { validator: validateNotDefault, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      form.oldPassword = ''
      form.newPassword = ''
      form.confirmPassword = ''
      setTimeout(() => formRef.value?.clearValidate(), 0)
    }
  },
)

function handleClose() {
  if (props.force) return
  emit('update:visible', false)
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await changePassword(form.oldPassword || 'admin123', form.newPassword)
    ElMessage.success('密码修改成功')
    emit('update:visible', false)
    emit('success')
  } catch { /* error handled by interceptor */ } finally {
    loading.value = false
  }
}
</script>
