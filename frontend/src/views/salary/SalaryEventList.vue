<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div style="display: flex; gap: 8px;">
          <el-input v-model="searchForm.person_id" placeholder="人员ID" style="width: 120px;" />
          <el-input v-model="searchForm.belong_month" placeholder="归属月份 YYYY-MM" style="width: 160px;" />
          <el-select v-model="searchForm.event_type" placeholder="事件类型" clearable style="width: 140px;">
            <el-option label="绩效调整" value="绩效调整" />
            <el-option label="提成" value="提成" />
            <el-option label="奖惩" value="奖惩" />
            <el-option label="借款扣除" value="借款扣除" />
            <el-option label="个税扣除" value="个税扣除" />
            <el-option label="其他" value="其他" />
          </el-select>
        </div>
        <el-button type="primary" @click="dialogVisible = true">新增事件</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="姓名" min-width="80">
          <template #default="{ row }">{{ row.person?.name }}</template>
        </el-table-column>
        <el-table-column prop="belong_month" label="归属月份" min-width="100" />
        <el-table-column prop="event_type" label="事件类型" min-width="100" />
        <el-table-column prop="event_name" label="事件名称" min-width="120" />
        <el-table-column prop="amount" label="金额/系数" min-width="100" />
        <el-table-column prop="remark" label="备注" min-width="140" />
        <el-table-column label="操作" min-width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑事件' : '新增事件'" width="500px" @closed="resetForm">
      <el-form :model="form" label-width="100px">
        <el-form-item label="人员ID" required>
          <el-input v-model.number="form.person_id" />
        </el-form-item>
        <el-form-item label="归属月份" required>
          <el-input v-model="form.belong_month" placeholder="YYYY-MM" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type">
            <el-option label="绩效调整" value="绩效调整" />
            <el-option label="提成" value="提成" />
            <el-option label="奖惩" value="奖惩" />
            <el-option label="借款扣除" value="借款扣除" />
            <el-option label="个税扣除" value="个税扣除" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="金额/系数">
          <el-input-number v-model="form.amount" :precision="2" />
        </el-form-item>
        <el-form-item label="事件名称">
          <el-input v-model="form.event_name" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
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
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/salary'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const isEdit = ref(false)
const searchForm = reactive({ person_id: '', belong_month: '', event_type: '' })

const defaultForm = { id: 0, person_id: 0, belong_month: '', event_type: '其他', amount: 0, event_name: '', remark: '' }
const form = reactive({ ...defaultForm })

const queryParams = computed(() => {
  const params: any = { page: page.value, page_size: pageSize.value }
  if (searchForm.person_id) params.person_id = searchForm.person_id
  if (searchForm.belong_month) params.belong_month = searchForm.belong_month
  if (searchForm.event_type) params.event_type = searchForm.event_type
  return params
})

function resetForm() { Object.assign(form, defaultForm) }
function openEdit(row: any) { isEdit.value = true; Object.assign(form, { ...row }); dialogVisible.value = true }

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getSalaryEventList(queryParams.value)
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) { await api.updateSalaryEvent(form.id, form); ElMessage.success('修改成功') }
    else { await api.createSalaryEvent(form); ElMessage.success('新增成功') }
    dialogVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确定删除该事件？', '确认删除', { type: 'warning' })
  await api.deleteSalaryEvent(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

fetchData()
</script>
