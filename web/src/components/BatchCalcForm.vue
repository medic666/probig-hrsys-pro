<template>
  <el-form label-width="90px" class="batch-calc-form">
    <el-form-item label="月份" required>
      <el-date-picker v-model="form.month" type="month" value-format="YYYY-MM" style="width:100%" />
    </el-form-item>
    <el-form-item label="人员">
      <PersonDomainSelect v-model="form.personIds" />
      <div class="hint">不选则核算全部在职人员</div>
    </el-form-item>
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="doSubmit">开始核算</el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import PersonDomainSelect from '@/components/PersonDomainSelect.vue'

// 批量核算统一表单（考勤/工资共用）：月份 + 人员域多选，提交函数由页面注入。
// 核算结果为三态提示：有结果（含 0）/ 空结果（置空）/ 失败（需人工干预）。
const props = withDefaults(
  defineProps<{
    submitFn: (data: any) => Promise<any>
  }>(),
  { submitFn: undefined },
)

const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const saving = ref(false)
const form = reactive({
  month: '',
  personIds: [] as number[],
})

async function doSubmit() {
  if (!form.month) { ElMessage.warning('请选择月份'); return }
  saving.value = true
  try {
    const d = await props.submitFn({ month: form.month, person_ids: form.personIds })
    ElMessage.success(`核算完成: 有结果${d.has_value}条, 空结果${d.empty}条, 失败${d.fail}条`)
    emit('saved')
  } catch { /* handled */ } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.hint {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
