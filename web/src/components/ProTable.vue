<template>
  <div>
    <div v-if="$slots.search" class="search-bar">
      <slot name="search" />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>
    <div v-if="$slots.toolbar" class="tool-bar">
      <slot name="toolbar" />
    </div>
    <el-table
      v-loading="loading"
      :data="data"
      :border="border"
      stripe
      @selection-change="handleSelectionChange"
      v-bind="$attrs"
    >
      <el-table-column v-if="showSelection" type="selection" width="50" />
      <slot />
    </el-table>
    <div v-if="total === 0 && !loading" style="padding: 40px 0; text-align: center;">
      <el-empty description="暂无数据" />
    </div>
    <div v-if="total > 0" style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSearch"
        @current-change="handleSearch"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  data: any[]
  total: number
  loading?: boolean
  border?: boolean
  showSelection?: boolean
}>(), {
  loading: false,
  border: true,
  showSelection: false,
})

const emit = defineEmits<{
  search: [params: { pageNum: number; pageSize: number }]
  'update:selection': [selection: any[]]
}>()

const currentPage = ref(1)
const pageSize = ref(20)

function handleSearch() {
  emit('search', { pageNum: currentPage.value, pageSize: pageSize.value })
}

function handleReset() {
  currentPage.value = 1
  emit('search', { pageNum: 1, pageSize: pageSize.value })
}

function handleSelectionChange(selection: any[]) {
  emit('update:selection', selection)
}

defineExpose({ handleSearch, currentPage, pageSize })
</script>
