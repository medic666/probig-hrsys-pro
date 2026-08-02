<template>
  <div class="page-container">
    <PageHeader title="人员管理">
      <template #actions>
        <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      </template>
    </PageHeader>

    <PageToolbar>
      <el-button type="primary" size="small" @click="handleAdd">新增人员</el-button>
      <el-button size="small" @click="handleExport">导出</el-button>
      <el-button size="small" @click="trashVisible = true">回收站</el-button>
    </PageToolbar>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchPersons" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="handleDetail(row)">查看详情</el-button>
          <el-button type="success" link size="small" @click="openProfileEdit(row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <CardGrid ref="cardGridRef" :fetch-fn="fetchCards">
        <template #default="{ item }">
          <PersonCard :person="item" @click="handleDetail" />
        </template>
      </CardGrid>
    </template>

    <ProFormDialog v-model:visible="addVisible" title="新增人员" mode="add" :form-fields="personFormFields" :rules="formRules" :submit-api="submitAdd" @success="onFormSuccess" />

    <el-dialog v-model="detailVisible" title="人员详情" width="760px" @close="detailRow = null">
      <template v-if="detailRow">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基础信息" name="info">
            <el-button size="small" type="primary" style="margin-bottom:10px" @click="openProfileEdit(detailRow)">编辑档案</el-button>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="姓名">{{ detailRow.name }}</el-descriptions-item>
              <el-descriptions-item label="身份证号">{{ detailRow.id_card || '-' }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ genderMap[detailRow.gender] || '-' }}</el-descriptions-item>
              <el-descriptions-item label="生日">{{ detailRow.birthday || '-' }}</el-descriptions-item>
              <el-descriptions-item label="民族">{{ detailRow.nation || '-' }}</el-descriptions-item>
              <el-descriptions-item label="籍贯">{{ detailRow.native_place || '-' }}</el-descriptions-item>
              <el-descriptions-item label="住址" :span="2">{{ detailRow.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="政治面貌">{{ detailRow.political_status || '-' }}</el-descriptions-item>
              <el-descriptions-item label="婚姻状态">{{ maritalMap[detailRow.marital_status] || '-' }}</el-descriptions-item>
              <el-descriptions-item label="别名" :span="2">{{ detailRow.alias || '-' }}</el-descriptions-item>
            </el-descriptions>

            <h4 class="sub-title">电话</h4>
            <el-table :data="detailRow.phones" border size="small" class="sub-table">
              <el-table-column prop="phone" label="号码" />
              <el-table-column prop="phone_type" label="类型" width="100" />
            </el-table>

            <h4 class="sub-title">邮箱</h4>
            <el-table :data="detailRow.emails" border size="small" class="sub-table">
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="email_type" label="类型" width="100" />
            </el-table>

            <h4 class="sub-title">银行卡</h4>
            <el-table :data="detailRow.bank_cards" border size="small" class="sub-table">
              <el-table-column prop="bank_name" label="开户行" />
              <el-table-column prop="account_number" label="账号" />
              <el-table-column prop="account_holder" label="持卡人" />
            </el-table>

            <h4 class="sub-title">紧急联系人</h4>
            <el-table :data="detailRow.emergency_contacts" border size="small" class="sub-table">
              <el-table-column prop="contact_name" label="联系人" />
              <el-table-column prop="contact_phone" label="联系电话" />
              <el-table-column prop="sort" label="序号" width="80" />
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="职务信息" name="position">
            <div v-if="positionLoading" v-loading="positionLoading" style="min-height:100px" />
            <template v-else>
              <h4>当前职务</h4>
              <el-descriptions v-if="currentPosition" :column="3" border size="small">
                <el-descriptions-item label="在职状态"><StatusTag :status="currentPosition.is_active ? 'calculated' : 'data_changed'" :text="currentPosition.is_active ? '在职' : '已离职'" /></el-descriptions-item>
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
              </el-descriptions>
              <el-empty v-else description="暂无职务信息" :image-size="40" />

              <h4 style="margin-top:16px">变动历史（职务事件）</h4>
              <el-table :data="positionEvents" border size="small">
                <el-table-column prop="event_type" label="事件类型" width="110" />
                <el-table-column prop="effective_date" label="生效日期" width="120" />
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button type="primary" link size="small" @click="editPositionEvent(row)">编辑</el-button>
                    <el-button type="danger" link size="small" @click="removePositionEvent(row)">删除</el-button>
                    <el-button type="warning" link size="small" @click="attachFileId = row.id; attachVisible = true">附件</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <div class="pos-pager">
                <el-pagination
                  v-model:current-page="posPage"
                  v-model:page-size="posPageSize"
                  :total="posTotal"
                  :page-sizes="[5, 10, 20]"
                  layout="total, sizes, prev, pager, next"
                  @size-change="loadPositionEvents"
                  @current-change="loadPositionEvents"
                />
              </div>
            </template>
          </el-tab-pane>

          <el-tab-pane label="假期余额" name="leave-balance">
            <LeaveBalanceDetail :person-id="detailRow.id" />
          </el-tab-pane>

          <el-tab-pane label="附件" name="files">
            <FileAttachPanel v-if="detailRow" :target-type="'person'" :target-id="detailRow.id" />
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-dialog>

    <PersonProfileEditDialog v-model:visible="profileVisible" :person="profileEditRow" @saved="onProfileSaved" />
    <PositionEventEditDialog v-model:visible="posEditVisible" :event="posEditRow" @saved="onPosEventSaved" />

    <RecycleBinDrawer v-model:visible="trashVisible" :fetch-api="fetchDeleted" :restore-api="restorePerson" :columns="trashColumns" @restored="onTrashRestored" />

    <el-dialog v-model="attachVisible" title="文件附件" width="500px">
      <FileAttachPanel :target-type="'position_event'" :target-id="attachFileId" />
    </el-dialog>

    <PersonScopeSwitch v-if="viewMode === 'cards'" v-model="scope" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import StatusTag from '@/components/StatusTag.vue'
import FileAttachPanel from '@/components/FileAttachPanel.vue'
import PersonCard from '@/components/cards/PersonCard.vue'
import CardGrid from '@/components/cards/CardGrid.vue'
import PageToolbar from '@/components/PageToolbar.vue'
import PersonScopeSwitch from '@/components/cards/PersonScopeSwitch.vue'
import LeaveBalanceDetail from '@/components/cards/LeaveBalanceDetail.vue'
import PersonProfileEditDialog from '@/components/person/PersonProfileEditDialog.vue'
import PositionEventEditDialog from '@/components/position/PositionEventEditDialog.vue'
import { getPersons, getPerson, createPerson, deletePerson, restorePerson, getDeletedPersons, getPersonCards, exportPersons } from '@/api/person'
import { getCurrentPosition, getPositionHistory } from '@/api/position-snapshot'
import { getPositionEvents, deletePositionEvent } from '@/api/position-event'
import { filterPersons, type PersonScope } from '@/utils/personScope'
import { downloadBlob } from '@/utils/download'

const tableRef = ref()
const viewMode = ref<'cards' | 'list'>('cards')
const cardGridRef = ref()
const scope = ref<PersonScope>('active')
const addVisible = ref(false)
const profileVisible = ref(false)
const profileEditRow = ref<any>(null)
const posEditVisible = ref(false)
const posEditRow = ref<any>(null)
const detailVisible = ref(false)
const detailRow = ref<any>(null)
const activeTab = ref('info')
const trashVisible = ref(false)
const positionLoading = ref(false)
const currentPosition = ref<any>(null)
const positionHistory = ref<any[]>([])
const positionEvents = ref<any[]>([])
const posPage = ref(1)
const posPageSize = ref(10)
const posTotal = ref(0)
const attachVisible = ref(false)
const attachFileId = ref<number | null>(null)

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
  const list = (await getPersonCards()) as { id: number; name: string }[]
  return k ? list.filter(p => p.name.includes(k)) : list
}

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
async function fetchCards() {
  const cards = (await getPersonCards()) as any[] || []
  return { list: filterPersons(cards, scope.value), total: cards.length }
}

function handleAdd() {
  addVisible.value = true
}

async function submitAdd(data: any) {
  return createPerson(data)
}

async function handleExport() {
  // 列表视图：按当前筛选导出；卡片视图：全量（不传筛选）
  const params: any = {}
  if (viewMode.value === 'list') {
    Object.assign(params, tableRef.value?.getSearchParams() || {})
  }
  const data = await exportPersons(params)
  downloadBlob(data)
}

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
    posPage.value = 1
    loadPositionEvents()
    detailVisible.value = true
  } catch { /* handled */ }
}

