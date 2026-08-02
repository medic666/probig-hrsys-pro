<template>
  <BusinessPage>
    <AuditLogDetailView :id="id" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import AuditLogDetailView from '@/components/audit/AuditLogDetailView.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const { id } = useBusinessPage()
</script>
