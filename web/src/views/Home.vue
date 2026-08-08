<template>
  <div class="home-page">
    <el-card>
      <div class="welcome-container">
        <h2>欢迎使用企业人事与行政管理系统</h2>
        <p class="subtitle">面向中小企业的轻量生产级人事与行政管理系统</p>
        <el-divider />
        <!-- 首页业务模块 = 菜单一级分组投影：可见性/命名/图标均由后端权限菜单驱动，
             无权限分组不在 menus 中 → 自动不渲染 -->
        <el-row v-if="featureGroups.length" :gutter="20">
          <el-col v-for="g in featureGroups" :key="g.path" :xs="12" :sm="6">
            <el-card shadow="hover" class="feature-card" @click="openGroup(g)">
              <el-icon :size="32" class="feature-icon">
                <component :is="getMenuIcon(g.icon)" />
              </el-icon>
              <h4>{{ g.title }}</h4>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-else description="暂无可用模块，请联系管理员配置权限" :image-size="80" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { usePermissionStore } from '@/stores/permission'
import { getMenuIcon } from '@/constants/menuIcons'

// 首页业务分组白名单（前端展示策略）：新增业务分组时，后端菜单加分组 + 本白名单加一行
const BUSINESS_GROUPS = ['/data', '/attendance-group', '/leave-group', '/salary-group']

const router = useRouter()
const permissionStore = usePermissionStore()

const featureGroups = computed(() =>
  permissionStore.menus
    .filter((m: any) => BUSINESS_GROUPS.includes(m.path) && m.children?.length)
    .map((m: any) => ({ path: m.path, title: m.title, icon: m.icon, firstChild: m.children[0].path })),
)

// 点击分组卡片 → 进入分组内第一个有权限的子模块
function openGroup(g: { firstChild: string }) {
  router.push(g.firstChild)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.home-page {
  .welcome-container {
    text-align: center;
    padding: 40px 0;

    h2 {
      font-size: 24px;
      color: #303133;
    }

    .subtitle {
      margin-top: 12px;
      color: #909399;
      font-size: 14px;
    }
  }

  .feature-list {
    margin-top: 20px;
  }

  .feature-card {
    text-align: center;
    cursor: pointer;
    margin-bottom: 12px;

    .feature-icon {
      color: $primary-color;
    }

    h4 {
      margin: 12px 0 8px;
      font-size: 16px;
      color: #303133;
    }

    @include hover-capable {
      &:hover {
        transform: translateY(-2px);
      }
    }
  }
}
</style>
