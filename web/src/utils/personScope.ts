export type PersonScope = 'active' | 'all'

export const ACTIVE_LEAVE_MONTHS = 3

export interface PersonScopeable {
  is_active?: boolean
  entry_date?: string | null
  leave_date?: string | null
}

// isActivePerson 活跃判定：未入职 + 在职 + 离职 ≤ 3 个月 = 活跃；离职 > 3 个月 = 非活跃
export function isActivePerson(p: PersonScopeable): boolean {
  if (p.is_active) return true
  if (p.leave_date) {
    const leave = new Date(p.leave_date)
    const now = new Date()
    const monthsDiff = (now.getFullYear() - leave.getFullYear()) * 12 + (now.getMonth() - leave.getMonth())
    return monthsDiff <= ACTIVE_LEAVE_MONTHS
  }
  return true // 无离职日期视为未入职/活跃
}

// filterPersons 按范围过滤人员卡片
export function filterPersons<T extends PersonScopeable>(cards: T[], scope: PersonScope): T[] {
  if (scope === 'all') return cards
  return cards.filter(isActivePerson)
}
