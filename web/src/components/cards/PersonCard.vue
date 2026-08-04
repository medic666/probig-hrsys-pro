<template>
  <div class="person-card" @click="$emit('click', person)">
    <span v-if="showWidget && dotColor" class="pc-dot" :class="'dot-' + dotColor" />
    <div class="pc-name">
      <span>{{ person.name }}</span>
      <slot v-if="showWidget && badgePosition === 'name'" name="badge" />
    </div>
    <div class="pc-meta">
      <div v-if="person.company_name" class="pc-line">公司：{{ person.company_name }}</div>
      <div v-if="person.department" class="pc-line">部门：{{ person.department }}</div>
      <div v-if="person.position" class="pc-line">职位：{{ person.position }}</div>
      <div v-if="!person.company_name && !person.department && !person.position" class="pc-line pc-empty">暂无职务信息</div>
      <div v-if="showWidget && badgePosition === 'meta'" class="pc-widget">
        <slot name="badge" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { isActivePerson } from '@/utils/personScope'

// 人员卡片：右上角颜色点徽章（dotColor，仅活跃人员佩戴，点恒在体系：gray 无数据/green 正常/
// orange 异常/red 特定状态）与内容小组件（#badge slot，仅活跃人员渲染，位置 badgePosition 参数化）。
// 佩戴规则在组件内部统一，页面零 v-if。
const props = withDefaults(
  defineProps<{
    person: {
      id: number
      name: string
      company_id?: number
      company_name?: string
      department?: string
      position?: string
      is_active?: boolean
      entry_date?: string | null
      leave_date?: string | null
    }
    dotColor?: '' | 'gray' | 'green' | 'orange' | 'red'
    badgePosition?: 'name' | 'meta'
  }>(),
  { dotColor: '', badgePosition: 'name' },
)
defineEmits<{ (e: 'click', person: any): void }>()

const showWidget = computed(() => isActivePerson(props.person))
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.person-card {
  position: relative;
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;

  @include hover-capable {
    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      transform: translateY(-2px);
    }
  }

  .pc-dot {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }

  .dot-gray {
    background: #c0c4cc;
  }

  .dot-green {
    background: #67c23a;
  }

  .dot-orange {
    background: #e6a23c;
  }

  .dot-red {
    background: #f56c6c;
  }

  .pc-name {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pc-meta {
    .pc-line {
      font-size: 12px;
      color: #606266;
      line-height: 20px;
    }

    .pc-empty {
      color: #c0c4cc;
    }
  }

  .pc-widget {
    margin-top: 8px;
  }
}
</style>
