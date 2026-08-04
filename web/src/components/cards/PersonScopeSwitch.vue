<template>
  <div class="person-scope-switch">
    <el-radio-group :model-value="modelValue" size="small" @update:model-value="handleChange">
      <el-radio-button value="active">活跃人员</el-radio-button>
      <el-radio-button value="all">全部人员</el-radio-button>
    </el-radio-group>
  </div>
</template>

<script setup lang="ts">
import type { PersonScope } from '@/utils/personScope'

defineProps<{ modelValue: PersonScope }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: PersonScope): void }>()

function handleChange(v: PersonScope) {
  emit('update:modelValue', v)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.person-scope-switch {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 100;

  @include mobile {
    right: 12px;
    bottom: calc(12px + env(safe-area-inset-bottom));
  }
}
</style>
