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
