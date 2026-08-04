import { describe, it, expect } from 'vitest'
import { hoursToDays, formatMoney, formatDate, formatDateTime } from '@/utils/index'

describe('hours/days 换算（8 小时 = 1 天）', () => {
  it('hoursToDays 往返一致', () => {
    expect(hoursToDays(8)).toBe(1)
    expect(hoursToDays(4)).toBe(0.5)
    expect(hoursToDays(10)).toBe(1.25)
    expect(hoursToDays(0)).toBe(0)
  })
})

describe('金额格式化', () => {
  it('formatMoney 保留两位小数', () => {
    expect(formatMoney(100)).toBe('100.00')
    expect(formatMoney(100.5)).toBe('100.50')
    expect(formatMoney(100.555)).toBe('100.56')
  })
})

describe('日期格式化', () => {
  it('formatDate 默认 YYYY-MM-DD', () => {
    expect(formatDate('2026-06-01')).toBe('2026-06-01')
  })

  it('formatDateTime 完整时间', () => {
    expect(formatDateTime('2026-06-01T08:30:00')).toBe('2026-06-01 08:30:00')
  })
})
