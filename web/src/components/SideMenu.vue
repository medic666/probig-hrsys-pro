<template>
  <el-menu
    :default-active="activeMenu"
    :collapse="collapse"
    :collapse-transition="false"
    router
    background-color="#304156"
    text-color="#bfcbd9"
    active-text-color="#409eff"
    @select="emit('select')"
  >
    <template v-for="menu in displayMenus" :key="menu.path">
      <el-sub-menu v-if="menu.children && menu.children.length" :index="menu.path">
        <template #title>
          <el-icon><component :is="getIcon(menu.icon)" /></el-icon>
          <span>{{ menu.title }}</span>
        </template>
        <el-menu-item v-for="child in menu.children" :key="child.path" :index="child.path">
          <span>{{ child.title }}</span>
        </el-menu-item>
      </el-sub-menu>
      <el-menu-item v-else :index="menu.path">
        <el-icon><component :is="getIcon(menu.icon)" /></el-icon>
        <span>{{ menu.title }}</span>
      </el-menu-item>
    </template>
  </el-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { usePermissionStore } from '@/stores/permission'
import {
  HomeFilled,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
} from '@element-plus/icons-vue'

// 侧边菜单（桌面侧栏与移动抽屉共用）：数据源为权限菜单，图标统一映射；
// select 事件在移动抽屉内用于选择后自动收起。
withDefaults(defineProps<{ collapse?: boolean }>(), { collapse: false })

const emit = defineEmits<{ (e: 'select'): void }>()

const route = useRoute()
const permissionStore = usePermissionStore()

const activeMenu = computed(() => route.path)

const iconMap: Record<string, any> = {
  HomeFilled,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
}

function getIcon(name: string) {
  return iconMap[name] || HomeFilled
}

const displayMenus = computed(() => {
  const storeMenus = permissionStore.menus
  if (storeMenus.length > 0) return storeMenus

  return [
    { path: '/home', title: '首页', icon: 'HomeFilled' },
  ]
})
</script>
