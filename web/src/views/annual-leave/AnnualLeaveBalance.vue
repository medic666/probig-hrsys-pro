<template>
  <div class="page-container"><div class="page-header"><h2>年假余额查询</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchData" :search-fields="searchFields">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">明细</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dv" title="额度明细" width="420px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="累计配发">{{ detail.grant }}</el-descriptions-item>
        <el-descriptions-item label="累计已休">{{ detail.consumed }}</el-descriptions-item>
        <el-descriptions-item label="累计人工调整">{{ detail.adjust }}</el-descriptions-item>
        <el-descriptions-item label="累计结转扣除">{{ detail.carryover }}</el-descriptions-item>
        <el-descriptions-item label="当前可用">{{ detail.balance }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import { getAnnualLeaveEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'
import request from '@/utils/request'
import { hoursToDays, formatDateTime } from '@/utils'

const tableRef=ref(); const dv=ref(false); const detail=ref<any>(null)

const columns=[
  {prop:'person_name',label:'人员',width:'100'},
  {prop:'balance_hours',label:'当前额度(天)',width:'120',formatter:(r:any)=>hoursToDays(r.balance_hours).toFixed(2)},
  {prop:'last_calc_at',label:'更新时间',width:'160',formatter:(r:any)=>formatDateTime(r.last_calc_at)},
]
const searchFields=[
  {prop:'person_id',label:'人员',type:'person-select' as const,fetchApi:fetchPersonOpts},
]

async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchData(p:any){
  return (await request.get('/annual-leave-balances',{params:p})) as any
}

async function showDetail(r:any){
  try{
    const data = (await request.get(`/persons/${r.person_id}/annual-leave-balance-detail`)) as any
    detail.value = data; dv.value = true
  }catch{/* */}
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
