<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3>职务事件 - {{ personName }}</h3>
        <el-button type="primary" @click="dialogVisible = true">新增事件</el-button>
      </div>
      <el-table :data="events" stripe>
        <el-table-column prop="event_name" label="事件名称" min-width="120" />
        <el-table-column prop="effective_date" label="生效日期" min-width="110" />
        <el-table-column prop="attendance_group" label="考勤组" min-width="100">
          <template #default="{ row }">{{ row.attendance_group || '-' }}</template>
        </el-table-column>
        <el-table-column prop="entry_date" label="入职日期" min-width="110">
          <template #default="{ row }">{{ row.entry_date || '-' }}</template>
        </el-table-column>
        <el-table-column prop="leave_date" label="离职日期" min-width="110">
          <template #default="{ row }">{{ row.leave_date || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑事件' : '新增事件'" width="700px" @closed="resetForm">
      <el-form :model="form" label-width="120px">
        <el-form-item label="事件名称" required>
          <el-input v-model="form.event_name" />
        </el-form-item>
        <el-form-item label="生效日期" required>
          <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="考勤组">
          <el-input v-model="form.attendance_group" />
        </el-form-item>
        <el-form-item label="入职日期">
          <el-date-picker v-model="form.entry_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="离职日期">
          <el-date-picker v-model="form.leave_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="享有年假">
          <el-switch v-model="form.has_annual_leave" />
        </el-form-item>
        <el-form-item label="享有全勤奖">
          <el-switch v-model="form.has_attendance_bonus" />
        </el-form-item>
        <el-form-item label="基本工资">
          <el-input-number v-model="form.base_salary" :precision="2" :min="0" />
        </el-form-item>
        <el-form-item label="绩效工资基数">
          <el-input-number v-model="form.performance_salary" :precision="2" :min="0" />
        </el-form-item>
        <el-form-item label="计薪天数">
          <el-input-number v-model="form.salary_days" :min="0" />
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
import { ref, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/position'
import * as personApi from '../../api/person'

const route = useRoute()
const personId = Number(route.params.id)
const personName = ref('')
const events = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)

const defaultForm = {
  id: 0, person_id: personId, event_name: '', effective_date: '',
  attendance_group: '', entry_date: '', leave_date: '',
  has_annual_leave: false, has_attendance_bonus: false,
  base_salary: 0, performance_salary: 0, salary_days: 0,
}
const form = reactive({ ...defaultForm })

function resetForm() { Object.assign(form, defaultForm) }

function openEdit(row: any) {
  isEdit.value = true
  Object.assign(form, { ...row, person_id: personId })
  dialogVisible.value = true
}

async function fetchData() {
  const [personRes, eventRes] = await Promise.all([
    personApi.getPerson(personId),
    api.getPositionEvents({ person_id: personId }),
  ])
  personName.value = personRes.data.name
  events.value = eventRes.data || []
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEdit.value) {
      await api.updatePositionEvent(form.id, form)
      ElMessage.success('修改成功')
    } else {
      await api.createPositionEvent(form)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确定删除该事件？', '确认删除', { type: 'warning' })
  await api.deletePositionEvent(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

fetchData()
</script>
