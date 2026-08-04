<template>
  <el-radio-group :model-value="modelValue" size="small" @update:model-value="handleChange">
    <el-radio-button :value="cardValue">{{ cardLabel }}</el-radio-button>
    <el-radio-button value="list">列表</el-radio-button>
  </el-radio-group>
</template>

<script setup lang="ts">
// 卡片/列表双视图切换（与 usePageView 配合）：
// cardValue 为卡片视图标识（'cards'/'blocks'），列表视图标识恒为 'list'；
// URL 同步由页面 usePageView 的 watch 完成，本组件仅渲染切换控件。
withDefaults(
  defineProps<{
    modelValue: string
    cardValue?: string
    cardLabel?: string
  }>(),
  { cardValue: 'cards', cardLabel: '卡片' },
)

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

function handleChange(v: string) {
  emit('update:modelValue', v)
}
</script>
