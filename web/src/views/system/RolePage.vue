<template>
  <BusinessPage>
    <RoleForm :id="id" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import RoleForm from '@/components/system/RoleForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const { id, goBack } = useBusinessPage()
</script>
