<template>
  <BusinessPage :title="title" back-to="/annual-leave-events">
    <AnnualLeaveEventForm :id="eventId" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AnnualLeaveEventForm from '@/components/annual-leave/AnnualLeaveEventForm.vue'

const route = useRoute()
const router = useRouter()

const eventId = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => eventId == null)
const title = computed(() => (isCreate.value ? '新增年假事件' : '年假事件详情'))

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/annual-leave-events')
  }
}
</script>
