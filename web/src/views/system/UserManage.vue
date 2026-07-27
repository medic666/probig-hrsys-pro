<template>
  <div class="page-container">
    <div class="page-header">
      <h2>用户管理</h2>
    </div>
    <ProTable
      ref="tableRef"
      :columns="columns"
      :fetch-api="fetchUsers"
      :search-fields="searchFields"
      :actions="actions"
      @action="handleAction"
    >
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-if="row.username !== 'admin'" type="warning" link size="small" @click="handleAssignRoles(row)">分配角色</el-button>
        <el-button type="info" link size="small" @click="handleResetPwd(row)">重置密码</el-button>
        <el-button v-if="row.username !== 'admin'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="dialogVisible"
      :title="dialogMode === 'add' ? '新增用户' : '编辑用户'"
      :mode="dialogMode"
      :form-fields="formFields"
      :rules="formRules"
      :submit-api="submitForm"
      :edit-data="editRow"
      @success="onFormSuccess"
    />

    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="420px"
    >
      <el-checkbox-group v-model="selectedRoles">
        <el-checkbox v-for="role in allRoles" :key="role.id" :label="role.id">
          {{ role.name }}
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSaving" @click="saveRoles">确定</el-button>
      </template>
    </el-dialog>

    <RecycleBinDrawer
      v-model:visible="trashVisible"
      :fetch-api="fetchDeletedUsers"
      :restore-api="restoreUser"
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
  getUsers,
  createUser,
  updateUser,
  deleteUser,
  resetPassword,
  assignUserRoles,
  getDeletedUsers,
  restoreUser,
} from '@/api/user'
import { getAllRoles } from '@/api/role'

interface FormField {
  prop: string
  label: string
  type: 'input' | 'number' | 'select' | 'date' | 'textarea' | 'switch'
  options?: { label: string; value: any }[]
  placeholder?: string
  span?: number
  defaultValue?: any
}

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editRow = ref<any>(null)
const trashVisible = ref(false)
const roleDialogVisible = ref(false)
const roleSaving = ref(false)
const allRoles = ref<any[]>([])
const selectedRoles = ref<number[]>([])
const currentUserID = ref(0)

const columns = [
  { prop: 'id', label: 'ID', width: '70' },
  { prop: 'username', label: '用户名', width: '150' },
  { prop: 'person_name', label: '关联人员', width: '120' },
  { prop: 'roles', label: '角色', formatter: (row: any) => (row.roles || []).join(', ') },
  { prop: 'is_active', label: '状态', width: '80', formatter: (row: any) => (row.is_active ? '启用' : '禁用') },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (row: any) => new Date(row.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop: 'username', label: '用户名', type: 'input' as const, placeholder: '模糊搜索' },
  {
    prop: 'is_active', label: '状态', type: 'select' as const,
    options: [{ label: '启用', value: true }, { label: '禁用', value: false }],
  },
]

const actions = [
  { key: 'add', label: '新增用户', type: 'primary' as const },
  { key: 'trash', label: '回收站', type: 'default' as const },
]

const formFields: FormField[] = [
  { prop: 'username', label: '用户名', type: 'input', placeholder: '请输入用户名' },
  { prop: 'password', label: '密码', type: 'input', placeholder: '新增时必填，编辑时留空', defaultValue: '' },
  { prop: 'is_active', label: '启用', type: 'switch', defaultValue: true },
]

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
}

const trashColumns = [
  { prop: 'id', label: 'ID', width: '70' },
  { prop: 'username', label: '用户名' },
  { prop: 'deleted_at', label: '删除时间', formatter: (row: any) => new Date(row.deleted_at).toLocaleString('zh-CN') },
]

async function fetchUsers(params: any) {
  const data = (await getUsers(params)) as any
  return data
}

async function fetchDeletedUsers(params: any) {
  const data = (await getDeletedUsers(params)) as any
  return data
}

async function submitForm(data: any) {
  if (dialogMode.value === 'add') {
    return createUser(data)
  }
  return updateUser(editRow.value.id, data)
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

async function handleAssignRoles(row: any) {
  currentUserID.value = row.id
  allRoles.value = (await getAllRoles()) as any
  const ids = row.roles ? [] : []
  selectedRoles.value = ids
  roleDialogVisible.value = true
}

async function saveRoles() {
  roleSaving.value = true
  try {
    await assignUserRoles(currentUserID.value, selectedRoles.value)
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
    tableRef.value?.refresh()
  } catch { /* error handled by interceptor */ } finally {
    roleSaving.value = false
  }
}

async function handleResetPwd(row: any) {
  try {
    await ElMessageBox.confirm(`确认重置用户「${row.username}」的密码？`, '提示', {
      type: 'warning',
      confirmButtonText: '确认重置',
    })
  } catch {
    return
  }
  try {
    const data = (await resetPassword(row.id)) as any
    ElMessage.success(data?.msg || '密码已重置')
  } catch { /* error handled by interceptor */ }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteUser(row.id)
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
