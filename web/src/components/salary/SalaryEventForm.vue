<template>
  <el-form label-width="90px">
    <el-form-item label="人员" required>
      <NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" :disabled="isEdit" />
    </el-form-item>
    <el-form-item label="归属月份" required>
      <el-date-picker v-model="form.belong_month" type="month" value-format="YYYY-MM" style="width:100%" />
    </el-form-item>
    <el-form-item label="事件类型" required>
      <el-select v-model="form.event_type" style="width:100%">
        <el-option v-for="t in types" :key="t" :label="t" :value="t" />
      </el-select>
    </el-form-item>
    <el-form-item label="值" required>
      <el-input-number v-model="form.amount" :precision="2" style="width:100%" />
    </el-form-item>
    <el-form-item label="备注">
      <el-input v-model="form.remark" type="textarea" :rows="2" />
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
import { getSalaryEvent, createSalaryEvent, updateSalaryEvent } from '@/api/salary'
import { getAllPersons } from '@/api/person'

// 新增=编辑统一表单：id 缺失 → 新增；{id} → 编辑
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)
const types = ['绩效系数', '提成', '奖惩', '借款还款', '个税扣除']
const form = reactive({ person_id: null as any, belong_month: '', event_type: '绩效系数', amount: 1, remark: '' })

onMounted(async () => {
  if (isEdit.value) {
    try {
      const row = (await getSalaryEvent(props.id!)) as any
      form.person_id = row.person_id
      form.belong_month = row.belong_month || ''
      form.event_type = row.event_type || '绩效系数'
      form.amount = row.amount ?? 0
      form.remark = row.remark || ''
    } catch { /* handled */ }
  }
})

async function fetchPersonOpts(k?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[]
  return k ? list.filter(p => p.name.includes(k)) : list
}

async function doSave() {
  if (!form.person_id) { ElMessage.warning('请选择人员'); return }
  if (!form.belong_month) { ElMessage.warning('请选择归属月份'); return }
  saving.value = true
  try {
    const data = {
      person_id: form.person_id,
      belong_month: form.belong_month,
      event_type: form.event_type,
      amount: form.amount,
      remark: form.remark,
    }
    if (isEdit.value) {
      await updateSalaryEvent(props.id!, data)
    } else {
      await createSalaryEvent(data)
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
