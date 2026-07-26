<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-select v-model="search.leave_type" placeholder="假期类型" clearable style="width:140px;">
        <el-option label="年假" value="annual_leave" />
        <el-option label="调休" value="time_off" />
      </el-select>
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column label="假期类型" width="80">
        <template #default="{ row }">{{ row.leave_type === 'annual_leave' ? '年假' : '调休' }}</template>
      </el-table-column>
      <el-table-column label="可用额度(天)" width="120">
        <template #default="{ row }">{{ hoursToDays(row.balance_hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" @click="showDetail(row)">查看明细</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="detailVisible" title="额度明细" width="500px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="累计配发(天)">{{ hoursToDays(detail.grant ?? 0) }}</el-descriptions-item>
        <el-descriptions-item label="累计调整(天)">{{ hoursToDays(detail.adjust ?? 0) }}</el-descriptions-item>
        <el-descriptions-item label="累计结转扣减(天)">{{ hoursToDays(detail.carryover ?? 0) }}</el-descriptions-item>
        <el-descriptions-item label="补班累计(天)">{{ hoursToDays(detail.time_off ?? 0) }}</el-descriptions-item>
        <el-descriptions-item label="已休(天)">{{ hoursToDays(detail.used ?? 0) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getBalanceList, getBalanceDetail } from '@/api/leaveAccount'
import { hoursToDays } from '@/utils/unit'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '', leave_type: '' })

const detailVisible = ref(false)
const detail = reactive({ grant: 0, adjust: 0, carryover: 0, time_off: 0, used: 0 })

async function fetchData() {
  loading.value = true
  try {
    const data = await getBalanceList({ pageNum: pageNum.value, pageSize: pageSize.value, person_id: search.person_id, leave_type: search.leave_type })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.person_id = ''; search.leave_type = ''; pageNum.value = 1; fetchData() }

async function showDetail(row: any) {
  try {
    const data = await getBalanceDetail(row.person_id, row.leave_type)
    Object.assign(detail, data)
    detailVisible.value = true
  } catch (e) {}
}

onMounted(fetchData)
</script>
