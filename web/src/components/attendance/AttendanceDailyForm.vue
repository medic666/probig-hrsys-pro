<template>
  <el-form label-width="90px">
    <el-row :gutter="16">
      <el-col :xs="24" :sm="8">
        <el-form-item label="人员" required>
          <NameSelect v-model="form.person_id" placeholder="选择人员" :disabled="isEdit || readonly" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-form-item label="日期" required>
          <el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" :disabled="isEdit || readonly" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-form-item label="状态">
          <el-radio-group v-model="form.status" size="small" :disabled="readonly">
            <el-radio-button value="confirmed">已确认</el-radio-button>
            <el-radio-button value="pending">待确认</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-form-item label="打卡时间"><el-input v-model="form.punch_time" placeholder="如 08:30,18:00" :disabled="readonly" /></el-form-item>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-form-item label="备注"><el-input v-model="form.remark" :disabled="readonly" /></el-form-item>
      </el-col>
    </el-row>
  </el-form>

  <AttendanceDetailsEditor v-model="details" :readonly="readonly" />
  <div v-if="!readonly && isEdit" class="hint">保存后该组将成为当日最新版本（序号递增），当日其它组将标记为待确认；待确认记录不参与投影核算。</div>
  <div v-else-if="!readonly" class="hint">录入将新增一条考勤组；若当天已有记录，已有记录将标记为待确认（最新记录优先，同日仅最新组可确认）；待确认记录不参与投影核算。</div>
  <div v-else class="hint">只读视图：从日记工时模块进入，仅查看当日考勤明细。</div>

  <div v-if="!readonly" class="form-footer">
    <el-button @click="$emit('cancel')">取消</el-button>
    <el-button v-permission="PERM.attendanceWrite" type="primary" :loading="saving" @click="doSave">确定</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import AttendanceDetailsEditor from '@/components/attendance/AttendanceDetailsEditor.vue'
import { getAttendanceEvent, createAttendanceEvent, confirmAttendanceDaily } from '@/api/attendance'
import { PERM } from '@/constants/permission'
// 新增=编辑=查看统一表单：id 缺失 → 新增；{id} → 编辑（回显整日明细）。
// 状态自选（已确认/待确认）：新增随创建提交，编辑随确认接口提交。
// readonly 只读模式：日记工时模块进入的查看入口，无编辑/确认能力。
const props = withDefaults(defineProps<{ id?: number | null; readonly?: boolean }>(), { id: null, readonly: false })
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const saving = ref(false)

const form = reactive({ person_id: null as any, event_date: '', status: 'confirmed', punch_time: '', remark: '' })
const details = ref<any[]>([])

onMounted(async () => {
  if (isEdit.value) {
    try {
      const d = (await getAttendanceEvent(props.id!)) as any
      form.person_id = d.person_id
      form.event_date = d.event_date
      form.status = d.status || 'confirmed'
      form.punch_time = d.punch_time || ''
      form.remark = d.remark || ''
      details.value = JSON.parse(JSON.stringify(d.details || []))
    } catch { /* handled */ }
  }
})

async function doSave() {
  if (!form.person_id) { ElMessage.warning('请选择人员'); return }
  if (!form.event_date) { ElMessage.warning('请选择日期'); return }
  saving.value = true
  try {
    if (isEdit.value) {
      // 编辑保存 = 按所选状态提交整日（与卡片"确认"同一入口，卡片不传状态默认已确认）
      await confirmAttendanceDaily(props.id!, details.value, form.punch_time, form.remark, form.status)
    } else {
      await createAttendanceEvent({
        person_id: form.person_id,
        event_date: form.event_date,
        status: form.status,
        punch_time: form.punch_time,
        remark: form.remark,
        details: details.value,
      })
    }
    ElMessage.success(isEdit.value ? '保存成功' : '录入成功')
    emit('saved')
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
.form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
