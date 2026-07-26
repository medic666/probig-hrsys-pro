<template>
  <div class="page-container">
    <div class="tool-bar">
      <el-button type="primary" v-permission="'leave:carryover'" @click="openCarryover">年假周年批量结转</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="batch_no" label="批次号" width="180" />
      <el-table-column prop="business_period" label="业务周期" width="120" />
      <el-table-column prop="operator_name" label="操作人" width="100" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">{{ statusMap[row.status] || row.status }}</template>
      </el-table-column>
      <el-table-column prop="total_count" label="处理人数" width="80" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="showEvents(row.id)">查看明细</el-button>
          <el-button v-if="row.status === 2" size="small" type="danger" v-permission="'leave:carryover'" @click="handleCancel(row.id)">反结账</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="carryoverVisible" title="周年结转" width="400px">
      <el-form label-width="100px">
        <el-form-item label="目标月份" required>
          <el-input v-model="carryoverMonth" placeholder="YYYY-MM" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="carryoverVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCarryover">执行结转</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="eventsVisible" title="批次事件明细" width="600px">
      <el-table :data="batchEvents" border stripe>
        <el-table-column prop="person_id" label="人员ID" width="80" />
        <el-table-column label="事件类型" width="80">
          <template #default="{ row }">{{ row.event_type }}</template>
        </el-table-column>
        <el-table-column label="变动(天)" width="100">
          <template #default="{ row }">{{ hoursToDays(row.hours) }}</template>
        </el-table-column>
        <el-table-column prop="effective_date" label="生效日期" width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBatches, getBatchEvents, executeCarryover, cancelCarryover } from '@/api/leaveAccount'
import { hoursToDays } from '@/utils/unit'

const statusMap: Record<number, string> = { 1: '待执行', 2: '已生效', 3: '已冲销', 4: '执行失败' }

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)

const carryoverVisible = ref(false)
const carryoverMonth = ref('')

const eventsVisible = ref(false)
const batchEvents = ref([])

async function fetchData() {
  loading.value = true
  try {
    const data = await getBatches({ pageNum: pageNum.value, pageSize: pageSize.value })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function openCarryover() { carryoverVisible.value = true }

async function handleCarryover() {
  if (!carryoverMonth.value) { ElMessage.warning('请输入目标月份'); return }
  try {
    const data = await executeCarryover({ target_month: carryoverMonth.value })
    ElMessage.success(`结转完成，处理${data.processed_count}人`)
    carryoverVisible.value = false; fetchData()
  } catch (e) {}
}

async function showEvents(batchId: number) {
  try {
    batchEvents.value = await getBatchEvents(batchId)
    eventsVisible.value = true
  } catch (e) {}
}

async function handleCancel(id: number) {
  await ElMessageBox.confirm('确定要反结账吗？', '确认', { type: 'warning' })
  try { await cancelCarryover(id); ElMessage.success('反结账成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
