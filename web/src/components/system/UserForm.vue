<template>
  <el-form label-width="100px">
    <el-form-item label="用户名" required>
      <el-input v-model="form.username" :disabled="isEdit" />
    </el-form-item>
    <el-form-item v-if="!isEdit" label="密码">
      <el-input v-model="form.password" placeholder="留空使用默认密码 admin123" />
    </el-form-item>
    <el-form-item label="关联人员">
      <NameSelect v-model="form.person_id" />
    </el-form-item>
    <el-form-item label="启用">
      <el-switch v-model="form.is_active" />
    </el-form-item>
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="doSave">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { getUser, createUser, updateUser } from '@/api/user'

// 新增=编辑统一表单：id 缺失 → 新增；{id} → 编辑
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)
const form = reactive({ username: '', password: '', person_id: null as any, is_active: true })

onMounted(async () => {
  if (isEdit.value) {
    try {
      const row = (await getUser(props.id!)) as any
      form.username = row.username || ''
      form.person_id = row.person_id || null
      form.is_active = !!row.is_active
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.username) { ElMessage.warning('请填写用户名'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateUser(props.id!, { username: form.username, person_id: form.person_id, is_active: form.is_active })
    } else {
      await createUser({ username: form.username, password: form.password, person_id: form.person_id, is_active: form.is_active })
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
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
