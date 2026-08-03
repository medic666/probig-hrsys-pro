<template>
  <div v-loading="loading" style="min-height:120px">
    <div class="card-grid">
      <slot v-for="item in items" :item="item" />
    </div>
    <el-empty v-if="!loading && items.length === 0" :description="emptyText" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// 通用卡片网格：fetchFn 只负责拉取全量数据（一次），filterFn 负责展示过滤——
// 纯响应式派生（依赖外部状态的闭包，外部状态变化即重算，不重复请求后端）。
// 与 TimeCardPanel 的 computed 派生模式同构。
const props = withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    filterFn?: (items: any[]) => any[]
    emptyText?: string
    pageSize?: number
  }>(),
  {
    emptyText: '暂无数据',
    pageSize: 100,
    filterFn: undefined,
  },
)

const allItems = ref<any[]>([])
const loading = ref(false)
const items = computed(() => (props.filterFn ? props.filterFn(allItems.value) : allItems.value))

async function load() {
  loading.value = true
  try {
    const d = await props.fetchFn({ pageNum: 1, pageSize: props.pageSize })
    allItems.value = d.list || []
  } catch {
    allItems.value = []
  } finally {
    loading.value = false
  }
}

function reload() {
  load()
}

onMounted(() => {
  load()
})

defineExpose({ reload })
</script>

<style lang="scss" scoped>
.card-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
</style>
