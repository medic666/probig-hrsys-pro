<template>
  <div class="page-container">
    <div class="page-header"><h2>公司管理</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchCompanies" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
        <el-button type="success" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog v-model:visible="dialogVisible" :title="dialogMode === 'add' ? '新增公司' : '编辑公司'" :mode="dialogMode" :form-fields="companyFormFields" :rules="formRules" :submit-api="submitForm" :edit-data="editRow" @success="onFormSuccess" />

    <el-dialog v-model="detailVisible" title="公司详情" width="600px" @close="detailRow = null">
      <template v-if="detailRow">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基础信息" name="info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="公司名称">{{ detailRow.name }}</el-descriptions-item>
              <el-descriptions-item label="统一社会信用代码">{{ detailRow.credit_code || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地址" :span="2">{{ detailRow.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ detailRow.contact_phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="开户行">{{ detailRow.bank_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="银行账号" :span="2">{{ detailRow.bank_account || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane label="附件" name="files">
            <FileAttachPanel v-if="detailRow" :target-type="'company'" :target-id="detailRow.id" />
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restoreCompany" :columns="trashColumns" @restored="onTrashRestored" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getCompanies, getCompany, createCompany, updateCompany, deleteCompany, restoreCompany, getDeletedCompanies, getAllCompanies, exportCompanies } from '@/api/company'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editRow = ref<any>(null)
const detailVisible = ref(false)
const detailRow = ref<any>(null)
const activeTab = ref('info')
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

const actions = [
  { key: 'add', label: '新增公司', type: 'primary' as const },
  { key: 'export', label: '导出', type: 'default' as const },
  { key: 'trash', label: '回收站', type: 'default' as const },
]

const companyFormFields = [
  { prop: 'name', label: '公司名称', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'credit_code', label: '统一社会信用代码', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'address', label: '地址', type: 'input' as const, placeholder: '请输入', span: 24 },
  { prop: 'contact_phone', label: '联系电话', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'bank_name', label: '开户行', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'bank_account', label: '银行账号', type: 'input' as const, placeholder: '请输入', span: 12 },
]

const formRules = {
  name: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
}

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '公司名称' },
  { prop: 'credit_code', label: '统一社会信用代码' },
]

async function fetchCompanies(params: any) { return (await getCompanies(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedCompanies(params)) as any }

function handleAction(key: string) {
  if (key === 'add') { dialogMode.value = 'add'; editRow.value = null; dialogVisible.value = true }
  else if (key === 'export') { handleExport() }
  else if (key === 'trash') { trashVisible.value = true }
}

async function handleExport() {
  const data = await exportCompanies({})
  downloadBlob(data)
}

function handleEdit(row: any) { dialogMode.value = 'edit'; editRow.value = row; dialogVisible.value = true }

async function handleDetail(row: any) {
  try {
    detailRow.value = (await getCompany(row.id)) as any
    activeTab.value = 'info'
    detailVisible.value = true
  } catch { /* handled */ }
}

async function submitForm(data: any) {
  if (dialogMode.value === 'add') return createCompany(data)
  return updateCompany(editRow.value.id, data)
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' }) } catch { return }
  try { await deleteCompany(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh() } catch { /* handled */ }
}

function onFormSuccess() { tableRef.value?.refresh() }
function onTrashRestored() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
</style>
