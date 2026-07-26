import { defineStore } from 'pinia'
import { ref } from 'vue'

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([])
  const menus = ref<any[]>([])

  function setPermissions(perms: string[]) {
    permissions.value = perms
  }

  function hasPermission(permKey: string): boolean {
    if (permissions.value.length === 0) return false
    return permissions.value.includes(permKey)
  }

  function setMenus(menuList: any[]) {
    menus.value = menuList
  }

  return { permissions, menus, setPermissions, hasPermission, setMenus }
})
