<script setup lang="ts" generic="T">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'

interface ColumnConfig {
  prop: string
  label: string
  width?: string | number
  minWidth?: string | number
  fixed?: string | boolean
  sortable?: boolean
  formatter?: (row: T, column: ColumnConfig) => string
  slot?: string
}

interface SearchField {
  prop: string
  label: string
  type: 'input' | 'select' | 'date' | 'daterange' | 'month' | 'name-select'
  placeholder?: string
  options?: { label: string; value: string | number }[]
  nameType?: string
}

interface Props {
  columns: ColumnConfig[]
  searchFields: SearchField[]
  api: (params: Record<string, unknown>) => Promise<{ list: T[]; total: number }>
  defaultParams?: Record<string, unknown>
  showSelection?: boolean
  rowKey?: string
  showActions?: boolean
  actionWidth?: string | number
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  defaultParams: () => ({}),
  showSelection: false,
  rowKey: 'id',
  showActions: true,
  actionWidth: 200,
  pageSize: 10
})

const emit = defineEmits<{
  (e: 'selection-change', rows: T[]): void
  (e: 'row-click', row: T): void
}>()

const tableData = ref<T[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(props.pageSize)
const selectedRows = ref<T[]>([])

const searchForm = reactive<Record<string, unknown>>({})
const searchFormRef = ref()

function buildParams() {
  const params: Record<string, unknown> = {
    pageNum: currentPage.value,
    pageSize: pageSize.value,
    ...props.defaultParams
  }
  for (const key of Object.keys(searchForm)) {
    if (searchForm[key] !== '' && searchForm[key] !== undefined && searchForm[key] !== null) {
      params[key] = searchForm[key]
    }
  }
  return params
}

async function fetchData() {
  loading.value = true
  try {
    const res = await props.api(buildParams())
    tableData.value = res.list
    total.value = res.total
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchData()
}

function handleReset() {
  for (const key of Object.keys(searchForm)) {
    searchForm[key] = undefined
  }
  currentPage.value = 1
  fetchData()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchData()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  fetchData()
}

function handleSelectionChange(rows: T[]) {
  selectedRows.value = rows
  emit('selection-change', rows)
}

function handleRowClick(row: T) {
  emit('row-click', row)
}

function refresh() {
  fetchData()
}

defineExpose({ refresh, getData: fetchData, selectedRows })

onMounted(() => {
  fetchData()
})

watch(() => props.defaultParams, () => {
  currentPage.value = 1
  fetchData()
}, { deep: true })
</script>

<template>
  <div class="pro-table">
    <div v-if="searchFields.length > 0" class="pro-table-search">
      <el-form ref="searchFormRef" :model="searchForm" inline>
        <el-form-item
          v-for="field in searchFields"
          :key="field.prop"
          :label="field.label"
        >
          <el-input
            v-if="field.type === 'input'"
            v-model="searchForm[field.prop]"
            :placeholder="field.placeholder || '请输入'"
            clearable
            style="width: 180px"
          />
          <el-select
            v-else-if="field.type === 'select'"
            v-model="searchForm[field.prop]"
            :placeholder="field.placeholder || '请选择'"
            clearable
            style="width: 180px"
          >
            <el-option
              v-for="opt in field.options"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-date-picker
            v-else-if="field.type === 'date'"
            v-model="searchForm[field.prop]"
            type="date"
            :placeholder="field.placeholder || '选择日期'"
            value-format="YYYY-MM-DD"
            style="width: 180px"
          />
          <el-date-picker
            v-else-if="field.type === 'month'"
            v-model="searchForm[field.prop]"
            type="month"
            :placeholder="field.placeholder || '选择月份'"
            value-format="YYYY-MM"
            style="width: 180px"
          />
          <el-date-picker
            v-else-if="field.type === 'daterange'"
            v-model="searchForm[field.prop]"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="pro-table-body">
      <el-table
        :data="tableData"
        v-loading="loading"
        :row-key="rowKey"
        @selection-change="handleSelectionChange"
        @row-click="handleRowClick"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column
          v-if="showSelection"
          type="selection"
          width="55"
          :reserve-selection="true"
        />
        <template v-for="col in columns" :key="col.prop">
          <el-table-column
            v-if="!col.slot"
            :prop="col.prop"
            :label="col.label"
            :width="col.width"
            :min-width="col.minWidth"
            :fixed="col.fixed"
            :sortable="col.sortable"
            :formatter="col.formatter"
          />
          <el-table-column
            v-else
            :label="col.label"
            :width="col.width"
            :min-width="col.minWidth"
            :fixed="col.fixed"
          >
            <template #default="scope">
              <slot :name="col.slot" :row="scope.row" :$index="scope.$index" />
            </template>
          </el-table-column>
        </template>
      </el-table>
    </div>

    <div class="pro-table-pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.pro-table {
  background: #fff;
  padding: 16px;
  border-radius: 4px;

  &-search {
    margin-bottom: 16px;
  }

  &-body {
    margin-bottom: 16px;
  }

  &-pagination {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
