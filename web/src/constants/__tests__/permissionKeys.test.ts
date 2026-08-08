/// <reference types="vite/client" />
import { describe, it, expect } from 'vitest'
import { PERM } from '@/constants/permission'

// v-permission 键一致性防错网：扫描 src 下全部 .vue/.ts 中 PERM.xxx 引用，
// 断言键必须定义于 PERM 常量（vue-tsc 对自定义指令绑定表达式不做属性检查，
// 此测试封死该盲区——键名拼错/旧键残留必然测试红）。
describe('权限键引用一致性', () => {
  it('所有 PERM.xxx 引用必须定义于 PERM 常量', () => {
    const keys = new Set(Object.keys(PERM))
    // ?raw 取源码原文（覆盖模板与脚本），eager 全量载入
    const modules = import.meta.glob('/src/**/*.{vue,ts}', {
      eager: true,
      query: '?raw',
      import: 'default',
    }) as Record<string, string>

    const bad: string[] = []
    for (const [path, content] of Object.entries(modules)) {
      for (const m of content.matchAll(/PERM\.(\w+)/g)) {
        if (!keys.has(m[1])) bad.push(`${path}: PERM.${m[1]}`)
      }
    }
    expect(bad).toEqual([])
  })
})
