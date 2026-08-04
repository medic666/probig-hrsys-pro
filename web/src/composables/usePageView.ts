import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// viewModeOf 视图标识推导（初始化与 URL 回退同步共用）：
// 列表标识恒为 'list'，其余（缺省/未知/异常值）一律回退卡片标识 cardValue。
export function viewModeOf<TCard extends string>(value: unknown, cardValue: TCard): TCard | 'list' {
  return value === 'list' ? 'list' : cardValue
}

// usePageView 页面卡片/列表双视图统一状态（URL 唯一状态源）：
// cardValue 为卡片视图标识（如 'cards' / 'blocks'），列表视图标识恒为 'list'。
// 视图切换 push 进历史（浏览器返回回到上一个视图），URL 回退时反向同步 viewMode；
// 卡片钻取（person/name/month）与列表筛选（各字段键）在各自 query 作用域内互不干扰。
export function usePageView<TCard extends string>(cardValue: TCard) {
  const route = useRoute()
  const router = useRouter()

  const viewMode = ref<TCard | 'list'>(viewModeOf(route.query.view, cardValue))
  const isList = computed(() => viewMode.value === 'list')

  watch(viewMode, (v) => {
    router.push({ query: { ...route.query, view: String(v) } })
  })

  // URL 回退（浏览器返回）时同步 viewMode；自身 push 触发时目标值与当前一致 → 跳过，无循环
  watch(
    () => route.query.view,
    (v) => {
      const target = viewModeOf(v, cardValue)
      if (viewMode.value !== target) viewMode.value = target
    },
  )

  return { viewMode, isList }
}
