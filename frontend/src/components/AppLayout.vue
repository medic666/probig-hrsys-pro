<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background: #1f2d3d">
      <div class="logo">
        <span>企业管理系统</span>
      </div>
      <el-menu
        :default-active="route.path"
        background-color="#1f2d3d"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>工作台</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('personnel')" index="/personnel">
          <el-icon><User /></el-icon>
          <span>人员管理</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('organization')" index="/organization">
          <el-icon><OfficeBuilding /></el-icon>
          <span>组织管理</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('attendance')" index="/attendance">
          <el-icon><Calendar /></el-icon>
          <span>假勤管理</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('salary')" index="/salary">
          <el-icon><Money /></el-icon>
          <span>工资管理</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('file')" index="/files">
          <el-icon><FolderOpened /></el-icon>
          <span>文件管理</span>
        </el-menu-item>
        <el-menu-item v-if="auth.hasAnyPermission('audit')" index="/audit">
          <el-icon><DocumentChecked /></el-icon>
          <span>操作审计</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="display: flex; align-items: center; justify-content: space-between; background: #fff; border-bottom: 1px solid #dcdfe6">
        <span style="font-size: 16px; font-weight: 500">{{ pageTitle }}</span>
        <div style="display: flex; align-items: center; gap: 12px">
          <span>{{ auth.user?.real_name }}</span>
          <el-button text @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Monitor, User, OfficeBuilding, Calendar, Money, FolderOpened, DocumentChecked } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '工作台',
    '/personnel': '人员管理',
    '/organization': '组织管理',
    '/attendance': '假勤管理',
    '/salary': '工资管理',
    '/files': '文件管理',
    '/audit': '操作审计',
  }
  return map[route.path] || (route.matched[route.matched.length - 1]?.meta?.title as string) || ''
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  letter-spacing: 2px;
}
</style>
