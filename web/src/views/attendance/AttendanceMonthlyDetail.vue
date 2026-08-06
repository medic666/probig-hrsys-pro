<template>
  <BusinessPage :title="`${personName} · ${month} 月度考勤核算`" back-to="/attendance-monthly">
    <div v-loading="loading" class="detail-wrap">
      <AttendanceCalcDescriptions :calc="row" :show-status="true" :status="row?.status" empty-text="当月无核算记录" />
    </div>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceCalcDescriptions from '@/components/attendance/AttendanceCalcDescriptions.vue'
import { getMonthlyList } from '@/api/attendance'

const route = useRoute()
const personId = Number(route.params.personId)
const month = String(route.params.month)
const personName = String(route.query.name || '')

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
