<template>
  <div class="page-container"><div class="page-header"><h2>待确认考勤事件</h2></div>
    <ProTable :columns="columns" :fetch-api="fetchPending" :search-fields="searchFields" :auto-load="true">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="editRow=row;details=JSON.parse(JSON.stringify(row.details||[]));dialogVisible=true">编辑确认</el-button>
      </template>
    </ProTable>
    <el-dialog v-model="dialogVisible" title="编辑并确认" width="600px">
      <el-descriptions v-if="editRow" :column="2" border size="small">
        <el-descriptions-item label="人员">{{ editRow.person_name }}</el-descriptions-item>
        <el-descriptions-item label="日期">{{ editRow.event_date }}</el-descriptions-item>
        <el-descriptions-item label="打卡时间">{{ editRow.punch_time || '-' }}</el-descriptions-item>
      </el-descriptions>
      <el-table :data="details" border size="small" style="margin-top:12px">
        <el-table-column label="类型" width="80">
          <template #default="{row:$r, $index:idx}">
            <el-select v-model="$r.event_type" size="small" @change="v=>detailChanged(idx,'event_type',v)">
              <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t"/>
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="子类型" width="110">
          <template #default="{row:$r, $index:idx}">
            <el-select v-if="$r.event_type!=='打卡时间戳'" v-model="$r.sub_type" size="small" @change="v=>detailChanged(idx,'sub_type',v)">
              <el-option v-for="s in subTypeMap[$r.event_type]||[]" :key="s" :label="s" :value="s"/>
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="时长(小时)" width="100">
          <template #default="{row:$r, $index:idx}">
            <el-input-number v-if="$r.event_type!=='打卡时间戳'&&$r.event_type!=='违纪'" v-model="$r.hours" :min="0" :precision="1" size="small" @change="v=>detailChanged(idx,'hours',v)"/>
          </template>
        </el-table-column>
        <el-table-column label="分钟" width="80">
          <template #default="{row:$r, $index:idx}">
            <el-input-number v-if="$r.sub_type==='迟到'||$r.sub_type==='早退'" v-model="$r.minutes" :min="0" size="small" @change="v=>detailChanged(idx,'minutes',v)"/>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="120">
          <template #default="{row:$r, $index:idx}">
            <el-input v-model="$r.remark" size="small" @change="v=>detailChanged(idx,'remark',v)"/>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60">
          <template #default="{ $index:idx }">
            <el-button type="danger" link size="small" @click="details.splice(idx,1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button size="small" style="margin-top:8px" @click="details.push({event_type:'出勤',sub_type:'普通出勤',hours:8,minutes:0,remark:''})">+ 添加事件</el-button>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doConfirm">确认并提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import { getPendingDailies, confirmPendingDaily } from '@/api/attendance'
import { getAllPersons } from '@/api/person'

const dialogVisible=ref(false), saving=ref(false), editRow=ref<any>(null), details=ref<any[]>([])
const eventTypes=['出勤','休假','加班','违纪','打卡时间戳']
const subTypeMap:Record<string,string[]>={'出勤':['普通出勤','补班出勤','外勤出勤'],'休假':['调休','事假','病假','年假','法定假','福利假'],'加班':['工作日加班','节假日加班'],'违纪':['缺卡','迟到','早退']}
const columns=[{prop:'person_name',label:'人员',width:'80'},{prop:'event_date',label:'日期',width:'110'},{prop:'punch_time',label:'打卡时间',width:'110'},{prop:'status',label:'状态',width:'80',slot:'status'}]
const searchFields=[{prop:'person_id',label:'人员',type:'person-select' as const,fetchApi: fetchPersonOpts}]

async function fetchPersonOpts(k?:string){const l=await getAllPersons() as any[];return k?l.filter(p=>p.name.includes(k)):l}
async function fetchPending(p:any){
  return (await getPendingDailies(p)) as any
}
function detailChanged(_idx:number, _field:string, _val:any){/* reactive binding */}

async function doConfirm(){
  if(!editRow.value) return
  saving.value=true
  try{
    await confirmPendingDaily(editRow.value.id, details.value)
    ElMessage.success('确认成功');dialogVisible.value=false
    window.location.reload()
  }catch{/* */}finally{saving.value=false}
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
