<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.name" placeholder="姓名" clearable style="width:160px;" />
      <el-input v-model="search.id_card" placeholder="身份证号" clearable style="width:200px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'person:write'" @click="openDialog(false)">新增人员</el-button>
      <el-button v-permission="'person:write'" @click="recycleVisible = true">回收站</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="name" label="姓名" width="100" />
      <el-table-column prop="id_card" label="身份证号" width="180" />
      <el-table-column prop="gender" label="性别" width="60">
        <template #default="{ row }">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '' }}</template>
      </el-table-column>
      <el-table-column prop="nation" label="民族" width="80" />
      <el-table-column prop="address" label="住址" min-width="150" />
      <el-table-column prop="created_at" label="创建时间" width="160" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDetail(row.id)">详情</el-button>
          <el-button size="small" v-permission="'person:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" @click="$router.push(`/position?person_id=${row.id}`)">职务</el-button>
          <el-button size="small" type="danger" v-permission="'person:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑人员' : '新增人员'" width="600px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="入职日期">
          <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input v-model="form.id_card" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" clearable>
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="生日">
          <el-date-picker v-model="form.birthday" type="date" value-format="YYYY-MM-DD" />
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
        <el-form-item label="别名">
          <el-input v-model="form.alias" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="人员详情" width="800px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="身份证号">{{ detail.id_card }}</el-descriptions-item>
        <el-descriptions-item label="性别">{{ detail.gender === 1 ? '男' : detail.gender === 2 ? '女' : '' }}</el-descriptions-item>
        <el-descriptions-item label="生日">{{ detail.birthday }}</el-descriptions-item>
        <el-descriptions-item label="民族">{{ detail.nation }}</el-descriptions-item>
        <el-descriptions-item label="籍贯">{{ detail.native_place }}</el-descriptions-item>
        <el-descriptions-item label="住址" :span="2">{{ detail.address }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detail" style="margin-top:16px;">
        <h4>联系方式</h4>
        <div v-for="p in detail.phones" :key="p.id" style="margin:4px 0;">📞 {{ p.phone }}</div>
        <div v-for="e in detail.emails" :key="e.id" style="margin:4px 0;">📧 {{ e.email }}</div>
      </div>
      <div v-if="detail" style="margin-top:16px;">
        <h4>银行卡</h4>
        <div v-for="c in detail.bank_cards" :key="c.id" style="margin:4px 0;">🏦 {{ c.bank_name }}: {{ c.card_no }}</div>
      </div>
    </el-dialog>

    <el-drawer v-model="recycleVisible" title="回收站" size="600px">
      <el-table :data="deletedList" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="姓名" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button size="small" @click="handleRestore(row.id)">恢复</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPersonList, getPerson, createPerson, updatePerson, deletePerson, restorePerson, getDeletedPersons } from '@/api/person'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ name: '', id_card: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({
  name: '', id_card: '', gender: undefined, birthday: '', nation: '',
  native_place: '', address: '', political_status: '', alias: '', marital_status: 0,
})
const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  id_card: [{ required: true, message: '请输入身份证号', trigger: 'blur' }],
}

const detailVisible = ref(false)
const detail = ref<any>(null)

const recycleVisible = ref(false)
const deletedList = ref([])

async function fetchData() {
  loading.value = true
  try {
    const data = await getPersonList({ pageNum: pageNum.value, pageSize: pageSize.value, ...search })
    list.value = data.list
    total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() {
  search.name = ''
  search.id_card = ''
  pageNum.value = 1
  fetchData()
}

function resetForm() {
  Object.assign(form, { name: '', id_card: '', gender: undefined, birthday: '', nation: '', native_place: '', address: '', political_status: '', alias: '', marital_status: 0 })
  formRef.value?.resetFields()
}

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) {
    editId.value = row.id
    Object.assign(form, row)
  } else {
    editId.value = 0
    resetForm()
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updatePerson(editId.value, { ...form })
      ElMessage.success('更新成功')
    } else {
      await createPerson({ ...form })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try {
    await deletePerson(id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {}
}

async function handleRestore(id: number) {
  await ElMessageBox.confirm('确定要恢复吗？', '确认', { type: 'warning' })
  try {
    await restorePerson(id)
    ElMessage.success('恢复成功')
    recycleVisible.value = false
    fetchData()
  } catch (e) {}
}

async function showDetail(id: number) {
  try {
    detail.value = await getPerson(id)
    detailVisible.value = true
  } catch (e) {}
}

onMounted(fetchData)
</script>
