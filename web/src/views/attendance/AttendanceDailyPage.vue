<template>
  <BusinessPage :title="isConfirm ? '确认考勤事件' : undefined">
    <AttendanceDailyForm :id="dailyId" :confirm="isConfirm" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceDailyForm from '@/components/attendance/AttendanceDailyForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const route = useRoute()
const { id: dailyId, goBack } = useBusinessPage()
// 待确认记录经此页保存 → 确认语义（置为已确认）
const isConfirm = route.query.confirm === '1'
</script>
