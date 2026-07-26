import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import { usePermissionStore } from '@/stores/permission'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/views/layout/index.vue'),
    redirect: '/person',
    children: [
      {
        path: 'person',
        name: 'Person',
        component: () => import('@/views/person/index.vue'),
        meta: { title: '人员管理', permissionKey: 'person:read' },
      },
      {
        path: 'company',
        name: 'Company',
        component: () => import('@/views/company/index.vue'),
        meta: { title: '公司管理', permissionKey: 'company:read' },
      },
      {
        path: 'position',
        name: 'Position',
        component: () => import('@/views/position/index.vue'),
        meta: { title: '职务管理', permissionKey: 'position:read' },
      },
      {
        path: 'attendance/event',
        name: 'AttendanceEvent',
        component: () => import('@/views/attendance/event.vue'),
        meta: { title: '考勤事件', permissionKey: 'attendance:read' },
      },
      {
        path: 'attendance/daily',
        name: 'AttendanceDaily',
        component: () => import('@/views/attendance/daily.vue'),
        meta: { title: '日记工时', permissionKey: 'attendance:read' },
      },
      {
        path: 'attendance/salary',
        name: 'AttendanceSalary',
        component: () => import('@/views/attendance/salary.vue'),
        meta: { title: '假勤工资', permissionKey: 'attendance:read' },
      },
      {
        path: 'leave-account/event',
        name: 'LeaveEvent',
        component: () => import('@/views/leave-account/event.vue'),
        meta: { title: '假期事件', permissionKey: 'leave:read' },
      },
      {
        path: 'leave-account/balance',
        name: 'LeaveBalance',
        component: () => import('@/views/leave-account/balance.vue'),
        meta: { title: '假期余额', permissionKey: 'leave:read' },
      },
      {
        path: 'leave-account/carryover',
        name: 'LeaveCarryover',
        component: () => import('@/views/leave-account/carryover.vue'),
        meta: { title: '周年结转', permissionKey: 'leave:read' },
      },
      {
        path: 'salary/event',
        name: 'SalaryEvent',
        component: () => import('@/views/salary/event.vue'),
        meta: { title: '工资事件', permissionKey: 'salary:read' },
      },
      {
        path: 'salary/summary',
        name: 'SalarySummary',
        component: () => import('@/views/salary/summary.vue'),
        meta: { title: '工资汇总', permissionKey: 'salary:read' },
      },
      {
        path: 'file',
        name: 'File',
        component: () => import('@/views/file/index.vue'),
        meta: { title: '文件管理', permissionKey: 'file:read' },
      },
      {
        path: 'audit',
        name: 'Audit',
        component: () => import('@/views/audit/index.vue'),
        meta: { title: '审计日志', permissionKey: 'audit:read' },
      },
      {
        path: 'system/config',
        name: 'SystemConfig',
        component: () => import('@/views/system/config.vue'),
        meta: { title: '系统配置', permissionKey: 'system:read' },
      },
      {
        path: 'rbac/user',
        name: 'RbacUser',
        component: () => import('@/views/rbac/user.vue'),
        meta: { title: '用户管理', permissionKey: 'rbac:read' },
      },
      {
        path: 'rbac/role',
        name: 'RbacRole',
        component: () => import('@/views/rbac/role.vue'),
        meta: { title: '角色管理', permissionKey: 'rbac:read' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/login/error.vue'),
    meta: { title: '无权限' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/login/error.vue'),
    meta: { title: '404' },
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  if (to.path === '/login') {
    next()
    return
  }

  const userStore = useUserStore()
  const permStore = usePermissionStore()

  const token = localStorage.getItem('token')
  if (!token) {
    next('/login')
    return
  }

  if (!userStore.userInfo) {
    try {
      const userInfo = await userStore.fetchUserInfo()
      if (userInfo && userInfo.permissions) {
        permStore.setPermissions(userInfo.permissions)
      }
    } catch {
      next('/login')
      return
    }
  }

  const permKey = to.meta?.permissionKey as string | undefined
  if (permKey && !permStore.hasPermission(permKey)) {
    next('/403')
    return
  }

  document.title = (to.meta?.title as string) || 'HR管理系统'
  next()
})

export default router
