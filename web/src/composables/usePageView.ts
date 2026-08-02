import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

// usePageView 页面卡片/列表双视图统一状态（URL 唯一状态源）：
// cardValue 为卡片视图标识（如 'cards' / 'blocks'），列表视图标识恒为 'list'。
// 视图模式读写路由 query 的 view 参数（合并保留其它键），切换视图仅翻转该参数，
// 卡片钻取（person/name/month）与列表筛选（各字段键）在各自 query 作用域内互不干扰。
export function usePageView<TCard extends string>(cardValue: TCard) {
  const route = useRoute()
  const router = useRouter()

  const queryView = route.query.view
  const viewMode = ref<TCard | 'list'>((queryView === 'list' ? 'list' : cardValue) as TCard | 'list')
  const isList = computed(() => viewMode.value === 'list')

  watch(viewMode, (v) => {
    router.replace({ query: { ...route.query, view: String(v) } })
  })

  return { viewMode, isList }
}
