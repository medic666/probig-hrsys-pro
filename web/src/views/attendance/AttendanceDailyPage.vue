<template>
  <BusinessPage>
    <AttendanceDailyForm :id="dailyId" :readonly="readonly" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AttendanceDailyForm from '@/components/attendance/AttendanceDailyForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

// 新增=编辑=查看统一：保存即确认（事务提交整日并置为已确认）。
// ?readonly=1 只读模式：日记工时模块进入，仅查看当日考勤明细。
const route = useRoute()
const { id: dailyId, goBack } = useBusinessPage()
const readonly = computed(() => route.query.readonly === '1')
</script>
