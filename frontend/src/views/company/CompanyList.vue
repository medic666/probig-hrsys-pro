<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-input v-model="searchForm.keyword" placeholder="搜索公司名称/信用代码" style="width: 240px;" clearable @keyup.enter="fetchData" />
        <el-button type="primary" @click="dialogVisible = true">新增公司</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="name" label="公司名称" min-width="160" />
        <el-table-column prop="credit_code" label="统一社会信用代码" min-width="180" />
        <el-table-column prop="address" label="地址" min-width="180" />
        <el-table-column prop="contact_phone" label="联系电话" min-width="120" />
        <el-table-column label="操作" min-width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑公司' : '新增公司'" width="600px" @closed="resetForm">
      <el-form :model="form" label-width="140px">
        <el-form-item label="公司名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="统一社会信用代码">
          <el-input v-model="form.credit_code" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contact_phone" />
        </el-form-item>
        <el-form-item label="开户行">
          <el-input v-model="form.bank_name" />
        </el-form-item>
        <el-form-item label="银行账号">
          <el-input v-model="form.bank_account" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/company'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const isEdit = ref(false)
const searchForm = reactive({ keyword: '' })

const defaultForm = { id: 0, name: '', credit_code: '', address: '', contact_phone: '', bank_name: '', bank_account: '' }
const form = reactive({ ...defaultForm })

function resetForm() { Object.assign(form, defaultForm) }

function openEdit(row: any) {
  isEdit.value = true
  Object.assign(form, { ...row })
  dialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getCompanyList({ page: page.value, page_size: pageSize.value, keyword: searchForm.keyword })
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await api.updateCompany(form.id, form)
      ElMessage.success('修改成功')
    } else {
      await api.createCompany(form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除公司「${row.name}」？`, '确认删除', { type: 'warning' })
  await api.deleteCompany(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>
