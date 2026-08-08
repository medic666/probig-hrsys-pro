import type { Directive, App } from 'vue'
import { watch } from 'vue'
import { usePermissionStore } from '@/stores/permission'

export const permission: Directive = {
  mounted(el, binding) {
    const permissionStore = usePermissionStore()
    const apply = () => {
      const key = binding.value as string
      if (!key || !el.isConnected) return
      if (!permissionStore.hasPermission(key)) {
        el.parentNode?.removeChild(el)
      }
    }
    apply()
    // 权限数据加载完成/变化后重新评估（元素可能先于权限数据挂载，如异步卡片渲染）
    const stop = watch(() => permissionStore.permissions, apply)
    ;(el as any).__permStop = stop
  },
  unmounted(el) {
    ;(el as any).__permStop?.()
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permission)
}
