<template>
  <div class="main-layout">
    <el-container>
      <el-aside v-if="!isMobile" :width="isCollapsed ? '64px' : '220px'" class="layout-sidebar">
        <div class="sidebar-logo">
          <span v-if="!isCollapsed" class="logo-text">人事管理系统</span>
          <span v-else class="logo-text-small">HR</span>
        </div>
        <SideMenu :collapse="isCollapsed" />
      </el-aside>

      <el-container>
        <el-header class="layout-header">
          <div class="header-left">
            <el-icon v-if="isMobile" class="collapse-btn" @click="mobileDrawer = true">
              <Menu />
            </el-icon>
            <el-icon v-else class="collapse-btn" @click="toggleCollapse">
              <Fold v-if="!isCollapsed" />
              <Expand v-else />
            </el-icon>
            <span v-if="isMobile" class="header-title">{{ pageTitle }}</span>
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

    <el-drawer
      v-model="mobileDrawer"
      direction="ltr"
      size="min(80vw, 300px)"
      class="menu-drawer"
      :with-header="false"
    >
      <div class="drawer-logo">人事管理系统</div>
      <SideMenu @select="mobileDrawer = false" />
    </el-drawer>

    <ChangePasswordDialog v-model:visible="showPwdDialog" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { Menu, Fold, Expand, ArrowDown } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ChangePasswordDialog from '@/components/ChangePasswordDialog.vue'
import SideMenu from '@/components/SideMenu.vue'
import { useBreakpoint } from '@/composables/useBreakpoint'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()
const { isMobile } = useBreakpoint()
const isCollapsed = ref(false)
const showPwdDialog = ref(false)
const mobileDrawer = ref(false)

const pageTitle = computed(() => String(route.meta.title || ''))

// 移动抽屉内选择菜单后自动收起（路由变化即关闭）
watch(
  () => route.path,
  () => {
    mobileDrawer.value = false
  },
)

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      ElMessage.info('个人信息功能将在后续阶段实现')
      break
    case 'changePassword':
      showPwdDialog.value = true
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
    max-height: calc(100vh - 56px);
    overflow-y: auto;
    // 隐藏滚动条但保留滚轮/触屏滚动
    scrollbar-width: none;
    -ms-overflow-style: none;

    &::-webkit-scrollbar {
      display: none;
    }
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
    gap: 8px;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      color: #606266;

      &:hover {
        color: $primary-color;
      }
    }

    .header-title {
      font-size: 15px;
      font-weight: 600;
      color: #303133;
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

  @include mobile {
    padding: 0 12px;
  }
}

.layout-main {
  background: #f5f7fa;
  overflow-y: auto;

  @include mobile {
    padding: 12px;
  }
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
