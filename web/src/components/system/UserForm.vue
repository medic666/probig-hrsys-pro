<template>
  <el-form label-width="100px">
    <el-form-item label="用户名" required>
      <el-input v-model="form.username" :disabled="isEdit" />
    </el-form-item>
    <el-form-item v-if="!isEdit" label="密码">
      <el-input v-model="form.password" placeholder="留空使用默认密码 123456" />
    </el-form-item>
    <el-form-item label="关联人员">
      <NameSelect v-model="form.person_id" :disabled="isAdmin" />
      <span v-if="form.data_scope === 'own' && !form.person_id" class="scope-hint">「仅自己」范围必须关联人员</span>
    </el-form-item>
    <el-form-item label="数据范围">
      <el-radio-group v-model="form.data_scope" :disabled="isAdmin">
        <el-radio-button value="all">全部</el-radio-button>
        <el-radio-button value="own">仅自己</el-radio-button>
      </el-radio-group>
      <span v-if="isAdmin" class="scope-hint">超级管理员锁定「全部」</span>
    </el-form-item>
    <el-form-item label="启用">
      <el-switch v-model="form.is_active" :disabled="isAdmin" />
    </el-form-item>
    <el-alert v-if="isAdmin" type="warning" :closable="false" title="超级管理员账号不可编辑，仅可被其它拥有管理员权限的账户重置密码" style="margin-bottom:12px" />
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button v-if="!isAdmin" v-permission="PERM.userWrite" type="primary" :loading="saving" @click="doSave">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { getUser, createUser, updateUser } from '@/api/user'
import { PERM } from '@/constants/permission'

// 新增=编辑统一表单：id 缺失 → 新增；{id} → 编辑
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const isAdmin = computed(() => form.username === 'admin')
const saving = ref(false)
const form = reactive({ username: '', password: '', person_id: null as any, data_scope: 'all', is_active: true })

onMounted(async () => {
  if (isEdit.value) {
    try {
      const row = (await getUser(props.id!)) as any
      form.username = row.username || ''
      form.person_id = row.person_id || null
      form.data_scope = row.data_scope || 'all'
      form.is_active = !!row.is_active
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.username) { ElMessage.warning('请填写用户名'); return }
  if (form.data_scope === 'own' && !form.person_id) { ElMessage.warning('数据范围为「仅自己」时必须关联人员'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateUser(props.id!, { username: form.username, person_id: form.person_id, data_scope: form.data_scope, is_active: form.is_active })
    } else {
      await createUser({ username: form.username, password: form.password, person_id: form.person_id, data_scope: form.data_scope, is_active: form.is_active })
    }
    ElMessage.success(isEdit.value ? '保存成功' : '创建成功')
    emit('saved')
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.scope-hint {
  margin-left: 8px;
  font-size: 12px;
  color: #e6a23c;
}
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
