<template>
  <BusinessPage>
    <BatchAttendanceForm @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import BatchAttendanceForm from '@/components/attendance/BatchAttendanceForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

// 批量录入考勤事件：人员多选 × 时间段 × 状态 × 明细组（页面化，提交后返回列表）
const { goBack } = useBusinessPage()
</script>
