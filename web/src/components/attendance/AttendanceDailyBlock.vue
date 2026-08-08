<template>
  <div class="daily-block" :class="{ 'is-edited': edited, 'is-pending': daily.status === 'pending' }">
    <div class="block-header">
      <span class="person">{{ daily.person_name }}</span>
      <span class="date">{{ daily.event_date }}</span>
      <el-tag v-if="edited" type="warning" size="small">已修改待确认</el-tag>
      <el-tag v-else :type="daily.status === 'pending' ? 'warning' : 'success'" size="small">
        {{ daily.status === 'pending' ? '待确认' : '已确认' }}
      </el-tag>
    </div>
    <div class="block-events">
      <div v-for="(d, i) in daily.details" :key="i" class="event-row">
        <span class="ev-type">{{ d.event_type }}<span v-if="d.sub_type">-{{ d.sub_type }}</span></span>
        <span v-if="d.event_type === '违纪'" class="ev-meta">{{ d.minutes ? d.minutes + '分钟' : '' }}</span>
        <span v-else class="ev-meta">{{ hoursToDays(d.hours || 0).toFixed(2) }}天</span>
        <span v-if="d.remark" class="ev-remark">{{ d.remark }}</span>
      </div>
      <div v-if="!daily.details || daily.details.length === 0" class="ev-empty">无事件</div>
      <div v-if="daily.punch_time" class="ev-punch">打卡时间: {{ daily.punch_time }}</div>
    </div>
    <div class="block-actions">
      <slot name="extra-actions" />
      <el-button v-permission="PERM.attendanceWrite" size="small" type="primary" link @click="$emit('edit', daily)">编辑</el-button>
      <el-button v-if="daily.status === 'pending'" v-permission="PERM.attendanceWrite" size="small" type="success" link @click="$emit('confirm', daily)">确认</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { hoursToDays } from '@/utils'
import { PERM } from '@/constants/permission'

defineProps<{ daily: any; edited?: boolean }>()
defineEmits<{
  (e: 'edit', row: any): void
  (e: 'confirm', row: any): void
}>()
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.daily-block {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
  padding: 10px 12px;
  transition: box-shadow 0.2s;

  @include hover-capable {
    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }
  }

  &.is-edited {
    border-color: #e6a23c;
  }

  &.is-pending {
    border-color: #eebe77;
  }

  .block-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .person {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }

    .date {
      color: #909399;
      font-size: 12px;
    }
  }

  .block-events {
    .event-row {
      font-size: 12px;
      line-height: 22px;
      display: flex;
      align-items: center;
      gap: 8px;

      .ev-type {
        color: #303133;
        white-space: nowrap;
      }

      .ev-meta,
      .ev-time {
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

    .ev-empty {
      color: #c0c4cc;
      font-size: 12px;
      line-height: 22px;
    }

    .ev-punch {
      font-size: 12px;
      color: #909399;
      margin-top: 4px;
    }
  }

  .block-actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
