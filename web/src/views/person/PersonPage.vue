<template>
  <BusinessPage>
    <template v-if="isCreate">
      <PersonProfileForm :person="null" @saved="onCreated" @cancel="goBack" />
    </template>
    <template v-else>
      <template v-if="editMode">
        <PersonProfileForm :person="{ id: personId }" @saved="onEdited" @cancel="editMode = false" />
      </template>
      <template v-else>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基础信息" name="info">
            <div class="toolbar">
              <el-button v-permission="PERM.personWrite" type="primary" size="small" @click="editMode = true">编辑档案</el-button>
            </div>
            <AppDescriptions v-if="person" :column="2" border size="small">
              <el-descriptions-item label="姓名">{{ person.name }}</el-descriptions-item>
              <el-descriptions-item label="身份证号">{{ person.id_card || '-' }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ genderMap[person.gender] || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="生日">{{ person.birthday || '-' }}</el-descriptions-item>
              <el-descriptions-item label="民族">{{ person.nation || '-' }}</el-descriptions-item>
              <el-descriptions-item label="籍贯">{{ person.native_place || '-' }}</el-descriptions-item>
              <el-descriptions-item label="住址" :span="2">{{ person.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="政治面貌">{{ person.political_status || '-' }}</el-descriptions-item>
              <el-descriptions-item label="婚姻状态">{{ maritalMap[person.marital_status] || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="别名" :span="2">{{ person.alias || '-' }}</el-descriptions-item>
            </AppDescriptions>

            <h4 class="sub-title">电话</h4>
            <el-table :data="person?.phones || []" border size="small" class="sub-table">
              <el-table-column prop="phone" label="号码" />
              <el-table-column prop="phone_type" label="类型" width="100" />
            </el-table>

            <h4 class="sub-title">邮箱</h4>
            <el-table :data="person?.emails || []" border size="small" class="sub-table">
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="email_type" label="类型" width="100" />
            </el-table>

            <h4 class="sub-title">银行卡</h4>
            <el-table :data="person?.bank_cards || []" border size="small" class="sub-table">
              <el-table-column prop="bank_name" label="开户行" />
              <el-table-column prop="account_number" label="账号" />
              <el-table-column prop="account_holder" label="持卡人" />
            </el-table>

            <h4 class="sub-title">紧急联系人</h4>
            <el-table :data="person?.emergency_contacts || []" border size="small" class="sub-table">
              <el-table-column prop="contact_name" label="联系人" />
              <el-table-column prop="contact_phone" label="联系电话" />
              <el-table-column prop="sort" label="序号" width="80" />
            </el-table>
          </el-tab-pane>

          <el-tab-pane v-if="permissionStore.hasPermission(PERM.positionEventRead)" label="职务信息" name="position" lazy>
            <!-- 跨模块子视图：独立组件自加载（lazy 激活才渲染 → 无权限/未激活不发起请求） -->
            <PersonPositionTab :person-id="personId" />
          </el-tab-pane>

          <el-tab-pane v-if="permissionStore.hasPermission(PERM.annualLeaveEventRead)" label="假期余额" name="leave-balance">
            <LeaveBalanceDetail :person-id="personId" />
          </el-tab-pane>

          <el-tab-pane v-if="permissionStore.hasPermission(PERM.fileRead)" label="附件" name="files">
            <FileAttachPanel target-type="person" :target-id="personId" />
          </el-tab-pane>
        </el-tabs>
      </template>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import PersonProfileForm from '@/components/person/PersonProfileForm.vue'
import PersonPositionTab from '@/components/person/PersonPositionTab.vue'
import LeaveBalanceDetail from '@/components/cards/LeaveBalanceDetail.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getPerson } from '@/api/person'
import { PERM } from '@/constants/permission'
import { usePermissionStore } from '@/stores/permission'
import { useBusinessPage } from '@/composables/useBusinessPage'

const router = useRouter()
const permissionStore = usePermissionStore()
const { id: personId, isCreate, goBack } = useBusinessPage()

const genderMap: Record<number, string> = { 1: '男', 2: '女' }
const maritalMap: Record<number, string> = { 1: '已婚', 2: '未婚' }

const activeTab = ref('info')
const editMode = ref(false)
const person = ref<any>(null)
async function loadPerson() {
  try {
    person.value = (await getPerson(personId.value)) as any
  } catch { person.value = null }
}

onMounted(() => {
  if (personId.value != null) {
    loadPerson()
  }
})

function onCreated(id: number) {
  router.replace(`/person/${id}`)
}

function onEdited() {
  editMode.value = false
  loadPerson()
}
</script>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
.sub-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 14px 0 8px;
}
.sub-table {
  margin-bottom: 4px;
}
</style>
