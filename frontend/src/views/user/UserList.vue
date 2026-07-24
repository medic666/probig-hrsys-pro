<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-input v-model="searchForm.keyword" placeholder="搜索用户名" style="width: 200px;" clearable @keyup.enter="fetchData" />
        <el-button type="primary" @click="dialogVisible = true">新增用户</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column label="关联人员" min-width="100">
          <template #default="{ row }">{{ row.person?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="首次登录" min-width="90">
          <template #default="{ row }">{{ row.is_first_login ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="handleResetPwd(row)">重置密码</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="500px" @closed="resetForm">
      <el-form :model="form" label-width="100px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" required>
          <el-input v-model="password" type="password" show-password />
        </el-form-item>
        <el-form-item label="关联人员">
          <el-input v-model.number="form.person_id" placeholder="人员ID" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/user'
import * as authApi from '../../api/auth'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const isEdit = ref(false)
const password = ref('')
const searchForm = reactive({ keyword: '' })

const defaultForm = { id: 0, username: '', person_id: 0, status: 1 }
const form = reactive({ ...defaultForm })

function resetForm() { Object.assign(form, defaultForm); password.value = '' }

function openEdit(row: any) {
  isEdit.value = true
  Object.assign(form, { id: row.id, username: row.username, person_id: row.person_id || 0, status: row.status })
  dialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getUserList({ page: page.value, page_size: pageSize.value, keyword: searchForm.keyword })
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleSubmit() {
  if (!isEdit.value && !password.value) { ElMessage.warning('请输入密码'); return }
  submitting.value = true
  try {
    if (isEdit.value) {
      await api.updateUser(form.id, form)
      ElMessage.success('修改成功')
    } else {
      await api.createUser({ ...form, password: password.value })
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleResetPwd(row: any) {
  try {
    const { value: newPwd } = await ElMessageBox.prompt('请输入新密码', '重置密码', { type: 'warning' })
    if (newPwd) {
      await authApi.resetUserPassword({ user_id: row.id, new_password: newPwd })
      ElMessage.success('密码已重置')
    }
  } catch { /* canceled */ }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除用户「${row.username}」？`, '确认删除', { type: 'warning' })
  await api.deleteUser(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

fetchData()
</script>
