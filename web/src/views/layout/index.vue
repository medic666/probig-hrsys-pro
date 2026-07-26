<template>
  <el-container style="height:100vh;">
    <el-aside :width="isCollapse ? '64px' : '200px'" style="background:#304156;transition:width 0.3s;overflow:hidden;">
      <div style="height:60px;display:flex;align-items:center;justify-content:center;color:#fff;font-size:18px;white-space:nowrap;">
        <span v-if="!isCollapse">HR管理系统</span>
        <span v-else>HR</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        router
      >
        <template v-for="menu in menuList" :key="menu.path">
          <el-sub-menu v-if="menu.children" :index="menu.path">
            <template #title>
              <el-icon><component :is="menu.icon" /></el-icon>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item v-for="child in menu.children" :key="child.path" :index="child.path">
              {{ child.title }}
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
      <el-header style="height:60px;border-bottom:1px solid #e6e6e6;display:flex;align-items:center;justify-content:space-between;">
        <div>
          <el-button :icon="isCollapse ? 'Expand' : 'Fold'" text @click="isCollapse = !isCollapse" />
        </div>
        <el-dropdown>
          <span style="cursor:pointer;">
            {{ userStore.userInfo?.username || '' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="showChangePwd = true">修改密码</el-dropdown-item>
              <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main style="background:#f0f2f5;">
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <el-dialog v-model="showChangePwd" title="修改密码" width="400px">
    <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules">
      <el-form-item prop="oldPassword" label="原密码">
        <el-input v-model="pwdForm.oldPassword" type="password" show-password />
      </el-form-item>
      <el-form-item prop="newPassword" label="新密码">
        <el-input v-model="pwdForm.newPassword" type="password" show-password />
      </el-form-item>
      <el-form-item prop="confirmPassword" label="确认密码">
        <el-input v-model="pwdForm.confirmPassword" type="password" show-password />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showChangePwd = false">取消</el-button>
      <el-button type="primary" :loading="changing" @click="handleChangePwd">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { ArrowDown } from '@element-plus/icons-vue'

const route = useRoute()
const userStore = useUserStore()
const permStore = usePermissionStore()

const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

const allMenus = [
  {
    path: '/person',
    title: '人员管理',
    icon: 'User',
    permKey: 'person:read',
  },
  {
    path: '/company',
    title: '公司管理',
    icon: 'OfficeBuilding',
    permKey: 'company:read',
  },
  {
    path: '/position',
    title: '职务管理',
    icon: 'Stamp',
    permKey: 'position:read',
  },
  {
    path: '/attendance',
    title: '考勤管理',
    icon: 'Clock',
    children: [
      { path: '/attendance/event', title: '考勤事件', permKey: 'attendance:read' },
      { path: '/attendance/daily', title: '日记工时', permKey: 'attendance:read' },
      { path: '/attendance/salary', title: '假勤工资', permKey: 'attendance:read' },
    ],
  },
  {
    path: '/leave-account',
    title: '假期管理',
    icon: 'Sunny',
    children: [
      { path: '/leave-account/event', title: '假期事件', permKey: 'leave:read' },
      { path: '/leave-account/balance', title: '假期余额', permKey: 'leave:read' },
      { path: '/leave-account/carryover', title: '周年结转', permKey: 'leave:read' },
    ],
  },
  {
    path: '/salary',
    title: '工资管理',
    icon: 'Money',
    children: [
      { path: '/salary/event', title: '工资事件', permKey: 'salary:read' },
      { path: '/salary/summary', title: '工资汇总', permKey: 'salary:read' },
    ],
  },
  { path: '/file', title: '文件管理', icon: 'Folder', permKey: 'file:read' },
  { path: '/audit', title: '审计日志', icon: 'DocumentChecked', permKey: 'audit:read' },
  {
    path: '/rbac',
    title: '权限管理',
    icon: 'Setting',
    children: [
      { path: '/rbac/user', title: '用户管理', permKey: 'rbac:read' },
      { path: '/rbac/role', title: '角色管理', permKey: 'rbac:read' },
    ],
  },
  { path: '/system/config', title: '系统配置', icon: 'Tools', permKey: 'system:read' },
]

function filterMenu(menus: any[]): any[] {
  return menus.filter(m => {
    if (m.children) {
      m.children = filterMenu(m.children)
      return m.children.length > 0
    }
    return permStore.hasPermission(m.permKey)
  })
}

const menuList = computed(() => filterMenu(allMenus))

const showChangePwd = ref(false)
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })
const pwdRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
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
const changing = ref(false)
const pwdFormRef = ref()

async function handleChangePwd() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  changing.value = true
  try {
    await userStore.changePassword(pwdForm.oldPassword, pwdForm.newPassword)
    ElMessage.success('密码修改成功')
    showChangePwd.value = false
  } catch (e) {
  } finally {
    changing.value = false
  }
}

function handleLogout() {
  userStore.logout()
}
</script>
