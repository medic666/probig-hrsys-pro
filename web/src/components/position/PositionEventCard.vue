<template>
  <div class="position-event-card">
    <div class="pec-header">
      <span class="pec-person">{{ event.person_name }}</span>
      <el-tag size="small">{{ event.event_type }}</el-tag>
    </div>
    <div class="pec-line">生效日期：{{ event.effective_date }}</div>
    <div class="pec-line" :class="{ 'pec-empty': !event.remark }">{{ event.remark ? '备注：' + event.remark : '暂无备注' }}</div>
    <div class="pec-actions">
      <el-button size="small" type="primary" link @click="$emit('edit', event)">编辑</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ event: any }>()
defineEmits<{ (e: 'edit', event: any): void }>()
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.position-event-card {
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

  .pec-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .pec-person {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }
  }

  .pec-line {
    font-size: 12px;
    line-height: 22px;
    color: #606266;

    &.pec-empty {
      color: #c0c4cc;
    }
  }

  .pec-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 6px;
    border-top: 1px dashed #ebeef5;
    padding-top: 6px;
  }
}
</style>
