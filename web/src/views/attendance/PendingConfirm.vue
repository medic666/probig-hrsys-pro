<template>
  <div class="page-container"><div class="page-header"><h2>待确认考勤事件</h2></div>
    <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchPending" :search-fields="searchFields" :auto-load="true">
      <template #actions="{ row }">
        <el-button v-permission="PERM.attendanceEventWrite" type="primary" link size="small" @click="editConfirm(row)">编辑确认</el-button>
      </template>
    </ProTable>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable.vue'
import { getPendingDailies } from '@/api/attendance'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref()

const columns=[{prop:'person_name',label:'人员',width:'80'},{prop:'event_date',label:'日期',width:'110'},{prop:'punch_time',label:'打卡时间',width:'110'},{prop:'status',label:'状态',width:'80',slot:'status'}]

async function fetchPending(p:any){
  return (await getPendingDailies(p)) as any
}

// 编辑确认 = 进入该日考勤页（保存即确认，置为已确认）
function editConfirm(row: any) {
  router.push(`/attendance-events/${row.id}`)
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
