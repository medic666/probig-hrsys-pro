<template>
  <div class="pro-table">
    <div v-if="searchFields && searchFields.length" class="search-bar">
      <el-form :model="searchForm" inline>
        <el-form-item v-for="field in searchFields" :key="field.prop" :label="field.label">
          <template v-if="field.type === 'input'">
            <el-input v-model="searchForm[field.prop]" :placeholder="field.placeholder || '请输入'" clearable style="width: 180px" />
          </template>
          <template v-else-if="field.type === 'select'">
            <el-select v-model="searchForm[field.prop]" :placeholder="field.placeholder || '请选择'" clearable style="width: 180px">
              <el-option v-for="opt in field.options" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </template>
          <template v-else-if="field.type === 'date-range'">
            <el-date-picker v-model="searchForm[field.prop]" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 260px" value-format="YYYY-MM-DD" />
          </template>
          <template v-else-if="field.type === 'month-range'">
            <el-date-picker v-model="searchForm[field.prop]" type="monthrange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 260px" value-format="YYYY-MM" />
          </template>
          <template v-else-if="field.type === 'person-select' || field.type === 'name-select'">
            <NameSelect v-model="searchForm[field.prop]" :fetch-api="field.fetchApi" :placeholder="field.placeholder || '选择'" style="width:200px" />
          </template>
          <template v-else-if="field.type === 'month'">
            <el-date-picker v-model="searchForm[field.prop]" type="month" value-format="YYYY-MM" :placeholder="field.placeholder || '选择月份'" style="width:200px" />
          </template>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-if="showActions" class="action-bar">
      <el-button v-for="action in actions" :key="action.key" :type="action.type || 'primary'" @click="handleAction(action.key)">
        <el-icon v-if="action.icon"><component :is="action.icon" /></el-icon>
        {{ action.label }}
      </el-button>
    </div>

    <div v-if="batchActions && batchActions.length && selectedRows.length" class="batch-bar">
      <span class="batch-info">已选择 {{ selectedRows.length }} 项</span>
      <el-button v-for="action in batchActions" :key="action.key" :type="action.type || 'default'" @click="handleBatchAction(action.key)">
        {{ action.label }}
      </el-button>
    </div>

    <el-table
      ref="tableRef"
      v-loading="loading"
      :data="tableData"
      :element-loading-text="loadingText"
      border
      stripe
      class="data-table"
      @selection-change="onSelectionChange"
    >
      <el-table-column v-if="showSelection" type="selection" width="50" />
      <el-table-column
        v-for="col in columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :fixed="col.fixed"
        :min-width="col.minWidth"
      >
        <template v-if="col.slot" #default="scope">
          <slot :name="col.slot" :row="scope.row" />
        </template>
        <template v-else #default="scope">
          {{ col.formatter ? col.formatter(scope.row) : scope.row[col.prop] }}
        </template>
      </el-table-column>
      <el-table-column v-if="$slots.actions" label="操作" fixed="right" min-width="180">
        <template #default="scope">
          <slot name="actions" :row="scope.row" />
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && tableData.length === 0" description="暂无数据" :image-size="80" />

    <div v-if="tableData.length > 0" class="pagination-bar">
      <el-pagination
        v-model:current-page="currentPageNum"
        v-model:page-size="currentPageSize"
        :page-sizes="pageSizes"
        :total="total"
        :layout="paginationLayout"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import type { Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NameSelect from '@/components/NameSelect.vue'
import { useBreakpoint } from '@/composables/useBreakpoint'

export interface TableColumn {
  prop: string
  label: string
  width?: string | number
  minWidth?: string | number
  fixed?: 'left' | 'right'
  formatter?: (row: any) => string
  slot?: string
}

export interface SearchField {
  prop: string
  label: string
  type: 'input' | 'select' | 'date-range' | 'month-range' | 'person-select' | 'name-select' | 'month'
  options?: { label: string; value: any }[]
  placeholder?: string
  fetchApi?: (keyword?: string) => Promise<{ id: number; name: string }[]>
  // 数组型字段（date-range）的后端参数名；缺省为 `${prop}Start`/`${prop}End`
  startKey?: string
  endKey?: string
}

export interface ActionButton {
  key: string
  label: string
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'default'
  icon?: Component
}

const props = withDefaults(
  defineProps<{
    columns: TableColumn[]
    fetchApi: (params?: any) => Promise<{ list: any[]; total: number }>
    searchFields?: SearchField[]
    actions?: ActionButton[]
    batchActions?: ActionButton[]
    defaultSearch?: Record<string, any>
    showSelection?: boolean
    pageSizes?: number[]
    autoLoad?: boolean
    // URL 驱动：筛选与分页状态同步路由 query（view/卡片钻取等其它键保留休眠），
    // 导航往返、视图切换后完整恢复
    urlDriven?: boolean
  }>(),
  {
    searchFields: () => [],
    actions: () => [],
    batchActions: () => [],
    defaultSearch: () => ({}),
    showSelection: false,
    pageSizes: () => [10, 20, 50, 100],
    autoLoad: true,
    urlDriven: false,
  },
)

const emit = defineEmits<{
  (e: 'action', key: string): void
  (e: 'batch-action', key: string, rows: any[]): void
  (e: 'selection-change', rows: any[]): void
}>()

const tableRef = ref()
const loading = ref(false)
const loadingText = ref('加载中...')
const tableData = ref<any[]>([])
const total = ref(0)
const currentPageNum = ref(1)
const currentPageSize = ref(20)
const selectedRows = ref<any[]>([])

const searchForm = reactive<Record<string, any>>({})

initSearchForm()

function initSearchForm() {
  for (const key of Object.keys(searchForm)) {
    delete searchForm[key]
  }
  if (props.searchFields) {
    for (const field of props.searchFields) {
      searchForm[field.prop] = props.defaultSearch?.[field.prop] ?? (
        field.type === 'date-range' ? [] :
        field.type === 'person-select' ? null : ''
      )
    }
  }
}

function fieldKeys(key: string) {
  const field = props.searchFields.find((f) => f.prop === key)
  return {
    startKey: field?.startKey || `${key}Start`,
    endKey: field?.endKey || `${key}End`,
  }
}

const route = useRoute()
const router = useRouter()

// writeUrl 将当前筛选与分页合并写入路由 query（urlDriven 模式，保留其它键休眠）
function writeUrl() {
  if (!props.urlDriven) return
  const query: Record<string, any> = { ...route.query }
  query.pageNum = currentPageNum.value
  query.pageSize = currentPageSize.value
  for (const key of Object.keys(searchForm)) {
    const val = searchForm[key]
    if (val === '' || val === null || val === undefined) {
      delete query[key]
      continue
    }
    if (Array.isArray(val)) {
      const { startKey, endKey } = fieldKeys(key)
      if (val[0]) query[startKey] = val[0]
      else delete query[startKey]
      if (val[1]) query[endKey] = val[1]
      else delete query[endKey]
    } else {
      query[key] = val
    }
  }
  router.replace({ query })
}

// readUrl 从路由 query 恢复筛选与分页
function readUrl() {
  if (!props.urlDriven) return
  const q = route.query as Record<string, any>
  const pn = Number(q.pageNum)
  if (pn > 0) currentPageNum.value = pn
  const ps = Number(q.pageSize)
  if (ps > 0) currentPageSize.value = ps
  for (const field of props.searchFields) {
    if (field.type === 'date-range') {
      const sk = field.startKey || `${field.prop}Start`
      const ek = field.endKey || `${field.prop}End`
      if (q[sk] || q[ek]) {
        searchForm[field.prop] = [q[sk] || '', q[ek] || '']
      }
    } else if (q[field.prop] !== undefined) {
      searchForm[field.prop] = q[field.prop]
    }
  }
}

async function loadData() {
  loading.value = true
  try {
    const params: any = {
      pageNum: currentPageNum.value,
      pageSize: currentPageSize.value,
    }
    for (const key of Object.keys(searchForm)) {
      const val = searchForm[key]
      if (val !== '' && val !== null && val !== undefined) {
        if (Array.isArray(val)) {
          const { startKey, endKey } = fieldKeys(key)
          params[startKey] = val[0] || ''
          params[endKey] = val[1] || ''
        } else {
          params[key] = val
        }
      }
    }
    const result = await props.fetchApi(params)
    tableData.value = result.list || []
    total.value = result.total || 0
    writeUrl()
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPageNum.value = 1
  loadData()
}

function handleReset() {
  initSearchForm()
  currentPageNum.value = 1
  loadData()
}

function handlePageChange() {
  loadData()
}

function handleSizeChange() {
  currentPageNum.value = 1
  loadData()
}

function handleAction(key: string) {
  emit('action', key)
}

function handleBatchAction(key: string) {
  emit('batch-action', key, selectedRows.value)
}

function onSelectionChange(rows: any[]) {
  selectedRows.value = rows
  emit('selection-change', rows)
}

function refresh() {
  loadData()
}

function clearSelection() {
  tableRef.value?.clearSelection()
}

function getSelected() {
  return selectedRows.value
}

function getSearchParams() {
  const params: Record<string, any> = {}
  for (const key of Object.keys(searchForm)) {
    const val = searchForm[key]
    if (val !== '' && val !== null && val !== undefined) {
      if (Array.isArray(val)) {
        const { startKey, endKey } = fieldKeys(key)
        params[startKey] = val[0] || ''
        params[endKey] = val[1] || ''
      } else {
        params[key] = val
      }
    }
  }
  return params
}

const showActions = computed(() => props.actions && props.actions.length > 0)

// 分页自适应：移动端精简（仅上一页/页码/下一页），桌面完整布局
const { isMobile } = useBreakpoint()
const paginationLayout = computed(() =>
  isMobile.value ? 'prev, pager, next' : 'total, sizes, prev, pager, next, jumper',
)

defineExpose({ refresh, clearSelection, getSelected, getSearchParams })

onMounted(() => {
  initSearchForm()
  readUrl()
  if (props.autoLoad) {
    loadData()
  }
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.pro-table {
  .search-bar {
    background: #fff;
    padding: 16px 16px 0;
    border-radius: 4px;
    margin-bottom: 12px;

    :deep(.el-form-item) {
      margin-bottom: 16px;
    }
  }

  .action-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }

  .batch-bar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    padding: 10px 16px;
    background: #ecf5ff;
    border-radius: 4px;
    margin-bottom: 12px;

    .batch-info {
      color: #409eff;
      font-size: 13px;
    }
  }

  .data-table {
    background: #fff;
    border-radius: 4px;
  }

  .pagination-bar {
    display: flex;
    justify-content: flex-end;
    padding: 16px 0 0;
    background: #fff;
    margin-top: -1px;
    border-radius: 0 0 4px 4px;

    @include mobile {
      justify-content: center;
      padding-bottom: 4px;
    }
  }
}
</style>
