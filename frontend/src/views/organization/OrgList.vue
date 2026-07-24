<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <el-input v-model="search" placeholder="搜索组织名称" clearable style="width: 240px" @change="fetchData" />
      <el-button v-if="auth.hasPermission('organization', 'write')" type="primary" @click="openEventDialog()">新增组织</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading" @row-click="(r: OrgSnapshot) => goDetail(r)">
      <el-table-column prop="company_name" label="组织名称" min-width="160" />
      <el-table-column prop="credit_code" label="统一社会信用代码" width="180" />
      <el-table-column prop="address" label="地址" min-width="180" />
      <el-table-column prop="phone" label="联系电话" width="120" />
      <el-table-column prop="bank_name" label="开户行" width="120" />
      <el-table-column prop="bank_account" label="银行账号" width="160" />
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" @click.stop="goDetail(row)">详情</el-button>
          <el-button v-if="auth.hasPermission('organization', 'write')" text size="small" type="primary" @click.stop="openEventDialog(row)">编辑</el-button>
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

    <OrgEventForm
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
import { listOrganizations } from '../../api/organization'
import type { OrganizationSnapshot as OrgSnapshot } from '../../types'
import OrgEventForm from './OrgEventForm.vue'

const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const list = ref<OrgSnapshot[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')

const eventDialogVisible = ref(false)
const selectedEntityId = ref(0)
const editSnapshot = ref<OrgSnapshot | null>(null)

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await listOrganizations({ search: search.value, page: page.value, page_size: pageSize })
    list.value = res.data.list
    total.value = res.data.total
  } catch { ElMessage.error('加载失败') } finally { loading.value = false }
}

function goDetail(row: OrgSnapshot) { router.push(`/organization/${row.entity_id}`) }

function openEventDialog(snapshot?: OrgSnapshot) {
  editSnapshot.value = snapshot || null
  selectedEntityId.value = snapshot?.entity_id || 0
  eventDialogVisible.value = true
}
</script>
