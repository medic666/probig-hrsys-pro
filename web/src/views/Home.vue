<template>
  <div class="home-page">
    <el-card>
      <div class="welcome-container">
        <h2>欢迎使用企业人事与行政管理系统</h2>
        <p class="subtitle">面向中小企业的轻量生产级人事与行政管理系统</p>
        <el-divider />
        <div class="feature-list">
          <el-row :gutter="20">
            <el-col v-for="item in features" :key="item.title" :span="6">
              <el-card shadow="hover" class="feature-card">
                <el-icon :size="32" :color="item.color">
                  <component :is="item.icon" />
                </el-icon>
                <h4>{{ item.title }}</h4>
                <p>{{ item.desc }}</p>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { User, OfficeBuilding, Clock, Money } from '@element-plus/icons-vue'

const features = [
  { title: '人员管理', desc: '人员信息维护与查询', icon: User, color: '#409eff' },
  { title: '公司管理', desc: '公司信息与文件管理', icon: OfficeBuilding, color: '#67c23a' },
  { title: '考勤管理', desc: '考勤事件录入与核算', icon: Clock, color: '#e6a23c' },
  { title: '薪资核算', desc: '工资事件管理与核算', icon: Money, color: '#f56c6c' },
]
</script>

<style lang="scss" scoped>
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

    h4 {
      margin: 12px 0 8px;
      font-size: 16px;
      color: #303133;
    }

    p {
      font-size: 12px;
      color: #909399;
    }
  }
}
</style>
