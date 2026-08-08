<template>
  <AppDescriptions :column="column" border size="small">
    <slot name="prefix" />
    <el-descriptions-item v-for="f in fields" :key="f.key" :label="f.label">
      {{ f.formatter ? f.formatter(data) : data[f.key] }}
    </el-descriptions-item>
    <slot name="suffix" />
  </AppDescriptions>
</template>

<script setup lang="ts">
import AppDescriptions from '@/components/AppDescriptions.vue'
import type { FieldDef } from '@/constants/fields'

// 字段描述标准模块：由统一字段表（constants/fields.ts）驱动渲染明细项，
// 明细/版本展开/追溯等展示与列表、导出共用同一份字段定义。
defineProps<{
  fields: FieldDef[]
  data: any
  column?: number
}>()
</script>
