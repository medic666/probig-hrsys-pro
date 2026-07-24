<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工资汇总</span>
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
        <el-table-column prop="attendance_salary" label="出勤工资" width="110" />
        <el-table-column prop="overtime_salary" label="加班工资" width="110" />
        <el-table-column prop="attendance_bonus" label="全勤奖" width="100" />
        <el-table-column prop="performance_salary" label="绩效工资" width="110" />
        <el-table-column prop="total_allowance" label="补贴合计" width="110" />
        <el-table-column prop="total_adjustment" label="调整项" width="100" />
        <el-table-column prop="total_deduction" label="代扣合计" width="110" />
        <el-table-column prop="final_salary" label="实发工资" width="110" fixed="right">
          <template #default="{ row }"><b style="color:#409EFF">{{ row.final_salary }}</b></template>
        </el-table-column>
        <el-table-column prop="is_locked" label="锁定" width="70">
          <template #default="{ row }">{{ row.is_locked ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
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
  const res = await request.get('/salary-summaries', { params: { belong_month: calcMonth.value } })
  list.value = res.data.list
  loading.value = false
}

async function handleCalc() {
  if (!calcMonth.value) { ElMessage.error('请选择月份'); return }
  await request.post('/salary-summaries/calc', { belong_month: calcMonth.value, person_ids: [] })
  ElMessage.success('核算完成')
  fetchList()
}

async function toggleLock(row: any) {
  await request.put(`/salary-summaries/${row.id}/lock`, { is_locked: !row.is_locked })
  ElMessage.success(row.is_locked ? '已解锁' : '已锁定')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
