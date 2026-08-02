<template>
  <div v-loading="loading" style="min-height:120px">
    <div class="card-grid">
      <slot v-for="item in items" :item="item" />
    </div>
    <el-empty v-if="!loading && items.length === 0" :description="emptyText" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const props = withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    emptyText?: string
    pageSize?: number
  }>(),
  {
    emptyText: '暂无数据',
    pageSize: 100,
  },
)

const items = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const d = await props.fetchFn({ pageNum: 1, pageSize: props.pageSize })
    items.value = d.list || []
  } catch {
    items.value = []
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
