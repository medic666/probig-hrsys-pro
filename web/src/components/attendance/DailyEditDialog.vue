<template>
  <el-dialog :model-value="visible" title="编辑当日事件" width="720px" @close="handleClose">
    <template v-if="daily">
      <el-descriptions :column="3" border size="small">
        <el-descriptions-item label="人员">{{ daily.person_name }}</el-descriptions-item>
        <el-descriptions-item label="日期">{{ daily.event_date }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ daily.status === 'pending' ? '待确认' : '已确认' }}</el-descriptions-item>
      </el-descriptions>

      <el-form style="margin-top:12px" label-width="80px">
        <el-form-item label="打卡时间">
          <el-input v-model="punchTime" placeholder="如 08:30,18:00" style="width:240px" />
        </el-form-item>
      </el-form>

      <el-table :data="details" border size="small">
        <el-table-column label="类型" width="90">
          <template #default="{ row: $r }">
            <el-select v-model="$r.event_type" size="small" @change="onTypeChange($r)">
              <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="子类型" width="120">
          <template #default="{ row: $r }">
            <el-select v-if="$r.event_type !== '违纪'" v-model="$r.sub_type" size="small">
              <el-option v-for="s in subTypeMap[$r.event_type] || []" :key="s" :label="s" :value="s" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="时长(小时)" width="105">
          <template #default="{ row: $r }">
            <el-input-number
              v-if="$r.event_type !== '违纪'"
              v-model="$r.hours"
              :min="0"
              :precision="1"
              size="small"
            />
          </template>
        </el-table-column>
        <el-table-column label="分钟" width="85">
          <template #default="{ row: $r }">
            <el-input-number
              v-if="$r.sub_type === '迟到' || $r.sub_type === '早退'"
              v-model="$r.minutes"
              :min="0"
              size="small"
            />
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="130">
          <template #default="{ row: $r }">
            <el-input v-model="$r.remark" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60">
          <template #default="{ $index: idx }">
            <el-button type="danger" link size="small" @click="details.splice(idx, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:8px" @click="addEvent">+ 添加事件</el-button>
      <div class="edit-hint">保存后状态保持不变，确认操作请在方块上点击"确认"提交整日。</div>
    </template>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="saving" @click="doSave">保存（暂存）</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { saveAttendanceDetails } from '@/api/attendance'

const props = defineProps<{
  visible: boolean
  daily: any
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'saved'): void
}>()

const eventTypes = ['出勤', '休假', '加班', '违纪']
const subTypeMap: Record<string, string[]> = {
  '出勤': ['普通出勤', '补班出勤', '外勤出勤'],
  '休假': ['调休', '事假', '病假', '年假', '法定假', '福利假'],
  '加班': ['工作日加班', '节假日加班'],
  '违纪': ['缺卡', '迟到', '早退'],
}

const details = ref<any[]>([])
const punchTime = ref('')
const saving = ref(false)

watch(
  () => props.visible,
  (v) => {
    if (v && props.daily) {
      details.value = JSON.parse(JSON.stringify(props.daily.details || []))
      punchTime.value = props.daily.punch_time || ''
    }
  },
)

function handleClose() {
  emit('update:visible', false)
}

function addEvent() {
  details.value.push({ event_type: '出勤', sub_type: '普通出勤', hours: 8, minutes: 0, remark: '' })
}

function onTypeChange(row: any) {
  row.sub_type = ''
  row.hours = row.event_type === '违纪' ? 0 : 8
  row.minutes = 0
}

async function doSave() {
  saving.value = true
  try {
    await saveAttendanceDetails(props.daily.id, {
      details: details.value,
      punch_time: punchTime.value,
      remark: props.daily.remark || '',
    })
    ElMessage.success('已保存')
    handleClose()
    emit('saved')
  } catch {
    /* handled */
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.edit-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
