<template>
  <BusinessPage>
    <PositionEventForm :event="editEvent" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import PositionEventForm from '@/components/position/PositionEventForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

// /position-events/create → 新增；/position-events/:id → 编辑（meta 驱动标题与返回）
const { id, goBack } = useBusinessPage()
const editEvent = computed(() => (id.value != null ? { id: id.value } : null))
</script>
