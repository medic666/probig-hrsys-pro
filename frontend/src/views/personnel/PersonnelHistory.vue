<template>
  <div>
    <div style="display: flex; justify-content: space-between; margin-bottom: 12px">
      <span style="font-weight: 500">事件记录 (共 {{ total }} 条)</span>
      <el-button v-if="auth.hasPermission('personnel', 'write')" size="small" type="primary" @click="openCreateEventForm()">新增事件</el-button>
    </div>

    <el-table :data="events" border size="small" v-loading="loading">
      <el-table-column prop="effective_date" label="生效日期" width="110">
        <template #default="{ row }">{{ row.effective_date?.slice(0, 10) }}</template>
      </el-table-column>
      <el-table-column prop="event_type" label="事件类型" width="80">
        <template #default="{ row }">
          <el-tag :type="row.event_type === 'create' ? 'success' : row.event_type === 'delete' ? 'danger' : 'primary'" size="small">
            {{ { create: '创建', update: '更新', delete: '删除' }[row.event_type] || row.event_type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="姓名" width="80" />
      <el-table-column label="基本工资" width="100">
        <template #default="{ row }">{{ row.base_salary.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="计薪天数" width="80">
        <template #default="{ row }">{{ row.pay_days }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="auth.hasPermission('personnel', 'write') && row.event_type !== 'delete'" text size="small" @click="openEditEventForm(row)">编辑</el-button>
          <el-button v-if="auth.hasPermission('personnel', 'delete')" text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      small
      style="margin-top: 12px; justify-content: flex-end"
      @current-change="fetchData"
    />

    <PersonnelEventForm
      v-model:visible="eventFormVisible"
      :entity-id="entityId"
      :edit-snapshot="editSnapshot"
      :event-id="editEventId"
      @success="fetchData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../../stores/auth'
import { listPersonnelEvents, deletePersonnelEvent } from '../../api/personnel'
import type { PersonnelEvent, PersonnelSnapshot } from '../../types'
import PersonnelEventForm from './PersonnelEventForm.vue'

const props = defineProps<{ entityId: number }>()
const auth = useAuthStore()

const loading = ref(false)
const events = ref<PersonnelEvent[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10

const eventFormVisible = ref(false)
const editSnapshot = ref<PersonnelSnapshot | null>(null)
const editEventId = ref<number | undefined>(undefined)

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await listPersonnelEvents({ entity_id: props.entityId, page: page.value, page_size: pageSize })
    events.value = res.data.list
    total.value = res.data.total
  } catch {} finally {
    loading.value = false
  }
}

function openCreateEventForm() {
  editSnapshot.value = null
  editEventId.value = undefined
  eventFormVisible.value = true
}

function openEditEventForm(event: PersonnelEvent) {
  editSnapshot.value = event as any
  editEventId.value = event.id
  eventFormVisible.value = true
}

async function handleDelete(row: PersonnelEvent) {
  try {
    await ElMessageBox.confirm('确定删除该事件吗？快照链将重新计算', '确认删除', { type: 'warning' })
    await deletePersonnelEvent(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch {}
}
</script>
