import { describe, it, expect } from 'vitest'
import { routes } from '@/router/routes'
import { PERM } from '@/constants/permission'

// 路由权限键一致性防错网：所有 meta.permissionKey 必须定义于 PERM 常量
// （新增业务模块时若 meta 键拼错/漏定义，本测试立即失败）
describe('路由权限键一致性', () => {
  it('所有 meta.permissionKey 必须定义于 PERM 常量', () => {
    const permValues = new Set(Object.values(PERM))
    const walk = (rs: any[]) => {
      for (const r of rs) {
        if (r.meta?.permissionKey) {
          expect(
            permValues.has(r.meta.permissionKey),
            `路由 ${r.path} 的 permissionKey ${r.meta.permissionKey} 未定义于 PERM`,
          ).toBe(true)
        }
        if (r.children) walk(r.children)
      }
    }
    walk(routes)
  })
})
