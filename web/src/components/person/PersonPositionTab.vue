<template>
  <div>
    <div v-if="positionLoading" v-loading="positionLoading" style="min-height:100px" />
    <template v-else>
      <h4>当前职务</h4>
      <AppDescriptions v-if="currentPosition" :column="3" border size="small">
        <el-descriptions-item label="在职状态">{{ currentPosition.is_active ? '在职' : '已离职' }}</el-descriptions-item>
        <el-descriptions-item label="入职日期">{{ currentPosition.entry_date || '-' }}</el-descriptions-item>
        <el-descriptions-item label="离职日期">{{ currentPosition.leave_date || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公司组">{{ currentPosition.company_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ currentPosition.department || '-' }}</el-descriptions-item>
        <el-descriptions-item label="职位">{{ currentPosition.position || '-' }}</el-descriptions-item>
        <el-descriptions-item label="考勤组">{{ currentPosition.attendance_group || '-' }}</el-descriptions-item>
        <el-descriptions-item label="享有年假">{{ currentPosition.has_annual_leave ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="享有全勤奖">{{ currentPosition.has_attendance_bonus ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="基本工资">{{ currentPosition.base_salary || '-' }}</el-descriptions-item>
        <el-descriptions-item label="绩效工资基数">{{ currentPosition.performance_salary || '-' }}</el-descriptions-item>
        <el-descriptions-item label="计薪天数">{{ currentPosition.salary_days || '-' }}</el-descriptions-item>
        <el-descriptions-item label="职位津贴">{{ currentPosition.post_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="餐补">{{ currentPosition.meal_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="房补">{{ currentPosition.housing_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="交通补贴">{{ currentPosition.transport_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="高温补贴">{{ currentPosition.high_temp_allowance || '-' }}</el-descriptions-item>
        <el-descriptions-item label="保险补偿">{{ currentPosition.insurance_compensation || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公积金补偿">{{ currentPosition.fund_compensation || '-' }}</el-descriptions-item>
        <el-descriptions-item label="社保代扣">{{ currentPosition.social_security_deduct || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公积金代扣">{{ currentPosition.housing_fund_deduct || '-' }}</el-descriptions-item>
      </AppDescriptions>
      <el-empty v-else description="暂无职务信息" :image-size="40" />

      <h4 class="sub-title">变动历史（职务事件）</h4>
      <el-table :data="positionEvents" border size="small">
        <el-table-column prop="event_type" label="事件类型" width="110" />
        <el-table-column prop="effective_date" label="生效日期" width="120" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button v-permission="PERM.positionEventWrite" type="primary" link size="small" @click="router.push(`/position-events/${row.id}?edit=1`)">编辑</el-button>
            <el-button v-permission="PERM.positionEventWrite" type="danger" link size="small" @click="removePositionEvent(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDescriptions from '@/components/AppDescriptions.vue'
import { getCurrentPosition } from '@/api/position-snapshot'
import { getPositionEvents, deletePositionEvent } from '@/api/position-event'
import { PERM } from '@/constants/permission'

// 人员详情「职务信息」子视图：独立组件、数据自加载。
// 由父页面以 v-if（权限门控）+ lazy（激活才渲染）挂载——无权限或未激活时绝不发起请求。
const props = defineProps<{ personId: number }>()

const router = useRouter()
const positionLoading = ref(false)
const currentPosition = ref<any>(null)
const positionEvents = ref<any[]>([])

async function loadPosition() {
  positionLoading.value = true
  try {
    currentPosition.value = (await getCurrentPosition(props.personId)) as any
  } catch {
    currentPosition.value = null
  }
  try {
    const d = (await getPositionEvents({ person_id: props.personId, pageNum: 1, pageSize: 100 })) as any
    positionEvents.value = d.list || []
  } catch {
    positionEvents.value = []
  }
  positionLoading.value = false
}

async function removePositionEvent(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除该职务事件（${row.event_type} ${row.effective_date}）？删除后将重建职务快照。`, '提示', { type: 'warning' })
  } catch { return }
  try {
    await deletePositionEvent(row.id)
    ElMessage.success('删除成功')
    loadPosition()
  } catch { /* handled */ }
}

onMounted(loadPosition)
</script>

<style scoped>
.sub-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 14px 0 8px;
}
</style>
