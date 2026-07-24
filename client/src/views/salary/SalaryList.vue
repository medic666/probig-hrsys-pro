<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工资事件</span>
          <el-button type="primary" @click="showDialog()">新增工资事件</el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="filters.person_id" placeholder="人员ID" style="width:120px" />
        <el-input v-model="filters.belong_month" type="month" placeholder="归属月份" style="width:150px" />
        <el-select v-model="filters.event_type" placeholder="事件类型" clearable style="width:130px">
          <el-option label="绩效调整" value="绩效调整" /><el-option label="提成" value="提成" /><el-option label="奖惩" value="奖惩" /><el-option label="借款扣除" value="借款扣除" /><el-option label="个税扣除" value="个税扣除" /><el-option label="其他" value="其他" />
        </el-select>
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column label="人员" min-width="100">
          <template #default="{ row }">{{ row.person?.name || row.person_id }}</template>
        </el-table-column>
        <el-table-column prop="belong_month" label="归属月份" width="100" />
        <el-table-column prop="event_type" label="类型" width="100" />
        <el-table-column prop="event_name" label="事件名称" width="120" />
        <el-table-column prop="amount" label="金额/系数" width="100" />
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

    <el-dialog v-model="dialogVisible" title="新增工资事件" width="450px">
      <el-form ref="salaryFormRef" :model="salaryForm" :rules="salaryRules">
        <el-form-item label="人员ID" prop="person_id"><el-input-number v-model="salaryForm.person_id" style="width:100%" /></el-form-item>
        <el-form-item label="归属月份" prop="belong_month"><el-input v-model="salaryForm.belong_month" type="month" style="width:100%" /></el-form-item>
        <el-form-item label="事件类型" prop="event_type">
          <el-select v-model="salaryForm.event_type" style="width:100%">
            <el-option label="绩效调整" value="绩效调整" /><el-option label="提成" value="提成" /><el-option label="奖惩" value="奖惩" /><el-option label="借款扣除" value="借款扣除" /><el-option label="个税扣除" value="个税扣除" /><el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="事件名称" prop="event_name"><el-input v-model="salaryForm.event_name" /></el-form-item>
        <el-form-item label="金额/系数" prop="amount"><el-input-number v-model="salaryForm.amount" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="salaryForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitSalary">确定</el-button>
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
const filters = reactive({ person_id: '', belong_month: '', event_type: '' })

const dialogVisible = ref(false); const salaryFormRef = ref()
const salaryForm = reactive({ person_id: undefined as number | undefined, belong_month: '', event_type: '其他', event_name: '', amount: 0, remark: '' })
const salaryRules = {
  person_id: [{ required: true }],
  belong_month: [{ required: true }],
  event_type: [{ required: true }],
  event_name: [{ required: true }],
  amount: [{ required: true }],
}

async function fetchList() {
  loading.value = true
  const params: any = { page: page.value, page_size: pageSize.value }
  if (filters.person_id) params.person_id = filters.person_id
  if (filters.belong_month) params.belong_month = filters.belong_month
  if (filters.event_type) params.event_type = filters.event_type
  const res = await request.get('/salary-events', { params })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function showDialog() {
  Object.assign(salaryForm, { person_id: undefined, belong_month: '', event_type: '其他', event_name: '', amount: 0, remark: '' })
  dialogVisible.value = true
}

async function submitSalary() {
  const valid = await salaryFormRef.value?.validate().catch(() => false)
  if (!valid) return
  await request.post('/salary-events', salaryForm)
  dialogVisible.value = false; fetchList(); ElMessage.success('新增成功')
}

async function handleDelete(id: number) {
  await request.delete(`/salary-events/${id}`)
  fetchList(); ElMessage.success('删除成功')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.search-bar { margin-bottom: 16px; display: flex; gap: 10px; flex-wrap: wrap; }
</style>
