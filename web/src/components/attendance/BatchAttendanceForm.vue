<template>
  <el-form label-width="100px" class="batch-attendance-form">
    <el-form-item label="人员" required>
      <PersonDomainSelect v-model="form.person_ids" />
    </el-form-item>
    <el-form-item label="时间段" required>
      <el-date-picker v-model="form.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width:100%" />
    </el-form-item>
    <el-form-item label="状态">
      <el-radio-group v-model="form.status" size="small">
        <el-radio-button value="confirmed">已确认</el-radio-button>
        <el-radio-button value="pending">待确认</el-radio-button>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="打卡时间">
      <el-input v-model="form.punch_time" placeholder="如 08:30,18:00" style="width:240px" />
    </el-form-item>
    <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
  </el-form>

  <div class="details-block">
    <div class="details-title">事件明细</div>
    <AttendanceDetailsEditor v-model="form.details" />
  </div>
  <div class="hint">批量录入 = 同一组事件明细 × 所选人员 × 时间段内每一天，每天新增一条考勤组；当天已有记录将标记为待确认（最新记录优先）。</div>
    <div class="form-footer">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button v-permission="PERM.attendanceWrite" type="primary" :loading="saving" @click="doSubmit">确定</el-button>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import AttendanceDetailsEditor from '@/components/attendance/AttendanceDetailsEditor.vue'
import PersonDomainSelect from '@/components/PersonDomainSelect.vue'
import { createBatchAttendanceEvents } from '@/api/attendance'
import { PERM } from '@/constants/permission'

// 批量录入考勤事件：人员多选（多域筛选）× 时间段 × 状态 × 明细组，提交后返回列表页。
// 与录入考勤共用明细编辑器，同一组明细应用到每一天（当天已有记录整体覆盖）。
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const saving = ref(false)
const form = reactive({
  person_ids: [] as number[],
  dateRange: [] as string[],
  status: 'confirmed',
  punch_time: '',
  remark: '',
  details: [] as any[],
})

async function doSubmit() {
  if (!form.person_ids.length) { ElMessage.warning('请选择人员'); return }
  if (!form.dateRange || form.dateRange.length !== 2) { ElMessage.warning('请选择时间段'); return }
  if (!form.details.length) { ElMessage.warning('请至少添加一条事件明细'); return }
  saving.value = true
  try {
    const d = await createBatchAttendanceEvents({
      person_ids: form.person_ids,
      start_date: form.dateRange[0],
      end_date: form.dateRange[1],
      status: form.status,
      punch_time: form.punch_time,
      remark: form.remark,
      details: form.details,
    }) as any
    ElMessage.success(`批量录入完成: 成功${d.success}条, 失败${d.fail}条`)
    emit('saved')
  } catch { /* handled */ } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.details-block {
  margin-top: 16px;
}

.details-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

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
