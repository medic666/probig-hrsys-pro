<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, Setting, Refresh } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listUsers, createUser, updateUser, deleteUser, resetPassword, assignRoles } from '@/api/rbac'
import type { UserListParams } from '@/api/rbac'
import type { User } from '@/api/rbac'

const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增用户')
const editingRow = ref<User | null>(null)
const roleVisible = ref(false)
const roleUserId = ref<number>(0)
const selectedRoleIds = ref<number[]>([])

const searchFields = [
  { prop: 'username', label: '用户名', type: 'input' as const },
  { prop: 'is_active', label: '状态', type: 'select' as const, options: [
    { label: '启用', value: 1 }, { label: '禁用', value: 0 }
  ]}
]

const columns = [
  { prop: 'username', label: '用户名' },
  { prop: 'person_name', label: '关联人员' },
  { prop: 'is_active', label: '状态' },
  { prop: 'created_at', label: '创建时间' },
  { slot: 'actions', label: '操作', width: 240, fixed: 'right' as const }
]

const formFields = [
  { prop: 'username', label: '用户名', type: 'input' as const, required: true },
  { prop: 'password', label: '密码', type: 'input' as const, required: true },
  { prop: 'person_id', label: '关联人员', type: 'name-select' as const, nameType: 'person' }
]

async function fetchList(params: Record<string, unknown>) {
  return listUsers(params as unknown as UserListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增用户'
  formVisible.value = true
}

function handleEdit(row: User) {
  editingRow.value = row
  formTitle.value = '编辑用户'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { username: editingRow.value.username, person_id: editingRow.value.person_id }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updateUser({ id: editingRow.value.id, ...data } as any)
  } else {
    await createUser(data as any)
  }
}

async function handleDelete(row: User) {
  try {
    await ElMessageBox.confirm(`确定要删除用户「${row.username}」吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

async function handleResetPassword(row: User) {
  try {
    await ElMessageBox.confirm(`确定要重置用户「${row.username}」的密码吗？`, '确认重置', { type: 'warning' })
  } catch {
    return
  }
  await resetPassword(row.id)
  ElMessage.success('密码已重置为默认密码')
}

function handleAssignRoles(row: User) {
  roleUserId.value = row.id
  selectedRoleIds.value = row.roles?.map((r) => r.id) || []
  roleVisible.value = true
}

async function handleRoleSubmit() {
  await assignRoles({ user_id: roleUserId.value, role_ids: selectedRoleIds.value })
  ElMessage.success('角色分配成功')
  roleVisible.value = false
  tableRef.value?.refresh()
}

function handleFormSuccess() {
  tableRef.value?.refresh()
}
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增用户</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
        <el-button type="primary" link :icon="Setting" @click="handleAssignRoles(row)">分配角色</el-button>
        <el-button type="warning" link :icon="Refresh" @click="handleResetPassword(row)">重置密码</el-button>
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

    <el-dialog v-model="roleVisible" title="分配角色" width="400px">
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="id in [1, 2]"
          :key="id"
          :value="id"
          :label="`角色${id}`"
        />
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRoleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>
