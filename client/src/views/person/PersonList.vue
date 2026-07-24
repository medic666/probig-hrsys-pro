<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>人员列表</span>
          <el-button type="primary" @click="showAddDialog">新增人员</el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="query" placeholder="搜索姓名/别名" style="width: 240px" clearable @clear="fetchList" @keyup.enter="fetchList" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading" @row-click="goDetail">
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="alias" label="别名" min-width="100" />
        <el-table-column prop="gender" label="性别" width="70">
          <template #default="{ row }">{{ row.gender === 0 ? '男' : row.gender === 1 ? '女' : '-' }}</template>
        </el-table-column>
        <el-table-column prop="nation" label="民族" width="80" />
        <el-table-column prop="political_status" label="政治面貌" width="100" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click.stop="goDetail(row)">详情</el-button>
            <el-button size="small" type="primary" @click.stop="showEditDialog(row)">编辑</el-button>
            <el-popconfirm title="确定删除该人员？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger" @click.stop>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑人员' : '新增人员'" width="600px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="别名">
          <el-input v-model="form.alias" />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input v-model="form.id_card" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" placeholder="请选择" clearable style="width:100%">
            <el-option :value="0" label="男" />
            <el-option :value="1" label="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="生日">
          <el-date-picker v-model="form.birthday" type="date" placeholder="选择日期" style="width:100%" value-format="YYYY-MM-DD" />
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
        <el-form-item label="婚姻状态">
          <el-select v-model="form.marital_status" placeholder="请选择" clearable style="width:100%">
            <el-option :value="0" label="未婚" />
            <el-option :value="1" label="已婚" />
          </el-select>
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
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const router = useRouter()
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const query = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()

const form = reactive({
  name: '', alias: '', id_card: '', gender: null as number | null,
  birthday: '', nation: '', native_place: '', address: '',
  political_status: '', marital_status: null as number | null,
})

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  id_card: [{ required: true, message: '请输入身份证号', trigger: 'blur' }],
}

function resetForm() {
  Object.assign(form, {
    name: '', alias: '', id_card: '', gender: null,
    birthday: '', nation: '', native_place: '', address: '',
    political_status: '', marital_status: null,
  })
  isEdit.value = false
  editId.value = 0
}

async function fetchList() {
  loading.value = true
  try {
    const res = await request.get('/persons', { params: { query: query.value, page: page.value, page_size: pageSize.value } })
    list.value = res.data.list
    total.value = res.data.total
  } catch {}
  loading.value = false
}

function goDetail(row: any) {
  router.push(`/person/${row.id}`)
}

function showAddDialog() {
  resetForm()
  dialogVisible.value = true
}

function showEditDialog(row: any) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    name: row.name, alias: row.alias || '', id_card: row.id_card || '',
    gender: row.gender, birthday: row.birthday || '', nation: row.nation || '',
    native_place: row.native_place || '', address: row.address || '',
    political_status: row.political_status || '', marital_status: row.marital_status,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    const payload: any = { ...form }
    if (!payload.birthday) delete payload.birthday
    if (isEdit.value) {
      await request.put(`/persons/${editId.value}`, payload)
      ElMessage.success('修改成功')
    } else {
      await request.post('/persons', payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch {}
}

async function handleDelete(id: number) {
  await request.delete(`/persons/${id}`)
  ElMessage.success('删除成功')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-bar { margin-bottom: 16px; display: flex; gap: 10px; }
</style>
