<template>
  <el-button class="page-back-btn" size="small" @click="$emit('click')">
    <el-icon><ArrowLeft /></el-icon>
    返回
  </el-button>
</template>

<script setup lang="ts">
import { ArrowLeft } from '@element-plus/icons-vue'

defineEmits<{ (e: 'click'): void }>()
</script>
