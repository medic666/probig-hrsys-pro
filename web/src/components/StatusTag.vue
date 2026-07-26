<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  type: 'success' | 'warning' | 'danger' | 'info'
}

const props = defineProps<Props>()

const tagType = computed(() => {
  switch (props.type) {
    case 'success': return 'success'
    case 'warning': return 'warning'
    case 'danger': return 'danger'
    case 'info': return 'info'
    default: return 'info'
  }
})
</script>

<template>
  <el-tag :type="tagType" size="small">
    <slot />
  </el-tag>
</template>
