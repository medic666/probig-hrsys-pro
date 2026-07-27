import type { Directive, App } from 'vue'
import { usePermissionStore } from '@/stores/permission'

export const permission: Directive = {
  mounted(el, binding) {
    const permissionStore = usePermissionStore()
    const key = binding.value as string
    if (!key) return
    if (!permissionStore.hasPermission(key)) {
      el.parentNode?.removeChild(el)
    }
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permission)
}
