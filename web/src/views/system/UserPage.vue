<template>
  <BusinessPage>
    <template v-if="isCreate || editMode">
      <UserForm :id="id" @saved="onSaved" @cancel="onCancel" />
    </template>
    <template v-else>
      <div v-if="!viewOnly" class="toolbar">
        <el-button v-permission="PERM.userWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <div v-loading="loading">
        <AppDescriptions v-if="detail" :column="2" border size="small">
          <el-descriptions-item label="用户名">{{ detail.username }}</el-descriptions-item>
          <el-descriptions-item label="关联人员">{{ detail.person_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="启用">{{ detail.is_active ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="角色">{{ roleNames.join('、') || '-' }}</el-descriptions-item>
        </AppDescriptions>
        <el-empty v-else-if="!loading" description="用户不存在" :image-size="60" />
      </div>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import UserForm from '@/components/system/UserForm.vue'
import { getUser } from '@/api/user'
import { getAllRoles } from '@/api/role'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

// /system/users/create → 新增；/system/users/:id → 默认查看态（?edit=1 直达编辑态）。
const { id, isCreate, goBack } = useBusinessPage()

const { viewOnly, editMode, enterEdit, exitEdit } = usePageEdit(PERM.userWrite)
const loading = ref(false)
const detail = ref<any>(null)
const roleNames = ref<string[]>([])

onMounted(async () => {
  if (id.value == null) return
  loading.value = true
  try {
    detail.value = (await getUser(id.value)) as any
    const roles = (await getAllRoles()) as any[] || []
    const idSet = new Set((detail.value?.role_ids || []) as number[])
    roleNames.value = roles.filter((r: any) => idSet.has(r.id)).map((r: any) => r.name)
  } catch {
    detail.value = null
  } finally {
    loading.value = false
  }
})

async function onSaved() {
  if (isCreate.value) {
    goBack()
    return
  }
  exitEdit()
  try {
    detail.value = (await getUser(id.value!)) as any
  } catch {
    detail.value = null
  }
}

function onCancel() {
  if (isCreate.value) goBack()
  else exitEdit()
}
</script>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
</style>
