<template>
  <div class="login-page">
    <div class="login-card">
      <h1 style="text-align:center;margin-bottom:24px;">企业人事管理系统</h1>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="0">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="handleLogin">登录</el-button>
        </el-form-item>
      </el-form>
    </div>
    <el-dialog v-model="showChangePwd" title="首次登录请修改密码" :close-on-click-modal="false" :show-close="false" width="400px">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules">
        <el-form-item prop="newPassword">
          <el-input v-model="pwdForm.newPassword" type="password" placeholder="新密码" show-password />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input v-model="pwdForm.confirmPassword" type="password" placeholder="确认密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" :loading="changingPwd" @click="handleChangePwd">确认修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '@/api/rbac'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'

const router = useRouter()
const userStore = useUserStore()
const permStore = usePermissionStore()

const form = reactive({ username: 'admin', password: 'admin123' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}
const loading = ref(false)
const formRef = ref()

const showChangePwd = ref(false)
const pwdForm = reactive({ newPassword: '', confirmPassword: '' })
const pwdRules = {
  newPassword: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, cb: any) => {
        if (value !== pwdForm.newPassword) cb(new Error('两次密码不一致'))
        else cb()
      },
      trigger: 'blur',
    },
  ],
}
const changingPwd = ref(false)
const pwdFormRef = ref()

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const data = await login(form)
    userStore.setToken(data.token)
    const userInfo = await userStore.fetchUserInfo()
    if (userInfo) {
      permStore.setPermissions(userInfo.permissions || [])
      if (userInfo.is_first_login) {
        showChangePwd.value = true
      } else {
        router.push('/')
      }
    }
  } catch (e) {
    // error shown by interceptor
  } finally {
    loading.value = false
  }
}

async function handleChangePwd() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (pwdForm.newPassword === 'admin123') {
    ElMessage.error('新密码不能与默认密码相同')
    return
  }
  changingPwd.value = true
  try {
    await userStore.changePassword('admin123', pwdForm.newPassword)
    ElMessage.success('密码修改成功')
    showChangePwd.value = false
    router.push('/')
  } catch (e) {
    // error shown by interceptor
  } finally {
    changingPwd.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  background: #fff;
  border-radius: 8px;
  padding: 40px;
  width: 400px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}
</style>
