import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { pinia } from './stores'
import { setupPermissionDirective } from './directives/permission'

import './styles/index.scss'

// Edge 浏览器缺陷修复：vue-router 在页面隐藏(visibilitychange)时调用 history.replaceState
// 保存滚动位置，Edge 会将其误判为"页面激活"信号，导致窗口被强制拉回前台、任务栏图标
// 持续闪烁。页面隐藏期间拦截 replaceState 调用以规避该缺陷（仅影响隐藏时的一次滚动保存）。
// 参考：vue-router/dist/vue-router.mjs beforeUnloadListener
const originalReplaceState = window.history.replaceState
window.history.replaceState = function (this: History, ...args: Parameters<History['replaceState']>) {
  if (document.visibilityState === 'hidden') {
    return
  }
  return originalReplaceState.apply(this, args)
}

const app = createApp(App)

app.use(ElementPlus, { locale: zhCn })
app.use(router)
app.use(pinia)
setupPermissionDirective(app)

app.mount('#app')
