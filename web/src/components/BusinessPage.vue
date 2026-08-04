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
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '@/components/PageHeader.vue'
import PageBackButton from '@/components/PageBackButton.vue'
import { resolveBackTarget } from '@/composables/useBusinessPage'

// 业务逻辑页外壳：标题与返回回退目标缺省读路由 meta（title / backTo），页面可显式覆盖。
// 返回目标：query.back（路由守卫注入的来源页）→ backTo（模块列表）→ 首页。
const props = withDefaults(
  defineProps<{
    title?: string
    backTo?: string
  }>(),
  { title: '', backTo: '' },
)

const route = useRoute()
const router = useRouter()

const title = computed(() => props.title || String(route.meta.title || '业务详情'))
const backTo = computed(() => props.backTo || String(route.meta.backTo || '/'))
const backTarget = computed(() => resolveBackTarget(route, backTo.value))

function goBack() {
  router.replace(backTarget.value)
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
