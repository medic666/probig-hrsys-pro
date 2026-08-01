<template>
  <div class="month-card">
    <div class="mc-title">{{ title }}</div>
    <div class="mc-grid">
      <div
        v-for="m in months"
        :key="m.month"
        class="mc-cell"
        :class="'level-' + m.level"
        @click="$emit('select', m.month)"
      >
        <div class="mc-month">{{ monthLabel(m.month) }}</div>
        <div v-if="m.count" class="mc-count">{{ m.count }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  months: { month: string; level: 'green' | 'orange' | 'gray'; count?: number }[]
  title?: string
}>()
defineEmits<{ (e: 'select', month: string): void }>()

function monthLabel(m: string) {
  const [, mm] = m.split('-')
  return `${Number(mm)}月`
}
</script>

<style lang="scss" scoped>
.month-card {
  .mc-title {
    font-size: 14px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 10px;
  }

  .mc-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    max-width: 640px;

    .mc-cell {
      border: 1px solid #e4e7ed;
      border-radius: 6px;
      padding: 12px;
      text-align: center;
      cursor: pointer;
      transition: box-shadow 0.2s;

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      }

      .mc-month {
        font-size: 14px;
        color: #303133;
      }

      .mc-count {
        font-size: 11px;
        color: #909399;
        margin-top: 2px;
      }

      &.level-green {
        background: #f0f9eb;
        border-color: #b3e19d;
      }

      &.level-orange {
        background: #fdf6ec;
        border-color: #eebe77;
      }

      &.level-gray {
        background: #f4f4f5;
        border-color: #e4e7ed;
      }
    }
  }
}
</style>
