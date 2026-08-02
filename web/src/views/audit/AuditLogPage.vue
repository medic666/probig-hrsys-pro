<template>
  <BusinessPage title="审计日志详情" back-to="/audit-logs">
    <AuditLogDetailView :id="logId" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AuditLogDetailView from '@/components/audit/AuditLogDetailView.vue'

const route = useRoute()
const logId = computed(() => Number(route.params.id))
</script>
