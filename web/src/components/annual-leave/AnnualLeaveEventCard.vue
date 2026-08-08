<template>
  <div class="annual-leave-card" @click="$emit('click', event)">
    <div class="alc-header">
      <span class="alc-person">{{ event.person_name }}</span>
      <el-tag size="small">{{ typeText }}</el-tag>
      <el-tag v-if="event.source_type === 'attendance'" size="small" type="info">考勤休假</el-tag>
    </div>
    <div class="alc-line">时长：{{ hoursToDays(event.hours || 0).toFixed(2) }} 天</div>
    <div class="alc-line">生效日期：{{ event.effective_date }}</div>
    <div class="alc-line" :class="{ 'alc-empty': !event.remark }">{{ event.remark ? '备注：' + event.remark : '暂无备注' }}</div>
    <div class="alc-actions">
      <el-button v-permission="PERM.annualLeaveEventWrite" size="small" type="primary" link @click.stop="$emit('edit', event)">编辑</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { hoursToDays } from '@/utils'
import { annualLeaveTypeText } from '@/constants/annual-leave'
import { PERM } from '@/constants/permission'

const props = defineProps<{ event: any }>()
defineEmits<{
  (e: 'click', event: any): void
  (e: 'edit', event: any): void
}>()

const typeText = computed(() => annualLeaveTypeText(props.event.event_type))
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.annual-leave-card {
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

  .alc-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 8px;

    .alc-person {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }
  }

  .alc-line {
    font-size: 12px;
    line-height: 22px;
    color: #606266;

    &.alc-empty {
      color: #c0c4cc;
    }
  }

  .alc-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
