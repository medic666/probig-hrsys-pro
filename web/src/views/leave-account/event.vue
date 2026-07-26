<template>
  <div class="page-container">
    <div class="search-bar">
      <el-input v-model="search.person_id" placeholder="人员ID" clearable style="width:120px;" />
      <el-select v-model="search.leave_type" placeholder="假期类型" clearable style="width:140px;">
        <el-option label="年假" value="annual_leave" />
        <el-option label="调休" value="time_off" />
      </el-select>
      <el-select v-model="search.source_type" placeholder="事件来源" clearable style="width:140px;">
        <el-option label="人工录入" value="manual" />
        <el-option label="系统结转" value="system_period" />
      </el-select>
      <el-button type="primary" @click="fetchData">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>
    <div class="tool-bar">
      <el-button type="primary" v-permission="'leave:write'" @click="openDialog">人工调整</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="person_id" label="人员ID" width="80" />
      <el-table-column label="假期类型" width="80">
        <template #default="{ row }">{{ row.leave_type === 'annual_leave' ? '年假' : '调休' }}</template>
      </el-table-column>
      <el-table-column label="事件类型" width="80">
        <template #default="{ row }">{{ eventTypeMap[row.event_type] || row.event_type }}</template>
      </el-table-column>
      <el-table-column label="变动(天)" width="100">
        <template #default="{ row }">{{ hoursToDays(row.hours ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="来源" width="80">
        <template #default="{ row }">{{ row.source_type === 'manual' ? '人工' : '系统' }}</template>
      </el-table-column>
      <el-table-column prop="effective_date" label="生效日期" width="120" />
      <el-table-column prop="remark" label="备注" min-width="150" />
      <el-table-column label="操作" width="80">
        <template #default="{ row }">
          <el-button v-if="row.source_type !== 'system_period'" size="small" type="danger" v-permission="'leave:delete'" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="pageNum" v-model:page-size="pageSize"
      :total="total" layout="total, sizes, prev, pager, next, jumper"
      style="margin-top:16px;justify-content:flex-end;"
      @current-change="fetchData" @size-change="fetchData"
    />

    <el-dialog v-model="dialogVisible" title="人工调整额度" width="400px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="人员ID" required>
          <el-input-number v-model="form.person_id" :min="1" />
        </el-form-item>
        <el-form-item label="假期类型" required>
          <el-select v-model="form.leave_type">
            <el-option label="年假" value="annual_leave" />
            <el-option label="调休" value="time_off" />
          </el-select>
        </el-form-item>
        <el-form-item label="变动(天)" required>
          <el-input-number v-model="form.hours_days" :precision="2" :step="0.5" />
        </el-form-item>
        <el-form-item label="生效日期">
          <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLeaveEvents, createLeaveEvent, deleteLeaveEvent } from '@/api/leaveAccount'
import { daysToHours, hoursToDays } from '@/utils/unit'

const eventTypeMap: Record<string, string> = {
  grant: '配发', adjust: '调整', carryover_deduct: '结转扣减', time_off_accrue: '补班累计',
}

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(20)
const search = reactive({ person_id: '', leave_type: '', source_type: '' })

const dialogVisible = ref(false)
const formRef = ref()
const form = reactive({ person_id: undefined as number | undefined, leave_type: 'annual_leave', hours_days: 0, effective_date: '', remark: '' })

async function fetchData() {
  loading.value = true
  try {
    const data = await getLeaveEvents({
      pageNum: pageNum.value, pageSize: pageSize.value,
      person_id: search.person_id, leave_type: search.leave_type, source_type: search.source_type,
    })
    list.value = data.list; total.value = data.total
  } catch (e) {} finally { loading.value = false }
}

function resetSearch() { search.person_id = ''; search.leave_type = ''; search.source_type = ''; pageNum.value = 1; fetchData() }

function openDialog() { dialogVisible.value = true }

async function handleSubmit() {
  try {
    await createLeaveEvent({
      person_id: form.person_id, leave_type: form.leave_type,
      hours: daysToHours(form.hours_days), effective_date: form.effective_date, remark: form.remark,
    })
    ElMessage.success('创建成功'); dialogVisible.value = false; fetchData()
  } catch (e) {}
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定要删除吗？', '确认', { type: 'warning' })
  try { await deleteLeaveEvent(id); ElMessage.success('删除成功'); fetchData() } catch (e) {}
}

onMounted(fetchData)
</script>
