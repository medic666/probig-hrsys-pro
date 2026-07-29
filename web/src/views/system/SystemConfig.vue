<template>
  <div class="page-container"><div class="page-header"><h2>系统配置</h2></div>
    <el-table :data="configList" border>
      <el-table-column prop="key" label="配置键" width="240"/>
      <el-table-column prop="name" label="名称" width="160"/>
      <el-table-column label="值">
        <template #default="{row}">
          <el-input v-if="row.value_type==='string'" v-model="row.draft" size="small"/>
          <el-input-number v-else-if="row.value_type==='number'" v-model="row.draft" size="small" :precision="2"/>
          <el-checkbox-group v-else-if="row.value_type==='select'" v-model="row.draftArr" size="small">
            <el-checkbox v-for="o in row.options" :key="o" :label="o" :value="o"/>
          </el-checkbox-group>
        </template>
      </el-table-column>
      <el-table-column prop="desc" label="说明"/>
    </el-table>
    <div style="margin-top:16px;display:flex;justify-content:flex-end">
      <el-button type="primary" @click="saveAll">保存配置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

interface ConfigItem { key:string;name:string;value:string;desc:string;value_type:string;options?:string[];draft:any;draftArr?:string[] }

const configList = ref<ConfigItem[]>([])

onMounted(async()=>{
  const raw=(await request.get('/system-configs')) as Record<string,string>
  const info:Record<string,Record<string,string>>={
    'system.work_hours_per_day':{name:'计薪小时基准',desc:'每日标准计薪小时数',type:'number'},
    'attendance.sick_leave_ratio':{name:'病假系数',desc:'病假折算记出勤工时的系数',type:'number'},
    'attendance.overtime_workday_ratio':{name:'工作日加班系数',desc:'工作日加班工资倍数',type:'number'},
    'attendance.overtime_holiday_ratio':{name:'节假日加班系数',desc:'节假日加班工资倍数',type:'number'},
    'attendance.full_attendance_bonus':{name:'全勤奖日标准',desc:'全勤奖每日标准金额',type:'number'},
    'attendance.high_temp_months':{name:'高温补贴发放月份',desc:'高温补贴发放月份列表',type:'select',opts:'06,07,08,09'},
    'annual_leave.yearly_hours':{name:'年假年度额度',desc:'每年标准年假小时数',type:'number'},
  }
  const list:ConfigItem[]=[]
  for(const k of Object.keys(info)){
    const i=info[k]
    const opts=i.opts?i.opts.split(','):undefined
    let draft=raw[k]||''
    if(i.type==='select'&&opts){try{draft=JSON.parse(draft)}catch{draft=[]}}else if(!isNaN(Number(draft))){draft=Number(draft)}
    list.push({key:k,name:i.name||k,desc:i.desc||'',value:raw[k]||'',value_type:i.type||'string',options:opts,draft,draftArr:draft as any})
  }
  configList.value=list
})

async function saveAll(){
  let count=0
  for(const item of configList.value){
    let newVal=item.draft
    if(item.value_type==='select') newVal=JSON.stringify(item.draftArr||[])
    else newVal=String(item.draft)
    if(newVal!==item.value){
      try{await request.put(`/system-configs/${item.key}`,{value:newVal});item.value=newVal;count++}catch{/* */}
    }
  }
  ElMessage.success(count>0?`已保存 ${count} 项配置`:'无变更')
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
