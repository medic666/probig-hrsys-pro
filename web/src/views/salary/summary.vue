<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-input v-model="search.belong_month" placeholder="月份(YYYY-MM)" clearable style="width:160px;" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'salary:calc'" @click="openCalcDialog">批量核算</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="belong_month" label="月份" width="100" />
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column label="出勤工资" width="110">
        <template #default="{ row }">{{ row.attendance_salary?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="加班工资" width="110">
        <template #default="{ row }">{{ ((row.overtime_workday_salary ?? 0) + (row.overtime_holiday_salary ?? 0)).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="绩效工资" width="100">
        <template #default="{ row }">{{ row.performance_salary?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="补贴合计" width="100">
        <template #default="{ row }">{{ ((row.post_allowance ?? 0) + (row.meal_allowance ?? 0) + (row.housing_allowance ?? 0) + (row.transport_allowance ?? 0)).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="调整项合计" width="100">
        <template #default="{ row }">{{ row.total_adjustment?.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="代扣合计" width="100">
        <template #default="{ row }">{{ ((row.social_security_deduct ?? 0) + (row.housing_fund_deduct ?? 0) + (row.tax_deduct ?? 0)).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="实发工资" width="110">
        <template #default="{ row }">
          <strong>{{ row.final_salary?.toFixed(2) }}</strong>
        </template>
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

    <el-dialog v-model="calcVisible" title="批量核算工资" width="400px">
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
import { getSalarySummaries, calculateSalarySummaries } from '@/api/salary'
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
    const data = await getSalarySummaries({ pageNum: pageNum.value, pageSize: pageSize.value, ...search })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.person_id = ''; search.belong_month = ''; pageNum.value = 1; fetchData() }

function openCalcDialog() { calcVisible.value = true }

async function handleCalc() {
  if (!calcMonth.value) { ElMessage.warning('请输入核算月份'); return }
  const ids = calcPersonIds.value ? calcPersonIds.value.split(',').map(Number).filter(Boolean) : []
  try {
    const data = await calculateSalarySummaries({ person_ids: ids, belong_month: calcMonth.value })
    ElMessage.success(`核算完成: 成功${data.success_count}条, 跳过${data.skipped?.length || 0}条`)
    calcVisible.value = false; fetchData()
  } catch (e) {}
}

onMounted(fetchData)
</script>
