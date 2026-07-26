<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  status: string
}

const props = defineProps<Props>()

const statusConfig = computed(() => {
  switch (props.status) {
    case 'calculated':
    case '已核算':
      return { type: 'success' as const, text: '已核算' }
    case 'stale':
    case '数据已变动':
      return { type: 'warning' as const, text: '数据已变动' }
    case 'not_calculated':
    case '未核算':
    default:
      return { type: 'info' as const, text: '未核算' }
  }
})
</script>

<template>
  <el-tag :type="statusConfig.type" size="small">
    {{ statusConfig.text }}
  </el-tag>
</template>
