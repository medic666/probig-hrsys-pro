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
        <el-button v-if="!row.is_default" v-permission="PERM.roleWrite" type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button v-if="!row.is_default" v-permission="PERM.roleWrite" type="warning" link size="small" @click="handleAssignPerms(row)">权限分配</el-button>
        <el-button v-if="!row.is_default" v-permission="PERM.roleWrite" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="permDialogVisible" title="权限分配" width="640px">
      <el-form label-width="80px">
        <el-form-item label="数据范围">
          <el-radio-group v-model="permDataScope" size="small">
            <el-radio-button value="all">全部</el-radio-button>
            <el-radio-button value="own">仅自己</el-radio-button>
          </el-radio-group>
          <span class="scope-hint">「仅自己」时用户只能交互本人关联人员的数据</span>
        </el-form-item>
      </el-form>
      <el-table :data="permModules" border size="small" max-height="420">
        <el-table-column prop="name" label="模块" min-width="140" />
        <el-table-column v-for="a in permActionColumns" :key="a.action" :label="a.label" width="70" align="center">
          <template #default="{ row }">
            <el-checkbox
              v-if="permOf(row, a.action)"
              :model-value="isPermChecked(permOf(row, a.action)!.id)"
              @change="(v: any) => togglePerm(permOf(row, a.action)!.id, !!v)"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="全选" width="70" align="center">
          <template #default="{ row }">
            <el-checkbox
              :model-value="isModuleAllChecked(row)"
              :indeterminate="isModulePartialChecked(row)"
              @change="(v: any) => toggleModule(row, !!v)"
            />
          </template>
        </el-table-column>
      </el-table>
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
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef = ref()
const trashVisible = ref(false)
const permDialogVisible = ref(false)
const permSaving = ref(false)
const permModules = ref<any[]>([])
const permDataScope = ref('all')
const currentRoleID = ref(0)
const checkedPermIDs = ref<Set<number>>(new Set())

// 动作列固定顺序（与后端 PermissionActionNames 对齐：查看/编辑/核算/导出）
const permActionColumns = [
  { action: 'read', label: '查看' },
  { action: 'write', label: '编辑' },
  { action: 'calculate', label: '核算' },
  { action: 'export', label: '导出' },
]

function permOf(module: any, action: string) {
  return module.actions.find((p: any) => p.key.endsWith('.' + action))
}
function isPermChecked(id: number) {
  return checkedPermIDs.value.has(id)
}
function togglePerm(id: number, checked: boolean) {
  const next = new Set(checkedPermIDs.value)
  if (checked) next.add(id)
  else next.delete(id)
  checkedPermIDs.value = next
}
function isModuleAllChecked(module: any) {
  return module.actions.every((p: any) => checkedPermIDs.value.has(p.id))
}
function isModulePartialChecked(module: any) {
  const checked = module.actions.filter((p: any) => checkedPermIDs.value.has(p.id)).length
  return checked > 0 && checked < module.actions.length
}
function toggleModule(module: any, checked: boolean) {
  const next = new Set(checkedPermIDs.value)
  for (const p of module.actions) {
    if (checked) next.add(p.id)
    else next.delete(p.id)
  }
  checkedPermIDs.value = next
}

const columns = [
  { prop: 'id', label: 'ID', width: '70' },
  { prop: 'name', label: '角色名称', width: '180' },
  { prop: 'data_scope', label: '数据范围', width: '100', formatter: (row: any) => (row.data_scope === 'own' ? '仅自己' : '全部') },
  { prop: 'remark', label: '备注' },
  { prop: 'is_default', label: '默认', width: '80', formatter: (row: any) => (row.is_default ? '是' : '否') },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (row: any) => new Date(row.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop: 'name', label: '角色名称', type: 'input' as const, placeholder: '模糊搜索' },
]

const actions = [
  { key: 'add', label: '新增角色', type: 'primary' as const, permission: PERM.roleWrite },
  { key: 'trash', label: '回收站', type: 'default' as const, permission: PERM.roleWrite },
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
  permDataScope.value = row.data_scope === 'own' ? 'own' : 'all'
  permDialogVisible.value = true
  try {
    const perms = (await getPermissions()) as any
    const ids = (await getRolePermissions(row.id)) as number[]
    permModules.value = perms || []
    checkedPermIDs.value = new Set(ids)
  } catch { permDialogVisible.value = false }
}

async function savePerms() {
  const checked = Array.from(checkedPermIDs.value)
  permSaving.value = true
  try {
    await assignRolePermissions(currentRoleID.value, checked, permDataScope.value)
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
.scope-hint {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
