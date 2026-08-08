import {
  HomeFilled,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
} from '@element-plus/icons-vue'

// 菜单图标映射（与后端 buildMenuTree 下发的 icon 名对齐）：
// 侧边菜单与首页模块卡片共用，新增菜单图标只改这一处。
export const menuIcons: Record<string, any> = {
  HomeFilled,
  User,
  OfficeBuilding,
  Clock,
  Money,
  Document,
  Setting,
}

export function getMenuIcon(name: string) {
  return menuIcons[name] || HomeFilled
}
