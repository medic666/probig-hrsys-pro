import { ref } from 'vue'

export type PeriodLevel = 'green' | 'orange' | 'gray'

export interface PeriodStat {
  period: string
  level: PeriodLevel
  count?: number
}

export type PeriodAggregate = 'month' | 'year'

export interface PeriodStatsOptions {
  personId: number
  fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
  dateField: string
  statusField?: string
  monthField?: string
  pendingValues?: string[]
  aggregate?: PeriodAggregate
}

// usePeriodStats 时段聚合统计（月度/年度同构）：
// 事件源：有待确认(橙) / 有事件(绿) / 无(灰)；快照：stale(橙) / 有核算(绿) / 无(灰)。
// month 窗口 = [max(最近 11 个月起点, startMonth 入职月), 当月] + 未来 1 个月；
// year 窗口 = [入职年(缺省今年-1), 今年] + 明年。start 之前的时段一律不展示。
export function usePeriodStats(options: PeriodStatsOptions) {
  const periods = ref<PeriodStat[]>([])
  const loading = ref(false)

  const aggregate: PeriodAggregate = options.aggregate || 'month'

  function periodKeyOf(d: Date): string {
    if (aggregate === 'year') return `${d.getFullYear()}`
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
  }

  function buildPeriods(start?: string): PeriodStat[] {
    const now = new Date()
    const list: PeriodStat[] = []
    if (aggregate === 'year') {
      const startYear = start ? parseInt(start, 10) : now.getFullYear() - 1
      const endYear = now.getFullYear() + 1
      for (let y = startYear; y <= endYear; y++) {
        list.push({ period: String(y), level: 'gray' })
      }
      return list
    }
    // 最近 11 个月（含当月）+ 未来 1 个月，共 12 格；入职月之后的才展示
    for (let i = 10; i >= -1; i--) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
      const key = periodKeyOf(d)
      if (start && key < start) continue
      list.push({ period: key, level: 'gray' })
    }
    return list
  }

  async function load(personId?: number, start?: string) {
    loading.value = true
    const pid = personId ?? options.personId
    const stats = buildPeriods(start)
    const countMap = new Map<string, number>()
    const orangeMap = new Map<string, boolean>()
    try {
      const now = new Date()
      let startStr: string
      let endStr: string
      if (aggregate === 'year') {
        const startYear = start ? parseInt(start, 10) : now.getFullYear() - 1
        startStr = `${startYear}-01-01`
        endStr = `${now.getFullYear() + 1}-12-31`
      } else {
        const windowStart = new Date(now.getFullYear(), now.getMonth() - 10, 1)
        const rangeStart = start ? new Date(start + '-01') : windowStart
        if (rangeStart < windowStart) rangeStart.setTime(windowStart.getTime())
        const rangeEnd = new Date(now.getFullYear(), now.getMonth() + 2, 0)
        startStr = `${rangeStart.getFullYear()}-${String(rangeStart.getMonth() + 1).padStart(2, '0')}-01`
        endStr = `${rangeEnd.getFullYear()}-${String(rangeEnd.getMonth() + 1).padStart(2, '0')}-${String(rangeEnd.getDate()).padStart(2, '0')}`
      }

      const params: any = { pageNum: 1, pageSize: 100 }
      if (pid > 0) {
        params.person_id = pid
      }
      if (options.dateField) {
        params.date_start = startStr
        params.date_end = endStr
      }
      const pendingSet = new Set(options.pendingValues || ['pending'])
      const keyLen = aggregate === 'year' ? 4 : 7

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
            key = String(row[options.monthField]).slice(0, keyLen)
          } else {
            const raw = row[options.dateField] || ''
            key = String(raw).slice(0, keyLen)
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
      const count = countMap.get(m.period) || 0
      if (count > 0) {
        m.count = count
        m.level = orangeMap.get(m.period) ? 'orange' : 'green'
      }
    }
    periods.value = stats
    loading.value = false
    return stats
  }

  return { periods, loading, load }
}
