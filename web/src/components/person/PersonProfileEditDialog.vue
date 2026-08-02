<template>
  <el-dialog :model-value="visible" :title="isAdd ? '新增人员' : '编辑人员档案'" width="720px" @close="handleClose">
    <el-form label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="姓名" required><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="身份证号"><el-input v-model="form.id_card" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="性别">
              <el-select v-model="form.gender" placeholder="请选择" style="width:100%">
                <el-option label="未设置" :value="0" />
                <el-option label="男" :value="1" />
                <el-option label="女" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="生日"><el-date-picker v-model="form.birthday" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="民族"><el-input v-model="form.nation" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="籍贯"><el-input v-model="form.native_place" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="住址"><el-input v-model="form.address" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="政治面貌"><el-input v-model="form.political_status" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="婚姻状态">
              <el-select v-model="form.marital_status" placeholder="请选择" style="width:100%">
                <el-option label="未设置" :value="0" />
                <el-option label="已婚" :value="1" />
                <el-option label="未婚" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="别名"><el-input v-model="form.alias" /></el-form-item></el-col>
        </el-row>
      </el-form>

      <el-divider content-position="left">电话</el-divider>
      <el-table :data="form.phones" border size="small">
        <el-table-column label="号码">
          <template #default="{ row }"><el-input v-model="row.phone" size="small" /></template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ $index }"><el-button type="danger" link size="small" @click="form.phones.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:6px" @click="form.phones.push({ id: 0, phone: '', phone_type: 'mobile' })">+ 添加电话</el-button>

      <el-divider content-position="left">邮箱</el-divider>
      <el-table :data="form.emails" border size="small">
        <el-table-column label="邮箱">
          <template #default="{ row }"><el-input v-model="row.email" size="small" /></template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ $index }"><el-button type="danger" link size="small" @click="form.emails.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:6px" @click="form.emails.push({ id: 0, email: '', email_type: 'personal' })">+ 添加邮箱</el-button>

      <el-divider content-position="left">银行卡</el-divider>
      <el-table :data="form.bank_cards" border size="small">
        <el-table-column label="开户行">
          <template #default="{ row }"><el-input v-model="row.bank_name" size="small" /></template>
        </el-table-column>
        <el-table-column label="账号">
          <template #default="{ row }"><el-input v-model="row.account_number" size="small" /></template>
        </el-table-column>
        <el-table-column label="持卡人">
          <template #default="{ row }"><el-input v-model="row.account_holder" size="small" /></template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ $index }"><el-button type="danger" link size="small" @click="form.bank_cards.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:6px" @click="form.bank_cards.push({ id: 0, bank_name: '', account_number: '', account_holder: '' })">+ 添加银行卡</el-button>

      <el-divider content-position="left">紧急联系人</el-divider>
      <el-table :data="form.emergency_contacts" border size="small">
        <el-table-column label="联系人">
          <template #default="{ row }"><el-input v-model="row.contact_name" size="small" /></template>
        </el-table-column>
        <el-table-column label="联系电话">
          <template #default="{ row }"><el-input v-model="row.contact_phone" size="small" /></template>
        </el-table-column>
        <el-table-column label="序号" width="90">
          <template #default="{ row }"><el-input-number v-model="row.sort" :min="1" size="small" style="width:100%" /></template>
        </el-table-column>
        <el-table-column label="操作" width="70">
          <template #default="{ $index }"><el-button type="danger" link size="small" @click="form.emergency_contacts.splice($index, 1)">删除</el-button></template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:6px" @click="form.emergency_contacts.push({ id: 0, contact_name: '', contact_phone: '', sort: 1 })">+ 添加联系人</el-button>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="saving" @click="doSave">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { upsertPersonProfile } from '@/api/person'

const props = defineProps<{
  visible: boolean
  person: any
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'saved'): void
}>()

// 新增=编辑统一语义：person.id 缺失即新增
const isAdd = computed(() => !props.person?.id)

const saving = ref(false)
const form = reactive<any>({})

// 每次打开先重置表单（新增/编辑统一契约），存在 person 再回填
const emptyForm = () => ({
  name: '',
  id_card: '',
  gender: 0,
  birthday: '',
  nation: '',
  native_place: '',
  address: '',
  political_status: '',
  marital_status: 0,
  alias: '',
  phones: [],
  emails: [],
  bank_cards: [],
  emergency_contacts: [],
})

function fillForm(p: any) {
  Object.assign(form, {
    name: p.name || '',
    id_card: p.id_card || '',
    gender: p.gender ?? 0,
    birthday: p.birthday || '',
    nation: p.nation || '',
    native_place: p.native_place || '',
    address: p.address || '',
    political_status: p.political_status || '',
    marital_status: p.marital_status ?? 0,
    alias: p.alias || '',
    phones: JSON.parse(JSON.stringify(p.phones || [])),
    emails: JSON.parse(JSON.stringify(p.emails || [])),
    bank_cards: JSON.parse(JSON.stringify(p.bank_cards || [])),
    emergency_contacts: JSON.parse(JSON.stringify(p.emergency_contacts || [])),
  })
}

watch(
  () => props.visible,
  (v) => {
    if (!v) return
    Object.assign(form, emptyForm())
    if (props.person) {
      fillForm(props.person)
    }
  },
)

function handleClose() {
  emit('update:visible', false)
}

async function doSave() {
  if (!form.name) { ElMessage.warning('请填写姓名'); return }
  saving.value = true
  try {
    await upsertPersonProfile({
      id: props.person?.id || 0,
      name: form.name,
      id_card: form.id_card,
      gender: form.gender,
      birthday: form.birthday || null,
      nation: form.nation,
      native_place: form.native_place,
      address: form.address,
      political_status: form.political_status,
      marital_status: form.marital_status,
      alias: form.alias,
      phones: form.phones,
      emails: form.emails,
      bank_cards: form.bank_cards,
      emergency_contacts: form.emergency_contacts,
    })
    ElMessage.success(isAdd.value ? '创建成功' : '保存成功')
    handleClose()
    emit('saved')
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>
