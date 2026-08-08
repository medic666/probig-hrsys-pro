<template>
  <BusinessPage>
    <template v-if="isCreate">
      <CompanyForm :company="null" @saved="goBack" @cancel="goBack" />
    </template>
    <template v-else>
      <div v-if="!editMode" class="toolbar">
        <el-button v-permission="PERM.companyWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <CompanyForm v-if="editMode" :company="{ id: companyId }" @saved="onEdited" @cancel="onCancel" />
      <template v-else>
        <AppDescriptions v-if="company" :column="2" border>
          <el-descriptions-item label="公司名称">{{ company.name }}</el-descriptions-item>
          <el-descriptions-item label="统一社会信用代码">{{ company.credit_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地址" :span="2">{{ company.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ company.contact_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开户行">{{ company.bank_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="银行账号" :span="2">{{ company.bank_account || '-' }}</el-descriptions-item>
        </AppDescriptions>
        <el-divider content-position="left">附件</el-divider>
        <FileAttachPanel target-type="company" :target-id="companyId" />
      </template>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import CompanyForm from '@/components/company/CompanyForm.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getCompany } from '@/api/company'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

const { id: companyId, isCreate, goBack } = useBusinessPage()

const company = ref<any>(null)
// ?edit=1 直达编辑态；默认查看态
const { editMode, enterEdit, exitEdit } = usePageEdit(PERM.companyWrite)

onMounted(async () => {
  if (companyId.value != null) {
    try {
      company.value = (await getCompany(companyId.value)) as any
    } catch { company.value = null }
  }
})

function onEdited() {
  exitEdit()
  getCompany(companyId.value).then((d: any) => { company.value = d }).catch(() => {})
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
