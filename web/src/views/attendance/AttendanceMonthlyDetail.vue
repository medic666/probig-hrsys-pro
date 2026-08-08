<template>
  <BusinessPage :title="`${personName} · ${month} 月度考勤核算`" back-to="/attendance-monthly">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="月度核算" name="calc">
        <div v-loading="loading" class="detail-wrap">
          <AttendanceCalcDescriptions :calc="row" :show-status="true" :status="row?.status" :show-calc-at="true" empty-text="当月无核算记录" />
        </div>
      </el-tab-pane>
      <el-tab-pane
        v-if="permissionStore.hasPermission(PERM.attendanceEventRead) && permissionStore.hasPermission(PERM.attendanceDailyRead)"
        label="全链路追溯"
        name="trace"
        lazy
      >
        <!-- 跨模块子视图：独立组件自加载（lazy 激活才渲染 → 无权限/未激活不发起请求） -->
        <AttendanceMonthlyTrace :person-id="personId" :month="month" />
      </el-tab-pane>
    </el-tabs>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceCalcDescriptions from '@/components/attendance/AttendanceCalcDescriptions.vue'
import AttendanceMonthlyTrace from '@/components/attendance/AttendanceMonthlyTrace.vue'
import { getMonthlyList } from '@/api/attendance'
import { PERM } from '@/constants/permission'
import { usePermissionStore } from '@/stores/permission'

const route = useRoute()
const permissionStore = usePermissionStore()
const personId = Number(route.params.personId)
const month = String(route.params.month)
const personName = String(route.query.name || '')
const activeTab = ref('calc')

const loading = ref(false)
const row = ref<any>(null)

onMounted(async () => {
  loading.value = true
  try {
    const d = (await getMonthlyList({ person_id: personId, month, pageNum: 1, pageSize: 1 })) as any
    row.value = d.list?.[0] || null
  } catch {
    row.value = null
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.detail-wrap {
  min-height: 120px;
}
</style>
