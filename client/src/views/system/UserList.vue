<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="showDialog()">新增用户</el-button>
        </div>
      </template>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column label="关联人员" min-width="100">
          <template #default="{ row }">{{ row.person?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">{{ row.status === 1 ? '启用' : '禁用' }}</template>
        </el-table-column>
        <el-table-column label="角色" min-width="150">
          <template #default="{ row }">{{ row.roles?.map((r: any) => r.name).join(', ') }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showDialog(row)">编辑</el-button>
            <el-button size="small" @click="showRoleDialog(row)">分配角色</el-button>
            <el-button size="small" type="warning" @click="handleResetPwd(row)">重置密码</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
              <template #reference><el-button size="small" type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top:16px"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="450px" @closed="resetForm">
      <el-form ref="userFormRef" :model="userForm" :rules="userRules" label-width="80px">
        <el-form-item label="用户名" prop="username"><el-input v-model="userForm.username" /></el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password"><el-input v-model="userForm.password" type="password" /></el-form-item>
        <el-form-item label="关联人员"><el-input-number v-model="userForm.person_id" style="width:100%" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="userForm.status" style="width:100%"><el-option :value="1" label="启用" /><el-option :value="0" label="禁用" /></el-select></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialog" title="分配角色" width="450px">
      <el-checkbox-group v-model="selectedRoles">
        <el-checkbox v-for="r in allRoles" :key="r.id" :label="r.id" :value="r.id">{{ r.name }}</el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialog = false">取消</el-button>
        <el-button type="primary" @click="submitRoles">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false); const isEdit = ref(false); const editId = ref(0); const userFormRef = ref()
const userForm = reactive({ username: '', password: '', person_id: undefined as number | undefined, status: 1 })
const userRules = { username: [{ required: true }] }

const roleDialog = ref(false); const currentUserId = ref(0); const selectedRoles = ref<number[]>([]); const allRoles = ref<any[]>([])

function resetForm() {
  Object.assign(userForm, { username: '', password: '', person_id: undefined, status: 1 })
  isEdit.value = false; editId.value = 0
}

async function fetchList() {
  loading.value = true
  const res = await request.get('/users', { params: { page: page.value, page_size: pageSize.value } })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function showDialog(row?: any) {
  if (row) {
    isEdit.value = true; editId.value = row.id
    Object.assign(userForm, { username: row.username, password: '', person_id: row.person_id, status: row.status })
  } else { resetForm() }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await userFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const payload: any = { ...userForm }
  if (isEdit.value) {
    delete payload.password
    await request.put(`/users/${editId.value}`, payload)
  } else {
    await request.post('/users', payload)
  }
  dialogVisible.value = false; fetchList(); ElMessage.success('保存成功')
}

async function handleDelete(id: number) {
  await request.delete(`/users/${id}`)
  fetchList(); ElMessage.success('删除成功')
}

async function handleResetPwd(row: any) {
  try {
    const { value } = await ElMessageBox.prompt('请输入新密码', '重置密码', { inputType: 'password' })
    await request.put(`/user/${row.id}/reset-password`, { new_password: value })
    ElMessage.success('密码已重置')
  } catch {}
}

async function showRoleDialog(row: any) {
  currentUserId.value = row.id
  const [userRes, rolesRes] = await Promise.all([
    request.get(`/users/${row.id}`),
    request.get('/roles/all'),
  ])
  allRoles.value = rolesRes.data
  selectedRoles.value = (userRes.data.roles || []).map((r: any) => r.id)
  roleDialog.value = true
}

async function submitRoles() {
  await request.put(`/users/${currentUserId.value}/roles`, { role_ids: selectedRoles.value })
  roleDialog.value = false; fetchList(); ElMessage.success('角色分配成功')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
