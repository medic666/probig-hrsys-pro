<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>公司列表</span>
          <el-button type="primary" @click="showDialog()">新增公司</el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="query" placeholder="搜索名称/信用代码" style="width: 240px" clearable @keyup.enter="fetchList" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="name" label="公司名称" min-width="150" />
        <el-table-column prop="credit_code" label="统一社会信用代码" min-width="180" />
        <el-table-column prop="address" label="地址" min-width="200" />
        <el-table-column prop="contact_phone" label="联系电话" width="130" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showDialog(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
              <template #reference><el-button size="small" type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑公司' : '新增公司'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="140px">
        <el-form-item label="公司名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="统一社会信用代码" prop="credit_code"><el-input v-model="form.credit_code" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="联系电话"><el-input v-model="form.contact_phone" /></el-form-item>
        <el-form-item label="开户行"><el-input v-model="form.bank_name" /></el-form-item>
        <el-form-item label="银行账号"><el-input v-model="form.bank_account" /></el-form-item>
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
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20)
const loading = ref(false); const query = ref('')
const dialogVisible = ref(false); const isEdit = ref(false); const editId = ref(0); const formRef = ref()

const form = reactive({ name: '', credit_code: '', address: '', contact_phone: '', bank_name: '', bank_account: '' })
const rules = {
  name: [{ required: true, message: '请输入公司名称' }],
  credit_code: [{ required: true, message: '请输入信用代码' }],
}

function resetForm() {
  Object.assign(form, { name: '', credit_code: '', address: '', contact_phone: '', bank_name: '', bank_account: '' })
  isEdit.value = false; editId.value = 0
}

async function fetchList() {
  loading.value = true
  const res = await request.get('/companies', { params: { query: query.value, page: page.value, page_size: pageSize.value } })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function showDialog(row?: any) {
  if (row) {
    isEdit.value = true; editId.value = row.id
    Object.assign(form, row)
  } else {
    resetForm()
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) {
    await request.put(`/companies/${editId.value}`, form)
  } else {
    await request.post('/companies', form)
  }
  dialogVisible.value = false; fetchList(); ElMessage.success('保存成功')
}

async function handleDelete(id: number) {
  await request.delete(`/companies/${id}`)
  fetchList(); ElMessage.success('删除成功')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-bar { margin-bottom: 16px; display: flex; gap: 10px; }
</style>
