<template>
  <div>
    <el-card>
      <template #header><span>操作审计</span></template>

      <div class="search-bar">
        <el-input v-model="filters.operator_id" placeholder="操作人ID" style="width:120px" />
        <el-select v-model="filters.target_type" placeholder="对象类型" clearable style="width:140px">
          <el-option label="人员" value="person" /><el-option label="公司" value="company" />
          <el-option label="假勤事件" value="attendance_event" /><el-option label="工资事件" value="salary_event" />
          <el-option label="职务事件" value="position_event" /><el-option label="文件" value="file" />
          <el-option label="用户" value="user" /><el-option label="角色" value="role" />
        </el-select>
        <el-select v-model="filters.action" placeholder="操作类型" clearable style="width:120px">
          <el-option label="新增" value="新增" /><el-option label="修改" value="修改" />
          <el-option label="删除" value="删除" /><el-option label="恢复" value="恢复" />
          <el-option label="核算" value="核算" /><el-option label="配置修改" value="配置修改" />
        </el-select>
        <el-input v-model="filters.start_date" type="date" placeholder="开始日期" style="width:150px" />
        <el-input v-model="filters.end_date" type="date" placeholder="结束日期" style="width:150px" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="operator_name" label="操作人" width="100" />
        <el-table-column prop="target_type" label="对象类型" width="100" />
        <el-table-column prop="target_id" label="对象ID" width="80" />
        <el-table-column prop="action" label="操作" width="80" />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column label="快照" width="80">
          <template #default="{ row }">
            <el-button size="small" @click="showSnapshot(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchList" style="margin-top:16px"
      />
    </el-card>

    <el-dialog v-model="snapshotDialog" title="操作快照" width="700px">
      <el-tabs>
        <el-tab-pane label="操作前">
          <pre style="max-height:400px;overflow:auto;background:#f5f5f5;padding:10px">{{ beforeSnapshot }}</pre>
        </el-tab-pane>
        <el-tab-pane label="操作后">
          <pre style="max-height:400px;overflow:auto;background:#f5f5f5;padding:10px">{{ afterSnapshot }}</pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import request from '@/utils/request'

const list = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20)
const loading = ref(false)
const filters = reactive({ operator_id: '', target_type: '', action: '', start_date: '', end_date: '' })

const snapshotDialog = ref(false); const beforeSnapshot = ref(''); const afterSnapshot = ref('')

async function fetchList() {
  loading.value = true
  const params: any = { page: page.value, page_size: pageSize.value }
  if (filters.operator_id) params.operator_id = filters.operator_id
  if (filters.target_type) params.target_type = filters.target_type
  if (filters.action) params.action = filters.action
  if (filters.start_date) params.start_date = filters.start_date
  if (filters.end_date) params.end_date = filters.end_date
  const res = await request.get('/audit-logs', { params })
  list.value = res.data.list; total.value = res.data.total; loading.value = false
}

function showSnapshot(row: any) {
  beforeSnapshot.value = JSON.stringify(JSON.parse(row.before_snapshot || '{}'), null, 2)
  afterSnapshot.value = JSON.stringify(JSON.parse(row.after_snapshot || '{}'), null, 2)
  snapshotDialog.value = true
}

onMounted(fetchList)
</script>

<style scoped>
.search-bar { margin-bottom: 16px; display: flex; gap: 10px; flex-wrap: wrap; }
</style>
