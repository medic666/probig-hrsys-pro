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
              <el-button type="primary" size="small" @click="editMode = true">编辑档案</el-button>
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

          <el-tab-pane label="职务信息" name="position">
            <div v-if="positionLoading" v-loading="positionLoading" style="min-height:100px" />
            <template v-else>
              <h4>当前职务</h4>
              <AppDescriptions v-if="currentPosition" :column="3" border size="small">
                <el-descriptions-item label="在职状态">{{ currentPosition.is_active ? '在职' : '已离职' }}</el-descriptions-item>
                <el-descriptions-item label="入职日期">{{ currentPosition.entry_date || '-' }}</el-descriptions-item>
                <el-descriptions-item label="离职日期">{{ currentPosition.leave_date || '-' }}</el-descriptions-item>
                <el-descriptions-item label="公司组">{{ currentPosition.company_name || '-' }}</el-descriptions-item>
                <el-descriptions-item label="部门">{{ currentPosition.department || '-' }}</el-descriptions-item>
                <el-descriptions-item label="职位">{{ currentPosition.position || '-' }}</el-descriptions-item>
                <el-descriptions-item label="考勤组">{{ currentPosition.attendance_group || '-' }}</el-descriptions-item>
                <el-descriptions-item label="享有年假">{{ currentPosition.has_annual_leave ? '是' : '否' }}</el-descriptions-item>
                <el-descriptions-item label="享有全勤奖">{{ currentPosition.has_attendance_bonus ? '是' : '否' }}</el-descriptions-item>
                <el-descriptions-item label="基本工资">{{ currentPosition.base_salary || '-' }}</el-descriptions-item>
                <el-descriptions-item label="绩效工资基数">{{ currentPosition.performance_salary || '-' }}</el-descriptions-item>
                <el-descriptions-item label="计薪天数">{{ currentPosition.salary_days || '-' }}</el-descriptions-item>
                <el-descriptions-item label="职位津贴">{{ currentPosition.post_allowance || '-' }}</el-descriptions-item>
                <el-descriptions-item label="餐补">{{ currentPosition.meal_allowance || '-' }}</el-descriptions-item>
                <el-descriptions-item label="房补">{{ currentPosition.housing_allowance || '-' }}</el-descriptions-item>
                <el-descriptions-item label="交通补贴">{{ currentPosition.transport_allowance || '-' }}</el-descriptions-item>
                <el-descriptions-item label="高温补贴">{{ currentPosition.high_temp_allowance || '-' }}</el-descriptions-item>
                <el-descriptions-item label="保险补偿">{{ currentPosition.insurance_compensation || '-' }}</el-descriptions-item>
                <el-descriptions-item label="公积金补偿">{{ currentPosition.fund_compensation || '-' }}</el-descriptions-item>
                <el-descriptions-item label="社保代扣">{{ currentPosition.social_security_deduct || '-' }}</el-descriptions-item>
                <el-descriptions-item label="公积金代扣">{{ currentPosition.housing_fund_deduct || '-' }}</el-descriptions-item>
              </AppDescriptions>
              <el-empty v-else description="暂无职务信息" :image-size="40" />

              <h4 class="sub-title">变动历史（职务事件）</h4>
              <el-table :data="positionEvents" border size="small">
                <el-table-column prop="event_type" label="事件类型" width="110" />
                <el-table-column prop="effective_date" label="生效日期" width="120" />
                <el-table-column label="操作" width="160">
                  <template #default="{ row }">
                    <el-button type="primary" link size="small" @click="router.push(`/position-events/${row.id}`)">编辑</el-button>
                    <el-button type="danger" link size="small" @click="removePositionEvent(row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </template>
          </el-tab-pane>

          <el-tab-pane label="假期余额" name="leave-balance">
            <LeaveBalanceDetail :person-id="personId" />
          </el-tab-pane>

          <el-tab-pane label="附件" name="files">
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
import { ElMessage, ElMessageBox } from 'element-plus'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import PersonProfileForm from '@/components/person/PersonProfileForm.vue'
import LeaveBalanceDetail from '@/components/cards/LeaveBalanceDetail.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getPerson } from '@/api/person'
import { getCurrentPosition } from '@/api/position-snapshot'
import { getPositionEvents, deletePositionEvent } from '@/api/position-event'
import { useBusinessPage } from '@/composables/useBusinessPage'

const router = useRouter()
const { id: personId, isCreate, goBack } = useBusinessPage()

const genderMap: Record<number, string> = { 1: '男', 2: '女' }
const maritalMap: Record<number, string> = { 1: '已婚', 2: '未婚' }

const activeTab = ref('info')
const editMode = ref(false)
const person = ref<any>(null)
const positionLoading = ref(false)
const currentPosition = ref<any>(null)
const positionEvents = ref<any[]>([])

async function loadPerson() {
  try {
    person.value = (await getPerson(personId.value)) as any
  } catch { person.value = null }
}

async function loadPosition() {
  positionLoading.value = true
  try {
    currentPosition.value = (await getCurrentPosition(personId.value)) as any
  } catch {
    currentPosition.value = null
  }
  try {
    const d = (await getPositionEvents({ person_id: personId.value, pageNum: 1, pageSize: 100 })) as any
    positionEvents.value = d.list || []
  } catch {
    positionEvents.value = []
  }
  positionLoading.value = false
}

onMounted(() => {
  if (personId.value != null) {
    loadPerson()
    loadPosition()
  }
})

async function removePositionEvent(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除该职务事件（${row.event_type} ${row.effective_date}）？删除后将重建职务快照。`, '提示', { type: 'warning' })
  } catch { return }
  try {
    await deletePositionEvent(row.id)
    ElMessage.success('删除成功')
    loadPosition()
  } catch { /* handled */ }
}

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
