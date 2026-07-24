import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('../components/AppLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('../views/Dashboard.vue'),
          meta: { module: 'personnel', action: 'read' },
        },
        {
          path: 'personnel',
          name: 'PersonnelList',
          component: () => import('../views/personnel/PersonnelList.vue'),
          meta: { module: 'personnel', action: 'read' },
        },
        {
          path: 'personnel/:id',
          name: 'PersonnelDetail',
          component: () => import('../views/personnel/PersonnelDetail.vue'),
          meta: { module: 'personnel', action: 'read' },
        },
        {
          path: 'organization',
          name: 'OrgList',
          component: () => import('../views/organization/OrgList.vue'),
          meta: { module: 'organization', action: 'read' },
        },
        {
          path: 'organization/:id',
          name: 'OrgDetail',
          component: () => import('../views/organization/OrgDetail.vue'),
          meta: { module: 'organization', action: 'read' },
        },
        {
          path: 'attendance',
          name: 'AttendanceSummary',
          component: () => import('../views/attendance/AttendanceView.vue'),
          meta: { module: 'attendance', action: 'read' },
        },
        {
          path: 'salary',
          name: 'SalarySummary',
          component: () => import('../views/salary/SalaryView.vue'),
          meta: { module: 'salary', action: 'read' },
        },
        {
          path: 'files',
          name: 'FileManagement',
          component: () => import('../views/file/FileManagement.vue'),
          meta: { module: 'file', action: 'read' },
        },
        {
          path: 'audit',
          name: 'AuditLog',
          component: () => import('../views/audit/AuditLog.vue'),
          meta: { module: 'audit', action: 'read' },
        },
      ],
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.public) {
    next()
    return
  }

  if (auth.token && !auth.user) {
    await auth.fetchUser()
    await auth.fetchPermissions()
  }

  if (!auth.token) {
    next('/login')
    return
  }

  const mod = to.meta.module as string
  const act = to.meta.action as string
  if (mod && act && !auth.hasPermission(mod, act)) {
    next('/login')
    return
  }

  next()
})

export default router
