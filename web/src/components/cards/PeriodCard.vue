<template>
  <div class="period-card">
    <div class="pc-title">{{ title }}</div>
    <div class="pc-grid">
      <div
        v-for="p in periods"
        :key="p.period"
        class="pc-cell"
        :class="'level-' + p.level"
        @click="$emit('select', p.period)"
      >
        <div class="pc-label">{{ periodLabel(p.period, aggregate) }}</div>
        <div v-if="p.count" class="pc-count">{{ p.count }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PeriodStat, PeriodAggregate } from '@/composables/usePeriodStats'

// 时段聚合卡片：月度（"N月"）与年度（"YYYY年"）同构渲染，颜色语义统一
// （绿=有事件/已核算、橙=待确认/过期、灰=无）。
defineProps<{
  periods: PeriodStat[]
  title?: string
  aggregate?: PeriodAggregate
}>()
defineEmits<{ (e: 'select', period: string): void }>()

function periodLabel(period: string, aggregate: PeriodAggregate = 'month') {
  if (aggregate === 'year') return `${period}年`
  const [, mm] = period.split('-')
  return `${Number(mm)}月`
}
</script>

<style lang="scss" scoped>
.period-card {
  .pc-title {
    font-size: 14px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 10px;
  }

  .pc-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    max-width: 640px;

    .pc-cell {
      border: 1px solid #e4e7ed;
      border-radius: 6px;
      padding: 12px;
      text-align: center;
      cursor: pointer;
      transition: box-shadow 0.2s;

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      }

      .pc-label {
        font-size: 14px;
        color: #303133;
      }

      .pc-count {
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
