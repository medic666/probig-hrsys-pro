<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.username" placeholder="用户名" clearable style="width:180px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'rbac:write'" @click="openDialog(false)">新增用户</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="username" label="用户名" width="140" />
      <el-table-column prop="roles" label="角色" min-width="200" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="160" />
      <el-table-column label="操作" width="300">
        <template #default="{ row }">
          <el-button size="small" v-permission="'rbac:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" v-permission="'rbac:write'" @click="handleResetPwd(row.id)">重置密码</el-button>
          <el-button size="small" v-permission="'rbac:write'" @click="handleToggleStatus(row.id)">{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
          <el-button size="small" type="danger" v-permission="'rbac:write'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="form.role_ids">
            <el-checkbox v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, createUser, updateUser, deleteUser, toggleUserStatus, resetUserPassword, getRoles } from '@/api/rbac'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ username: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({ username: '', password: '', role_ids: [] as number[] })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
}
const roles = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const data = await getUserList({ pageNum: pageNum.value, pageSize: pageSize.value, username: search.username })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.username = ''; pageNum.value = 1; fetchData() }

async function fetchRoles() { try { roles.value = await getRoles() } catch (e) {} }

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) {
    editId.value = row.id
    form.username = row.username
    form.role_ids = row.roles ? row.roles.split(', ').map((r: string) => roles.value.find(x => x.name === r)?.id).filter(Boolean) : []
  } else {
    editId.value = 0
    form.username = ''; form.password = ''; form.role_ids = []
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updateUser(editId.value, { username: form.username, role_ids: form.role_ids })
      ElMessage.success('更新成功')
    } else {
      await createUser({ username: form.username, password: form.password, role_ids: form.role_ids })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleResetPwd(id: number) {
  await ElMessageBox.confirm('确定要重置密码为 admin123 吗？', '确认', { type: 'warning' })
  try { await resetUserPassword(id); ElMessage.success('密码已重置为 admin123') } catch (e) {}
}

async function handleToggleStatus(id: number) {
  try { await toggleUserStatus(id); ElMessage.success('操作成功'); fetchData() } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteUser(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(() => { fetchData(); fetchRoles() })
</script>