async function loadPositionEvents() {
  if (!detailRow.value) return
  const d = (await getPositionEvents({ person_id: detailRow.value.id, pageNum: posPage.value, pageSize: posPageSize.value })) as any
  positionEvents.value = d.list || []
  posTotal.value = d.total || 0
}

function openProfileEdit(row: any) {
  profileEditRow.value = row
  profileVisible.value = true
}

function onProfileSaved() {
  tableRef.value?.refresh()
  cardGridRef.value?.reload()
  handleDetail({ id: profileEditRow.value?.id ?? detailRow.value?.id })
}

function editPositionEvent(row: any) {
  posEditRow.value = row
  posEditVisible.value = true
}

async function removePositionEvent(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除该职务事件（${row.event_type} ${row.effective_date}）？删除后将重建职务快照。`, '提示', { type: 'warning' })
  } catch { return }
  try {
    await deletePositionEvent(row.id)
    ElMessage.success('删除成功')
    loadPositionEvents()
    handleDetail({ id: detailRow.value.id })
  } catch { /* handled */ }
}

function onPosEventSaved() {
  loadPositionEvents()
  handleDetail({ id: detailRow.value.id })
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确认删除「${row.name}」？`, '提示', { type: 'warning' }) } catch { return }
  try { await deletePerson(row.id); ElMessage.success('删除成功'); tableRef.value?.refresh(); cardGridRef.value?.reload() } catch { /* handled */ }
}

function onFormSuccess() { tableRef.value?.refresh(); cardGridRef.value?.reload() }
function onTrashRestored() { tableRef.value?.refresh(); cardGridRef.value?.reload() }
</script>

<style lang="scss" scoped>
.page-container { padding: 0; background: transparent; }


.sub-title { font-size: 14px; font-weight: 600; color: #303133; margin: 14px 0 8px; }
.sub-table { margin-bottom: 4px; }
.pos-pager { display: flex; justify-content: flex-end; margin-top: 8px; }
</style>
