<template>
  <div class="salary-event-card" @click="$emit('click', event)">
    <div class="sec-header">
      <span class="sec-person">{{ event.person_name }}</span>
      <el-tag size="small">{{ event.event_type }}</el-tag>
    </div>
    <div class="sec-line">金额：{{ formatMoney(event.amount || 0) }} 元</div>
    <div class="sec-line">归属月份：{{ event.belong_month }}</div>
    <div class="sec-line" :class="{ 'sec-empty': !event.remark }">{{ event.remark ? '备注：' + event.remark : '暂无备注' }}</div>
    <div class="sec-actions">
      <el-button v-permission="PERM.salaryEventWrite" size="small" type="primary" link @click.stop="$emit('edit', event)">编辑</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatMoney } from '@/utils'
import { PERM } from '@/constants/permission'

// 工资事件原子卡片：本体点击=查看详情；封面编辑按钮按权限渲染
defineProps<{ event: any }>()
defineEmits<{
  (e: 'click', event: any): void
  (e: 'edit', event: any): void
}>()
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.salary-event-card {
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

  .sec-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 8px;

    .sec-person {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }
  }

  .sec-line {
    font-size: 12px;
    line-height: 22px;
    color: #606266;

    &.sec-empty {
      color: #c0c4cc;
    }
  }

  .sec-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
