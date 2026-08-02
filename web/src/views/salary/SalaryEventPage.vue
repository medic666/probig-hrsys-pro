<template>
  <BusinessPage :title="title" back-to="/salary-events">
    <SalaryEventForm :id="eventId" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import SalaryEventForm from '@/components/salary/SalaryEventForm.vue'

const route = useRoute()
const router = useRouter()

const eventId = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => eventId == null)
const title = computed(() => (isCreate.value ? '新增工资事件' : '工资事件详情'))

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/salary-events')
  }
}
</script>
