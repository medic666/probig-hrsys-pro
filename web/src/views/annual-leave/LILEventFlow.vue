<template>
  <div class="page-container"><div class="page-header"><h2>调休事件流水</h2></div>
    <ProTable :columns="columns" :fetch-api="fetchEvents" :search-fields="searchFields" />
  </div>
</template>

<script setup lang="ts">
import ProTable from '@/components/ProTable.vue'
import { getLILEvents } from '@/api/annual-leave'
import { getAllPersons } from '@/api/person'

const columns = [
  { prop:'id', label:'ID', width:'60' },{ prop:'person_name', label:'人员', width:'80' },
  { prop:'sub_type', label:'类型', width:'100' },{ prop:'event_date', label:'日期', width:'110' },
  { prop:'hours', label:'时长(小时)', width:'90' },{ prop:'remark', label:'备注' },
]
const searchFields = [
  { prop:'person_id', label:'人员', type:'person-select' as const, fetchApi: fetchOpts },
]

async function fetchOpts(k?:string){ const l=await getAllPersons() as any[]; return k?l.filter(p=>p.name.includes(k)):l }
async function fetchEvents(p:any){
  const d=(await getLILEvents(p)) as any
  const persons=(await getAllPersons()) as any[]||[]
  const nm:Record<number,string>={}; persons.forEach((x:any)=>nm[x.id]=x.name)
  return { list:(d.list||[]).map((r:any)=>({...r,person_name:nm[r.person_id]||'-'})), total:d.total||0 }
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
