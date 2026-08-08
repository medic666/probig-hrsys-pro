<template>
  <div class="lil-card" @click="$emit('click', event)">
    <div class="lc-header">
      <span class="lc-person">{{ event.person_name }}</span>
      <el-tag size="small" :type="event.sub_type === '补班出勤' ? 'success' : 'warning'">{{ event.sub_type }}</el-tag>
    </div>
    <div class="lc-line">时长：{{ hoursToDays(event.hours || 0).toFixed(2) }} 天</div>
    <div class="lc-line">日期：{{ event.event_date }}</div>
    <div class="lc-line" :class="{ 'lc-empty': !event.remark }">{{ event.remark ? '备注：' + event.remark : '暂无备注' }}</div>
    <div class="lc-actions">
      <el-button v-permission="PERM.attendanceEventWrite" size="small" type="primary" link @click.stop="$emit('edit', event)">编辑</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { hoursToDays } from '@/utils'
import { PERM } from '@/constants/permission'

// 调休事件卡片：本体点击=查看当日考勤详情；编辑=考勤事件编辑（调休由考勤事件派生）
defineProps<{ event: any }>()
defineEmits<{
  (e: 'click', event: any): void
  (e: 'edit', event: any): void
}>()
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.lil-card {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
  padding: 10px 12px;
  cursor: pointer;
  transition: box-shadow 0.2s;

  @include hover-capable {
    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }
  }

  .lc-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 8px;

    .lc-person {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }
  }

  .lc-line {
    font-size: 12px;
    line-height: 22px;
    color: #606266;

    &.lc-empty {
      color: #c0c4cc;
    }
  }

  .lc-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
