<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Fold, Expand, ArrowDown } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import type { MenuItem } from '@/stores/permission'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

const isCollapse = ref(false)
const changePwdVisible = ref(false)
const changePwdLoading = ref(false)

const newPassword = ref('')
const confirmPassword = ref('')

const menuList = computed(() => permissionStore.menus)

const activeMenu = computed(() => {
  return route.path
})

function toggleCollapse() {
  isCollapse.value = !isCollapse.value
}

function handleMenuSelect(path: string) {
  router.push(path)
}

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      break
    case 'changePassword':
      changePwdVisible.value = true
      break
    case 'logout':
      handleLogout()
      break
  }
}

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  userStore.logout()
  permissionStore.clearPermissions()
  router.replace('/login')
}

async function submitChangePassword() {
  if (!newPassword.value || newPassword.value.length < 6) {
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    return
  }
  changePwdLoading.value = true
  try {
    await userStore.changePassword('', newPassword.value)
    changePwdVisible.value = false
    newPassword.value = ''
    confirmPassword.value = ''
  } catch {
    // error handled by interceptor
  } finally {
    changePwdLoading.value = false
  }
}

function hasSubmenu(menu: MenuItem): boolean {
  return !!(menu.children && menu.children.length > 0)
}
</script>

<template>
  <div class="layout-container">
    <el-container class="layout-main">
      <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
        <div class="aside-header">
          <span v-if="!isCollapse" class="aside-title">人事管理系统</span>
          <span v-else class="aside-title-collapsed">HR</span>
        </div>

        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409eff"
          @select="handleMenuSelect"
        >
          <template v-for="menu in menuList" :key="menu.path">
            <el-sub-menu v-if="hasSubmenu(menu)" :index="menu.path">
              <template #title>
                <el-icon><component :is="menu.icon" /></el-icon>
                <span>{{ menu.title }}</span>
              </template>
              <el-menu-item
                v-for="child in menu.children"
                :key="child.path"
                :index="child.path"
              >
                <el-icon><component :is="child.icon" /></el-icon>
                <span>{{ child.title }}</span>
              </el-menu-item>
            </el-sub-menu>

            <el-menu-item v-else :index="menu.path">
              <el-icon><component :is="menu.icon" /></el-icon>
              <span>{{ menu.title }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <el-container>
        <el-header class="layout-header">
          <div class="header-left">
            <el-icon class="collapse-btn" @click="toggleCollapse">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </div>
          <div class="header-right">
            <el-dropdown trigger="click" @command="handleCommand">
              <span class="user-info">
                {{ userStore.userInfo?.person_name || userStore.userInfo?.username || '用户' }}
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

        <el-main class="layout-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>

    <el-dialog
      v-model="changePwdVisible"
      title="修改密码"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="新密码">
          <el-input v-model="newPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="confirmPassword" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changePwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="changePwdLoading" @click="submitChangePassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.layout-container {
  height: 100%;
}

.layout-main {
  height: 100%;
}

.layout-aside {
  background-color: #304156;
  overflow: hidden;
  transition: width 0.3s;

  .aside-header {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 18px;
    font-weight: bold;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .aside-title-collapsed {
    font-size: 16px;
  }

  .el-menu {
    border-right: none;
  }
}

.layout-header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;

  .header-left {
    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      color: #666;

      &:hover {
        color: #409eff;
      }
    }
  }

  .header-right {
    .user-info {
      cursor: pointer;
      color: #333;
      display: flex;
      align-items: center;
      gap: 4px;

      &:hover {
        color: #409eff;
      }
    }
  }
}

.layout-content {
  background: #f5f7fa;
  padding: 16px;
  overflow-y: auto;
}
</style>
