import { ref, computed, getCurrentInstance, onUnmounted, type Ref } from 'vue'

// useBreakpoint 响应式断点（与 styles/variables.scss 断点同源，单一事实源）：
// 移动 <768px / 平板 768-1199px / 桌面 ≥1200px；isTouch 供触屏 hover 降级等判断。
// matchMedia 在 setup 即评估（无需等待挂载），组件卸载自动注销监听。
export type Breakpoint = 'mobile' | 'tablet' | 'desktop'

export function breakpointOf(width: number): Breakpoint {
  if (width < 768) return 'mobile'
  if (width < 1200) return 'tablet'
  return 'desktop'
}

function useMediaQuery(query: string): Ref<boolean> {
  const matches = ref(false)
  const mql = typeof matchMedia === 'function' ? matchMedia(query) : undefined
  if (!mql) return matches

  const onChange = (e: MediaQueryListEvent) => {
    matches.value = e.matches
  }
  matches.value = mql.matches
  mql.addEventListener('change', onChange)

  const instance = getCurrentInstance()
  if (instance) {
    onUnmounted(() => {
      mql.removeEventListener('change', onChange)
    })
  }
  return matches
}

export function useBreakpoint() {
  const isMobile = useMediaQuery('(max-width: 767px)')
  const isTablet = useMediaQuery('(min-width: 768px) and (max-width: 1199px)')
  const isDesktop = useMediaQuery('(min-width: 1200px)')
  const isTouch = useMediaQuery('(hover: none), (pointer: coarse)')

  const bp = computed<Breakpoint>(() => {
    if (isMobile.value) return 'mobile'
    if (isTablet.value) return 'tablet'
    return 'desktop'
  })

  return { isMobile, isTablet, isDesktop, isTouch, bp }
}
