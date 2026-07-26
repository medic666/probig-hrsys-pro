import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getPermissions as getPermissionsApi } from '@/api/auth'

export interface MenuItem {
  path: string
  title: string
  icon: string
  permissionKey: string
  children?: MenuItem[]
}

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([])
  const menus = ref<MenuItem[]>([])

  const allMenus: MenuItem[] = [
    { path: '/home', title: '首页', icon: 'HomeFilled', permissionKey: '' },
    { path: '/person', title: '人员管理', icon: 'User', permissionKey: 'person:read' },
    { path: '/company', title: '公司管理', icon: 'OfficeBuilding', permissionKey: 'company:read' },
    { path: '/position', title: '职务管理', icon: 'Briefcase', permissionKey: 'position:read' },
    {
      path: '/attendance',
      title: '考勤管理',
      icon: 'Clock',
      permissionKey: 'attendance:read',
      children: [
        { path: '/attendance/event', title: '考勤事件', icon: 'Edit', permissionKey: 'attendance:read' },
        { path: '/attendance/daily', title: '日记工时', icon: 'Calendar', permissionKey: 'attendance:read' },
        { path: '/attendance/salary', title: '月度假勤工资', icon: 'Money', permissionKey: 'attendance:read' }
      ]
    },
    {
      path: '/leave-account',
      title: '假期管理',
      icon: 'Sunny',
      permissionKey: 'leave_account:read',
      children: [
        { path: '/leave-account/event', title: '额度事件', icon: 'Edit', permissionKey: 'leave_account:read' },
        { path: '/leave-account/balance', title: '额度查询', icon: 'Search', permissionKey: 'leave_account:read' },
        { path: '/leave-account/carryover', title: '周年结转', icon: 'Refresh', permissionKey: 'leave_account:read' }
      ]
    },
    {
      path: '/salary',
      title: '工资管理',
      icon: 'Coin',
      permissionKey: 'salary:read',
      children: [
        { path: '/salary/event', title: '工资事件', icon: 'Edit', permissionKey: 'salary:read' },
        { path: '/salary/summary', title: '月度工资', icon: 'Tickets', permissionKey: 'salary:read' }
      ]
    },
    { path: '/file', title: '文件管理', icon: 'FolderOpened', permissionKey: 'file:read' },
    { path: '/audit', title: '审计日志', icon: 'DocumentChecked', permissionKey: 'audit:read' },
    { path: '/system/config', title: '系统配置', icon: 'Setting', permissionKey: 'system:read' },
    {
      path: '/rbac',
      title: '权限管理',
      icon: 'Lock',
      permissionKey: 'rbac:read',
      children: [
        { path: '/rbac/user', title: '用户管理', icon: 'UserFilled', permissionKey: 'rbac:read' },
        { path: '/rbac/role', title: '角色管理', icon: 'Avatar', permissionKey: 'rbac:read' }
      ]
    }
  ]

  function filterMenus(perms: string[]): MenuItem[] {
    return allMenus
      .filter((menu) => !menu.permissionKey || perms.includes(menu.permissionKey))
      .map((menu) => ({
        ...menu,
        children: menu.children
          ? menu.children.filter((child) => !child.permissionKey || perms.includes(child.permissionKey))
          : undefined
      }))
  }

  async function fetchPermissions() {
    const perms = await getPermissionsApi()
    permissions.value = perms
    menus.value = filterMenus(perms)
  }

  function hasPermission(key: string): boolean {
    if (!key) return true
    return permissions.value.includes(key)
  }

  function clearPermissions() {
    permissions.value = []
    menus.value = []
  }

  return {
    permissions,
    menus,
    fetchPermissions,
    hasPermission,
    clearPermissions
  }
})
