<template>
  <div class="page-container">
    <div class="tool-bar">
      <el-button type="primary" v-permission="'rbac:write'" @click="openDialog(false)">新增角色</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="name" label="角色名称" width="200" />
      <el-table-column prop="remark" label="备注" min-width="200" />
      <el-table-column label="操作" width="250">
        <template #default="{ row }">
          <el-button size="small" v-permission="'rbac:write'" @click="openPermDialog(row)">分配权限</el-button>
          <el-button size="small" v-permission="'rbac:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" type="danger" v-if="!row.is_admin" v-permission="'rbac:write'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="400px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permVisible" title="分配权限" width="600px">
      <el-tree
        ref="treeRef"
        :data="permTree"
        show-checkbox
        node-key="id"
        :default-checked-keys="checkedPerms"
        :props="{ label: 'perm_name' }"
      />
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSetPerms">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoles, createRole, updateRole, deleteRole, getRolePermissions, setRolePermissions, getAllPermissions } from '@/api/rbac'

const loading = ref(false)
const list = ref([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({ name: '', remark: '' })
const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

const permVisible = ref(false)
const permRoleId = ref(0)
const treeRef = ref()
const permTree = ref<any[]>([])
const checkedPerms = ref<number[]>([])

async function fetchData() {
  loading.value = true
  try { list.value = await getRoles() } catch (e) {} finally { loading.value = false }
}

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) { editId.value = row.id; form.name = row.name; form.remark = row.remark }
  else { editId.value = 0; form.name = ''; form.remark = '' }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) { await updateRole(editId.value, { name: form.name, remark: form.remark }); ElMessage.success('更新成功') }
    else { await createRole({ name: form.name, remark: form.remark }); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function openPermDialog(row: any) {
  permRoleId.value = row.id
  try {
    const perms = await getAllPermissions()
    const checked = await getRolePermissions(row.id)
    checkedPerms.value = checked || []

    const tree: any[] = []
    const moduleMap: Record<string, any> = {}
    if (typeof perms === 'object' && !Array.isArray(perms)) {
      for (const [module, items] of Object.entries(perms)) {
        tree.push({
          id: module, perm_name: module, children: (items as any[]).map((p: any) => ({ id: p.id, perm_name: p.perm_name })),
        })
      }
    } else if (Array.isArray(perms)) {
      for (const p of perms) {
        if (!moduleMap[p.module]) {
          moduleMap[p.module] = { id: p.module, perm_name: p.module, children: [] }
          tree.push(moduleMap[p.module])
        }
        moduleMap[p.module].children.push({ id: p.id, perm_name: p.perm_name })
      }
    }
    permTree.value = tree
    permVisible.value = true
  } catch (e) {}
}

async function handleSetPerms() {
  const keys = treeRef.value?.getCheckedKeys(true)
  const permIds = keys?.filter((k: any) => typeof k === 'number')
  try {
    await setRolePermissions(permRoleId.value, { perm_ids: permIds || [] })
    ElMessage.success('权限已保存')
    permVisible.value = false
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteRole(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
