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
import { getPersonALBalance, getAnnualLeaveEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'

const tableRef = ref(); const dv = ref(false); const detail = ref<any>(null)

const columns = [
  { prop:'person_name', label:'人员', width:'100' },
  { prop:'balance_hours', label:'当前额度(小时)', width:'120' },
  { prop:'last_calc_at', label:'更新时间', width:'110' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchPersonOpts },
]

async function fetchPersonOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchData(_p:any){
  const persons=(await getAllPersons()) as any[]||[]; const nm:Record<number,string>={}; persons.forEach((x:any)=>nm[x.id]=x.name)
  const list:any[]=[]
  for(const ps of persons){
    try{ const b=(await getPersonALBalance(ps.id)) as any; list.push({...b,person_name:ps.name}) }catch{ /* */ }
  }
  return {list, total:list.length}
}

async function showDetail(r:any){
  try{
    const events=(await getAnnualLeaveEvents({person_id:r.person_id})) as any
    let grant=0, adjust=0, carryover=0
    for(const e of (events.list||[])){
      if(e.event_type==='grant') grant+=e.hours
      else if(e.event_type==='adjust') adjust+=e.hours
      else if(e.event_type==='carryover_deduct') carryover+=e.hours
    }
    const consumed = (grant + adjust - carryover) - r.balance_hours
    detail.value={grant, consumed, adjust, carryover, balance: r.balance_hours}; dv.value=true
  }catch{ /* */ }
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
