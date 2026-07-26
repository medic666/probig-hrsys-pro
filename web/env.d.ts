/// <reference types="vite/client" />

declare module 'element-plus/dist/locale/zh-cn.mjs' {
  const zhCn: any
  export default zhCn
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}
