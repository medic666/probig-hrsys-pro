<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, Lock } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import { listRoles, createRole, updateRole, deleteRole, assignPermissions, listAllPermissions } from '@/api/rbac'
import type { RoleListParams, Role, PermissionTreeItem } from '@/api/rbac'

const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增角色')
const editingRow = ref<Role | null>(null)
const permVisible = ref(false)
const permRoleId = ref<number>(0)

const searchFields = [
  { prop: 'name', label: '角色名称', type: 'input' as const }
]

const columns = [
  { prop: 'name', label: '角色名称' },
  { prop: 'remark', label: '备注' },
  { prop: 'is_system', label: '系统角色' },
  { prop: 'created_at', label: '创建时间' },
  { slot: 'actions', label: '操作', width: 200, fixed: 'right' as const }
]

const formFields = [
  { prop: 'name', label: '角色名称', type: 'input' as const, required: true },
  { prop: 'remark', label: '备注', type: 'textarea' as const }
]

const permissionTree = ref<PermissionTreeItem[]>([])
const checkedPermIds = ref<number[]>([])

async function fetchList(params: Record<string, unknown>) {
  return listRoles(params as unknown as RoleListParams)
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增角色'
  formVisible.value = true
}

function handleEdit(row: Role) {
  if (row.is_system) {
    ElMessage.warning('系统角色不可编辑')
    return
  }
  editingRow.value = row
  formTitle.value = '编辑角色'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { name: editingRow.value.name, remark: editingRow.value.remark }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updateRole({ id: editingRow.value.id, ...data } as any)
  } else {
    await createRole(data as any)
  }
}

async function handleDelete(row: Role) {
  if (row.is_system) {
    ElMessage.warning('系统角色不可删除')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除角色「${row.name}」吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deleteRole(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

async function handleAssignPermissions(row: Role) {
  permRoleId.value = row.id
  const perms = await listAllPermissions()
  permissionTree.value = perms
  checkedPermIds.value = row.permissions?.map((p) => p.id) || []
  permVisible.value = true
}

async function handlePermSubmit() {
  await assignPermissions({ role_id: permRoleId.value, permission_ids: checkedPermIds.value })
  ElMessage.success('权限分配成功')
  permVisible.value = false
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
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增角色</el-button>
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
        <el-button type="primary" link :icon="Lock" @click="handleAssignPermissions(row)">分配权限</el-button>
        <el-button
          v-if="!row.is_system"
          type="danger" link :icon="Delete"
          @click="handleDelete(row)"
        >删除</el-button>
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

    <el-dialog v-model="permVisible" title="分配权限" width="600px">
      <el-tree
        :data="permissionTree"
        show-checkbox
        node-key="id"
        :default-checked-keys="checkedPermIds"
        :props="{ children: 'permissions', label: 'module_name' }"
        default-expand-all
        @check="(_node, data) => { checkedPermIds = data.checkedKeys as number[] }"
      />
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePermSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>
