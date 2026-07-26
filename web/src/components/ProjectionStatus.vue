<template>
  <el-tag :type="tagType" size="small">{{ label }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: number
}>()

const tagType = computed(() => {
  switch (props.status) {
    case 0: return 'info'
    case 1: return 'success'
    case 2: return 'warning'
    default: return 'info'
  }
})

const label = computed(() => {
  switch (props.status) {
    case 0: return '未核算'
    case 1: return '已核算'
    case 2: return '数据已变动'
    default: return '未知'
  }
})
</script>
