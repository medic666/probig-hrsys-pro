<template>
  <div>
    <el-card>
      <div style="display: flex; gap: 8px; margin-bottom: 16px;">
        <el-select v-model="searchForm.target_type" placeholder="对象类型" clearable style="width: 140px;">
          <el-option label="人员" value="person" />
          <el-option label="公司" value="company" />
          <el-option label="职务事件" value="position_event" />
          <el-option label="假勤事件" value="attendance_event" />
          <el-option label="工资事件" value="salary_event" />
        </el-select>
        <el-select v-model="searchForm.action" placeholder="操作类型" clearable style="width: 120px;">
          <el-option label="新增" value="新增" />
          <el-option label="修改" value="修改" />
          <el-option label="删除" value="删除" />
          <el-option label="恢复" value="恢复" />
          <el-option label="核算" value="核算" />
        </el-select>
        <el-date-picker v-model="searchForm.date_range" type="daterange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" value-format="YYYY-MM-DD" />
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="operator_name" label="操作人" min-width="80" />
        <el-table-column prop="action" label="操作类型" min-width="80" />
        <el-table-column prop="target_type" label="对象类型" min-width="100" />
        <el-table-column prop="target_id" label="对象ID" min-width="80" />
        <el-table-column prop="ip" label="IP" min-width="120" />
        <el-table-column prop="created_at" label="操作时间" min-width="160" />
        <el-table-column label="详情" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 16px; justify-content: flex-end;" v-model:current-page="page" :total="total" :page-size="pageSize" layout="total, prev, pager, next" @current-change="fetchData" />
    </el-card>

    <el-dialog v-model="detailVisible" title="操作详情" width="700px">
      <div v-if="detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="操作人">{{ detail.operator_name }}</el-descriptions-item>
          <el-descriptions-item label="操作类型">{{ detail.action }}</el-descriptions-item>
          <el-descriptions-item label="对象类型">{{ detail.target_type }}</el-descriptions-item>
          <el-descriptions-item label="对象ID">{{ detail.target_id }}</el-descriptions-item>
          <el-descriptions-item label="IP">{{ detail.ip }}</el-descriptions-item>
          <el-descriptions-item label="操作时间">{{ detail.created_at }}</el-descriptions-item>
          <el-descriptions-item label="操作前快照">
            <pre style="max-height: 200px; overflow: auto; font-size: 12px;">{{ JSON.stringify(JSON.parse(detail.before_snapshot || '{}'), null, 2) }}</pre>
          </el-descriptions-item>
          <el-descriptions-item label="操作后快照">
            <pre style="max-height: 200px; overflow: auto; font-size: 12px;">{{ JSON.stringify(JSON.parse(detail.after_snapshot || '{}'), null, 2) }}</pre>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import * as api from '../../api/audit'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detail = ref<any>(null)
const searchForm = reactive({ target_type: '', action: '', date_range: [] as string[] })

const queryParams = computed(() => {
  const p: any = { page: page.value, page_size: pageSize.value }
  if (searchForm.target_type) p.target_type = searchForm.target_type
  if (searchForm.action) p.action = searchForm.action
  if (searchForm.date_range?.[0]) p.start_date = searchForm.date_range[0]
  if (searchForm.date_range?.[1]) p.end_date = searchForm.date_range[1]
  return p
})

async function fetchData() {
  loading.value = true
  try {
    const res = await api.getAuditLogList(queryParams.value)
    list.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

function showDetail(row: any) {
  detail.value = row
  detailVisible.value = true
}

fetchData()
</script>
