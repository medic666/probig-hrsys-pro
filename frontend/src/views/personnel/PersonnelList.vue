<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div style="display: flex; gap: 12px">
        <el-input v-model="search" placeholder="搜索姓名" clearable style="width: 240px" @change="fetchData" />
      </div>
      <el-button v-if="auth.hasPermission('personnel', 'write')" type="primary" @click="openEventDialog()">新增人员</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" @row-click="goDetail">
      <el-table-column prop="name" label="姓名" min-width="100" />
      <el-table-column prop="attendance_group" label="考勤组" width="100" />
      <el-table-column label="基本工资" width="110">
        <template #default="{ row }">{{ row.base_salary.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="计薪天数" width="90">
        <template #default="{ row }">{{ row.pay_days }}</template>
      </el-table-column>
      <el-table-column label="入职日期" width="110">
        <template #default="{ row }">{{ row.hire_date || '-' }}</template>
      </el-table-column>
      <el-table-column label="生效日期" width="110">
        <template #default="{ row }">{{ row.effective_date?.slice(0, 10) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" @click.stop="goDetail(row)">详情</el-button>
          <el-button v-if="auth.hasPermission('personnel', 'write')" text size="small" type="primary" @click.stop="openEventDialog(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      style="margin-top: 16px; justify-content: flex-end"
      @current-change="fetchData"
    />

    <PersonnelEventForm
      v-model:visible="eventDialogVisible"
      :entity-id="selectedEntityId"
      :edit-snapshot="editSnapshot"
      @success="fetchData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../../stores/auth'
import { listPersonnel } from '../../api/personnel'
import type { PersonnelSnapshot } from '../../types'
import PersonnelEventForm from './PersonnelEventForm.vue'

const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const list = ref<PersonnelSnapshot[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')

const eventDialogVisible = ref(false)
const selectedEntityId = ref(0)
const editSnapshot = ref<PersonnelSnapshot | null>(null)

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await listPersonnel({ search: search.value, page: page.value, page_size: pageSize })
    list.value = res.data.list
    total.value = res.data.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function goDetail(row: PersonnelSnapshot) {
  router.push(`/personnel/${row.entity_id}`)
}

function openEventDialog(snapshot?: PersonnelSnapshot) {
  if (snapshot) {
    editSnapshot.value = snapshot
    selectedEntityId.value = snapshot.entity_id
  } else {
    editSnapshot.value = null
    selectedEntityId.value = 0
  }
  eventDialogVisible.value = true
}
</script>
