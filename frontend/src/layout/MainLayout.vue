<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background-color: #304156">
      <div style="height: 60px; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 18px; font-weight: bold; border-bottom: 1px solid rgba(255,255,255,0.1);">
        企业管理系统
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        style="border-right: none"
      >
        <template v-for="item in menuItems" :key="item.path">
          <el-menu-item v-if="!item.meta?.hidden" :index="item.path">
            <el-icon><component :is="item.meta?.icon" /></el-icon>
            <span>{{ item.meta?.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; border-bottom: 1px solid #e6e6e6; display: flex; align-items: center; justify-content: flex-end; padding: 0 20px;">
        <span style="margin-right: 16px;">{{ userStore.userInfo?.username }}</span>
        <el-button type="danger" text @click="handleLogout">退出登录</el-button>
      </el-header>
      <el-main style="background: #f0f2f5">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import * as authApi from '../api/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)

const menuItems = computed(() => {
  const routes = router.options.routes.find(r => r.path === '/')
  if (!routes?.children) return []
  return routes.children.filter(child => {
    if (!child.meta?.perm) return true
    return userStore.hasPermission(child.meta.perm as string)
  })
})

onMounted(async () => {
  if (userStore.token && !userStore.userInfo) {
    try {
      const res = await authApi.getUserInfo()
      userStore.userInfo = res.data.user
      userStore.permissions = res.data.permissions || []
    } catch {
      userStore.clearToken()
      router.push('/login')
    }
  }
})

function handleLogout() {
  userStore.clearToken()
  router.push('/login')
}
</script>
