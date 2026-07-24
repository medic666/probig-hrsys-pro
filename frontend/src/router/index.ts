import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Login from '../views/Login.vue'
import MainLayout from '../layout/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '工作台', icon: 'Odometer', perm: '' },
      },
      {
        path: 'persons',
        name: 'PersonList',
        component: () => import('../views/person/PersonList.vue'),
        meta: { title: '人员管理', icon: 'User', perm: 'person.read' },
      },
      {
        path: 'persons/:id/position-events',
        name: 'PositionEvents',
        component: () => import('../views/person/PositionEvents.vue'),
        meta: { title: '职务事件', icon: 'Document', perm: 'person.read', hidden: true },
      },
      {
        path: 'companies',
        name: 'CompanyList',
        component: () => import('../views/company/CompanyList.vue'),
        meta: { title: '公司管理', icon: 'OfficeBuilding', perm: 'company.read' },
      },
      {
        path: 'attendance/events',
        name: 'AttendanceEventList',
        component: () => import('../views/attendance/AttendanceEventList.vue'),
        meta: { title: '假勤事件', icon: 'Calendar', perm: 'attendance.read' },
      },
      {
        path: 'attendance/summary',
        name: 'AttendanceSummary',
        component: () => import('../views/attendance/AttendanceSummary.vue'),
        meta: { title: '考勤核算', icon: 'DataAnalysis', perm: 'attendance.read' },
      },
      {
        path: 'salary/events',
        name: 'SalaryEventList',
        component: () => import('../views/salary/SalaryEventList.vue'),
        meta: { title: '工资事件', icon: 'Money', perm: 'salary.read' },
      },
      {
        path: 'salary/summary',
        name: 'SalarySummary',
        component: () => import('../views/salary/SalarySummary.vue'),
        meta: { title: '工资核算', icon: 'Wallet', perm: 'salary.read' },
      },
      {
        path: 'files',
        name: 'FileList',
        component: () => import('../views/file/FileList.vue'),
        meta: { title: '文件管理', icon: 'Folder', perm: 'file.read' },
      },
      {
        path: 'audit',
        name: 'AuditLogList',
        component: () => import('../views/audit/AuditLogList.vue'),
        meta: { title: '操作审计', icon: 'Warning', perm: 'audit.read' },
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('../views/user/UserList.vue'),
        meta: { title: '用户管理', icon: 'Avatar', perm: 'user.read' },
      },
      {
        path: 'roles',
        name: 'RoleList',
        component: () => import('../views/user/RoleList.vue'),
        meta: { title: '角色管理', icon: 'Lock', perm: 'user.read' },
      },
      {
        path: 'config',
        name: 'SystemConfig',
        component: () => import('../views/system/SystemConfig.vue'),
        meta: { title: '系统配置', icon: 'Setting', perm: 'system.read' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
    return
  }
  if (to.path === '/login' && token) {
    next('/dashboard')
    return
  }
  next()
})

export default router
