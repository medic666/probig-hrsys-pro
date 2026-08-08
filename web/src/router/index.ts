import { createRouter, createWebHashHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { getMe } from '@/api/auth'

import { routes } from './routes'

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const whiteList = ['/login', '/404', '/403']

router.beforeEach(async (to, from, next) => {
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

  // 业务逻辑页进入时记录来源页完整路径（含列表筛选/钻取状态），供返回精确回退：
  // 返回 = replace(query.back) → 来源页；直达（无 back）→ meta.backTo 模块列表
  const backable = to.meta.backTo as string | undefined
  const excludedSources = ['/', '/login', '/403', '/404']
  if (
    backable &&
    !to.query.back &&
    from.path !== to.path &&
    !excludedSources.includes(from.path)
  ) {
    next({ ...to, query: { ...to.query, back: from.fullPath } })
    return
  }

  next()
})

export default router
