<template>
  <BusinessPage>
    <template v-if="isCreate || editMode">
      <AttendanceDailyForm :id="dailyId" @saved="onSaved" @cancel="onCancel" />
    </template>
    <template v-else>
      <div v-if="!viewOnly" class="toolbar">
        <el-button v-permission="PERM.attendanceEventWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <div v-loading="loading">
        <template v-if="detail">
          <AppDescriptions :column="2" border size="small">
            <el-descriptions-item label="人员">{{ detail.person_name }}</el-descriptions-item>
            <el-descriptions-item label="日期">{{ detail.event_date }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ detail.status === 'pending' ? '待确认' : '已确认' }}
            </el-descriptions-item>
            <el-descriptions-item label="打卡时间">{{ detail.punch_time || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
          </AppDescriptions>
          <el-divider content-position="left">事件明细</el-divider>
          <el-table :data="detail.details || []" border size="small">
            <el-table-column prop="event_type" label="类型" width="110" />
            <el-table-column prop="sub_type" label="子类型" width="130" />
            <el-table-column label="时长(天)" width="110">
              <template #default="{ row }">{{ row.event_type === '违纪' ? '-' : hoursToDays(row.hours || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="分钟" width="90">
              <template #default="{ row }">{{ row.minutes || '-' }}</template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" />
          </el-table>
        </template>
        <el-empty v-else-if="!loading" description="记录不存在" :image-size="60" />
      </div>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import AttendanceDailyForm from '@/components/attendance/AttendanceDailyForm.vue'
import { getAttendanceEvent } from '@/api/attendance'
import { hoursToDays } from '@/utils'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

// /attendance-events/create → 新增；/attendance-events/:id → 默认查看态（?edit=1 直达编辑态）。
// 查看态：只读展示（描述 + 明细纯文本表格，不置灰）；
// ?readonly=1（日记工时模块进入）或无权编辑 → 恒查看态，无编辑入口。
const route = useRoute()
const { id: dailyId, isCreate, goBack } = useBusinessPage()
const { viewOnly, editMode, enterEdit, exitEdit } = usePageEdit(
  PERM.attendanceEventWrite,
  () => route.query.readonly === '1',
)
const loading = ref(false)
const detail = ref<any>(null)


onMounted(async () => {
  if (dailyId.value == null) return
  loading.value = true
  try {
    detail.value = (await getAttendanceEvent(dailyId.value)) as any
  } catch {
    detail.value = null
  } finally {
    loading.value = false
  }
})

async function onSaved() {
  if (isCreate.value) {
    goBack()
    return
  }
  exitEdit()
  try {
    detail.value = (await getAttendanceEvent(dailyId.value!)) as any
  } catch {
    detail.value = null
  }
}

function onCancel() {
  if (isCreate.value) goBack()
  else exitEdit()
}
</script>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
</style>
