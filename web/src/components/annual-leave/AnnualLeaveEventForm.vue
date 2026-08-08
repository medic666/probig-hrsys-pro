<template>
  <el-form label-width="90px">
    <el-form-item label="人员" required>
      <NameSelect v-model="form.person_id" placeholder="选择" :disabled="isEdit" />
    </el-form-item>
    <el-form-item label="类型" required>
      <el-select v-model="form.event_type" style="width:100%">
        <el-option v-for="t in types" :key="t.value" :label="t.label" :value="t.value" />
      </el-select>
    </el-form-item>
    <el-form-item label="时长(小时)" required>
      <el-input-number v-model="form.hours" :precision="1" style="width:100%" />
    </el-form-item>
    <el-form-item label="生效日期" required>
      <el-date-picker v-model="form.effective_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
    </el-form-item>
    <el-form-item label="备注">
      <el-input v-model="form.remark" type="textarea" :rows="2" />
    </el-form-item>
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button v-permission="PERM.annualLeaveEventWrite" type="primary" :loading="saving" @click="doSave">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { getAnnualLeaveEvent, createAnnualLeaveEvent, updateAnnualLeaveEvent } from '@/api/annual-leave'
import { PERM } from '@/constants/permission'
// 新增=编辑统一表单：id 缺失 → 新增；{id} → 编辑
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)
// 类型下拉：label 中文、value 保持英文存储值
const types = [
  { label: '配发', value: 'grant' },
  { label: '人工调整', value: 'adjust' },
  { label: '结转扣除', value: 'carryover_deduct' },
]
const form = reactive({ person_id: null as any, event_type: 'grant', hours: 8, effective_date: '', remark: '' })

onMounted(async () => {
  if (isEdit.value) {
    try {
      const row = (await getAnnualLeaveEvent(props.id!)) as any
      form.person_id = row.person_id
      form.event_type = row.event_type || 'grant'
      form.hours = row.hours ?? 8
      form.effective_date = row.effective_date || ''
      form.remark = row.remark || ''
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.person_id) { ElMessage.warning('请选择人员'); return }
  if (!form.effective_date) { ElMessage.warning('请选择生效日期'); return }
  saving.value = true
  try {
    const data = {
      person_id: form.person_id,
      event_type: form.event_type,
      hours: form.hours,
      effective_date: form.effective_date,
      remark: form.remark,
    }
    if (isEdit.value) {
      await updateAnnualLeaveEvent(props.id!, data)
    } else {
      await createAnnualLeaveEvent(data)
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
