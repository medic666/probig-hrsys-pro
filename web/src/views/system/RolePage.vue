<template>
  <BusinessPage>
    <template v-if="isCreate || editMode">
      <RoleForm :id="id" @saved="onSaved" @cancel="onCancel" />
    </template>
    <template v-else>
      <div v-if="!viewOnly && !detail?.is_default" class="toolbar">
        <el-button v-permission="PERM.roleWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <div v-loading="loading">
        <AppDescriptions v-if="detail" :column="2" border size="small">
          <el-descriptions-item label="角色名称">{{ detail.name }}</el-descriptions-item>
          <el-descriptions-item label="数据范围">{{ detail.data_scope === 'own' ? '仅自己' : '全部' }}</el-descriptions-item>
          <el-descriptions-item label="默认角色">{{ detail.is_default ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
        </AppDescriptions>
        <el-empty v-else-if="!loading" description="角色不存在" :image-size="60" />
      </div>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import RoleForm from '@/components/system/RoleForm.vue'
import { getRole } from '@/api/role'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

// /system/roles/create → 新增；/system/roles/:id → 默认查看态（?edit=1 直达编辑态）。
const { id, isCreate, goBack } = useBusinessPage()

const { viewOnly, editMode, enterEdit, exitEdit } = usePageEdit(PERM.roleWrite)
const loading = ref(false)
const detail = ref<any>(null)

onMounted(async () => {
  if (id.value == null) return
  loading.value = true
  try {
    detail.value = (await getRole(id.value)) as any
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
    detail.value = (await getRole(id.value!)) as any
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
