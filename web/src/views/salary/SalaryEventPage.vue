<template>
  <BusinessPage>
    <SalaryEventForm :id="id" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import SalaryEventForm from '@/components/salary/SalaryEventForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const { id, goBack } = useBusinessPage()
</script>
