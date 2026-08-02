<template>
  <div class="company-card" @click="$emit('click', company)">
    <div class="cc-name">{{ company.name }}</div>
    <div v-if="company.credit_code" class="cc-line">信用代码：{{ company.credit_code }}</div>
    <div v-if="!company.credit_code" class="cc-line cc-empty">暂无信用代码</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  company: { id: number; name: string; credit_code?: string }
}>()
defineEmits<{ (e: 'click', company: any): void }>()
</script>

<style lang="scss" scoped>
.company-card {
  width: 220px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .cc-name {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
  }

  .cc-line {
    font-size: 12px;
    color: #606266;
    line-height: 20px;

    &.cc-empty {
      color: #c0c4cc;
    }
  }
}
</style>
