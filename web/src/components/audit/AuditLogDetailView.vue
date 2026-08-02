<template>
  <div v-loading="loading">
    <el-descriptions v-if="detail" :column="2" border size="small">
      <el-descriptions-item label="操作人">{{ detail.operator_name }}</el-descriptions-item>
      <el-descriptions-item label="操作类型">{{ detail.action }}</el-descriptions-item>
      <el-descriptions-item label="对象类型">{{ detail.target_type }}</el-descriptions-item>
      <el-descriptions-item label="对象名称">{{ detail.target_name }}</el-descriptions-item>
      <el-descriptions-item label="对象ID">{{ detail.target_id }}</el-descriptions-item>
      <el-descriptions-item label="IP">{{ detail.ip || '-' }}</el-descriptions-item>
      <el-descriptions-item label="操作时间" :span="2">{{ formatDateTime(detail.created_at) }}</el-descriptions-item>
    </el-descriptions>
    <el-empty v-else-if="!loading" description="记录不存在" :image-size="60" />

    <template v-if="detail">
      <el-divider content-position="left">操作前快照</el-divider>
      <el-descriptions v-if="beforeRows.length" :column="1" border size="small">
        <el-descriptions-item v-for="(v, k) in beforeRows" :key="k" :label="k">{{ v }}</el-descriptions-item>
      </el-descriptions>
      <p v-else style="color:#909399">(无)</p>

      <el-divider content-position="left">操作后快照</el-divider>
      <el-descriptions v-if="afterRows.length" :column="1" border size="small">
        <el-descriptions-item v-for="(v, k) in afterRows" :key="k" :label="k">{{ v }}</el-descriptions-item>
      </el-descriptions>
      <p v-else style="color:#909399">(无)</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getAuditLogDetail } from '@/api/audit'
import { formatDateTime } from '@/utils'

const props = defineProps<{ id: number }>()
const loading = ref(false)
const detail = ref<any>(null)

const techFields = ['id', 'ID', 'created_at', 'updated_at', 'deleted_at', 'last_calc_at', 'seq', 'DeletedAt', 'CreatedAt', 'UpdatedAt', 'LastCalcAt', 'path', 'Path']

function parseSnapshot(raw: string | null): Record<string, any> {
  if (!raw) return {}
  try {
    const obj = JSON.parse(raw)
    const result: Record<string, any> = {}
    for (const k of Object.keys(obj)) {
      if (techFields.includes(k) || k.startsWith('_') || k.endsWith('_at')) continue
      const v = obj[k]
      if (v !== null && v !== undefined && v !== '') result[k] = typeof v === 'object' ? JSON.stringify(v) : String(v)
    }
    return result
  } catch {
    return {}
  }
}

const beforeRows = computed(() => parseSnapshot(detail.value?.before_snapshot))
const afterRows = computed(() => parseSnapshot(detail.value?.after_snapshot))

onMounted(async () => {
  loading.value = true
  try {
    detail.value = (await getAuditLogDetail(props.id)) as any
  } catch {
    detail.value = null
  } finally {
    loading.value = false
  }
})
</script>
