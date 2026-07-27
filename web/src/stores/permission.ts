import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([])
  const menus = ref<any[]>([])

  function setPermissions(perms: string[]) {
    permissions.value = perms
  }

  function setMenus(menuList: any[]) {
    menus.value = menuList
  }

  function hasPermission(key: string): boolean {
    return permissions.value.includes(key)
  }

  function clearPermissions() {
    permissions.value = []
    menus.value = []
  }

  async function fetchPermissions() {
    const data = (await request.get('/user/permissions')) as {
      permissions: string[]
      menus: any[]
    }
    setPermissions(data?.permissions || [])
    setMenus(data?.menus || [])
  }

  return {
    permissions,
    menus,
    setPermissions,
    setMenus,
    hasPermission,
    clearPermissions,
    fetchPermissions,
  }
})
