<template>
  <el-form label-width="90px">
    <el-row :gutter="16">
      <el-col :span="8">
        <el-form-item label="人员" required>
          <NameSelect v-model="form.person_id" :fetch-api="fetchPersonOpts" placeholder="选择人员" :disabled="isEdit" />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="日期" required>
          <el-date-picker v-model="form.event_date" type="date" value-format="YYYY-MM-DD" style="width:100%" :disabled="isEdit" />
        </el-form-item>
      </el-col>
      <el-col :span="8">
        <el-form-item label="状态">
          <el-tag v-if="form.status" :type="form.status === 'pending' ? 'warning' : 'success'" size="small">
            {{ form.status === 'pending' ? '待确认' : '已确认' }}
          </el-tag>
          <span v-else style="color:#909399">新增后为已确认</span>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="打卡时间"><el-input v-model="form.punch_time" placeholder="如 08:30,18:00" /></el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-col>
    </el-row>
  </el-form>

  <el-table :data="details" border size="small">
    <el-table-column label="类型" width="110">
      <template #default="{ row }">
        <el-select v-model="row.event_type" size="small" style="width:100%" @change="onTypeChange(row)">
          <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
        </el-select>
      </template>
    </el-table-column>
    <el-table-column label="子类型" width="140">
      <template #default="{ row }">
        <el-select v-if="row.event_type !== '违纪'" v-model="row.sub_type" size="small" style="width:100%">
          <el-option v-for="s in subTypeMap[row.event_type] || []" :key="s" :label="s" :value="s" />
        </el-select>
      </template>
    </el-table-column>
    <el-table-column label="时长(小时)" width="120">
      <template #default="{ row }">
        <el-input-number v-if="row.event_type !== '违纪'" v-model="row.hours" :min="0" :precision="1" size="small" style="width:100%" />
      </template>
    </el-table-column>
    <el-table-column label="分钟" width="100">
      <template #default="{ row }">
        <el-input-number v-if="row.sub_type === '迟到' || row.sub_type === '早退'" v-model="row.minutes" :min="0" size="small" style="width:100%" />
      </template>
    </el-table-column>
    <el-table-column label="备注" min-width="140">
      <template #default="{ row }">
        <el-input v-model="row.remark" size="small" />
      </template>
    </el-table-column>
    <el-table-column label="操作" width="70">
      <template #default="{ $index }">
        <el-button type="danger" link size="small" @click="details.splice($index, 1)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <el-button size="small" style="margin-top:8px" @click="addEvent">+ 添加事件</el-button>
  <div class="hint">编辑=查看：打开即回显整日全部事件原值；保存即事务提交。{{ isConfirm ? '待确认记录保存后将置为已确认。' : '' }}</div>

  <div class="form-footer">
    <el-button @click="$emit('cancel')">取消</el-button>
    <el-button type="primary" :loading="saving" @click="doSave">确定</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import NameSelect from '@/components/NameSelect.vue'
import { getAttendanceEvent, createAttendanceEvent, updateAttendanceEvent, confirmPendingDaily } from '@/api/attendance'
import { getAllPersons } from '@/api/person'

// 新增=编辑=查看统一表单：id 缺失 → 新增；{id} → 编辑（回显整日明细）；
// confirm=true 时保存走"确认"语义（待确认 → 已确认）
const props = defineProps<{ id?: number | null; confirm?: boolean }>()
const emit = defineEmits<{ (e: 'saved'): void; (e: 'cancel'): void }>()

const isEdit = computed(() => props.id != null)
const isConfirm = computed(() => props.confirm === true)
const saving = ref(false)

const eventTypes = ['出勤', '休假', '加班', '违纪']
const subTypeMap: Record<string, string[]> = {
  '出勤': ['普通出勤', '补班出勤', '外勤出勤'],
  '休假': ['调休', '事假', '病假', '年假', '法定假', '福利假'],
  '加班': ['工作日加班', '节假日加班'],
  '违纪': ['缺卡', '迟到', '早退'],
}

const form = reactive({ person_id: null as any, event_date: '', status: '', punch_time: '', remark: '' })
const details = ref<any[]>([])

onMounted(async () => {
  if (isEdit.value) {
    try {
      const d = (await getAttendanceEvent(props.id!)) as any
      form.person_id = d.person_id
      form.event_date = d.event_date
      form.status = d.status || ''
      form.punch_time = d.punch_time || ''
      form.remark = d.remark || ''
      details.value = JSON.parse(JSON.stringify(d.details || []))
    } catch { /* handled */ }
  }
})

function addEvent() {
  details.value.push({ event_type: '出勤', sub_type: '普通出勤', hours: 8, minutes: 0, remark: '' })
}

function onTypeChange(row: any) {
  row.sub_type = ''
  row.hours = row.event_type === '违纪' ? 0 : 8
  row.minutes = 0
}

async function fetchPersonOpts(k?: string) {
  const list = (await getAllPersons()) as { id: number; name: string }[]
  return k ? list.filter(p => p.name.includes(k)) : list
}

async function doSave() {
  if (!form.person_id) { ElMessage.warning('请选择人员'); return }
  if (!form.event_date) { ElMessage.warning('请选择日期'); return }
  saving.value = true
  try {
    if (isConfirm.value) {
      await confirmPendingDaily(props.id!, details.value, form.punch_time, form.remark)
    } else if (isEdit.value) {
      await updateAttendanceEvent(props.id!, { punch_time: form.punch_time, remark: form.remark, details: details.value })
    } else {
      await createAttendanceEvent({
        person_id: form.person_id,
        event_date: form.event_date,
        punch_time: form.punch_time,
        remark: form.remark,
        details: details.value,
      })
    }
    ElMessage.success('保存成功')
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
