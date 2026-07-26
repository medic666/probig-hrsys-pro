<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-date-picker v-model="searchDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="work_date" label="日期" width="120" />
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column label="记出勤(天)" width="100">
        <template #default="{ row }">{{ hoursToDays(row.work_hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="工作日加班(天)" width="120">
        <template #default="{ row }">{{ hoursToDays(row.overtime_workday_hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="节假日加班(天)" width="120">
        <template #default="{ row }">{{ hoursToDays(row.overtime_holiday_hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="事假" width="70">
        <template #default="{ row }">{{ row.has_personal_leave ? '是' : '否' }}</template>
      </el-table-column>
      <el-table-column prop="violation_count" label="违纪次数" width="80" />
      <el-table-column prop="remark" label="备注" min-width="150" />
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getDailyList } from '@/api/attendance'
import { hoursToDays } from '@/utils/unit'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '' })
const searchDateRange = ref<string[]>([])

async function fetchData() {
  loading.value = true
  try {
    const data = await getDailyList({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: search.person_id,
      start_date: searchDateRange.value?.[0] || '',
      end_date: searchDateRange.value?.[1] || '',
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.person_id = ''; searchDateRange.value = []; pageNum.value = 1; fetchData() }

onMounted(fetchData)
</script>
