<script setup lang="ts">
import { ref } from 'vue'
import { Download, View } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import { listAuditLogs, getAuditDetail } from '@/api/audit'
import type { AuditListParams, AuditLog } from '@/api/audit'

const tableRef = ref()
const detailVisible = ref(false)
const detailData = ref<AuditLog | null>(null)

const searchFields = [
  { prop: 'operator_name', label: '操作人', type: 'input' as const },
  { prop: 'action', label: '操作类型', type: 'select' as const, options: [
    { label: '新增', value: '新增' }, { label: '修改', value: '修改' },
    { label: '删除', value: '删除' }, { label: '恢复', value: '恢复' }
  ]},
  { prop: 'target_type', label: '对象类型', type: 'input' as const },
  { prop: 'created_at_start', label: '时间起', type: 'date' as const },
  { prop: 'created_at_end', label: '时间止', type: 'date' as const }
]

const columns = [
  { prop: 'operator_name', label: '操作人' },
  { prop: 'action', label: '操作类型' },
  { prop: 'target_name', label: '操作对象' },
  { prop: 'target_type', label: '对象类型' },
  { prop: 'created_at', label: '操作时间' },
  { slot: 'actions', label: '操作', width: 120, fixed: 'right' as const }
]

async function fetchList(params: Record<string, unknown>) {
  return listAuditLogs(params as unknown as AuditListParams)
}

async function handleViewDetail(row: AuditLog) {
  const detail = await getAuditDetail(row.id)
  detailData.value = detail
  detailVisible.value = true
}
</script>

<template>
  <div class="page-container">
    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchList"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="View" @click="handleViewDetail(row)">详情</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="detailVisible" title="审计日志详情" width="900px">
      <div v-if="detailData" class="audit-detail">
        <div class="detail-row">
          <span class="detail-label">操作人：</span>
          <span>{{ detailData.operator_name }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">操作类型：</span>
          <span>{{ detailData.action }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">操作对象：</span>
          <span>{{ detailData.target_name }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">操作时间：</span>
          <span>{{ detailData.created_at }}</span>
        </div>
        <div class="snapshot-container">
          <div class="snapshot-left">
            <h4>操作前</h4>
            <pre>{{ detailData.before_snapshot || '无' }}</pre>
          </div>
          <div class="snapshot-right">
            <h4>操作后</h4>
            <pre>{{ detailData.after_snapshot || '无' }}</pre>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.audit-detail {
  .detail-row {
    margin-bottom: 8px;
    .detail-label {
      font-weight: bold;
      color: #606266;
    }
  }
  .snapshot-container {
    display: flex;
    gap: 16px;
    margin-top: 16px;
    .snapshot-left,
    .snapshot-right {
      flex: 1;
      h4 {
        margin-bottom: 8px;
        color: #303133;
      }
      pre {
        background: #f5f7fa;
        padding: 12px;
        border-radius: 4px;
        max-height: 300px;
        overflow-y: auto;
        font-size: 12px;
        white-space: pre-wrap;
        word-break: break-all;
      }
    }
  }
}
</style>
