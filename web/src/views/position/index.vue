<template>
  <div class="page-container">
    <div class="search-bar">
      <NameSelect v-model="searchPersonId" style="width:180px;" />
      <el-date-picker v-model="searchDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
      <el-input v-model="search.event_name" placeholder="事件名称" clearable style="width:140px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'position:write'" @click="openDialog(false)">新增事件</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column prop="event_name" label="事件名称" width="100" />
      <el-table-column prop="effective_date" label="生效日期" width="120" />
      <el-table-column prop="base_salary" label="基本工资" width="100" />
      <el-table-column prop="performance_salary" label="绩效基数" width="100" />
      <el-table-column prop="salary_days" label="计薪天数" width="80" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" v-permission="'position:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'position:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑职务事件' : '新增职务事件'" width="700px" @close="resetForm">
      <el-form ref="formRef" :model="form" label-width="140px">
        <el-form-item label="人员" required>
          <NameSelect v-model="form.person_id" style="width:100%;" />
        </el-form-item>
        <el-form-item label="事件名称" required>
          <el-select v-model="form.event_name">
            <el-option v-for="n in eventNames" :key="n" :label="n" :value="n" />
          </el-select>
        </el-form-item>
        <el-form-item label="生效日期" required>
          <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-divider content-position="left">变更字段（勾选即提交，不勾选不提交）</el-divider>
        <el-row :gutter="16">
          <el-col :span="12" v-for="f in fields" :key="f.key">
            <el-checkbox v-model="f.checked" style="margin-bottom:8px;">{{ f.label }}</el-checkbox>
            <el-input-number v-if="f.checked && f.type === 'number'" v-model="form[f.key]" :precision="2" style="width:100%;" />
            <el-input v-else-if="f.checked && f.type === 'string'" v-model="form[f.key]" style="width:100%;" />
            <el-switch v-else-if="f.checked && f.type === 'bool'" v-model="form[f.key]" />
            <el-select v-else-if="f.checked && f.type === 'int'" v-model="form[f.key]" style="width:100%;">
              <el-option v-for="i in 31" :key="i" :label="String(i)" :value="i" />
            </el-select>
          </el-col>
        </el-row>
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
import { getPositionEvents, createPositionEvent, updatePositionEvent, deletePositionEvent } from '@/api/position'
import NameSelect from '@/components/NameSelect.vue'

const eventNames = ['入职', '转正', '调岗', '调薪', '离职']
const fieldDefs = [
  { key: 'attendance_group', label: '考勤组', type: 'string' },
  { key: 'has_annual_leave', label: '享有年假', type: 'bool' },
  { key: 'has_attendance_bonus', label: '享有全勤奖', type: 'bool' },
  { key: 'base_salary', label: '基本工资', type: 'number' },
  { key: 'performance_salary', label: '绩效工资基数', type: 'number' },
  { key: 'salary_days', label: '计薪天数', type: 'int' },
  { key: 'post_allowance', label: '职位津贴', type: 'number' },
  { key: 'meal_allowance', label: '餐补', type: 'number' },
  { key: 'housing_allowance', label: '房补', type: 'number' },
  { key: 'transport_allowance', label: '交通补贴', type: 'number' },
  { key: 'high_temp_allowance', label: '高温补贴', type: 'number' },
  { key: 'insurance_compensation', label: '保险补偿', type: 'number' },
  { key: 'fund_compensation', label: '公积金补偿', type: 'number' },
  { key: 'social_security_deduct', label: '社保代扣', type: 'number' },
  { key: 'housing_fund_deduct', label: '公积金代扣', type: 'number' },
]
const fields = reactive(fieldDefs.map(f => ({ ...f, checked: false })))

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '', event_name: '' })
const searchDateRange = ref<string[]>([])
const searchPersonId = ref<number | undefined>()

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive<Record<string, any>>({
  person_id: undefined, event_name: '', effective_date: '',
})

async function fetchData() {
  loading.value = true
  try {
    const data = await getPositionEvents({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: searchPersonId.value || '', event_name: search.event_name,
      start_date: searchDateRange.value?.[0] || '',
      end_date: searchDateRange.value?.[1] || '',
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() {
  searchPersonId.value = undefined; search.event_name = ''; searchDateRange.value = []; pageNum.value = 1; fetchData()
}

function resetForm() {
  form.person_id = undefined; form.event_name = ''; form.effective_date = ''
  fields.forEach(f => { f.checked = false; delete form[f.key] })
}

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit; resetForm()
  if (edit && row) {
    editId.value = row.id
    form.person_id = row.person_id; form.event_name = row.event_name; form.effective_date = row.effective_date
    fields.forEach(f => {
      if (row[f.key] !== null && row[f.key] !== undefined && row[f.key] !== '') {
        f.checked = true; form[f.key] = row[f.key]
      }
    })
  } else { editId.value = 0 }
  dialogVisible.value = true
}

async function handleSubmit() {
  const submitData: any = {}
  if (form.person_id) submitData.person_id = form.person_id
  if (form.event_name) submitData.event_name = form.event_name
  if (form.effective_date) submitData.effective_date = form.effective_date
  fields.filter(f => f.checked).forEach(f => { submitData[f.key] = form[f.key] })
  try {
    if (isEdit.value) {
      await updatePositionEvent(editId.value, submitData)
      ElMessage.success('更新成功')
    } else {
      await createPositionEvent(submitData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deletePositionEvent(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(() => {
  const query = new URLSearchParams(location.hash.split('?')[1] || '')
  const pid = query.get('person_id')
  if (pid) { search.person_id = pid }
  fetchData()
})
</script>
