<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>假勤事件</span>
          <div>
            <el-button type="primary" @click="showSingleDialog()">单条录入</el-button>
            <el-button type="success" @click="showBatchDialog()">批量录入</el-button>
          </div>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="filters.person_id" placeholder="人员ID" style="width:120px" />
        <el-select v-model="filters.event_type" placeholder="事件类型" clearable style="width:120px">
          <el-option label="出勤" value="出勤" />
          <el-option label="休假" value="休假" />
          <el-option label="加班" value="加班" />
          <el-option label="违纪" value="违纪" />
          <el-option label="年假调整" value="年假调整" />
        </el-select>
        <el-input v-model="filters.start_date" type="date" placeholder="开始日期" style="width:150px" />
        <el-input v-model="filters.end_date" type="date" placeholder="结束日期" style="width:150px" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="event_date" label="日期" width="120" />
        <el-table-column label="人员" min-width="100">
          <template #default="{ row }">{{ row.person?.name || row.person_id }}</template>
        </el-table-column>
        <el-table-column prop="event_type" label="类型" width="80" />
        <el-table-column prop="sub_type" label="子类型" width="100" />
        <el-table-column prop="hours" label="时长(h)" width="80" />
        <el-table-column prop="remark" label="备注" min-width="150" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
              <template #reference><el-button size="small" type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top:16px"
      />
    </el-card>

    <el-dialog v-model="singleDialog" title="单条录入" width="500px">
      <el-form ref="singleFormRef" :model="singleForm" :rules="singleRules">
        <el-form-item label="人员ID" prop="person_id"><el-input-number v-model="singleForm.person_id" style="width:100%" /></el-form-item>
        <el-form-item label="日期" prop="event_date"><el-date-picker v-model="singleForm.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型" prop="event_type">
          <el-select v-model="singleForm.event_type" style="width:100%">
            <el-option label="出勤" value="出勤" /><el-option label="休假" value="休假" /><el-option label="加班" value="加班" /><el-option label="违纪" value="违纪" /><el-option label="年假调整" value="年假调整" />
          </el-select>
        </el-form-item>
        <el-form-item label="子类型" prop="sub_type">
          <el-select v-model="singleForm.sub_type" style="width:100%">
            <el-option label="普通出勤" value="普通出勤" /><el-option label="补班出勤" value="补班出勤" />
            <el-option label="调休" value="调休" /><el-option label="事假" value="事假" /><el-option label="病假" value="病假" /><el-option label="年假" value="年假" /><el-option label="法定假" value="法定假" /><el-option label="福利假" value="福利假" />
            <el-option label="工作日加班" value="工作日加班" /><el-option label="节假日加班" value="节假日加班" />
            <el-option label="缺卡" value="缺卡" /><el-option label="迟到" value="迟到" /><el-option label="早退" value="早退" />
            <el-option label="年假配发" value="年假配发" /><el-option label="年假结转" value="年假结转" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(h)"><el-input-number v-model="singleForm.hours" :precision="1" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="singleForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleDialog = false">取消</el-button>
        <el-button type="primary" @click="submitSingle">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialog" title="批量录入" width="500px">
      <el-form :model="batchForm">
        <el-form-item label="人员ID（逗号分隔）"><el-input v-model="batchPersonIds" placeholder="1,2,3" /></el-form-item>
        <el-form-item label="开始日期"><el-date-picker v-model="batchForm.start_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="结束日期"><el-date-picker v-model="batchForm.end_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型"><el-select v-model="batchForm.event_type" style="width:100%"><el-option label="出勤" value="出勤" /><el-option label="休假" value="休假" /><el-option label="加班" value="加班" /></el-select></el-form-item>
        <el-form-item label="子类型"><el-select v-model="batchForm.sub_type" style="width:100%"><el-option label="普通出勤" value="普通出勤" /><el-option label="调休" value="调休" /><el-option label="事假" value="事假" /></el-select></el-form-item>
        <el-form-item label="时长(h)"><el-input-number v-model="batchForm.hours" :precision="1" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="batchForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchDialog = false">取消</el-button>
        <el-button type="primary" @click="submitBatch">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20)
const loading = ref(false)
const filters = reactive({ person_id: '', event_type: '', start_date: '', end_date: '' })

const singleDialog = ref(false); const singleFormRef = ref()
const singleForm = reactive({ person_id: undefined as number | undefined, event_date: '', event_type: '', sub_type: '', hours: undefined as number | undefined, remark: '' })
const singleRules = { person_id: [{ required: true }], event_date: [{ required: true }], event_type: [{ required: true }], sub_type: [{ required: true }] }

const batchDialog = ref(false); const batchPersonIds = ref('')
const batchForm = reactive({ start_date: '', end_date: '', event_type: '', sub_type: '', hours: undefined as number | undefined, remark: '' })

async function fetchList() {
  loading.value = true
  const params: any = { page: page.value, page_size: pageSize.value }
  if (filters.person_id) params.person_id = filters.person_id
  if (filters.event_type) params.event_type = filters.event_type
  if (filters.start_date) params.start_date = filters.start_date
  if (filters.end_date) params.end_date = filters.end_date
  const res = await request.get('/attendance-events', { params })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function showSingleDialog() {
  Object.assign(singleForm, { person_id: undefined, event_date: '', event_type: '', sub_type: '', hours: undefined, remark: '' })
  singleDialog.value = true
}

async function submitSingle() {
  const valid = await singleFormRef.value?.validate().catch(() => false)
  if (!valid) return
  await request.post('/attendance-events', singleForm)
  singleDialog.value = false; fetchList(); ElMessage.success('录入成功')
}

async function submitBatch() {
  const ids = batchPersonIds.value.split(',').map(s => Number(s.trim())).filter(Boolean)
  if (ids.length === 0) { ElMessage.error('请填写人员ID'); return }
  await request.post('/attendance-events/batch', { ...batchForm, person_ids: ids })
  batchDialog.value = false; fetchList(); ElMessage.success('批量录入成功')
}

async function handleDelete(id: number) {
  await request.delete(`/attendance-events/${id}`)
  fetchList(); ElMessage.success('删除成功')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-bar { margin-bottom: 16px; display: flex; gap: 10px; flex-wrap: wrap; }
</style>
