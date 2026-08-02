<template>
  <div class="person-card" @click="$emit('click', person)">
    <div class="pc-name">
      <span>{{ person.name }}</span>
      <el-tag v-if="statusText" :type="statusType" size="small" class="pc-status">{{ statusText }}</el-tag>
    </div>
    <div class="pc-meta">
      <div v-if="person.company_name" class="pc-line">公司：{{ person.company_name }}</div>
      <div v-if="person.department" class="pc-line">部门：{{ person.department }}</div>
      <div v-if="person.position" class="pc-line">职位：{{ person.position }}</div>
      <div v-if="!person.company_name && !person.department && !person.position" class="pc-line pc-empty">暂无职务信息</div>
      <slot name="extra" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
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
}>()
defineEmits<{ (e: 'click', person: any): void }>()

// 在职状态：有快照且 is_active → 在职；有快照且 !is_active → 已离职；无快照（entry_date 为空）→ 未入职
const statusText = computed(() => {
  if (props.person.entry_date == null && !props.person.is_active) return '未入职'
  return props.person.is_active ? '在职' : '已离职'
})
const statusType = computed(() => {
  if (props.person.is_active) return 'success' as const
  if (props.person.entry_date == null) return 'info' as const
  return 'danger' as const
})
</script>

<style lang="scss" scoped>
.person-card {
  width: 220px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .pc-name {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    gap: 8px;

    .pc-status {
      font-weight: 400;
    }
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
}
</style>
