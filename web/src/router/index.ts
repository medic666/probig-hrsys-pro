import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { getMe } from '@/api/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: { title: '首页', permissionKey: 'home.read' },
      },
      {
        path: 'system/users',
        name: 'UserManage',
        component: () => import('@/views/system/UserManage.vue'),
        meta: { title: '用户管理', permissionKey: 'user.read' },
      },
      {
        path: 'system/roles',
        name: 'RoleManage',
        component: () => import('@/views/system/RoleManage.vue'),
        meta: { title: '角色管理', permissionKey: 'role.read' },
      },
      {
        path: 'person',
        name: 'PersonManage',
        component: () => import('@/views/person/PersonManage.vue'),
        meta: { title: '人员管理', permissionKey: 'person.read' },
      },
      {
        path: 'company',
        name: 'CompanyManage',
        component: () => import('@/views/company/CompanyManage.vue'),
        meta: { title: '公司管理', permissionKey: 'company.read' },
      },
      {
        path: 'position-event',
        name: 'PositionEventManage',
        component: () => import('@/views/position/PositionEventManage.vue'),
        meta: { title: '职务事件', permissionKey: 'position_event.read' },
      },
      {
        path: 'position-events/create',
        name: 'PositionEventCreate',
        component: () => import('@/views/position/PositionEventPage.vue'),
        meta: { title: '新增职务事件', permissionKey: 'position_event.read' },
      },
      {
        path: 'position-events/:id',
        name: 'PositionEventEdit',
        component: () => import('@/views/position/PositionEventPage.vue'),
        meta: { title: '编辑职务事件', permissionKey: 'position_event.read' },
      },
      {
        path: 'attendance',
        name: 'AttendanceEventManage',
        component: () => import('@/views/attendance/AttendanceEventManage.vue'),
        meta: { title: '考勤事件', permissionKey: 'attendance.read' },
      },
      {
        path: 'attendance-pending',
        name: 'PendingConfirm',
        component: () => import('@/views/attendance/PendingConfirm.vue'),
        meta: { title: '待确认考勤', permissionKey: 'attendance.read' },
      },
      {
        path: 'attendance-daily',
        name: 'AttendanceDailyQuery',
        component: () => import('@/views/attendance/AttendanceDailyQuery.vue'),
        meta: { title: '日记工时', permissionKey: 'attendance.read' },
      },
      {
        path: 'attendance-monthly',
        name: 'AttendanceMonthlyCalc',
        component: () => import('@/views/attendance/AttendanceMonthlyCalc.vue'),
        meta: { title: '月度考勤核算', permissionKey: 'attendance.read' },
      },
      {
        path: 'attendance-monthly/:personId/:month',
        name: 'AttendanceMonthlyDetail',
        component: () => import('@/views/attendance/AttendanceMonthlyDetail.vue'),
        meta: { title: '月度考勤核算详情', permissionKey: 'attendance.read' },
      },
      {
        path: 'annual-leave-events',
        name: 'AnnualLeaveEventFlow',
        component: () => import('@/views/annual-leave/AnnualLeaveEventFlow.vue'),
        meta: { title: '年假事件', permissionKey: 'annual_leave.read' },
      },
      {
        path: 'annual-leave-balance',
        name: 'AnnualLeaveBalance',
        component: () => import('@/views/annual-leave/AnnualLeaveBalance.vue'),
        meta: { title: '年假余额', permissionKey: 'annual_leave.read' },
      },
      {
        path: 'annual-leave-carryover',
        name: 'AnnualLeaveCarryover',
        component: () => import('@/views/annual-leave/AnnualLeaveCarryover.vue'),
        meta: { title: '周年结转', permissionKey: 'annual_leave.read' },
      },
      {
        path: 'lil-events',
        name: 'LILEventFlow',
        component: () => import('@/views/annual-leave/LILEventFlow.vue'),
        meta: { title: '调休事件', permissionKey: 'annual_leave.read' },
      },
      {
        path: 'lil-balance',
        name: 'LILBalance',
        component: () => import('@/views/annual-leave/LILBalance.vue'),
        meta: { title: '调休余额', permissionKey: 'annual_leave.read' },
      },
      {
        path: 'salary-events',
        name: 'SalaryEventManage',
        component: () => import('@/views/salary/SalaryEventManage.vue'),
        meta: { title: '工资事件', permissionKey: 'salary.read' },
      },
      {
        path: 'salary-summaries',
        name: 'SalarySummary',
        component: () => import('@/views/salary/SalarySummary.vue'),
        meta: { title: '月度工资汇总', permissionKey: 'salary.read' },
      },
      {
        path: 'files',
        name: 'FileManage',
        component: () => import('@/views/file/FileManage.vue'),
        meta: { title: '文件管理', permissionKey: 'file.read' },
      },
      {
        path: 'audit-logs',
        name: 'AuditLogList',
        component: () => import('@/views/audit/AuditLogList.vue'),
        meta: { title: '审计日志', permissionKey: 'audit.read' },
      },
      {
        path: 'system/config',
        name: 'SystemConfig',
        component: () => import('@/views/system/SystemConfig.vue'),
        meta: { title: '系统配置', permissionKey: 'system_config.read' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/Forbidden.vue'),
    meta: { title: '无权限' },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/NotFound.vue'),
    meta: { title: '页面不存在' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const whiteList = ['/login', '/404', '/403']

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()
  const permissionStore = usePermissionStore()

  if (whiteList.includes(to.path)) {
    if (to.path === '/login' && userStore.token) {
      next('/')
      return
    }
    next()
    return
  }

  if (!userStore.token) {
    next(`/login?redirect=${to.path}`)
    return
  }

  if (permissionStore.permissions.length === 0) {
    try {
      const data = await getMe()
      permissionStore.setPermissions(data.permissions || [])
      permissionStore.setMenus(data.menus || [])
      if ((data as any).is_first_login) {
        userStore.clearUser()
        permissionStore.clearPermissions()
        next('/login')
        return
      }
    } catch {
      userStore.clearUser()
      permissionStore.clearPermissions()
      next('/login')
      return
    }
  }

  const permKey = to.meta.permissionKey as string | undefined
  if (permKey && !permissionStore.hasPermission(permKey)) {
    next('/403')
    return
  }

  next()
})

export default router
