<template>
  <div class="page-container">
    <div class="search-bar">
      <NameSelect v-model="searchPersonId" style="width:180px;" />
      <el-input v-model="search.belong_month" placeholder="归属月份(YYYY-MM)" clearable style="width:180px;" />
      <el-select v-model="search.event_type" placeholder="事件类型" clearable style="width:140px;">
        <el-option label="绩效系数" value="绩效系数" />
        <el-option label="提成" value="提成" />
        <el-option label="奖惩" value="奖惩" />
        <el-option label="个税扣除" value="个税扣除" />
        <el-option label="其他" value="其他" />
      </el-select>
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'salary:write'" @click="openDialog(false)">新增事件</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column prop="belong_month" label="归属月份" width="100" />
      <el-table-column prop="event_type" label="事件类型" width="100" />
      <el-table-column prop="event_name" label="事件名称" width="120" />
      <el-table-column prop="amount" label="金额/系数" width="100" />
      <el-table-column prop="remark" label="备注" min-width="150" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" v-permission="'salary:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'salary:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑工资事件' : '新增工资事件'" width="500px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="人员" required>
          <NameSelect v-model="form.person_id" style="width:100%;" />
        </el-form-item>
        <el-form-item label="归属月份" required>
          <el-date-picker v-model="form.belong_month" type="month" value-format="YYYY-MM" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type">
            <el-option label="绩效系数" value="绩效系数" />
            <el-option label="提成" value="提成" />
            <el-option label="奖惩" value="奖惩" />
            <el-option label="借款还款" value="借款还款" />
            <el-option label="个税扣除" value="个税扣除" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="事件名称">
          <el-input v-model="form.event_name" />
        </el-form-item>
        <el-form-item label="金额/系数" required>
          <el-input-number v-model="form.amount" :precision="2" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
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
import { getSalaryEvents, createSalaryEvent, updateSalaryEvent, deleteSalaryEvent } from '@/api/salary'
import NameSelect from '@/components/NameSelect.vue'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ belong_month: '', event_type: '' })
const searchPersonId = ref<number | undefined>()

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({ person_id: undefined as number | undefined, belong_month: '', event_type: '', event_name: '', amount: 0, remark: '' })

async function fetchData() {
  loading.value = true
  try {
    const data = await getSalaryEvents({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: searchPersonId.value || '', ...search
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { searchPersonId.value = undefined; search.belong_month = ''; search.event_type = ''; pageNum.value = 1; fetchData() }

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) { editId.value = row.id; Object.assign(form, row) }
  else { editId.value = 0; Object.assign(form, { person_id: undefined, belong_month: '', event_type: '', event_name: '', amount: 0, remark: '' }) }
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    if (isEdit.value) { await updateSalaryEvent(editId.value, { ...form }); ElMessage.success('更新成功') }
    else { await createSalaryEvent({ ...form }); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteSalaryEvent(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
