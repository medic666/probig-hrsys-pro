<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>考勤汇总</span>
          <div>
            <el-input v-model="calcMonth" type="month" style="width:180px;margin-right:10px" placeholder="选择月份" />
            <el-button type="primary" @click="handleCalc">开始核算</el-button>
          </div>
        </div>
      </template>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column label="人员" min-width="100">
          <template #default="{ row }">{{ row.person?.name || row.person_id }}</template>
        </el-table-column>
        <el-table-column prop="belong_month" label="月份" width="100" />
        <el-table-column prop="work_days" label="出勤天数" width="90" />
        <el-table-column prop="sick_leave_days" label="病假" width="70" />
        <el-table-column prop="personal_leave_days" label="事假" width="70" />
        <el-table-column prop="annual_leave_days" label="年假" width="70" />
        <el-table-column prop="overtime_workday_hours" label="加班(h)" width="90" />
        <el-table-column prop="violation_count" label="违纪" width="70" />
        <el-table-column prop="is_locked" label="锁定" width="70">
          <template #default="{ row }">{{ row.is_locked ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button size="small" :type="row.is_locked ? 'warning' : 'success'" @click="toggleLock(row)">
              {{ row.is_locked ? '解锁' : '锁定' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([])
const loading = ref(false)
const calcMonth = ref('')

async function fetchList() {
  loading.value = true
  const res = await request.get('/attendance-summaries', { params: { belong_month: calcMonth.value } })
  list.value = res.data.list
  loading.value = false
}

async function handleCalc() {
  if (!calcMonth.value) { ElMessage.error('请选择月份'); return }
  await request.post('/attendance-summaries/calc', { belong_month: calcMonth.value, person_ids: [] })
  ElMessage.success('核算完成')
  fetchList()
}

async function toggleLock(row: any) {
  await request.put(`/attendance-summaries/${row.id}/lock`, { is_locked: !row.is_locked })
  ElMessage.success(row.is_locked ? '已解锁' : '已锁定')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
