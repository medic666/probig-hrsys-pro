<template>
  <BusinessPage>
    <AnnualLeaveEventForm :id="id" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import AnnualLeaveEventForm from '@/components/annual-leave/AnnualLeaveEventForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const { id, goBack } = useBusinessPage()
</script>
