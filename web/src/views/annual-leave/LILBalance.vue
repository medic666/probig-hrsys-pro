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
import { getPersonLILBalance, getLILEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'

const dv = ref(false); const detail = ref<any>(null)

const columns = [
  { prop:'person_name', label:'人员', width:'100' },
  { prop:'balance_hours', label:'当前额度(小时)', width:'120' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchOpts },
]

async function fetchOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchData(_p:any){
  const persons=(await getAllPersons()) as any[]||[]; const list:any[]=[]
  for(const ps of persons){
    try{ const b=(await getPersonLILBalance(ps.id)) as any; list.push({...b,person_name:ps.name}) }catch{ /* */ }
  }
  return {list, total:list.length}
}
async function showDetail(r:any){
  try{
    const events=(await getLILEvents({person_id:r.person_id})) as any
    let makeup=0, lil=0
    for(const e of (events.list||[])){ if(e.sub_type==='补班出勤') makeup+=e.hours; else lil+=e.hours }
    detail.value={makeup,lil,balance:r.balance_hours}; dv.value=true
  }catch{ /* */ }
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
