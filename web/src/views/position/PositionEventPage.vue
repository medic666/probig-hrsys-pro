<template>
  <BusinessPage :title="isCreate ? '新增职务事件' : '编辑职务事件'" back-to="/position-event">
    <PositionEventForm :event="editEvent" @saved="onSaved" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import PositionEventForm from '@/components/position/PositionEventForm.vue'

const route = useRoute()
const router = useRouter()

// /position-events/create → 新增；/position-events/:id → 编辑
const id = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => id == null)
const editEvent = ref(id != null ? { id } : null)

function onSaved() {
  goBack()
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/position-event')
  }
}
</script>
