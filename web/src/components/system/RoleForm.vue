<template>
  <el-form label-width="100px">
    <el-form-item label="角色名称" required>
      <el-input v-model="form.name" />
    </el-form-item>
    <el-form-item label="备注">
      <el-input v-model="form.remark" type="textarea" :rows="3" />
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
import { getRole, createRole, updateRole } from '@/api/role'

// 新增=编辑统一表单：id 缺失 → 新增；{id} → 编辑
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)
const form = reactive({ name: '', remark: '' })

onMounted(async () => {
  if (isEdit.value) {
    try {
      const row = (await getRole(props.id!)) as any
      form.name = row.name || ''
      form.remark = row.remark || ''
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.name) { ElMessage.warning('请填写角色名称'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateRole(props.id!, form)
    } else {
      await createRole(form)
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
