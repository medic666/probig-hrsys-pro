<template>
  <div class="page-container">
    <PageHeader title="公司管理">
      <template #actions>
                <ViewModeSwitch v-model="viewMode" card-value="cards" />
      </template>
    </PageHeader>
    <PageToolbar :right-visible="isList">
      <el-button type="primary" size="small" @click="handleAdd">新增公司</el-button>
      <template #right>
        <el-button size="small" @click="handleExport">导出</el-button>
        <el-button size="small" @click="trashVisible = true">回收站</el-button>
      </template>
    </PageToolbar>
    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchCompanies" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
          <el-button type="success" link size="small" @click="handleDetail(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </ProTable>
    </template>
    <template v-else>
      <CardGrid ref="cardGridRef" :fetch-fn="fetchCompanyCards">
        <template #default="{ item }">
          <CompanyCard :company="item" @click="handleDetail" />
        </template>
      </CardGrid>
    </template>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restoreCompany" :columns="trashColumns" @restored="onTrashRestored" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import CompanyCard from '@/components/cards/CompanyCard.vue'
import CardGrid from '@/components/cards/CardGrid.vue'
import PageHeader from '@/components/PageHeader.vue'
import ViewModeSwitch from '@/components/ViewModeSwitch.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import { getCompanies, deleteCompany, restoreCompany, getDeletedCompanies, getAllCompanies, exportCompanies } from '@/api/company'

import { usePageView } from '@/composables/usePageView'
import { useExport } from '@/composables/useExport'

const router = useRouter()
const tableRef = ref()
const { viewMode, isList } = usePageView('cards')
const cardGridRef = ref()
const trashVisible = ref(false)

const columns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '公司名称', width: '180' },
  { prop: 'credit_code', label: '统一社会信用代码', width: '180' },
  { prop: 'contact_phone', label: '联系电话', width: '130' },
  { prop: 'bank_name', label: '开户行', width: '160' },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (r: any) => new Date(r.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop:'id', label:'公司名称', type:'name-select' as const, fetchApi: fetchCompanyOptions },
  { prop:'credit_code', label:'统一社会信用代码', type:'input' as const, placeholder:'模糊搜索' },
]

async function fetchCompanyOptions(k?: string) {
  const list = (await getAllCompanies()) as { id: number; name: string }[] || []
  return k ? list.filter(c => c.name.includes(k)) : list
}

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '公司名称' },
  { prop: 'credit_code', label: '统一社会信用代码' },
]

async function fetchCompanies(params: any) { return (await getCompanies(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedCompanies(params)) as any }

async function fetchCompanyCards() {
  const d = (await getCompanies({ pageNum: 1, pageSize: 100 })) as any
  return { list: d.list || [], total: d.total || 0 }
}

// 新增=编辑=查看统一走公司业务逻辑页
function handleAdd() {
  router.push('/company/create')
}

function handleDetail(row: any) {
  router.push(`/company/${row.id}`)
}

// 导出严格关联列表视图的当前筛选（导出按钮仅列表视图展示）
const { run: handleExport } = useExport(exportCompanies, () => tableRef.value?.getSearchParams() || {})

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' }) } catch { return }
  try { await deleteCompany(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh(); cardGridRef.value?.reload() } catch { /* handled */ }
}

function onTrashRestored() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }
</style>
