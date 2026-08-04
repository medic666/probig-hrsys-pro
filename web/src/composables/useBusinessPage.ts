import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// resolveBackTarget 返回目标解析：优先 query.back（进入业务页时由路由守卫注入的
// 来源页完整路径，含列表筛选/钻取状态）；back 必须是站内路径（/ 开头）且不等于
// 当前页（防自环）；否则回退 fallback（meta.backTo 模块列表）。
export function resolveBackTarget(
  route: { query: Record<string, unknown>; fullPath: string },
  fallback: string,
): string {
  const q = route.query.back
  const back = typeof q === 'string' && q.startsWith('/') && q !== route.fullPath ? q : ''
  return back || fallback || '/'
}

// useBusinessPage 业务逻辑页统一胶水：从路由取 id、推导新增/编辑模式、统一返回逻辑。
// 返回目标取 query.back（来源页）→ meta.backTo（模块列表）→ 首页，与 BusinessPage 标题/返回行为同源。
export function useBusinessPage() {
  const route = useRoute()
  const router = useRouter()

  const id = computed(() => (route.params.id ? Number(route.params.id) : null))
  const isCreate = computed(() => id.value == null)
  const backTo = computed(() => String(route.meta.backTo || '/'))
  const backTarget = computed(() => resolveBackTarget(route, backTo.value))

  function goBack() {
    router.replace(backTarget.value)
  }

  return { id, isCreate, backTo, backTarget, goBack }
}
