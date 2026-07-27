<template>
  <div class="login-page">
    <div class="login-card">
      <h2>企业人事与行政管理系统</h2>
      <el-form ref="formRef" :model="form" :rules="rules" size="large">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" show-password prefix-icon="Lock" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin"> 登录 </el-button>
        </el-form-item>
      </el-form>
    </div>
    <ChangePasswordDialog v-model:visible="showPwdDialog" :force="true" @success="onPwdChanged" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { login } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import ChangePasswordDialog from '@/components/ChangePasswordDialog.vue'

const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const showPwdDialog = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const data = await login({ username: form.username, password: form.password })
    userStore.setToken(data.token)
    userStore.setUserInfo(data.user)
    userStore.isFirstLogin = data.is_first_login
    permissionStore.setPermissions(data.permissions)
    permissionStore.setMenus(data.menus)

    if (data.is_first_login) {
      showPwdDialog.value = true
    } else {
      ElMessage.success('登录成功')
      router.push('/')
    }
  } catch { /* error handled by interceptor */ } finally {
    loading.value = false
  }
}

function onPwdChanged() {
  router.push('/')
}
</script>

<style lang="scss" scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);

  h2 {
    text-align: center;
    margin-bottom: 32px;
    color: #303133;
    font-size: 20px;
  }

  .login-btn {
    width: 100%;
  }
}
</style>
