<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.operator_name" placeholder="操作人" clearable style="width:150px;" />
      <el-select v-model="search.action" placeholder="操作类型" clearable style="width:120px;">
        <el-option label="新增" value="新增" />
        <el-option label="修改" value="修改" />
        <el-option label="删除" value="删除" />
        <el-option label="核算" value="核算" />
        <el-option label="结转" value="结转" />
      </el-select>
      <el-date-picker v-model="searchDateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="operator_name" label="操作人" width="100" />
      <el-table-column prop="action" label="操作类型" width="80" />
      <el-table-column prop="target_type" label="对象类型" width="100" />
      <el-table-column prop="target_name" label="对象名称" width="120" />
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="created_at" label="操作时间" width="160" />
      <el-table-column label="操作" width="80">
        <template #default="{ row }">
          <el-button size="small" @click="showDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="detailVisible" title="审计详情" width="800px">
      <el-row :gutter="16">
        <el-col :span="12">
          <h4>操作前</h4>
          <pre style="max-height:400px;overflow:auto;background:#f5f5f5;padding:12px;border-radius:4px;">{{ fmtJSON(detailRow.before_snapshot) }}</pre>
        </el-col>
        <el-col :span="12">
          <h4>操作后</h4>
          <pre style="max-height:400px;overflow:auto;background:#f5f5f5;padding:12px;border-radius:4px;">{{ fmtJSON(detailRow.after_snapshot) }}</pre>
        </el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getAuditList } from '@/api/audit'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ operator_name: '', action: '' })
const searchDateRange = ref<string[]>([])

const detailVisible = ref(false)
const detailRow = reactive({ before_snapshot: '', after_snapshot: '' })

function fmtJSON(json: string) {
  if (!json) return '无数据'
  try { return JSON.stringify(JSON.parse(json), null, 2) } catch (e) { return json }
}

async function fetchData() {
  loading.value = true
  try {
    const data = await getAuditList({
      pageNum: pageNum.value, pageSize: pageSize.value,
      operator_name: search.operator_name, action: search.action,
      start_date: searchDateRange.value?.[0] || '',
      end_date: searchDateRange.value?.[1] || '',
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.operator_name = ''; search.action = ''; searchDateRange.value = []; pageNum.value = 1; fetchData() }

function showDetail(row: any) {
  Object.assign(detailRow, row)
  detailVisible.value = true
}

onMounted(fetchData)
</script>
