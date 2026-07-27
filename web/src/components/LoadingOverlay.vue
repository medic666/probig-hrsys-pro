<template>
  <div v-loading="loading" :element-loading-text="text" class="loading-overlay">
    <slot v-if="!loading" />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    loading: boolean
    text?: string
  }>(),
  {
    text: '加载中...',
  },
)
</script>

<style lang="scss" scoped>
.loading-overlay {
  min-height: 200px;
}
</style>
