import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/layout/index.vue'),
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/home/index.vue'),
        meta: { title: '首页', icon: 'HomeFilled', permissionKey: '' }
      },
      {
        path: 'person',
        name: 'Person',
        component: () => import('@/views/person/index.vue'),
        meta: { title: '人员管理', icon: 'User', permissionKey: 'person:read' }
      },
      {
        path: 'company',
        name: 'Company',
        component: () => import('@/views/company/index.vue'),
        meta: { title: '公司管理', icon: 'OfficeBuilding', permissionKey: 'company:read' }
      },
      {
        path: 'position',
        name: 'Position',
        component: () => import('@/views/position/index.vue'),
        meta: { title: '职务管理', icon: 'Briefcase', permissionKey: 'position:read' }
      },
      {
        path: 'attendance/event',
        name: 'AttendanceEvent',
        component: () => import('@/views/attendance/event.vue'),
        meta: { title: '考勤事件', icon: 'Edit', permissionKey: 'attendance:read' }
      },
      {
        path: 'attendance/daily',
        name: 'AttendanceDaily',
        component: () => import('@/views/attendance/daily.vue'),
        meta: { title: '日记工时', icon: 'Calendar', permissionKey: 'attendance:read' }
      },
      {
        path: 'attendance/salary',
        name: 'AttendanceSalary',
        component: () => import('@/views/attendance/salary.vue'),
        meta: { title: '月度假勤工资', icon: 'Money', permissionKey: 'attendance:read' }
      },
      {
        path: 'leave-account/event',
        name: 'LeaveAccountEvent',
        component: () => import('@/views/leave-account/event.vue'),
        meta: { title: '额度事件', icon: 'Edit', permissionKey: 'leave_account:read' }
      },
      {
        path: 'leave-account/balance',
        name: 'LeaveAccountBalance',
        component: () => import('@/views/leave-account/balance.vue'),
        meta: { title: '额度查询', icon: 'Search', permissionKey: 'leave_account:read' }
      },
      {
        path: 'leave-account/carryover',
        name: 'LeaveAccountCarryover',
        component: () => import('@/views/leave-account/carryover.vue'),
        meta: { title: '周年结转', icon: 'Refresh', permissionKey: 'leave_account:read' }
      },
      {
        path: 'salary/event',
        name: 'SalaryEvent',
        component: () => import('@/views/salary/event.vue'),
        meta: { title: '工资事件', icon: 'Edit', permissionKey: 'salary:read' }
      },
      {
        path: 'salary/summary',
        name: 'SalarySummary',
        component: () => import('@/views/salary/summary.vue'),
        meta: { title: '月度工资', icon: 'Tickets', permissionKey: 'salary:read' }
      },
      {
        path: 'file',
        name: 'File',
        component: () => import('@/views/file/index.vue'),
        meta: { title: '文件管理', icon: 'FolderOpened', permissionKey: 'file:read' }
      },
      {
        path: 'audit',
        name: 'Audit',
        component: () => import('@/views/audit/index.vue'),
        meta: { title: '审计日志', icon: 'DocumentChecked', permissionKey: 'audit:read' }
      },
      {
        path: 'system/config',
        name: 'SystemConfig',
        component: () => import('@/views/system/config.vue'),
        meta: { title: '系统配置', icon: 'Setting', permissionKey: 'system:read' }
      },
      {
        path: 'rbac/user',
        name: 'RbacUser',
        component: () => import('@/views/rbac/user.vue'),
        meta: { title: '用户管理', icon: 'UserFilled', permissionKey: 'rbac:read' }
      },
      {
        path: 'rbac/role',
        name: 'RbacRole',
        component: () => import('@/views/rbac/role.vue'),
        meta: { title: '角色管理', icon: 'Avatar', permissionKey: 'rbac:read' }
      }
    ]
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '403' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '404' }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach(async (to, _from, next) => {
  document.title = `${to.meta.title || '企业人事管理系统'}`

  if (to.name === 'Login') {
    next()
    return
  }

  const userStore = useUserStore()
  const permissionStore = usePermissionStore()

  if (!userStore.token) {
    next({ name: 'Login' })
    return
  }

  if (!userStore.userInfo) {
    try {
      await userStore.fetchUserInfo()
    } catch {
      userStore.logout()
      next({ name: 'Login' })
      return
    }
  }

  if (permissionStore.permissions.length === 0) {
    try {
      await permissionStore.fetchPermissions()
    } catch {
      next({ name: 'Forbidden' })
      return
    }
  }

  const permissionKey = to.meta.permissionKey as string | undefined
  if (permissionKey && !permissionStore.hasPermission(permissionKey)) {
    next({ name: 'Forbidden' })
    return
  }

  next()
})

export default router
