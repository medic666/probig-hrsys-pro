<template>
  <div>
    <div style="display: flex; gap: 12px; margin-bottom: 16px">
      <el-select v-model="filterTargetType" placeholder="操作对象类型" clearable style="width: 160px" @change="fetchData">
        <el-option label="人员事件" value="personnel_event" />
        <el-option label="组织事件" value="organization_event" />
        <el-option label="假勤事件" value="attendance_event" />
        <el-option label="工资事件" value="salary_event" />
        <el-option label="文件" value="file" />
      </el-select>
    </div>

    <el-table :data="logs" border v-loading="loading">
      <el-table-column prop="username" label="操作人" width="100" />
      <el-table-column prop="action" label="操作" width="70">
        <template #default="{ row }">
          <el-tag :type="actionType(row.action)" size="small">{{ actionLabel(row.action) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="target_type" label="对象类型" width="110">
        <template #default="{ row }">{{ targetLabel(row.target_type) }}</template>
      </el-table-column>
      <el-table-column prop="target_name" label="操作对象" min-width="120">
        <template #default="{ row }">
          <template v-if="row.target_name && row.target_summary">
            <div>{{ row.target_name }}</div>
            <div style="font-size: 12px; color: #909399">{{ row.target_summary }}</div>
          </template>
          <template v-else-if="row.target_name">{{ row.target_name }}</template>
          <span v-else style="color: #c0c4cc">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="操作时间" width="170">
        <template #default="{ row }">{{ row.created_at?.slice(0, 19).replace('T', ' ') }}</template>
      </el-table-column>
      <el-table-column label="变更详情" min-width="180">
        <template #default="{ row }">
          <template v-if="hasFieldChanges(row)">
            <el-popover placement="left" width="380" trigger="click">
              <template #reference>
                <el-button text size="small" type="primary">查看变更</el-button>
              </template>
              <div style="max-height: 400px; overflow: auto">
                <el-table :data="formatFieldChanges(row)" border size="small">
                  <el-table-column prop="field" label="字段" width="140" />
                  <el-table-column label="变更值" min-width="200">
                    <template #default="{ row: r }">
                      <template v-if="r.old !== undefined">
                        <span style="color: #f56c6c; text-decoration: line-through">{{ r.old }}</span>
                        <span style="margin: 0 4px">→</span>
                      </template>
                      <span style="color: #67c23a">{{ r.new }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-popover>
          </template>
          <el-button v-else text size="small" @click="showRawPayload(row)">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      style="margin-top: 16px; justify-content: flex-end"
      @current-change="fetchData"
    />

    <el-dialog v-model="payloadDialogVisible" title="操作详情" width="500px">
      <pre style="max-height: 400px; overflow: auto; font-size: 13px">{{ payloadContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listAuditLogs } from '../../api/audit'
import type { AuditLog } from '../../types'

const loading = ref(false)
const logs = ref<AuditLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const filterTargetType = ref('')
const payloadDialogVisible = ref(false)
const payloadContent = ref('')

const fieldLabelMap: Record<string, string> = {
  base_salary: '基本工资',
  performance_salary: '绩效工资',
  pay_days: '计薪天数',
  position_allowance: '职位津贴',
  meal_subsidy: '餐补',
  housing_subsidy: '房补',
  transport_subsidy: '交通补贴',
  heat_subsidy: '高温补贴',
  insurance_compensation: '保险补偿',
  housing_fund_compensation: '公积金补偿',
  social_insurance_deduct: '社保代扣',
  housing_fund_deduct: '公积金代扣',
  attendance_group: '考勤组',
  hire_date: '入职日期',
  credit_code: '信用代码',
  address: '地址',
  phone: '联系电话',
  bank_name: '开户行',
  bank_account: '银行账号',
  company_name: '组织名称',
  entity_id: '人员ID',
  amount: '金额',
  event_type: '事件类型',
  event_category: '事件类别',
  event_subtype: '事件子类',
  event_date: '事件日期',
  duration_days: '天数',
  description: '备注',
  period_start: '期间开始',
  period_end: '期间结束',
}

onMounted(() => fetchData())

async function fetchData() {
  loading.value = true
  try {
    const res = await listAuditLogs({
      page: page.value,
      page_size: pageSize,
      target_type: filterTargetType.value || undefined,
    })
    logs.value = res.data.list
    total.value = res.data.total
  } catch {} finally { loading.value = false }
}

function actionType(a: string) {
  return { create: 'success', update: 'warning', delete: 'danger' }[a] || 'info'
}

function actionLabel(a: string) {
  return { create: '创建', update: '更新', delete: '删除', calculate: '计算' }[a] || a
}

function targetLabel(t: string) {
  return {
    personnel_event: '人员事件', organization_event: '组织事件',
    attendance_event: '假勤事件', salary_event: '工资事件',
    file: '文件', file_association: '文件关联',
    attendance_summary: '假勤汇总', salary_summary: '工资汇总',
  }[t] || t
}

function hasFieldChanges(row: AuditLog): boolean {
  const p = row.payload
  if (!p) return false
  if (p.changed_fields && typeof p.changed_fields === 'object' && Object.keys(p.changed_fields).length > 0) return true
  return false
}

function formatFieldChanges(row: AuditLog): { field: string; old?: string; new: string }[] {
  const p = row.payload || {}
  const cf = p.changed_fields || {}
  const result: { field: string; old?: string; new: string }[] = []

  for (const [key, value] of Object.entries(cf as Record<string, any>)) {
    if (key === 'extended_info' && typeof value === 'object') {
      for (const [ek, ev] of Object.entries(value as Record<string, any>)) {
        result.push({
          field: fieldLabelMap['ext_' + ek] || ek,
          new: String(ev),
        })
      }
    } else {
      result.push({
        field: fieldLabelMap[key] || key,
        new: value !== undefined && value !== null ? String(value) : '-',
      })
    }
  }

  if (result.length === 0) {
    for (const [key, value] of Object.entries(p)) {
      if (['entity_id', 'name', 'company_name', 'event_type', 'effective_date', 'event_name'].includes(key)) continue
      result.push({
        field: fieldLabelMap[key] || key,
        new: value !== undefined && value !== null ? String(value) : '-',
      })
    }
  }

  return result
}

function showRawPayload(row: AuditLog) {
  payloadContent.value = JSON.stringify(row.payload, null, 2)
  payloadDialogVisible.value = true
}
</script>
