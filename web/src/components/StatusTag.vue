<template>
  <el-tag :type="tagType" size="small">
    {{ tagText }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String as () => 'not_calculated' | 'calculated' | 'data_changed',
    default: 'not_calculated',
  },
  text: {
    type: String,
    default: '',
  },
})

const tagType = computed(() => {
  switch (props.status) {
    case 'calculated':
      return 'success'
    case 'data_changed':
      return 'warning'
    default:
      return 'info'
  }
})

const tagText = computed(() => {
  if (props.text) return props.text
  switch (props.status) {
    case 'calculated':
      return '已核算'
    case 'data_changed':
      return '数据已变动'
    default:
      return '未核算'
  }
})
</script>
