<template>
  <BusinessPage>
    <template v-if="isCreate || editMode">
      <SalaryEventForm :id="id" @saved="onSaved" @cancel="onCancel" />
    </template>
    <template v-else>
      <div class="toolbar">
        <el-button v-permission="PERM.salaryEventWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <div v-loading="loading">
        <AppDescriptions v-if="detail" :column="2" border size="small">
          <el-descriptions-item label="人员">{{ personName }}</el-descriptions-item>
          <el-descriptions-item label="归属月份">{{ detail.belong_month }}</el-descriptions-item>
          <el-descriptions-item label="事件类型">{{ detail.event_type }}</el-descriptions-item>
          <el-descriptions-item label="值">{{ detail.amount }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
        </AppDescriptions>
        <el-empty v-else-if="!loading" description="事件不存在" :image-size="60" />
      </div>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import SalaryEventForm from '@/components/salary/SalaryEventForm.vue'
import { getSalaryEvent } from '@/api/salary'
import { getPersonOptions } from '@/api/reference'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

// /salary-events/create → 新增；/salary-events/:id → 默认查看态（?edit=1 直达编辑态）。
const { id, isCreate, goBack } = useBusinessPage()

const { editMode, enterEdit, exitEdit } = usePageEdit(PERM.salaryEventWrite)
const loading = ref(false)
const detail = ref<any>(null)
const personName = ref('')


onMounted(async () => {
  if (id.value == null) return
  loading.value = true
  try {
    detail.value = (await getSalaryEvent(id.value)) as any
    if (detail.value?.person_id) {
      const opts = (await getPersonOptions()) as any[] || []
      personName.value = opts.find((p: any) => p.id === detail.value.person_id)?.name || `#${detail.value.person_id}`
    }
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
    detail.value = (await getSalaryEvent(id.value!)) as any
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
