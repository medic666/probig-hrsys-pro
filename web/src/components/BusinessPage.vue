<template>
  <div class="business-page">
    <PageHeader :title="title">
      <template #back>
        <PageBackButton @click="goBack" />
      </template>
    </PageHeader>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import PageHeader from '@/components/PageHeader.vue'
import PageBackButton from '@/components/PageBackButton.vue'

const props = withDefaults(
  defineProps<{
    title: string
    // 无历史记录（直接输入 URL 进入）时的回退目标
    backTo: string
  }>(),
  { backTo: '/' },
)

const router = useRouter()

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace(props.backTo)
  }
}
</script>

<style scoped>
.business-page {
  .page-container {
    padding: 0;
    background: transparent;
  }
}
</style>
