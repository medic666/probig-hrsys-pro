<template>
  <div class="page-container">
    <div class="page-header"><h2>人员管理</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchPersons" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
        <el-button type="success" link size="small" @click="handleEdit(row)">编辑</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog v-model:visible="dialogVisible" :title="dialogMode === 'add' ? '新增人员' : '编辑人员'" :mode="dialogMode" :form-fields="personFormFields" :rules="formRules" :submit-api="submitForm" :edit-data="editRow" @success="onFormSuccess" />

    <el-dialog v-model="detailVisible" title="人员详情" width="700px" @close="detailRow = null">
      <template v-if="detailRow">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基础信息" name="info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="姓名">{{ detailRow.name }}</el-descriptions-item>
              <el-descriptions-item label="身份证号">{{ detailRow.id_card }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ genderMap[detailRow.gender] || '-' }}</el-descriptions-item>
              <el-descriptions-item label="生日">{{ detailRow.birthday || '-' }}</el-descriptions-item>
              <el-descriptions-item label="民族">{{ detailRow.nation || '-' }}</el-descriptions-item>
              <el-descriptions-item label="籍贯">{{ detailRow.native_place || '-' }}</el-descriptions-item>
              <el-descriptions-item label="住址" :span="2">{{ detailRow.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="政治面貌">{{ detailRow.political_status || '-' }}</el-descriptions-item>
              <el-descriptions-item label="婚姻状态">{{ maritalMap[detailRow.marital_status] || '-' }}</el-descriptions-item>
              <el-descriptions-item label="别名" :span="2">{{ detailRow.alias || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane label="职务信息" name="position">
            <div v-if="positionLoading" v-loading="positionLoading" style="min-height:100px" />
            <template v-else>
              <h4>当前职务</h4>
              <el-descriptions v-if="currentPosition" :column="2" border size="small">
                <el-descriptions-item label="在职状态"><StatusTag :status="currentPosition.is_active ? 'calculated' : 'data_changed'" :text="currentPosition.is_active ? '在职' : '已离职'" /></el-descriptions-item>
                <el-descriptions-item label="入职日期">{{ currentPosition.entry_date || '-' }}</el-descriptions-item>
                <el-descriptions-item label="公司组">{{ currentPosition.company_name || currentPosition.company_id || '-' }}</el-descriptions-item>
                <el-descriptions-item label="部门">{{ currentPosition.department || '-' }}</el-descriptions-item>
                <el-descriptions-item label="职位">{{ currentPosition.position || '-' }}</el-descriptions-item>
                <el-descriptions-item label="考勤组">{{ currentPosition.attendance_group || '-' }}</el-descriptions-item>
                <el-descriptions-item label="基本工资">{{ currentPosition.base_salary || '-' }}</el-descriptions-item>
                <el-descriptions-item label="绩效工资基数">{{ currentPosition.performance_salary || '-' }}</el-descriptions-item>
                <el-descriptions-item label="计薪天数">{{ currentPosition.salary_days || '-' }}</el-descriptions-item>
              </el-descriptions>
              <el-button v-if="currentPosition" size="small" style="margin-top:4px" @click="posDetailVisible=true">查看完整职务信息</el-button>
              <el-empty v-else description="暂无职务信息" :image-size="40" />

              <h4 style="margin-top:16px">变动历史</h4>
              <el-table :data="positionHistory" border size="small">
                <el-table-column prop="effective_start_date" label="起始日期" width="110" />
                <el-table-column prop="effective_end_date" label="结束日期" width="110">
                  <template #default="{ row: r }">{{ r.effective_end_date === '9999-12-31T00:00:00Z' ? '至今' : r.effective_end_date }}</template>
                </el-table-column>
                <el-table-column prop="base_salary" label="基本工资" width="100" />
                <el-table-column prop="is_active" label="在职" width="70">
                  <template #default="{ row: r }">{{ r.is_active ? '是' : '否' }}</template>
                </el-table-column>
                <el-table-column prop="attendance_group" label="考勤组" />
              </el-table>
            </template>
          </el-tab-pane>

          <el-tab-pane label="联系方式" name="contacts">
            <h4>电话</h4>
            <el-table :data="detailRow.phones" border size="small" class="sub-table">
              <el-table-column prop="phone" label="号码" />
              <el-table-column prop="phone_type" label="类型" width="100" />
              <el-table-column label="操作" width="100">
                <template #default="{ row: r }">
                  <el-button type="danger" link size="small" @click="delPhone(r.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="sub-add">
              <el-input v-model="newPhone" placeholder="电话号码" size="small" style="width:180px" /> <el-button size="small" @click="addPhone">添加电话</el-button>
            </div>

            <h4 style="margin-top:16px">邮箱</h4>
            <el-table :data="detailRow.emails" border size="small" class="sub-table">
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="email_type" label="类型" width="100" />
              <el-table-column label="操作" width="100">
                <template #default="{ row: r }">
                  <el-button type="danger" link size="small" @click="delEmail(r.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="sub-add"><el-input v-model="newEmail" placeholder="邮箱地址" size="small" style="width:240px" /> <el-button size="small" @click="addEmail">添加邮箱</el-button></div>
          </el-tab-pane>

          <el-tab-pane label="银行卡" name="bankcards">
            <el-table :data="detailRow.bank_cards" border size="small" class="sub-table">
              <el-table-column prop="bank_name" label="开户行" />
              <el-table-column prop="account_number" label="账号" />
              <el-table-column prop="account_holder" label="持卡人" />
              <el-table-column label="操作" width="100">
                <template #default="{ row: r }">
                  <el-button type="danger" link size="small" @click="delBankCard(r.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="sub-add">
              <el-input v-model="newBankName" placeholder="开户行" size="small" style="width:140px" />
              <el-input v-model="newBankAccount" placeholder="银行卡号" size="small" style="width:180px" />
              <el-input v-model="newBankHolder" placeholder="持卡人" size="small" style="width:120px" />
              <el-button size="small" @click="addBankCard">添加</el-button>
            </div>
          </el-tab-pane>

          <el-tab-pane label="紧急联系人" name="emergency">
            <el-table :data="detailRow.emergency_contacts" border size="small" class="sub-table">
              <el-table-column prop="contact_name" label="联系人" />
              <el-table-column prop="contact_phone" label="联系电话" />
              <el-table-column prop="sort" label="序号" width="80" />
              <el-table-column label="操作" width="100">
                <template #default="{ row: r }">
                  <el-button type="danger" link size="small" @click="delEmergencyContact(r.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="sub-add">
              <el-input v-model="newContactName" placeholder="联系人" size="small" style="width:140px" />
              <el-input v-model="newContactPhone" placeholder="联系电话" size="small" style="width:160px" />
              <el-input-number v-model="newContactSort" :min="1" size="small" style="width:90px" />
              <el-button size="small" @click="addEmergencyContact">添加</el-button>
            </div>
          </el-tab-pane>

          <el-tab-pane label="附件" name="files">
            <FileAttachPanel v-if="detailRow" :target-type="'person'" :target-id="detailRow.id" />
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restorePerson" :columns="trashColumns" @restored="onTrashRestored" />

    <el-dialog v-model="posDetailVisible" title="完整职务信息" width="600px">
      <el-descriptions v-if="currentPosition" :column="2" border size="small">
        <template #title>在职信息</template>
        <el-descriptions-item label="在职状态">{{ currentPosition.is_active ? '在职' : '已离职' }}</el-descriptions-item>
        <el-descriptions-item label="入职日期">{{ currentPosition.entry_date || '-' }}</el-descriptions-item>
        <el-descriptions-item label="离职日期">{{ currentPosition.leave_date || '-' }}</el-descriptions-item>
        <el-descriptions-item label="考勤组">{{ currentPosition.attendance_group || '-' }}</el-descriptions-item>
        <el-descriptions-item label="享有年假">{{ currentPosition.has_annual_leave ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="享有全勤奖">{{ currentPosition.has_attendance_bonus ? '是' : '否' }}</el-descriptions-item>
      </el-descriptions>
      <el-descriptions v-if="currentPosition" :column="2" border size="small" style="margin-top:12px" title="薪资基数">
        <el-descriptions-item label="基本工资">{{ currentPosition.base_salary || '-' }}</el-descriptions-item>
        <el-descriptions-item label="绩效工资基数">{{ currentPosition.performance_salary || '-' }}</el-descriptions-item>
        <el-descriptions-item label="计薪天数">{{ currentPosition.salary_days || '-' }}</el-descriptions-item>
      </el-descriptions>
      <el-descriptions v-if="currentPosition" :column="2" border size="small" style="margin-top:12px" title="补贴与代扣">
        <el-descriptions-item label="职位津贴">{{ currentPosition.post_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="餐补">{{ currentPosition.meal_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="房补">{{ currentPosition.housing_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="交通补贴">{{ currentPosition.transport_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="高温补贴">{{ currentPosition.high_temp_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="保险补偿">{{ currentPosition.insurance_compensation || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公积金补偿">{{ currentPosition.fund_compensation || '-' }}</el-descriptions-item>
        <el-descriptions-item label="社保代扣">{{ currentPosition.social_security_deduct || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公积金代扣">{{ currentPosition.housing_fund_deduct || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import StatusTag from '@/components/StatusTag.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import { getPersons, getPerson, createPerson, updatePerson, deletePerson, restorePerson, getDeletedPersons, addPersonPhone, deletePersonPhone, addPersonEmail, deletePersonEmail, addPersonBankCard, deletePersonBankCard, addPersonEmergencyContact, deletePersonEmergencyContact, getAllPersons, exportPersons } from '@/api/person'
import { getCurrentPosition, getPositionHistory } from '@/api/position-snapshot'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const editRow = ref<any>(null)
const detailVisible = ref(false)
const detailRow = ref<any>(null)
const activeTab = ref('info')
const trashVisible = ref(false)
const positionLoading = ref(false)
const currentPosition = ref<any>(null)
const positionHistory = ref<any[]>([])
const posDetailVisible = ref(false)

const newPhone = ref('')
const newEmail = ref('')
const newBankName = ref('')
const newBankAccount = ref('')
const newBankHolder = ref('')
const newContactName = ref('')
const newContactPhone = ref('')
const newContactSort = ref(1)

const genderMap: Record<number, string> = { 1: '男', 2: '女' }
const maritalMap: Record<number, string> = { 1: '已婚', 2: '未婚' }

const columns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '姓名', width: '100' },
  { prop: 'id_card', label: '身份证号', width: '180' },
  { prop: 'gender', label: '性别', width: '60', formatter: (r: any) => genderMap[r.gender] || '-' },
  { prop: 'phones', label: '联系电话', formatter: (r: any) => r.phones?.[0]?.phone || '-' },
  { prop: 'created_at', label: '创建时间', width: '160', formatter: (r: any) => new Date(r.created_at).toLocaleString('zh-CN') },
]

const searchFields = [
  { prop:'person_id', label:'姓名', type:'person-select' as const, fetchApi: fetchPersonOptions },
  { prop:'id_card', label:'身份证号', type:'input' as const, placeholder:'模糊搜索' },
]

async function fetchPersonOptions(k?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[] || []
  return k ? list.filter(p => p.name.includes(k)) : list
}

const actions = [
  { key: 'add', label: '新增人员', type: 'primary' as const },
  { key: 'export', label: '导出', type: 'default' as const },
  { key: 'trash', label: '回收站', type: 'default' as const },
]

const personFormFields = [
  { prop: 'name', label: '姓名', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'id_card', label: '身份证号', type: 'input' as const, placeholder: '请输入', span: 12 },
  { prop: 'gender', label: '性别', type: 'select' as const, options: [{ label: '男', value: 1 }, { label: '女', value: 2 }], span: 12 },
  { prop: 'birthday', label: '生日', type: 'date' as const, span: 12 },
  { prop: 'nation', label: '民族', type: 'input' as const, span: 12 },
  { prop: 'native_place', label: '籍贯', type: 'input' as const, span: 12 },
  { prop: 'address', label: '住址', type: 'input' as const, span: 24 },
  { prop: 'political_status', label: '政治面貌', type: 'input' as const, span: 12 },
  { prop: 'marital_status', label: '婚姻状态', type: 'select' as const, options: [{ label: '已婚', value: 1 }, { label: '未婚', value: 2 }], span: 12 },
  { prop: 'alias', label: '别名', type: 'input' as const, span: 12 },
]

const formRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
}

const trashColumns = [
  { prop: 'id', label: 'ID', width: '60' },
  { prop: 'name', label: '姓名' },
  { prop: 'id_card', label: '身份证号' },
]

async function fetchPersons(params: any) { return (await getPersons(params)) as any }
async function fetchDeleted(params: any) { return (await getDeletedPersons(params)) as any }

function handleAction(key: string) {
  if (key === 'add') { dialogMode.value = 'add'; editRow.value = null; dialogVisible.value = true }
  else if (key === 'export') { handleExport() }
  else if (key === 'trash') { trashVisible.value = true }
}

async function handleExport() {
  const data = await exportPersons({})
  downloadBlob(data)
}

function handleEdit(row: any) { dialogMode.value = 'edit'; editRow.value = row; dialogVisible.value = true }

async function handleDetail(row: any) {
  try {
    detailRow.value = (await getPerson(row.id)) as any
    positionLoading.value = true
    try {
      currentPosition.value = (await getCurrentPosition(row.id)) as any
      positionHistory.value = (await getPositionHistory(row.id)) as any[] || []
    } catch { currentPosition.value = null; positionHistory.value = [] }
    positionLoading.value = false
    activeTab.value = 'info'
    detailVisible.value = true
  } catch { /* handled */ }
}

async function submitForm(data: any) {
  if (dialogMode.value === 'add') return createPerson(data)
  return updatePerson(editRow.value.id, data)
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' }) } catch { return }
  try { await deletePerson(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh() } catch { /* handled */ }
}

async function addPhone() {
  if (!newPhone.value) return
  await addPersonPhone(detailRow.value.id, { phone: newPhone.value, phone_type: 'mobile' })
  ElMessage.success('添加成功')
  newPhone.value = ''
  handleDetail({ id: detailRow.value.id })
}
async function delPhone(id: number) {
  await deletePersonPhone(id)
  ElMessage.success('删除成功')
  handleDetail({ id: detailRow.value.id })
}
async function addEmail() {
  if (!newEmail.value) return
  await addPersonEmail(detailRow.value.id, { email: newEmail.value, email_type: 'personal' })
  ElMessage.success('添加成功')
  newEmail.value = ''
  handleDetail({ id: detailRow.value.id })
}
async function delEmail(id: number) {
  await deletePersonEmail(id)
  ElMessage.success('删除成功')
  handleDetail({ id: detailRow.value.id })
}
async function addBankCard() {
  if (!newBankAccount.value) return
  await addPersonBankCard(detailRow.value.id, { bank_name: newBankName.value, account_number: newBankAccount.value, account_holder: newBankHolder.value })
  ElMessage.success('添加成功')
  newBankName.value = ''; newBankAccount.value = ''; newBankHolder.value = ''
  handleDetail({ id: detailRow.value.id })
}
async function delBankCard(id: number) {
  await deletePersonBankCard(id)
  ElMessage.success('删除成功')
  handleDetail({ id: detailRow.value.id })
}
async function addEmergencyContact() {
  if (!newContactName.value) return
  await addPersonEmergencyContact(detailRow.value.id, { contact_name: newContactName.value, contact_phone: newContactPhone.value, sort: newContactSort.value || 1 })
  ElMessage.success('添加成功')
  newContactName.value = ''; newContactPhone.value = ''; newContactSort.value = 1
  handleDetail({ id: detailRow.value.id })
}
async function delEmergencyContact(id: number) {
  await deletePersonEmergencyContact(id)
  ElMessage.success('删除成功')
  handleDetail({ id: detailRow.value.id })
}
function onFormSuccess() { tableRef.value?.refresh() }
function onTrashRestored() { tableRef.value?.refresh() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
.sub-table { margin-bottom: 4px; }
.sub-add { display: flex; gap: 8px; margin-bottom: 8px; }
</style>
