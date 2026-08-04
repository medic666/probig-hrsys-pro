<template>
  <el-dialog v-model="visible" title="文件附件" width="500px">
    <FileAttachPanel :target-type="targetType" :target-id="targetId" />
  </el-dialog>
</template>

<script setup lang="ts">
import FileAttachPanel from '@/components/FileAttachPanel.vue'

// 文件附件通用对话框：各业务行"附件"按钮统一弹窗载体（target-type/target-id 由页面注入）
defineProps<{ targetType: string; targetId: number | null }>()

const visible = defineModel<boolean>('visible', { default: false })
</script>
