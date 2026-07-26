<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-input v-model="search.belong_month" placeholder="月份(YYYY-MM)" clearable style="width:160px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'attendance:calc'" @click="openCalcDialog">批量核算</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="belong_month" label="月份" width="100" />
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column label="出勤工资" width="120">
        <template #default="{ row }">{{ row.attendance_salary?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="工作日加班工资" width="130">
        <template #default="{ row }">{{ row.overtime_workday_salary?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="节假日加班工资" width="130">
        <template #default="{ row }">{{ row.overtime_holiday_salary?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="全勤奖" width="100">
        <template #default="{ row }">{{ row.attendance_bonus?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{}">
          <ProjectionStatus :status="1" />
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="calcVisible" title="批量核算" width="400px">
      <el-form label-width="100px">
        <el-form-item label="核算月份" required>
          <el-input v-model="calcMonth" placeholder="YYYY-MM" />
        </el-form-item>
        <el-form-item label="人员ID(可选)">
          <el-input v-model="calcPersonIds" placeholder="多个用逗号分隔，留空核算全部" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="calcVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCalc">开始核算</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAttendanceSalaryList, calculateAttendanceSalary } from '@/api/attendance'
import ProjectionStatus from '@/components/ProjectionStatus.vue'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '', belong_month: '' })

const calcVisible = ref(false)
const calcMonth = ref('')
const calcPersonIds = ref('')

async function fetchData() {
  loading.value = true
  try {
    const data = await getAttendanceSalaryList({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: search.person_id, belong_month: search.belong_month,
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.person_id = ''; search.belong_month = ''; pageNum.value = 1; fetchData() }

function openCalcDialog() { calcVisible.value = true }

async function handleCalc() {
  if (!calcMonth.value) { ElMessage.warning('请输入核算月份'); return }
  const personIds = calcPersonIds.value ? calcPersonIds.value.split(',').map(Number).filter(Boolean) : []
  try {
    const data = await calculateAttendanceSalary({ person_ids: personIds, belong_month: calcMonth.value })
    ElMessage.success(`核算完成: 成功${data.success_count}条, 跳过${data.skipped?.length || 0}条`)
    calcVisible.value = false; fetchData()
  } catch (e) {}
}

onMounted(fetchData)
</script>
