import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// useBusinessPage 业务逻辑页统一胶水：从路由取 id、推导新增/编辑模式、统一返回逻辑。
// 返回目标取路由 meta.backTo（无历史记录时回退），与 BusinessPage 标题/返回行为同源。
export function useBusinessPage() {
  const route = useRoute()
  const router = useRouter()

  const id = computed(() => (route.params.id ? Number(route.params.id) : null))
  const isCreate = computed(() => id.value == null)
  const backTo = computed(() => String(route.meta.backTo || '/'))

  function goBack() {
    if (window.history.length > 1) {
      router.back()
    } else {
      router.replace(backTo.value)
    }
  }

  return { id, isCreate, backTo, goBack }
}
