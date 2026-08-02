<template>
  <el-tag :type="statusType" size="small">{{ statusText }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  person: {
    is_active?: boolean
    entry_date?: string | null
    leave_date?: string | null
  }
}>()

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
