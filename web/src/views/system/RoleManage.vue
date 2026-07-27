<template>
  <div class="page-container">
    <div class="page-header">
      <h2>角色管理</h2>
    </div>
    <ProTable
      ref="tableRef"
      :columns="columns"
      :fetch-api="fetchRoles"
      :search-fields="searchFields"
      :actions="actions"
      @action="handleAction"
    >
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button type="warning" link size="small" @click="handleAssignPerms(row)">权限分配</el-button>
        <el-button v-if="!row.is_default" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="dialogVisible"
      :title="dialogMode === 'add' ? '新增角色' : '编辑角色'"
      :mode="dialogMode"
      :form-fields="formFields"
      :rules="formRules"
      :submit-api="submitForm"
      :edit-data="editRow"
      @success="onFormSuccess"
    />

    <el-dialog v-model="permDialogVisible" title="权限分配" width="500px">
      <el-tree
        ref="treeRef"
        :data="permTree"
        show-checkbox
        node-key="id"
        default-expand-all
        :default-checked-keys="checkedPerms"
      />
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="permSaving" @click="savePerms">确定</el-button>
      </template>
    </el-dialog>

    <RecycleBinDrawer
      v-model:visible="trashVisible"
      :fetch-api="fetchDeletedRoles"
      :restore-api="restoreRole"
      :columns="trashColumns"
      @restored="onTrashRestored"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import {
  getRoles,
  createRole,
  updateRole,
  deleteRole,
  getDeletedRoles,
  restoreRole,
  assignRolePermissions,
  getRolePermissions,
  getPermissions,
} from '@/api/role'

const tableRef = ref()
const treeRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editRow = ref<any>(null)
const trashVisible = ref(false)
const permDialogVisible = ref(false)
const permSaving = ref(false)
const permTree = ref<any[]>([])
const checkedPerms = ref<number[]>([])
const currentRoleID = ref(0)

const columns = [
  { prop: 'id', label: 'ID', width: '70' },
  { prop: 'name', label: '角色名称', width: '180' },
  { prop: 'remark', label: '备注' },
  { prop: 'is_default', label: '默认', width: '80', formatter: (row: any) => (row.is_default ? '是' : '否') },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (row: any) => new Date(row.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop: 'name', label: '角色名称', type: 'input' as const, placeholder: '模糊搜索' },
]

const actions = [
  { key: 'add', label: '新增角色', type: 'primary' as const },
  { key: 'trash', label: '回收站', type: 'default' as const },
]

const formFields = [
  { prop: 'name', label: '角色名称', type: 'input' as const, placeholder: '请输入角色名称' },
  { prop: 'remark', label: '备注', type: 'textarea' as const, placeholder: '请输入备注' },
]

const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

const trashColumns = [
  { prop: 'id', label: 'ID', width: '70' },
  { prop: 'name', label: '角色名称' },
  { prop: 'deleted_at', label: '删除时间', formatter: (row: any) => new Date(row.deleted_at).toLocaleString('zh-CN') },
]

async function fetchRoles(params: any) {
  return (await getRoles(params)) as any
}

async function fetchDeletedRoles(params: any) {
  return (await getDeletedRoles(params)) as any
}

async function submitForm(data: any) {
  if (dialogMode.value === 'add') {
    return createRole(data)
  }
  return updateRole(editRow.value.id, data)
}

function handleAction(key: string) {
  if (key === 'add') {
    dialogMode.value = 'add'
    editRow.value = null
    dialogVisible.value = true
  } else if (key === 'trash') {
    trashVisible.value = true
  }
}

function handleEdit(row: any) {
  dialogMode.value = 'edit'
  editRow.value = row
  dialogVisible.value = true
}

async function handleAssignPerms(row: any) {
  currentRoleID.value = row.id
  const perms = (await getPermissions()) as any
  const ids = (await getRolePermissions(row.id)) as any

  const tree: any[] = []
  const moduleMap: Record<string, { name: string; children: any[] }> = {}

  for (const [module, actions] of Object.entries(perms)) {
    if (!moduleMap[module]) {
      moduleMap[module] = { name: module, children: [] }
    }
    const permsAny = actions as any[]
    for (const p of permsAny) {
      moduleMap[module].children.push({ id: p.id, label: p.name })
    }
  }

  for (const key of Object.keys(moduleMap)) {
    tree.push({ id: key, label: moduleMap[key].name, children: moduleMap[key].children })
  }

  permTree.value = tree
  checkedPerms.value = ids as number[]
  permDialogVisible.value = true
}

async function savePerms() {
  const checked = (treeRef.value?.getCheckedKeys() || []) as number[]
  permSaving.value = true
  try {
    await assignRolePermissions(currentRoleID.value, checked)
    ElMessage.success('权限分配成功')
    permDialogVisible.value = false
    tableRef.value?.refresh()
  } catch { /* error handled by interceptor */ } finally {
    permSaving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch { /* error handled by interceptor */ }
}

function onFormSuccess() {
  tableRef.value?.refresh()
}

function onTrashRestored() {
  tableRef.value?.refresh()
}
</script>

<style lang="scss" scoped>
.page-container {
  padding: 0;
  background: transparent;
}
.page-header {
  margin-bottom: 16px;
  h2 { font-size: 18px; font-weight: 600; color: #303133; }
}
</style>
