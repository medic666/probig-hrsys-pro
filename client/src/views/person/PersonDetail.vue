<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>人员详情 - {{ person?.name }}</span>
          <el-button @click="$router.back()">返回列表</el-button>
        </div>
      </template>

      <el-tabs v-if="person" v-model="activeTab">
        <el-tab-pane label="基本信息" name="info">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="姓名">{{ person.name }}</el-descriptions-item>
            <el-descriptions-item label="别名">{{ person.alias || '-' }}</el-descriptions-item>
            <el-descriptions-item label="性别">{{ person.gender === 0 ? '男' : person.gender === 1 ? '女' : '-' }}</el-descriptions-item>
            <el-descriptions-item label="民族">{{ person.nation || '-' }}</el-descriptions-item>
            <el-descriptions-item label="籍贯">{{ person.native_place || '-' }}</el-descriptions-item>
            <el-descriptions-item label="政治面貌">{{ person.political_status || '-' }}</el-descriptions-item>
            <el-descriptions-item label="婚姻状态">{{ person.marital_status === 0 ? '未婚' : person.marital_status === 1 ? '已婚' : '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane label="电话" name="phones">
          <div style="margin-bottom:10px">
            <el-button size="small" type="primary" @click="showPhoneDialog()">新增电话</el-button>
          </div>
          <el-table :data="person.phones" border stripe>
            <el-table-column prop="phone" label="电话号码" />
            <el-table-column prop="remark" label="备注" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="showPhoneDialog(row)">编辑</el-button>
                <el-popconfirm title="确定删除？" @confirm="deletePhone(row.id)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="邮箱" name="emails">
          <div style="margin-bottom:10px">
            <el-button size="small" type="primary" @click="showEmailDialog()">新增邮箱</el-button>
          </div>
          <el-table :data="person.emails" border stripe>
            <el-table-column prop="email" label="邮箱" />
            <el-table-column prop="remark" label="备注" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="showEmailDialog(row)">编辑</el-button>
                <el-popconfirm title="确定删除？" @confirm="deleteEmail(row.id)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="银行卡" name="bank_cards">
          <div style="margin-bottom:10px">
            <el-button size="small" type="primary" @click="showBankCardDialog()">新增银行卡</el-button>
          </div>
          <el-table :data="person.bank_cards" border stripe>
            <el-table-column prop="bank_card" label="卡号" />
            <el-table-column prop="bank_name" label="开户行" />
            <el-table-column prop="remark" label="备注" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="showBankCardDialog(row)">编辑</el-button>
                <el-popconfirm title="确定删除？" @confirm="deleteBankCard(row.id)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="职务事件" name="position_events">
          <div style="margin-bottom:10px">
            <el-button size="small" type="primary" @click="showPositionDialog()">新增职务事件</el-button>
          </div>
          <el-table :data="positionEvents" border stripe>
            <el-table-column prop="event_name" label="事件名称" />
            <el-table-column prop="effective_date" label="生效日期" width="120" />
            <el-table-column prop="attendance_group" label="考勤组" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-popconfirm title="确定删除？" @confirm="deletePositionEvent(row.id)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="关联文件" name="files">
          <div style="margin-bottom:10px">
            <el-upload :action="`/api/files/upload`" :headers="uploadHeaders" :data="{ target_type: 'person', target_id: person.id }" :on-success="onUploadSuccess">
              <el-button size="small" type="primary">上传文件</el-button>
            </el-upload>
          </div>
          <el-table :data="files" border stripe>
            <el-table-column prop="name" label="文件名" />
            <el-table-column prop="mime_type" label="类型" width="120" />
            <el-table-column prop="size" label="大小" width="100">
              <template #default="{ row }">{{ (row.size / 1024).toFixed(1) }} KB</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="phoneDialog" title="电话" width="400px">
      <el-form ref="phoneFormRef" :model="phoneForm">
        <el-form-item label="号码"><el-input v-model="phoneForm.phone" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="phoneForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="phoneDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPhone">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="emailDialog" title="邮箱" width="400px">
      <el-form><el-form-item label="邮箱"><el-input v-model="emailForm.email" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="emailForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="emailDialog = false">取消</el-button>
        <el-button type="primary" @click="submitEmail">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="bankCardDialog" title="银行卡" width="400px">
      <el-form><el-form-item label="卡号"><el-input v-model="bankCardForm.bank_card" /></el-form-item>
        <el-form-item label="开户行"><el-input v-model="bankCardForm.bank_name" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="bankCardForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bankCardDialog = false">取消</el-button>
        <el-button type="primary" @click="submitBankCard">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="positionDialog" title="职务事件" width="500px">
      <el-form ref="posFormRef" :model="posForm">
        <el-form-item label="事件名称"><el-input v-model="posForm.event_name" /></el-form-item>
        <el-form-item label="生效日期"><el-date-picker v-model="posForm.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="考勤组"><el-input v-model="posForm.attendance_group" /></el-form-item>
        <el-form-item label="入职日期"><el-date-picker v-model="posForm.entry_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="基本工资"><el-input-number v-model="posForm.base_salary" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="计薪天数"><el-input-number v-model="posForm.salary_days" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="positionDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPosition">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const route = useRoute()
const personId = computed(() => Number(route.params.id))
const person = ref<any>(null)
const positionEvents = ref([])
const files = ref([])
const activeTab = ref('info')
const uploadHeaders = { Authorization: `Bearer ${localStorage.getItem('token')}` }

const phoneDialog = ref(false); const phoneForm = reactive({ phone: '', remark: '' }); let phoneEditId = 0
const emailDialog = ref(false); const emailForm = reactive({ email: '', remark: '' }); let emailEditId = 0
const bankCardDialog = ref(false); const bankCardForm = reactive({ bank_card: '', bank_name: '', remark: '' }); let bankCardEditId = 0
const positionDialog = ref(false); const posForm = reactive({ event_name: '', effective_date: '', attendance_group: '', entry_date: '', base_salary: undefined, salary_days: undefined } as any)

async function fetchPerson() {
  const res = await request.get(`/persons/${personId.value}`)
  person.value = res.data
}

async function fetchEvents() {
  const res = await request.get('/position-events', { params: { person_id: personId.value, page_size: 100 } })
  positionEvents.value = res.data.list
}

async function fetchFiles() {
  const res = await request.get('/file-relations', { params: { target_type: 'person', target_id: personId.value } })
  files.value = res.data || []
}

onMounted(async () => { await fetchPerson(); await fetchEvents(); await fetchFiles() })

function onUploadSuccess() { fetchFiles() }

function showPhoneDialog(row?: any) {
  if (row) { phoneEditId = row.id; phoneForm.phone = row.phone; phoneForm.remark = row.remark }
  else { phoneEditId = 0; phoneForm.phone = ''; phoneForm.remark = '' }
  phoneDialog.value = true
}
async function submitPhone() {
  if (phoneEditId) { await request.put(`/person-phones/${phoneEditId}`, phoneForm) }
  else { await request.post(`/persons/${personId.value}/phones`, phoneForm) }
  phoneDialog.value = false; fetchPerson(); ElMessage.success('保存成功')
}
async function deletePhone(id: number) { await request.delete(`/person-phones/${id}`); fetchPerson() }

function showEmailDialog(row?: any) {
  if (row) { emailEditId = row.id; emailForm.email = row.email; emailForm.remark = row.remark }
  else { emailEditId = 0; emailForm.email = ''; emailForm.remark = '' }
  emailDialog.value = true
}
async function submitEmail() {
  if (emailEditId) { await request.put(`/person-emails/${emailEditId}`, emailForm) }
  else { await request.post(`/persons/${personId.value}/emails`, emailForm) }
  emailDialog.value = false; fetchPerson(); ElMessage.success('保存成功')
}
async function deleteEmail(id: number) { await request.delete(`/person-emails/${id}`); fetchPerson() }

function showBankCardDialog(row?: any) {
  if (row) { bankCardEditId = row.id; bankCardForm.bank_card = row.bank_card; bankCardForm.bank_name = row.bank_name; bankCardForm.remark = row.remark }
  else { bankCardEditId = 0; bankCardForm.bank_card = ''; bankCardForm.bank_name = ''; bankCardForm.remark = '' }
  bankCardDialog.value = true
}
async function submitBankCard() {
  if (bankCardEditId) { await request.put(`/person-bank-cards/${bankCardEditId}`, bankCardForm) }
  else { await request.post(`/persons/${personId.value}/bank-cards`, bankCardForm) }
  bankCardDialog.value = false; fetchPerson()
}
async function deleteBankCard(id: number) { await request.delete(`/person-bank-cards/${id}`); fetchPerson() }

function showPositionDialog() { posForm.event_name = ''; posForm.effective_date = ''; posForm.attendance_group = ''; posForm.entry_date = ''; posForm.base_salary = undefined; posForm.salary_days = undefined; positionDialog.value = true }
async function submitPosition() {
  const payload: any = {}
  if (posForm.event_name) payload.event_name = posForm.event_name
  if (posForm.effective_date) payload.effective_date = posForm.effective_date
  if (posForm.attendance_group) payload.attendance_group = posForm.attendance_group
  if (posForm.entry_date) payload.entry_date = posForm.entry_date
  if (posForm.base_salary !== undefined) payload.base_salary = posForm.base_salary
  if (posForm.salary_days !== undefined) payload.salary_days = posForm.salary_days
  payload.person_id = personId.value
  await request.post('/position-events', payload)
  positionDialog.value = false; fetchEvents()
  ElMessage.success('职务事件已提交，快照正在重建')
}
async function deletePositionEvent(id: number) { await request.delete(`/position-events/${id}`); fetchEvents() }
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
