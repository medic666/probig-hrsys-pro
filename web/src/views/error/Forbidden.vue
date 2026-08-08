<template>
  <div class="forbidden-page">
    <el-result icon="warning" title="403" sub-title="抱歉，您没有权限访问此页面">
      <template #extra>
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="handleLogout">退出登录</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'

const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

function goHome() {
  router.push('/')
}

function handleLogout() {
  userStore.clearUser()
  permissionStore.clearPermissions()
  router.push('/login')
}
</script>

<style lang="scss" scoped>
.forbidden-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}
</style>
