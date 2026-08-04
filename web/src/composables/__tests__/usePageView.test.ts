import { describe, it, expect } from 'vitest'
import { viewModeOf } from '../usePageView'

describe('viewModeOf 视图标识推导（URL query.view → 视图模式）', () => {
  it('list 恒映射为列表视图', () => {
    expect(viewModeOf('list', 'cards')).toBe('list')
    expect(viewModeOf('list', 'blocks')).toBe('list')
  })

  it('卡片标识匹配 cardValue 时原样返回', () => {
    expect(viewModeOf('cards', 'cards')).toBe('cards')
    expect(viewModeOf('blocks', 'blocks')).toBe('blocks')
  })

  it('缺省/异常值回退卡片标识', () => {
    expect(viewModeOf(undefined, 'cards')).toBe('cards')
    expect(viewModeOf(null, 'cards')).toBe('cards')
    expect(viewModeOf('', 'cards')).toBe('cards')
    expect(viewModeOf('unknown', 'cards')).toBe('cards')
    expect(viewModeOf(['list'], 'cards')).toBe('cards')
    expect(viewModeOf(123, 'blocks')).toBe('blocks')
  })
})
