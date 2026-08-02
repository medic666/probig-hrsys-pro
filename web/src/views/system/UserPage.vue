<template>
  <BusinessPage>
    <UserForm :id="id" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import UserForm from '@/components/system/UserForm.vue'
import { useBusinessPage } from '@/composables/useBusinessPage'

const { id, goBack } = useBusinessPage()
</script>
