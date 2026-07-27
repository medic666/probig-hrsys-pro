<template>
  <div class="main-layout">
    <el-container>
      <el-aside :width="isCollapsed ? '64px' : '220px'" class="layout-sidebar">
        <div class="sidebar-logo">
          <span v-if="!isCollapsed" class="logo-text">人事管理系统</span>
          <span v-else class="logo-text-small">HR</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapsed"
          :collapse-transition="false"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409eff"
        >
          <template v-for="menu in displayMenus" :key="menu.path">
            <el-sub-menu v-if="menu.children && menu.children.length" :index="menu.path">
              <template #title>
                <el-icon><component :is="getIcon(menu.icon)" /></el-icon>
                <span>{{ menu.title }}</span>
              </template>
              <el-menu-item v-for="child in menu.children" :key="child.path" :index="child.path">
                <span>{{ child.title }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-menu-item v-else :index="menu.path">
              <el-icon><component :is="getIcon(menu.icon)" /></el-icon>
              <span>{{ menu.title }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <el-container>
        <el-header class="layout-header">
          <div class="header-left">
            <el-icon class="collapse-btn" @click="toggleCollapse">
              <Fold v-if="!isCollapsed" />
              <Expand v-else />
            </el-icon>
          </div>
          <div class="header-right">
            <el-dropdown trigger="click" @command="handleCommand">
              <span class="user-info">
                {{ userStore.userInfo?.name || '管理员' }}
                <el-icon><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                  <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
                  <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <el-main class="layout-main">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import {
  HomeFilled,
  Fold,
  Expand,
  ArrowDown,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()
const isCollapsed = ref(false)

const activeMenu = computed(() => route.path)

const iconMap: Record<string, any> = {
  HomeFilled,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
}

function getIcon(name: string) {
  return iconMap[name] || HomeFilled
}

const displayMenus = computed(() => {
  const storeMenus = permissionStore.menus
  if (storeMenus.length > 0) return storeMenus

  return [
    { path: '/home', title: '首页', icon: 'HomeFilled' },
  ]
})

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      ElMessage.info('个人信息功能将在后续阶段实现')
      break
    case 'changePassword':
      ElMessage.info('修改密码功能将在阶段四实现')
      break
    case 'logout':
      userStore.clearUser()
      permissionStore.clearPermissions()
      router.push('/login')
      break
  }
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.main-layout {
  height: 100vh;

  .el-container {
    height: 100%;
  }
}

.layout-sidebar {
  background-color: #304156;
  transition: width 0.3s;
  overflow: hidden;

  .sidebar-logo {
    height: $header-height;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 18px;
    font-weight: 600;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .logo-text-small {
      font-size: 20px;
      font-weight: 700;
    }
  }

  .el-menu {
    border-right: none;
  }
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);

  .header-left {
    display: flex;
    align-items: center;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      color: #606266;

      &:hover {
        color: $primary-color;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;

    .user-info {
      display: flex;
      align-items: center;
      gap: 4px;
      cursor: pointer;
      color: #606266;
      font-size: 14px;

      &:hover {
        color: $primary-color;
      }
    }
  }
}

.layout-main {
  background: #f5f7fa;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
