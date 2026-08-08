<template>
  <BusinessPage>
    <template v-if="isCreate || editMode">
      <PositionEventForm :event="editEvent" @saved="onSaved" @cancel="onCancel" />
    </template>
    <template v-else>
      <div class="toolbar">
        <el-button v-permission="PERM.positionEventWrite" type="primary" size="small" @click="enterEdit">编辑</el-button>
      </div>
      <div v-loading="loading">
        <template v-if="detail">
          <AppDescriptions :column="2" border size="small">
            <el-descriptions-item label="事件类型">{{ detail.event_type }}</el-descriptions-item>
            <el-descriptions-item label="生效日期">{{ detail.effective_date }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.entry_date" label="入职日期">{{ detail.entry_date }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.leave_date" label="离职日期">{{ detail.leave_date }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
          </AppDescriptions>

          <template v-if="detail.event_type === '入职' || detail.event_type === '调薪调岗'">
            <el-divider content-position="left">岗位/薪资信息</el-divider>
            <el-table :data="fieldRows" border size="small">
              <el-table-column prop="label" label="字段" width="180" />
              <el-table-column prop="value" label="值" />
            </el-table>
          </template>
        </template>
        <el-empty v-else-if="!loading" description="事件不存在" :image-size="60" />
      </div>
    </template>
  </BusinessPage>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import BusinessPage from '@/components/BusinessPage.vue'
import AppDescriptions from '@/components/AppDescriptions.vue'
import PositionEventForm from '@/components/position/PositionEventForm.vue'
import { adjustFieldGroups } from '@/constants/position-event'
import { getPositionEvent } from '@/api/position-event'
import { getCompanyOptions } from '@/api/reference'
import { useBusinessPage } from '@/composables/useBusinessPage'
import { usePageEdit } from '@/composables/usePageEdit'
import { PERM } from '@/constants/permission'

// /position-events/create → 新增；/position-events/:id → 默认查看态（?edit=1 直达编辑态）。
const { id, isCreate, goBack } = useBusinessPage()
const editEvent = computed(() => (id.value != null ? { id: id.value } : null))

const { editMode, enterEdit, exitEdit } = usePageEdit(PERM.positionEventWrite)
const loading = ref(false)
const detail = ref<any>(null)
const companyNames = ref<Record<number, string>>({})

// 查看态字段行：非空值字段 → {label, value}（company 映射名称、bool 是/否）
const fieldRows = computed(() => {
  if (!detail.value) return []
  const rows: { label: string; value: string }[] = []
  for (const g of adjustFieldGroups) {
    for (const f of g.fields) {
      const v = detail.value[f.key]
      if (v === null || v === undefined || v === '') continue
      let text = String(v)
      if (f.type === 'company') text = companyNames.value[v as number] || `公司#${v}`
      if (f.type === 'bool') text = v ? '是' : '否'
      rows.push({ label: f.label, value: text })
    }
  }
  return rows
})

onMounted(async () => {
  if (id.value == null) return
  loading.value = true
  try {
    const opts = (await getCompanyOptions()) as { id: number; name: string }[] || []
    for (const c of opts) companyNames.value[c.id] = c.name
    detail.value = (await getPositionEvent(id.value)) as any
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
  // 编辑保存 → 回到查看态并刷新数据
  exitEdit()
  try {
    detail.value = (await getPositionEvent(id.value!)) as any
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
