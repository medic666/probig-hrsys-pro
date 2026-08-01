import { ref } from 'vue'

export interface MonthStat {
  month: string
  level: 'green' | 'orange' | 'gray'
  count?: number
}

export interface MonthStatsOptions {
  personId: number
  fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
  dateField: string
  statusField?: string
  monthField?: string
  pendingValues?: string[]
  kind?: 'event' | 'snapshot'
}

// useMonthStats 最近 12 个月（含当月）的月度统计：
// 事件源：有待确认(橙) / 有事件(绿) / 无(灰)；快照：stale(橙) / 有核算(绿) / 无(灰)
export function useMonthStats(options: MonthStatsOptions) {
  const months = ref<MonthStat[]>([])
  const loading = ref(false)

  function monthKey(d: Date) {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
  }

  function buildMonths() {
    const now = new Date()
    const list: MonthStat[] = []
    for (let i = 11; i >= 0; i--) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
      list.push({ month: monthKey(d), level: 'gray' })
    }
    return list
  }

  async function load(personId?: number) {
    loading.value = true
    const pid = personId ?? options.personId
    const stats = buildMonths()
    const countMap = new Map<string, number>()
    const orangeMap = new Map<string, boolean>()
    try {
      const rangeStart = new Date()
      rangeStart.setMonth(rangeStart.getMonth() - 11)
      rangeStart.setDate(1)
      const start = `${rangeStart.getFullYear()}-${String(rangeStart.getMonth() + 1).padStart(2, '0')}-01`
      const end = `${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, '0')}-31`

      const params: any = { pageNum: 1, pageSize: 100 }
      if (pid > 0) {
        params.person_id = pid
      }
      if (options.dateField) {
        params.date_start = start
        params.date_end = end
      }
      const pendingSet = new Set(options.pendingValues || ['pending'])

      // 分页循环拉取
      let page = 1
      let total = Infinity
      while ((page - 1) * 100 < total && page <= 50) {
        const d = await options.fetchFn({ ...params, pageNum: page, pageSize: 100 })
        const list = d.list || []
        total = d.total || 0
        for (const row of list) {
          let key: string
          if (options.monthField && row[options.monthField]) {
            key = String(row[options.monthField]).slice(0, 7)
          } else {
            const raw = row[options.dateField] || ''
            key = String(raw).slice(0, 7)
          }
          countMap.set(key, (countMap.get(key) || 0) + 1)
          const status = row[options.statusField || 'status']
          if (pendingSet.has(String(status))) {
            orangeMap.set(key, true)
          }
        }
        page++
      }
    } catch {
      /* 加载失败保持灰 */
    }

    for (const m of stats) {
      const count = countMap.get(m.month) || 0
      if (count > 0) {
        m.count = count
        m.level = orangeMap.get(m.month) ? 'orange' : 'green'
      }
    }
    months.value = stats
    loading.value = false
    return stats
  }

  return { months, loading, load }
}
