<template>
  <div class="page-container"><div class="page-header"><h2>年假周年结转</h2></div>
    <el-button type="primary" style="margin-bottom:12px" @click="doCarryover">批量结转</el-button>
    <el-dialog v-model="cv" title="批量结转" width="400px">
      <el-form><el-form-item label="目标月份"><el-date-picker v-model="cm" type="month" value-format="YYYY-MM" style="width:100%"/></el-form-item></el-form>
      <template #footer><el-button @click="cv=false">取消</el-button><el-button type="primary" :loading="s" @click="exec">开始结转</el-button></template>
    </el-dialog>

    <el-table v-loading="bl" :data="batches" border>
      <el-table-column prop="batch_no" label="批次号" width="200"/>
      <el-table-column prop="business_period" label="业务周期" width="100"/>
      <el-table-column prop="person_names" label="处理人员" />
      <el-table-column prop="total_count" label="人数" width="70"/>
      <el-table-column prop="status" label="状态" width="90"><template #default="{row}">{{ {2:'已生效',3:'已冲销',4:'失败'}[row.status]||row.status }}</template></el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="160"/>
      <el-table-column label="操作" width="140">
        <template #default="{row}">
          <el-button type="primary" link size="small" @click="showEvents(row)">详情</el-button>
          <el-button v-if="row.status===2" type="danger" link size="small" @click="cancel(row)">反结账</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="ev" title="结转详情" width="500px">
      <el-table :data="batchEvents" border size="small">
        <el-table-column prop="person_name" label="人员" />
        <el-table-column prop="event_type" label="类型" width="120">
          <template #default="{row}">{{ {grant:'配发',carryover_deduct:'抵扣'}[row.event_type]||row.event_type }}</template>
        </el-table-column>
        <el-table-column prop="hours" label="时长(小时)" width="100" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { executeCarryover, cancelCarryover, getCarryoverBatches } from '@/api/annual-leave'
import request from '@/utils/request'

const cv=ref(false); const cm=ref(''); const s=ref(false); const batches=ref<any[]>([]); const bl=ref(false)
const ev=ref(false); const batchEvents=ref<any[]>([])

onMounted(()=>loadBatches())
async function loadBatches(){ bl.value=true; try{
  const data=(await getCarryoverBatches()) as any[]||[]
  for(const b of data){
    const evts=(await request.get(`/annual-leave-carryover/batches/${b.id}/events`)) as any[]||[]
    const names=(evts as any[]||[]).map((e:any)=>e.person_name).filter((n:string,i:number,a:string[])=>n&&a.indexOf(n)===i)
    b.person_names=names.join(', ')
  }
  batches.value=data
} catch{ /* */ } finally{ bl.value=false } }

function doCarryover(){ cm.value=''; cv.value=true }
async function exec(){
  if(!cm.value){ ElMessage.warning('请选择月份'); return }
  s.value=true
  try{ const d=await executeCarryover(cm.value) as any; ElMessage.success(`成功${d.success}人, 失败${d.fail}人`); cv.value=false; loadBatches() } catch{ /* */ } finally{ s.value=false }
}
async function cancel(r:any){
  try{ await ElMessageBox.confirm('确认反结账?','提示',{type:'warning'}) } catch{ return }
  try{ await cancelCarryover(r.id); ElMessage.success('已冲销'); loadBatches() } catch{ /* */ }
}
async function showEvents(r:any){
  batchEvents.value=(await request.get(`/annual-leave-carryover/batches/${r.id}/events`)) as any[]||[]
  ev.value=true
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
