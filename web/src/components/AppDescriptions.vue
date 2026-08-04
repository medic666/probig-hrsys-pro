<template>
  <el-descriptions :column="col" v-bind="$attrs">
    <slot />
  </el-descriptions>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useBreakpoint } from '@/composables/useBreakpoint'

// AppDescriptions 响应式列数描述组件：桌面按 column 传入，平板最多 2 列、移动端 1 列。
// 与 el-descriptions 用法完全一致（border/size/label-width 等透传）。
defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{ column?: number }>(), { column: 3 })

const { isMobile, isTablet } = useBreakpoint()

const col = computed(() => {
  if (isMobile.value) return 1
  if (isTablet.value) return Math.min(props.column, 2)
  return props.column
})
</script>
