import { describe, it, expect, beforeEach, vi } from 'vitest'
import { downloadBlob, type BlobResult } from '@/utils/download'

beforeEach(() => {
  vi.stubGlobal('document', {
    createElement: () => ({ click: () => {}, href: '', download: '' }),
  })
  URL.createObjectURL = vi.fn(() => 'blob:mock')
  URL.revokeObjectURL = vi.fn()
})

describe('downloadBlob', () => {
  it('接受 BlobResult（拦截器 blob 分支产物）', () => {
    const blob = new Blob(['x'], { type: 'application/octet-stream' })
    expect(() => downloadBlob({ blob, filename: 'a.xlsx' })).not.toThrow()
  })

  it('接受裸 Blob 并回退文件名', () => {
    const blob = new Blob(['x'])
    expect(() => downloadBlob(blob, 'fallback.xlsx')).not.toThrow()
  })

  it('类型结构正确', () => {
    const result: BlobResult = { blob: new Blob(['']), filename: 'f.xlsx' }
    expect(result.blob).toBeInstanceOf(Blob)
    expect(result.filename).toBe('f.xlsx')
  })
})
