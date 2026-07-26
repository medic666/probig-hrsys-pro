<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.name" placeholder="公司名称" clearable style="width:180px;" />
      <el-input v-model="search.credit_code" placeholder="统一社会信用代码" clearable style="width:220px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'company:write'" @click="openDialog(false)">新增公司</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="name" label="公司名称" width="200" />
      <el-table-column prop="credit_code" label="统一社会信用代码" width="200" />
      <el-table-column prop="address" label="地址" min-width="180" />
      <el-table-column prop="contact_phone" label="联系电话" width="130" />
      <el-table-column prop="bank_name" label="开户行" width="140" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" v-permission="'company:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'company:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑公司' : '新增公司'" width="500px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="140px">
        <el-form-item label="公司名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" prop="credit_code">
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
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCompanyList, createCompany, updateCompany, deleteCompany } from '@/api/company'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ name: '', credit_code: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({ name: '', credit_code: '', address: '', contact_phone: '', bank_name: '', bank_account: '' })
const rules = {
  name: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
  credit_code: [{ required: true, message: '请输入统一社会信用代码', trigger: 'blur' }],
}

async function fetchData() {
  loading.value = true
  try {
    const data = await getCompanyList({ pageNum: pageNum.value, pageSize: pageSize.value, ...search })
    list.value = data.list
    total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.name = ''; search.credit_code = ''; pageNum.value = 1; fetchData() }

function resetForm() { Object.assign(form, { name: '', credit_code: '', address: '', contact_phone: '', bank_name: '', bank_account: '' }); formRef.value?.resetFields() }

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) { editId.value = row.id; Object.assign(form, row) } else { editId.value = 0; resetForm() }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) { await updateCompany(editId.value, { ...form }); ElMessage.success('更新成功') }
    else { await createCompany({ ...form }); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteCompany(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
