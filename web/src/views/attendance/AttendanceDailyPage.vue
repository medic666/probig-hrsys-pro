<template>
  <BusinessPage :title="title" back-to="/attendance">
    <AttendanceDailyForm :id="dailyId" :confirm="isConfirm" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceDailyForm from '@/components/attendance/AttendanceDailyForm.vue'

const route = useRoute()
const router = useRouter()

const dailyId = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => dailyId == null)
// 待确认记录经此页保存 → 确认语义（置为已确认）
const isConfirm = route.query.confirm === '1'
const title = computed(() => (isCreate.value ? '新增考勤事件' : isConfirm.value ? '确认考勤事件' : '考勤事件详情'))

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/attendance')
  }
}
</script>
