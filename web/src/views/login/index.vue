<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'

const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

const formRef = ref()
const loading = ref(false)
const changePwdVisible = ref(false)
const changePwdLoading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const changePwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const changePwdFormRef = ref()

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const changePwdRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (value === changePwdForm.oldPassword) {
          callback(new Error('新密码不能与旧密码相同'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (value !== changePwdForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await userStore.login(loginForm.username, loginForm.password)
    if (res.is_first_login) {
      changePwdForm.oldPassword = loginForm.password
      changePwdVisible.value = true
      return
    }
    await doAfterLogin()
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function doAfterLogin() {
  await userStore.fetchUserInfo()
  await permissionStore.fetchPermissions()
  router.replace('/')
}

async function handleChangePassword() {
  const valid = await changePwdFormRef.value?.validate().catch(() => false)
  if (!valid) return

  changePwdLoading.value = true
  try {
    await userStore.changePassword(changePwdForm.oldPassword, changePwdForm.newPassword)
    ElMessage.success('密码修改成功')
    changePwdVisible.value = false
    await doAfterLogin()
  } catch {
    // error handled by interceptor
  } finally {
    changePwdLoading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <h2 class="login-title">企业人事与行政管理系统</h2>
      <el-form ref="formRef" :model="loginForm" :rules="loginRules" size="large">
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            prefix-icon="User"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" style="width: 100%" @click="handleLogin">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-dialog
      v-model="changePwdVisible"
      title="首次登录，请修改密码"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      width="450px"
    >
      <el-form ref="changePwdFormRef" :model="changePwdForm" :rules="changePwdRules" label-width="100px">
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="changePwdForm.newPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="changePwdForm.confirmPassword" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" :loading="changePwdLoading" @click="handleChangePassword">
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.login-title {
  text-align: center;
  font-size: 20px;
  color: #303133;
  margin-bottom: 32px;
}
</style>
