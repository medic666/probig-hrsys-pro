<template>
  <div class="day-card">
    <div class="dc-header">
      <span class="dc-date">{{ date }}</span>
      <slot name="status" />
    </div>
    <div class="dc-events">
      <div v-for="(e, i) in events" :key="i" class="dc-event">
        <span class="ev-type">{{ e.event_type }}<span v-if="e.sub_type">-{{ e.sub_type }}</span></span>
        <span v-if="e.event_type === '违纪'" class="ev-meta">{{ e.minutes ? e.minutes + '分钟' : '' }}</span>
        <span v-else class="ev-meta">{{ hoursToDays(e.hours || 0).toFixed(2) }}天</span>
        <span v-if="e.remark" class="ev-remark">{{ e.remark }}</span>
      </div>
      <div v-if="events.length === 0" class="dc-empty">无事件</div>
    </div>
    <div v-if="$slots.actions" class="dc-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { hoursToDays } from '@/utils'

defineProps<{
  date: string
  events: any[]
}>()
</script>

<style lang="scss" scoped>
.day-card {
  width: 260px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
  padding: 10px 12px;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .dc-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;

    .dc-date {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }
  }

  .dc-events {
    .dc-event {
      font-size: 12px;
      line-height: 22px;
      display: flex;
      align-items: center;
      gap: 8px;

      .ev-type {
        color: #303133;
        white-space: nowrap;
      }

      .ev-meta {
        color: #606266;
        white-space: nowrap;
      }

      .ev-remark {
        color: #909399;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .dc-empty {
      color: #c0c4cc;
      font-size: 12px;
      line-height: 22px;
    }
  }

  .dc-actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
