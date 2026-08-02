<template>
  <el-form label-width="110px">
    <el-form-item label="公司名称" required><el-input v-model="form.name" /></el-form-item>
    <el-form-item label="统一社会信用代码"><el-input v-model="form.credit_code" /></el-form-item>
    <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
    <el-form-item label="联系电话"><el-input v-model="form.contact_phone" /></el-form-item>
    <el-form-item label="开户行"><el-input v-model="form.bank_name" /></el-form-item>
    <el-form-item label="银行账号"><el-input v-model="form.bank_account" /></el-form-item>
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="doSave">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCompany, createCompany, updateCompany } from '@/api/company'

// 新增=编辑统一表单：company 为 null 或 {id} 缺失 → 新增；{id} → 编辑
const props = defineProps<{ company: any }>()
const emit = defineEmits<{ (e: 'saved', id: number): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.company?.id != null)
const saving = ref(false)
const form = reactive({
  name: '',
  credit_code: '',
  address: '',
  contact_phone: '',
  bank_name: '',
  bank_account: '',
})

onMounted(async () => {
  if (isEdit.value) {
    try {
      const c = (await getCompany(props.company.id)) as any
      Object.assign(form, {
        name: c.name || '',
        credit_code: c.credit_code || '',
        address: c.address || '',
        contact_phone: c.contact_phone || '',
        bank_name: c.bank_name || '',
        bank_account: c.bank_account || '',
      })
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.name) { ElMessage.warning('请填写公司名称'); return }
  saving.value = true
  try {
    let id = props.company?.id
    if (isEdit.value) {
      await updateCompany(id, form)
    } else {
      const d = (await createCompany(form)) as any
      id = d?.id
    }
    ElMessage.success(isEdit.value ? '保存成功' : '创建成功')
    emit('saved', id)
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
