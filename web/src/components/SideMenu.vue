<template>
  <el-menu
    :default-active="activeMenu"
    :collapse="collapse"
    :collapse-transition="false"
    :router="!manualNav"
    background-color="#304156"
    text-color="#bfcbd9"
    active-text-color="#409eff"
    @select="(index: string) => emit('select', index)"
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

// 侧边菜单（桌面侧栏与移动抽屉共用）：数据源为权限菜单，图标统一映射。
// manualNav=true（移动抽屉场景）时禁用 el-menu 自动路由，仅上抛选中的 index，
// 由外层先关闭抽屉、待关闭动画结束后再手动导航（避免导航瞬间抽屉入镜历史快照）。
withDefaults(defineProps<{ collapse?: boolean; manualNav?: boolean }>(), {
  collapse: false,
  manualNav: false,
})

const emit = defineEmits<{ (e: 'select', index: string): void }>()

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
