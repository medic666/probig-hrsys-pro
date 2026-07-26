<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, View, RefreshLeft } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listCompanies, createCompany, updateCompany, deleteCompany, restoreCompany, listCompanyTrash } from '@/api/company'
import type { Company, CompanyListParams } from '@/api/company'

const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增公司')
const editingRow = ref<Company | null>(null)
const trashVisible = ref(false)

const searchFields = [
  { prop: 'name', label: '公司名称', type: 'input' as const, placeholder: '请输入公司名称' },
  { prop: 'credit_code', label: '统一社会信用代码', type: 'input' as const, placeholder: '请输入信用代码' }
]

const columns = [
  { prop: 'name', label: '公司名称' },
  { prop: 'credit_code', label: '统一社会信用代码' },
  { prop: 'contact_phone', label: '联系电话' },
  { prop: 'address', label: '地址' },
  { slot: 'actions', label: '操作', width: 200, fixed: 'right' as const }
]

const formFields = [
  { prop: 'name', label: '公司名称', type: 'input' as const, required: true },
  { prop: 'credit_code', label: '统一社会信用代码', type: 'input' as const, required: true },
  { prop: 'address', label: '地址', type: 'textarea' as const },
  { prop: 'contact_phone', label: '联系电话', type: 'input' as const },
  { prop: 'bank_name', label: '开户行', type: 'input' as const },
  { prop: 'bank_account', label: '银行账号', type: 'input' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listCompanies(params as unknown as CompanyListParams)
}

async function fetchTrash(params: Record<string, unknown>) {
  return listCompanyTrash({ pageNum: params.pageNum as number, pageSize: params.pageSize as number, name: params.name as string })
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增公司'
  formVisible.value = true
}

function handleEdit(row: Company) {
  editingRow.value = row
  formTitle.value = '编辑公司'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { ...editingRow.value }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updateCompany(data as any)
  } else {
    await createCompany(data as any)
  }
}

async function handleDelete(row: Company) {
  try {
    await ElMessageBox.confirm(`确定要删除公司「${row.name}」吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteCompany(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

async function handleRestore(row: Company) {
  try {
    await ElMessageBox.confirm(`确定要恢复公司「${row.name}」吗？`, '确认恢复', { type: 'warning' })
  } catch {
    return
  }
  await restoreCompany(row.id)
  ElMessage.success('恢复成功')
  tableRef.value?.refresh()
}

function handleFormSuccess() {
  tableRef.value?.refresh()
}

const trashColumns = [
  { prop: 'name', label: '公司名称' },
  { prop: 'credit_code', label: '统一社会信用代码' },
  { slot: 'trashActions', label: '操作', width: 120, fixed: 'right' as const }
]
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增公司</el-button>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Delete" @click="trashVisible = true">回收站</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="View" @click="handleEdit(row)">详情</el-button>
        <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link :icon="Delete" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="formVisible"
      :title="formTitle"
      :form-fields="formFields"
      :initial-data="getInitialData()"
      :submit-api="handleSubmit"
      @success="handleFormSuccess"
    />

    <el-drawer v-model="trashVisible" title="回收站" size="800px">
      <ProTable
        :columns="trashColumns"
        :search-fields="searchFields.slice(0, 1)"
        :api="fetchTrash"
      >
        <template #trashActions="{ row }">
          <el-button type="primary" link :icon="RefreshLeft" @click="handleRestore(row)">恢复</el-button>
        </template>
      </ProTable>
    </el-drawer>
  </div>
</template>
