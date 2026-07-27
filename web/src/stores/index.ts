import { createPinia } from 'pinia'

export const pinia = createPinia()

export { useUserStore } from './user'
export { usePermissionStore } from './permission'
export { useConfigStore } from './config'
