import { App, DirectiveBinding } from 'vue'
import { usePermissionStore } from '@/stores/permission'

export function hasPermission(permKey: string): boolean {
  const store = usePermissionStore()
  return store.hasPermission(permKey)
}

export function setupPermission(app: App) {
  app.directive('permission', {
    mounted(el: HTMLElement, binding: DirectiveBinding<string>) {
      if (!binding.value) return
      const permissionStore = usePermissionStore()
      if (!permissionStore.hasPermission(binding.value)) {
        el.parentNode?.removeChild(el)
      }
    },
  })
}
