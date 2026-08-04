import { describe, it, expect, vi, afterEach } from 'vitest'
import { breakpointOf, useBreakpoint } from '../useBreakpoint'

function stubMatchMedia(queries: Record<string, boolean>) {
  vi.stubGlobal('matchMedia', (query: string) => ({
    matches: queries[query] ?? false,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

describe('breakpointOf 宽度映射（与 useBreakpoint 断点同源）', () => {
  it('三档断点映射正确', () => {
    expect(breakpointOf(375)).toBe('mobile')
    expect(breakpointOf(767)).toBe('mobile')
    expect(breakpointOf(768)).toBe('tablet')
    expect(breakpointOf(1199)).toBe('tablet')
    expect(breakpointOf(1200)).toBe('desktop')
    expect(breakpointOf(1920)).toBe('desktop')
  })
})

describe('useBreakpoint 媒体查询解析', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('移动端查询命中', () => {
    stubMatchMedia({ '(max-width: 767px)': true })
    const { isMobile, isTablet, isDesktop, bp } = useBreakpoint()
    expect(isMobile.value).toBe(true)
    expect(isTablet.value).toBe(false)
    expect(isDesktop.value).toBe(false)
    expect(bp.value).toBe('mobile')
  })

  it('平板查询命中', () => {
    stubMatchMedia({ '(min-width: 768px) and (max-width: 1199px)': true })
    const { isTablet, isMobile, bp } = useBreakpoint()
    expect(isTablet.value).toBe(true)
    expect(isMobile.value).toBe(false)
    expect(bp.value).toBe('tablet')
  })

  it('桌面查询命中', () => {
    stubMatchMedia({ '(min-width: 1200px)': true })
    const { isDesktop, isMobile, bp } = useBreakpoint()
    expect(isDesktop.value).toBe(true)
    expect(isMobile.value).toBe(false)
    expect(bp.value).toBe('desktop')
  })

  it('无 matchMedia 环境安全降级（全 false 不抛错）', () => {
    vi.unstubAllGlobals()
    const { isMobile, isTablet, isDesktop, isTouch } = useBreakpoint()
    expect(isMobile.value).toBe(false)
    expect(isTablet.value).toBe(false)
    expect(isDesktop.value).toBe(false)
    expect(isTouch.value).toBe(false)
  })
})
