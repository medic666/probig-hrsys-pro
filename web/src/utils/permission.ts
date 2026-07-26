import { usePermissionStore } from '@/stores/permission'

export function hasPermission(key: string): boolean {
  const permissionStore = usePermissionStore()
  return permissionStore.hasPermission(key)
}
