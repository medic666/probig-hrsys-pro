<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-input v-model="searchForm.keyword" placeholder="搜索姓名/别名" style="width: 240px;" clearable @keyup.enter="fetchData" />
        <el-button type="primary" @click="dialogVisible = true">新增人员</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="gender" label="性别" min-width="60">
          <template #default="{ row }">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '' }}</template>
        </el-table-column>
        <el-table-column prop="id_card" label="身份证号" min-width="160" />
        <el-table-column prop="nation" label="民族" min-width="80" />
        <el-table-column prop="native_place" label="籍贯" min-width="120" />
        <el-table-column prop="alias" label="别名" min-width="100" />
        <el-table-column label="操作" min-width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="$router.push(`/persons/${row.id}/position-events`)">职务</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        style="margin-top: 16px; justify-content: flex-end;"
        v-model:current-page="page" :total="total" :page-size="pageSize"
        layout="total, prev, pager, next" @current-change="fetchData"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑人员' : '新增人员'" width="700px" @closed="resetForm">
      <el-form :model="form" label-width="100px">
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="身份证号">
          <el-input v-model="form.id_card" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender">
            <el-option :value="1" label="男" />
            <el-option :value="2" label="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="别名">
          <el-input v-model="form.alias" />
        </el-form-item>
        <el-form-item label="民族">
          <el-input v-model="form.nation" />
        </el-form-item>
        <el-form-item label="籍贯">
          <el-input v-model="form.native_place" />
        </el-form-item>
        <el-form-item label="住址">
          <el-input v-model="form.address" />
        </el-form-item>
        <el-form-item label="政治面貌">
          <el-input v-model="form.political_status" />
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
import * as api from '../../api/person'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const isEdit = ref(false)
const searchForm = reactive({ keyword: '' })

const defaultForm = {
  id: 0, name: '', id_card: '', gender: 0, birthday: null,
  nation: '', native_place: '', address: '', political_status: '',
  marital_status: 0, alias: '', phones: [], emails: [], bank_cards: [],
}
const form = reactive({ ...defaultForm })

function resetForm() {
  Object.assign(form, defaultForm)
}

function openEdit(row: any) {
  isEdit.value = true
  Object.assign(form, { ...row, id_card: row.id_card || '' })
  dialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getPersonList({ page: page.value, page_size: pageSize.value, keyword: searchForm.keyword })
    list.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await api.updatePerson(form.id, form)
      ElMessage.success('修改成功')
    } else {
      await api.createPerson(form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除人员「${row.name}」？`, '确认删除', { type: 'warning' })
  await api.deletePerson(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>
