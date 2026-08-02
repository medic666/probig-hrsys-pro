<template>
  <BusinessPage :title="title" back-to="/system/roles">
    <RoleForm :id="roleId" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import RoleForm from '@/components/system/RoleForm.vue'

const route = useRoute()
const router = useRouter()

const roleId = route.params.id ? Number(route.params.id) : null
const isCreate = computed(() => roleId == null)
const title = computed(() => (isCreate.value ? '新增角色' : '编辑角色'))

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.replace('/system/roles')
  }
}
</script>
