<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-date-picker v-model="searchDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
      <el-select v-model="search.event_type" placeholder="事件类型" clearable style="width:120px;">
        <el-option label="出勤" value="出勤" />
        <el-option label="休假" value="休假" />
        <el-option label="加班" value="加班" />
        <el-option label="违纪" value="违纪" />
      </el-select>
      <el-select v-model="search.sub_type" placeholder="子类型" clearable style="width:140px;">
        <el-option label="普通出勤" value="普通出勤" />
        <el-option label="补班出勤" value="补班出勤" />
        <el-option label="事假" value="事假" />
        <el-option label="病假" value="病假" />
        <el-option label="年假" value="年假" />
        <el-option label="调休" value="调休" />
        <el-option label="工作日加班" value="工作日加班" />
        <el-option label="节假日加班" value="节假日加班" />
      </el-select>
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'attendance:write'" @click="openDialog(false)">新增事件</el-button>
      <el-button v-permission="'attendance:write'" @click="openCrossDay">跨天录入</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column prop="event_date" label="日期" width="120" />
      <el-table-column prop="event_type" label="事件类型" width="80" />
      <el-table-column prop="sub_type" label="子类型" width="100" />
      <el-table-column label="时长(天)" width="100">
        <template #default="{ row }">{{ hoursToDays(row.hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="150" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" v-permission="'attendance:write'" @click="openDialog(true, row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'attendance:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑考勤事件' : '新增考勤事件'" width="500px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="人员ID" required>
          <el-input-number v-model="form.person_id" :min="1" />
        </el-form-item>
        <el-form-item label="日期" required>
          <el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="form.event_type" @change="onEventTypeChange">
            <el-option label="出勤" value="出勤" />
            <el-option label="休假" value="休假" />
            <el-option label="加班" value="加班" />
            <el-option label="违纪" value="违纪" />
          </el-select>
        </el-form-item>
        <el-form-item label="子类型" required>
          <el-select v-model="form.sub_type">
            <el-option v-for="st in getSubTypes(form.event_type)" :key="st" :label="st" :value="st" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(天)" required>
          <el-input-number v-model="form.hours_days" :min="0" :precision="2" :step="0.5" />
        </el-form-item>
        <el-form-item label="是否特批">
          <el-switch v-model="form.is_special_approval" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="crossDayVisible" title="跨天录入" width="500px">
      <el-form ref="crossFormRef" :model="crossForm" label-width="100px">
        <el-form-item label="人员ID" required>
          <el-input-number v-model="crossForm.person_id" :min="1" />
        </el-form-item>
        <el-form-item label="日期范围" required>
          <el-date-picker v-model="crossForm.dateRange" type="daterange" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="事件类型" required>
          <el-select v-model="crossForm.event_type" @change="onCrossTypeChange">
            <el-option label="出勤" value="出勤" />
            <el-option label="休假" value="休假" />
            <el-option label="加班" value="加班" />
          </el-select>
        </el-form-item>
        <el-form-item label="子类型" required>
          <el-select v-model="crossForm.sub_type">
            <el-option v-for="st in getSubTypes(crossForm.event_type)" :key="st" :label="st" :value="st" />
          </el-select>
        </el-form-item>
        <el-form-item label="每天时长(天)" required>
          <el-input-number v-model="crossForm.hours_per_day_days" :min="0" :precision="2" :step="0.5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="crossDayVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCrossSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAttendanceEvents, createAttendanceEvent, updateAttendanceEvent, deleteAttendanceEvent, createCrossDayEvent } from '@/api/attendance'
import { daysToHours, hoursToDays } from '@/utils/unit'

const subTypesMap: Record<string, string[]> = {
  '出勤': ['普通出勤', '补班出勤', '外勤出勤'],
  '休假': ['调休', '事假', '病假', '年假', '法定假', '福利假'],
  '加班': ['工作日加班', '节假日加班'],
  '违纪': ['缺卡', '迟到', '早退'],
}

function getSubTypes(et: string) { return subTypesMap[et] || [] }

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '', event_type: '', sub_type: '' })
const searchDateRange = ref<string[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const formRef = ref()
const form = reactive({
  person_id: undefined as number | undefined, event_date: '', event_type: '', sub_type: '',
  hours_days: 0, is_special_approval: false, remark: '', punch_time: '',
})

const crossDayVisible = ref(false)
const crossForm = reactive({
  person_id: undefined as number | undefined, dateRange: [] as string[],
  event_type: '', sub_type: '', hours_per_day_days: 0, remark: '',
})
const crossFormRef = ref()

async function fetchData() {
  loading.value = true
  try {
    const data = await getAttendanceEvents({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: search.person_id, event_type: search.event_type, sub_type: search.sub_type,
      start_date: searchDateRange.value?.[0] || '',
      end_date: searchDateRange.value?.[1] || '',
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() {
  search.person_id = ''; search.event_type = ''; search.sub_type = ''
  searchDateRange.value = []; pageNum.value = 1; fetchData()
}

function onEventTypeChange() { form.sub_type = '' }
function onCrossTypeChange() { crossForm.sub_type = '' }

function openDialog(edit: boolean, row?: any) {
  isEdit.value = edit
  if (edit && row) {
    editId.value = row.id
    form.person_id = row.person_id; form.event_date = row.event_date
    form.event_type = row.event_type; form.sub_type = row.sub_type
    form.hours_days = hoursToDays(row.hours ?? 0)
    form.is_special_approval = row.is_special_approval
    form.remark = row.remark || ''
  } else {
    editId.value = 0
    form.person_id = undefined; form.event_date = ''; form.event_type = ''; form.sub_type = ''
    form.hours_days = 0; form.is_special_approval = false; form.remark = ''
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  const data: any = {
    person_id: form.person_id, event_date: form.event_date,
    event_type: form.event_type, sub_type: form.sub_type,
    hours: daysToHours(form.hours_days),
    is_special_approval: form.is_special_approval, remark: form.remark,
  }
  try {
    if (isEdit.value) { await updateAttendanceEvent(editId.value, data); ElMessage.success('更新成功') }
    else { await createAttendanceEvent(data); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchData()
  } catch (e) {}
}

function openCrossDay() { crossDayVisible.value = true }

async function handleCrossSubmit() {
  try {
    await createCrossDayEvent({
      person_id: crossForm.person_id,
      start_date: crossForm.dateRange[0], end_date: crossForm.dateRange[1],
      event_type: crossForm.event_type, sub_type: crossForm.sub_type,
      hours_per_day: daysToHours(crossForm.hours_per_day_days),
    })
    ElMessage.success('创建成功'); crossDayVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteAttendanceEvent(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
