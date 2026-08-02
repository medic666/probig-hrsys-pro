<template>
  <div>
    <h4 v-if="title" class="lb-title">{{ title }}</h4>
    <el-descriptions v-if="rows.length" :column="1" border size="small" label-width="110px">
      <el-descriptions-item v-for="r in rows" :key="r.label" :label="r.label">
        {{ r.value }}
      </el-descriptions-item>
    </el-descriptions>
    <el-empty v-else :description="emptyText" :image-size="50" />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    rows: { label: string; value: string }[]
    emptyText?: string
  }>(),
  { title: '', emptyText: '暂无记录' },
)
</script>

<style scoped>
.lb-title {
  font-size: 13px;
  color: #606266;
  margin: 0 0 8px;
}
.lb-title + .el-descriptions,
.lb-title + .el-empty {
  margin-bottom: 12px;
}
</style>
