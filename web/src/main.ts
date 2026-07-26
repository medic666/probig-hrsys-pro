import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import { usePermissionStore } from '@/stores/permission'
import router from '@/router'
import App from '@/App.vue'
import '@/assets/styles/global.scss'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.directive('permission', {
  mounted(el: HTMLElement, binding) {
    const permissionStore = usePermissionStore()
    if (binding.value && !permissionStore.hasPermission(binding.value)) {
      el.parentNode?.removeChild(el)
    }
  }
})

app.use(router)
app.use(ElementPlus)
app.mount('#app')
