<template>
  <el-drawer
    :model-value="visible"
    title="回收站"
    size="600px"
    @close="handleClose"
  >
    <el-table v-loading="loading" :data="tableData" border stripe>
      <el-table-column
        v-for="col in columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
      >
        <template #default="scope">
          {{ col.formatter ? col.formatter(scope.row) : scope.row[col.prop] }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="scope">
          <el-button type="primary" link size="small" @click="handleRestore(scope.row)">
            恢复
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && tableData.length === 0" description="回收站为空" :image-size="60" />

    <div v-if="tableData.length > 0" class="trash-pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        :total="total"
        :layout="paginationLayout"
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>
    <slot name="footer" />
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBreakpoint } from '@/composables/useBreakpoint'

export interface TrashColumn {
  prop: string
  label: string
  width?: string | number
  minWidth?: string | number
  formatter?: (row: any) => string
}

const props = defineProps<{
  visible: boolean
  fetchApi: (params: any) => Promise<{ list: any[]; total: number }>
  restoreApi: (id: number) => Promise<any>
  columns: TrashColumn[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'restored'): void
}>()

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 分页自适应：移动端精简布局，避免抽屉内溢出
const { isMobile } = useBreakpoint()
const paginationLayout = computed(() => (isMobile.value ? 'prev, pager, next' : 'total, sizes, prev, pager, next'))

async function loadData() {
  loading.value = true
  try {
    const result = await props.fetchApi({
      pageNum: currentPage.value,
      pageSize: pageSize.value,
    })
    tableData.value = result.list || []
    total.value = result.total || 0
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function handleRestore(row: any) {
  try {
    await ElMessageBox.confirm('确认恢复该记录？', '提示', {
      type: 'warning',
      confirmButtonText: '确认恢复',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  loading.value = true
  try {
    await props.restoreApi(row.id)
    ElMessage.success('恢复成功')
    emit('restored')
    await loadData()
  } catch {
    // error handled by request interceptor
  } finally {
    loading.value = false
  }
}

function handleClose() {
  emit('update:visible', false)
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      currentPage.value = 1
      loadData()
    }
  },
)
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.trash-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;

  @include mobile {
    justify-content: center;
  }
}
</style>
