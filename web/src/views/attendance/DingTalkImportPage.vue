<template>
  <BusinessPage>
    <DingTalkImportForm @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import DingTalkImportForm from '@/components/attendance/DingTalkImportForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

// 钉钉考勤导入三步向导（页面化，导入完成返回列表）
const { goBack } = useBusinessPage()
</script>
