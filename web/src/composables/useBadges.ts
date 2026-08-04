import { ref } from 'vue'
import request from '@/utils/request'

// 徽章是时点状态展示（如结转后应立即由橙变绿），必须实时——不做任何时间缓存；
// 仅合并同一时刻的重复请求（共享同一 Promise），请求完成后立即释放，下一次调用总是重新拉取。
const inflight = new Map<string, Promise<any>>()

export function fetchBadges<T = any>(cacheKey: string, fetchFn: () => Promise<T>): Promise<T> {
  const pending = inflight.get(cacheKey)
  if (pending) {
    return pending
  }
  const promise = fetchFn().finally(() => {
    inflight.delete(cacheKey)
  })
  inflight.set(cacheKey, promise)
  return promise
}

// useBadges：页面徽章（dotMap）/余额映射（balanceMap）统一加载，接口失败静默降级（无点/无徽章）
export function useBadges() {
  const dotMap = ref<Record<number, string>>({})
  const balanceMap = ref<Record<number, number>>({})

  async function loadDots(cacheKey: string, fetchFn: () => Promise<any>) {
    try {
      const badges = (await fetchBadges(cacheKey, fetchFn)) || []
      dotMap.value = Object.fromEntries(badges.map((b: any) => [b.person_id, b.level]))
    } catch {
      dotMap.value = {}
    }
  }

  async function loadBalances(cacheKey: string, url: string) {
    try {
      const d = (await fetchBadges(cacheKey, () =>
        request.get(url, { params: { pageNum: 1, pageSize: 100 } }),
      )) as any
      const map: Record<number, number> = {}
      for (const row of d?.list || []) {
        map[row.person_id] = row.balance_hours ?? 0
      }
      balanceMap.value = map
    } catch {
      balanceMap.value = {}
    }
  }

  return { dotMap, balanceMap, loadDots, loadBalances }
}
