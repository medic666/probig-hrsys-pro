<template>
  <div class="page-container">
    <PageHeader title="人员管理">
      <template #actions>
        <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>

    <PageToolbar :right-visible="isList">
      <el-button type="primary" size="small" @click="handleAdd">新增人员</el-button>
      <template #right>
        <el-button size="small" @click="handleExport">导出</el-button>
        <el-button size="small" @click="trashVisible = true">回收站</el-button>
      </template>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchPersons" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
          <el-button type="success" link size="small" @click="handleDetail(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <PersonCardGrid ref="cardGridRef" :fetch-fn="fetchCards" :dot-color-of="dotColorOf" @select="handleDetail" />
    </template>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restorePerson" :columns="trashColumns" @restored="onTrashRestored" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import PersonCardGrid from '@/components/cards/PersonCardGrid.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import { getPersons, deletePerson, restorePerson, getDeletedPersons, getPersonCards, exportPersons } from '@/api/person'
import { getAllCompanies } from '@/api/company'
import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'

const router = useRouter()
const tableRef = ref()
const { viewMode, isList } = usePageView('cards')
const cardGridRef = ref()
const trashVisible = ref(false)
const companyOptions = ref<{ label: string; value: number }[]>([])

const columns = [
  { prop: 'is_active', label: '在职状态', width: '90', formatter: (r: any) => (r.is_active === null ? '未入职' : r.is_active ? '在职' : '已离职') },
  { prop: 'name', label: '姓名', width: '100' },
  { prop: 'company_name', label: '公司', width: '120', formatter: (r: any) => r.company_name || '-' },
  { prop: 'department', label: '部门', width: '110', formatter: (r: any) => r.department || '-' },
  { prop: 'position', label: '职位', width: '110', formatter: (r: any) => r.position || '-' },
  { prop: 'entry_date', label: '入职时间', width: '110', formatter: (r: any) => r.entry_date || '-' },
]

const searchFields = computed(() => [
  { prop: 'company_id', label: '公司', type: 'select' as const, options: companyOptions.value },
  { prop: 'department', label: '部门', type: 'input' as const, placeholder: '模糊搜索' },
  {
    prop: 'status', label: '在职状态', type: 'select' as const,
    options: [
      { label: '在职', value: 'active' },
      { label: '已离职', value: 'left' },
      { label: '未入职', value: 'not_entered' },
    ],
  },
])

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '姓名' },
  { prop: 'id_card', label: '身份证号' },
]

onMounted(async () => {
  const companies = (await getAllCompanies()) as { id: number; name: string }[]
  companyOptions.value = companies.map((c) => ({ label: c.name, value: c.id }))
})

async function fetchPersons(params: any) { return (await getPersons(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedPersons(params)) as any }
async function fetchCards() {
  // 拉取全量人员卡片，活跃/全部过滤由 PersonCardGrid 响应式派生
  const cards = (await getPersonCards()) as any[] || []
  return { list: cards, total: cards.length }
}

// 在职状态颜色点：未入职灰 / 在职绿 / 已离职红（非活跃离职>3月由 PersonCard 内部自动不显示）
function dotColorOf(p: any): string {
  if (p.entry_date == null && !p.is_active) return 'gray'
  return p.is_active ? 'green' : 'red'
}

// 新增=编辑=查看统一走人员聚合页
function handleAdd() {
  router.push('/person/create')
}

function handleDetail(row: any) {
  router.push(`/person/${row.id}`)
}

// 导出严格关联列表视图的当前筛选（导出按钮仅列表视图展示）
const { run: handleExport } = useExport(exportPersons, () => tableRef.value?.getSearchParams() || {})

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' }) } catch { return }
  try { await deletePerson(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh(); cardGridRef.value?.reload() } catch { /* handled */ }
}

function onTrashRestored() { tableRef.value?.refresh(); cardGridRef.value?.reload() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }
</style>
