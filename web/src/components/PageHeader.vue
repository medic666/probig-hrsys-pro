<template>
  <div class="page-header">
    <h2>{{ title }}</h2>
    <div class="page-header-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;

  h2 {
    font-size: 18px;
    font-weight: 600;
    color: #303133;
    margin: 0;
  }

  .page-header-actions {
    margin-left: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
  }
}
</style>
