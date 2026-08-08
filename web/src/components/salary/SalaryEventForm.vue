<template>
  <el-form label-width="90px">
    <el-form-item label="人员" required>
      <PersonDomainSelect v-if="!isEdit" v-model="form.person_ids" />
      <NameSelect v-else v-model="form.person_id" :disabled="isEdit" />
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
      <el-button v-permission="PERM.salaryWrite" type="primary" :loading="saving" @click="doSave">确定</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import PersonDomainSelect from '@/components/PersonDomainSelect.vue'
import { getSalaryEvent, createSalaryEvent, updateSalaryEvent } from '@/api/salary'
import { PERM } from '@/constants/permission'
// 新增=编辑统一表单：id 缺失 → 新增（人员可多选，逐人创建）；{id} → 编辑（单人）
const props = defineProps<{ id?: number | null }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)
const types = ['绩效系数', '提成', '奖惩', '工资预支', '预支还款', '个税扣除']
const form = reactive({ person_ids: [] as number[], person_id: null as any, belong_month: '', event_type: '绩效系数', amount: 1, remark: '' })

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

async function doSave() {
  if (isEdit.value ? !form.person_id : !form.person_ids.length) { ElMessage.warning('请选择人员'); return }
  if (!form.belong_month) { ElMessage.warning('请选择归属月份'); return }
  saving.value = true
  try {
    const data = {
      belong_month: form.belong_month,
      event_type: form.event_type,
      amount: form.amount,
      remark: form.remark,
    }
    if (isEdit.value) {
      await updateSalaryEvent(props.id!, { ...data, person_id: form.person_id })
      ElMessage.success('保存成功')
    } else {
      // 多选创建：逐人创建，失败计数继续（与批量录入策略一致）
      let success = 0
      let fail = 0
      for (const pid of form.person_ids) {
        try {
          await createSalaryEvent({ ...data, person_id: pid })
          success++
        } catch {
          fail++
        }
      }
      ElMessage.success(`创建成功 ${success} 条，失败 ${fail} 条`)
    }
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
