<template>
  <div class="page-container">
    <div class="page-header">
      <h2>角色管理</h2>
    </div>
    <ProTable
ref="tableRef"
      :url-driven="true"
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

    <el-dialog v-model="permDialogVisible" title="权限分配" width="500px">
      <el-tree
        ref="treeRef"
        :data="permTree"
        show-checkbox
        node-key="id"
        default-expand-all
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
import { ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import {
  getRoles,
  deleteRole,
  getDeletedRoles,
  restoreRole,
  assignRolePermissions,
  getRolePermissions,
  getPermissions,
} from '@/api/role'

const router = useRouter()
const tableRef = ref()
const treeRef = ref()
const trashVisible = ref(false)
const permDialogVisible = ref(false)
const permSaving = ref(false)
const permTree = ref<any[]>([])
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

function handleAction(key: string) {
  if (key === 'add') {
    router.push('/system/roles/create')
  } else if (key === 'trash') {
    trashVisible.value = true
  }
}

function handleEdit(row: any) {
  router.push(`/system/roles/${row.id}`)
}

async function handleAssignPerms(row: any) {
  currentRoleID.value = row.id
  permDialogVisible.value = true
  try {
    const perms = (await getPermissions()) as any
    const ids = (await getRolePermissions(row.id)) as number[]

    const tree: any[] = []
    for (const [module, actions] of Object.entries(perms)) {
      const children = (actions as any[]).map((p: any) => ({ id: p.id, label: p.name }))
      tree.push({ id: 'm_' + module, label: module, children })
    }

    permTree.value = tree
    await nextTick()
    treeRef.value?.setCheckedKeys(ids)
  } catch { permDialogVisible.value = false }
}

async function savePerms() {
    const checked = (treeRef.value?.getCheckedKeys(true) || []) as number[]
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
