import { describe, it, expect } from 'vitest'
import { isActivePerson, filterPersons } from '../personScope'

describe('isActivePerson', () => {
  it('在职人员为活跃', () => {
    expect(isActivePerson({ is_active: true, entry_date: '2020-01-01', leave_date: null })).toBe(true)
  })

  it('未入职（无快照）为活跃', () => {
    expect(isActivePerson({ is_active: false, entry_date: null, leave_date: null })).toBe(true)
  })

  it('离职 3 个月内为活跃', () => {
    const leave = new Date()
    leave.setMonth(leave.getMonth() - 2)
    expect(isActivePerson({ is_active: false, entry_date: '2020-01-01', leave_date: leave.toISOString().slice(0, 10) })).toBe(true)
  })

  it('离职恰好 3 个月为活跃（边界）', () => {
    const leave = new Date()
    leave.setMonth(leave.getMonth() - 3)
    expect(isActivePerson({ is_active: false, entry_date: '2020-01-01', leave_date: leave.toISOString().slice(0, 10) })).toBe(true)
  })

  it('离职超过 3 个月为非活跃', () => {
    const leave = new Date()
    leave.setMonth(leave.getMonth() - 4)
    expect(isActivePerson({ is_active: false, entry_date: '2020-01-01', leave_date: leave.toISOString().slice(0, 10) })).toBe(false)
  })
})

describe('filterPersons', () => {
  const now = new Date()
  const leave2m = new Date()
  leave2m.setMonth(now.getMonth() - 2)
  const leave6m = new Date()
  leave6m.setMonth(now.getMonth() - 6)
  const cards = [
    { id: 1, is_active: true, leave_date: null },
    { id: 2, is_active: false, entry_date: null, leave_date: null },
    { id: 3, is_active: false, entry_date: '2020-01-01', leave_date: leave2m.toISOString().slice(0, 10) },
    { id: 4, is_active: false, entry_date: '2020-01-01', leave_date: leave6m.toISOString().slice(0, 10) },
  ]

  it('active 范围：仅活跃人员', () => {
    const ids = filterPersons(cards, 'active').map(c => c.id)
    expect(ids).toEqual([1, 2, 3])
  })

  it('all 范围：全部人员', () => {
    expect(filterPersons(cards, 'all')).toHaveLength(4)
  })
})
