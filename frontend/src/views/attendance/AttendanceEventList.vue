<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div style="display: flex; gap: 8px;">
          <el-select v-model="searchForm.event_type" placeholder="事件类型" clearable style="width: 120px;">
            <el-option label="出勤" value="出勤" />
            <el-option label="休假" value="休假" />
            <el-option label="加班" value="加班" />
            <el-option label="违纪" value="违纪" />
          </el-select>
          <el-date-picker v-model="searchForm.date_range" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
        </div>
        <el-button type="primary" @click="dialogVisible = true">新增事件</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="姓名" min-width="80">
          <template #default="{ row }">{{ row.person?.name }}</template>
        </el-table-column>
        <el-table-column prop="event_date" label="日期" min-width="110" />
        <el-table-column prop="event_type" label="类型" min-width="80" />
        <el-table-column prop="sub_type" label="子类型" min-width="100" />
        <el-table-column prop="hours" label="时长(小时)" min-width="100" />
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑事件' : '新增事件'" width="600px" @closed="resetForm">
      <el-form :model="form" label-width="100px">
        <el-form-item label="人员" required>
          <el-input v-model.number="form.person_id" placeholder="输入人员ID" />
        </el-form-item>
        <el-form-item label="日期" required>
          <el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type">
            <el-option label="出勤" value="出勤" />
            <el-option label="休假" value="休假" />
            <el-option label="加班" value="加班" />
            <el-option label="违纪" value="违纪" />
          </el-select>
        </el-form-item>
        <el-form-item label="子类型">
          <el-select v-model="form.sub_type">
            <el-option v-if="form.event_type === '出勤'" label="普通出勤" value="普通出勤" />
            <el-option v-if="form.event_type === '出勤'" label="补班出勤" value="补班出勤" />
            <el-option v-if="form.event_type === '休假'" label="调休" value="调休" />
            <el-option v-if="form.event_type === '休假'" label="事假" value="事假" />
            <el-option v-if="form.event_type === '休假'" label="病假" value="病假" />
            <el-option v-if="form.event_type === '休假'" label="年假" value="年假" />
            <el-option v-if="form.event_type === '休假'" label="法定假" value="法定假" />
            <el-option v-if="form.event_type === '休假'" label="福利假" value="福利假" />
            <el-option v-if="form.event_type === '加班'" label="工作日加班" value="工作日加班" />
            <el-option v-if="form.event_type === '加班'" label="节假日加班" value="节假日加班" />
            <el-option v-if="form.event_type === '违纪'" label="缺卡" value="缺卡" />
            <el-option v-if="form.event_type === '违纪'" label="迟到" value="迟到" />
            <el-option v-if="form.event_type === '违纪'" label="早退" value="早退" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(小时)">
          <el-input-number v-model="form.hours" :precision="1" :min="0" :max="24" />
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
import * as api from '../../api/attendance'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const isEdit = ref(false)
const searchForm = reactive({ event_type: '', date_range: [] as string[] })

const defaultForm = { id: 0, person_id: 0, event_date: '', event_type: '出勤', sub_type: '', hours: 0, remark: '', is_special_approval: false }
const form = reactive({ ...defaultForm })

const queryParams = computed(() => {
  const params: any = { page: page.value, page_size: pageSize.value }
  if (searchForm.event_type) params.event_type = searchForm.event_type
  if (searchForm.date_range?.[0]) params.start_date = searchForm.date_range[0]
  if (searchForm.date_range?.[1]) params.end_date = searchForm.date_range[1]
  return params
})

function resetForm() { Object.assign(form, defaultForm) }

function openEdit(row: any) {
  isEdit.value = true
  Object.assign(form, { ...row })
  dialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getAttendanceEventList(queryParams.value)
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await api.updateAttendanceEvent(form.id, form)
      ElMessage.success('修改成功')
    } else {
      await api.createAttendanceEvent(form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确定删除该事件？', '确认删除', { type: 'warning' })
  await api.deleteAttendanceEvent(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

fetchData()
</script>
