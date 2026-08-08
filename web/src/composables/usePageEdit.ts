import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePermissionStore } from '@/stores/permission'

// 业务详情页「查看态/编辑态」统一状态机：
// - URL query.edit 为编辑态单一状态源（同路由 query 变化不重挂载，watch 同步，刷新/回退可恢复）
// - 无 write 权限或 extraViewOnly（如考勤 ?readonly=1）→ 恒查看态，无编辑入口
// - enterEdit/exitEdit 同步 URL（保留 back 等既有 query）
// 各详情页只声明权限键与展示，状态机单一实现。
export function usePageEdit(writeKey: string, extraViewOnly?: () => boolean) {
  const route = useRoute()
  const router = useRouter()
  const permissionStore = usePermissionStore()

  const hasWrite = computed(() => permissionStore.hasPermission(writeKey))
  const viewOnly = computed(() => (extraViewOnly ? extraViewOnly() : false) || !hasWrite.value)
  const editMode = ref(false)
  watch(
    () => route.query.edit,
    (v) => {
      if (!viewOnly.value) editMode.value = v === '1'
    },
    { immediate: true },
  )

  function enterEdit() {
    editMode.value = true
    router.replace({ query: { ...route.query, edit: '1' } })
  }

  function exitEdit() {
    editMode.value = false
    const q = { ...route.query }
    delete q.edit
    router.replace({ query: q })
  }

  return { hasWrite, viewOnly, editMode, enterEdit, exitEdit }
}
