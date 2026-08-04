<template>
  <div class="page-toolbar">
    <div class="toolbar-left">
      <slot />
    </div>
    <div v-if="rightVisible" class="toolbar-right">
      <slot name="right" />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    // 右区（回收站/导出等列表视图精细操作）是否可见
    rightVisible?: boolean
  }>(),
  { rightVisible: true },
)
</script>

<style scoped>
.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;

  .toolbar-left,
  .toolbar-right {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .toolbar-right {
    margin-left: auto;
  }
}
</style>
