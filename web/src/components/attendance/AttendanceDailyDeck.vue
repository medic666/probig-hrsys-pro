<template>
  <div class="daily-deck" :class="{ 'is-stacked': items.length > 1 }">
    <div class="deck-face">
      <AttendanceDailyBlock
        :key="current"
        :daily="items[current]"
        :edited="edited"
        @edit="emit('edit', items[current])"
        @confirm="emit('confirm', items[current])"
      >
        <template v-if="items.length > 1" #extra-actions>
          <el-button size="small" link type="primary" @click="cycle">切换</el-button>
          <el-button v-permission="PERM.attendanceWrite" size="small" link type="danger" @click="handleDelete">删除</el-button>
        </template>
      </AttendanceDailyBlock>
      <span v-if="items.length > 1" class="deck-badge">共 {{ items.length }} 组</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import AttendanceDailyBlock from '@/components/attendance/AttendanceDailyBlock.vue'
import { PERM } from '@/constants/permission'

// 考勤日套卡：当日多组事件（同日多版本，API 按 seq 倒序，items[0] 为最新有效组）时
// 以扑克牌堆叠效果展示，提供「切换」（滚动查看各组）与「删除」（软删除当前组）。
// 切换/删除与面卡自身的编辑/确认共用同一操作排；每组面卡均可编辑/转正为当日最新版。
const props = withDefaults(
  defineProps<{
    items: any[]
    edited?: boolean
  }>(),
  { edited: false },
)

const emit = defineEmits<{
  (e: 'edit', item: any): void
  (e: 'confirm', item: any): void
  (e: 'delete', item: any): void
}>()

const current = ref(0)

// 父级重载（确认/编辑/删除后 items 重新生成）时自动回到第 1 组（最新组）：
// 被操作组提升为新版后位于 items[0]，归零保证画面仍显示刚操作的那张卡
watch(
  () => props.items,
  () => {
    current.value = 0
  },
)

function cycle() {
  current.value = (current.value + 1) % props.items.length
}

async function handleDelete() {
  const item = props.items[current.value]
  try {
    await ElMessageBox.confirm(
      `确认删除 ${item.person_name} ${item.event_date} 第 ${item.seq} 组考勤？删除后当日有效记录将按剩余组重算。`,
      '提示',
      { type: 'warning' },
    )
  } catch {
    return
  }
  emit('delete', item)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.daily-deck {
  position: relative;

  .deck-face {
    position: relative;
    z-index: 2;

    // 切换面卡时的轻量滑入动效（:key 触发重渲染）
    :deep(.daily-block) {
      animation: deck-switch 0.18s ease;
    }
  }

  &.is-stacked {
    // 为下层牌预留底部露出区（不左右扩张，不顶邻卡）
    padding-bottom: 14px;

    // 下层牌（牌背质感）：统一尺寸 + 对角阶梯错位 + 轻微反向旋转，
    // 右缘/底缘分别收在 100%-4px 与 100% 处，等距对齐、整体收于容器内
    &::before,
    &::after {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      width: calc(100% - 8px);
      height: calc(100% - 8px);
      border: 1px solid #e6e8ec;
      border-radius: 6px;
      background: linear-gradient(160deg, #fafbfc 0%, #eef1f5 100%);
      box-shadow: 0 2px 6px rgba(31, 45, 61, 0.08);
      z-index: 0;
      pointer-events: none;
      transition: transform 0.25s ease;
    }

    &::after {
      transform: translate(4px, 4px) rotate(-1.5deg);
    }

    &::before {
      transform: translate(8px, 8px) rotate(1deg);
    }

    // hover 时下层牌向外散开，营造"拿起一叠牌"的互动感（仅悬停设备）
    @include hover-capable {
      &:hover::after {
        transform: translate(6px, 6px) rotate(-2.5deg);
      }

      &:hover::before {
        transform: translate(11px, 11px) rotate(1.8deg);
      }
    }

    .deck-badge {
      position: absolute;
      top: -8px;
      right: 8px;
      z-index: 3;
      padding: 1px 8px;
      border-radius: 10px;
      background: #409eff;
      color: #fff;
      font-size: 11px;
      line-height: 18px;
      box-shadow: 0 1px 4px rgba(64, 158, 255, 0.4);
    }
  }
}

@keyframes deck-switch {
  from {
    opacity: 0.4;
    transform: translateX(6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}
</style>
