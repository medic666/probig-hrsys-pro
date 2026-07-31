<template>
  <div class="page-container"><div class="page-header"><h2>调休余额查询</h2></div>
    <ProTable :columns="columns" :fetch-api="fetchData" :search-fields="searchFields">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">明细</el-button>
      </template>
    </ProTable>
    <el-dialog v-model="dv" title="额度明细" width="400px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="累计补班">{{ detail.makeup }}</el-descriptions-item>
        <el-descriptions-item label="累计调休">{{ detail.lil }}</el-descriptions-item>
        <el-descriptions-item label="当前可用">{{ detail.balance }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import { getLILEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'
import request from '@/utils/request'
import { hoursToDays } from '@/utils'

const dv=ref(false); const detail=ref<any>(null)

const columns=[
  {prop:'person_name',label:'人员',width:'100'},
  {prop:'balance_hours',label:'当前额度(天)',width:'120',formatter:(r:any)=>hoursToDays(r.balance_hours).toFixed(2)},
]
const searchFields=[{prop:'person_id',label:'人员',type:'person-select' as const,fetchApi:fetchOpts}]

async function fetchOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchData(p:any){
  return (await request.get('/lil-balances',{params:p})) as any
}
async function showDetail(r:any){
  try{
    const data = (await request.get(`/persons/${r.person_id}/lil-balance-detail`)) as any
    detail.value = data; dv.value = true
  }catch{/* */}
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
