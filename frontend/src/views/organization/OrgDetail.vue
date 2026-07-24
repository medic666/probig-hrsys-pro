<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/organization')" content="组织详情" style="margin-bottom: 16px" />

    <el-tabs v-model="activeTab">
      <el-tab-pane label="当前信息" name="info">
        <el-descriptions v-if="snapshot" :column="2" border>
          <el-descriptions-item label="组织名称">{{ snapshot.company_name }}</el-descriptions-item>
          <el-descriptions-item label="统一社会信用代码">{{ snapshot.credit_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地址">{{ snapshot.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ snapshot.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开户行">{{ snapshot.bank_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="银行账号">{{ snapshot.bank_account || '-' }}</el-descriptions-item>
          <el-descriptions-item label="生效日期">{{ snapshot.effective_date?.slice(0, 10) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <el-tab-pane label="事件记录" name="events">
        <OrgHistory :entity-id="entityId" />
      </el-tab-pane>

      <el-tab-pane label="历史快照" name="history">
        <el-timeline v-if="history.length">
          <el-timeline-item v-for="snap in history" :key="snap.id" :timestamp="snap.effective_date?.slice(0, 10)" placement="top">
            <el-card>
              <p><strong>组织名称:</strong> {{ snap.company_name }}</p>
              <p><strong>地址:</strong> {{ snap.address || '-' }}</p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无历史快照" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getOrganization, getOrganizationHistory } from '../../api/organization'
import type { OrganizationSnapshot } from '../../types'
import OrgHistory from './OrgHistory.vue'

const route = useRoute()
const entityId = Number(route.params.id)
const loading = ref(false)
const snapshot = ref<OrganizationSnapshot | null>(null)
const history = ref<OrganizationSnapshot[]>([])
const activeTab = ref('info')

onMounted(async () => {
  loading.value = true
  try {
    const [r1, r2] = await Promise.all([getOrganization(entityId), getOrganizationHistory(entityId)])
    snapshot.value = r1.data
    history.value = r2.data
  } catch {} finally { loading.value = false }
})
</script>
