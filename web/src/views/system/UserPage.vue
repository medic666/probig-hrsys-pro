<template>
  <BusinessPage :title="title" back-to="/system/users">
    <UserForm :id="userId" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import UserForm from '@/components/system/UserForm.vue'

const route = useRoute()
const router = useRouter()

const userId = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => userId == null)
const title = computed(() => (isCreate.value ? '新增用户' : '编辑用户'))

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/system/users')
  }
}
</script>
