import { describe, it, expect } from 'vitest'
import { resolveBackTarget } from '../useBusinessPage'

function route(query: Record<string, unknown>, fullPath = '/person/5') {
  return { query, fullPath }
}

describe('resolveBackTarget 返回目标解析（来源页语义）', () => {
  it('优先返回 query.back（站内路径，含列表状态）', () => {
    expect(resolveBackTarget(route({ back: '/person?view=cards&status=active' }), '/person')).toBe(
      '/person?view=cards&status=active',
    )
  })

  it('back 非站内路径（非 / 开头）回退 fallback', () => {
    expect(resolveBackTarget(route({ back: 'javascript:alert(1)' }), '/person')).toBe('/person')
    expect(resolveBackTarget(route({ back: 'https://evil.example' }), '/person')).toBe('/person')
  })

  it('back 与当前页相同（自环）回退 fallback', () => {
    expect(resolveBackTarget(route({ back: '/person/5' }, '/person/5'), '/person')).toBe('/person')
  })

  it('无 back 回退 fallback', () => {
    expect(resolveBackTarget(route({}), '/person')).toBe('/person')
  })

  it('back 为数组等异常值回退 fallback', () => {
    expect(resolveBackTarget(route({ back: ['/a', '/b'] }), '/person')).toBe('/person')
    expect(resolveBackTarget(route({ back: null }), '/person')).toBe('/person')
  })

  it('fallback 为空时回退首页', () => {
    expect(resolveBackTarget(route({}), '')).toBe('/')
  })
})
