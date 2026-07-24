import { createRouter, createWebHashHistory } from 'vue-router'
import Layout from '@/components/layout/Layout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/login/Login.vue'), meta: { title: '登录' } },
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        { path: '/dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/Dashboard.vue'), meta: { title: '首页' } },
        { path: '/person', name: 'Person', component: () => import('@/views/person/PersonList.vue'), meta: { title: '人员管理' } },
        { path: '/person/deleted', name: 'PersonDeleted', component: () => import('@/views/person/PersonDeleted.vue'), meta: { title: '回收站' } },
        { path: '/person/:id', name: 'PersonDetail', component: () => import('@/views/person/PersonDetail.vue'), meta: { title: '人员详情' } },
        { path: '/company', name: 'Company', component: () => import('@/views/company/CompanyList.vue'), meta: { title: '公司管理' } },
        { path: '/attendance', name: 'Attendance', component: () => import('@/views/attendance/AttendanceList.vue'), meta: { title: '假勤管理' } },
        { path: '/attendance-summary', name: 'AttendanceSummary', component: () => import('@/views/attendance/AttendanceSummary.vue'), meta: { title: '考勤汇总' } },
        { path: '/salary', name: 'Salary', component: () => import('@/views/salary/SalaryList.vue'), meta: { title: '工资管理' } },
        { path: '/salary-summary', name: 'SalarySummary', component: () => import('@/views/salary/SalarySummary.vue'), meta: { title: '工资汇总' } },
        { path: '/file', name: 'File', component: () => import('@/views/file/FileList.vue'), meta: { title: '文件管理' } },
        { path: '/audit', name: 'Audit', component: () => import('@/views/audit/AuditList.vue'), meta: { title: '操作审计' } },
        { path: '/system', name: 'SystemConfig', component: () => import('@/views/system/SystemConfig.vue'), meta: { title: '系统配置' } },
        { path: '/user', name: 'User', component: () => import('@/views/system/UserList.vue'), meta: { title: '用户管理' } },
        { path: '/role', name: 'Role', component: () => import('@/views/system/RoleList.vue'), meta: { title: '角色管理' } },
        { path: '/profile', name: 'Profile', component: () => import('@/views/profile/Profile.vue'), meta: { title: '个人中心' } },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  document.title = (to.meta.title as string) || '企业人事与行政管理系统'
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
