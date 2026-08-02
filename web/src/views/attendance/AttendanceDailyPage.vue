<template>
  <BusinessPage>
    <AttendanceDailyForm :id="dailyId" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceDailyForm from '@/components/attendance/AttendanceDailyForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

// 新增=编辑=查看统一：保存即确认（事务提交整日并置为已确认）
const { id: dailyId, goBack } = useBusinessPage()
</script>
