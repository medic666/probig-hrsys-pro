<template>
  <div class="page-container"><div class="page-header"><h2>系统配置</h2></div>
    <el-table :data="configList" border>
      <el-table-column prop="key" label="配置键" width="220"/>
      <el-table-column prop="name" label="名称" width="160"/>
      <el-table-column label="值">
        <template #default="{row}">
          <el-input v-if="row.value_type==='string'" v-model="row.value" size="small" @change="(v:any)=>save(row.key,v)"/>
          <el-input-number v-else-if="row.value_type==='number'" v-model="row.value" size="small" @change="(v:any)=>save(row.key,v)"/>
          <el-switch v-else-if="row.value_type==='bool'" v-model="row.value" @change="(v:any)=>save(row.key,v)"/>
          <el-select v-else-if="row.value_type==='select'" v-model="row.value" size="small" @change="(v:any)=>save(row.key,v)">
            <el-option v-for="o in row.options" :key="o" :label="o" :value="o"/>
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="desc" label="说明"/>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

interface ConfigItem { key:string;name:string;value:string;desc:string;value_type:string;options?:string[] }

const configList = ref<ConfigItem[]>([])

onMounted(async()=>{
  const raw=(await request.get('/system-configs')) as Record<string,string>
  const info:Record<string,Record<string,string>>={
    'system.page_size':{name:'分页默认条数',desc:'列表分页默认每页条数',type:'number'},
    'system.export_max':{name:'导出最大条数',desc:'导出Excel最大记录条数',type:'number'},
    'system.upload_max_size':{name:'文件上传大小限制(MB)',desc:'单个文件上传最大大小',type:'number'},
    'system.upload_path':{name:'文件存储根路径',desc:'上传文件存储目录',type:'string'},
    'system.work_hours_per_day':{name:'计薪小时基准',desc:'每日标准计薪小时数',type:'number'},
    'attendance.min_leave_unit':{name:'最小请假单位',desc:'请假最小时间单位(小时)',type:'number'},
    'attendance.sick_leave_ratio':{name:'病假系数',desc:'病假折算记出勤工时的系数',type:'number'},
    'attendance.overtime_workday_ratio':{name:'工作日加班系数',desc:'工作日加班工资倍数',type:'number'},
    'attendance.overtime_holiday_ratio':{name:'节假日加班系数',desc:'节假日加班工资倍数',type:'number'},
    'attendance.full_attendance_bonus':{name:'全勤奖日标准',desc:'全勤奖每日标准金额',type:'number'},
    'attendance.high_temp_months':{name:'高温补贴发放月份',desc:'高温补贴发放月份列表',type:'select',opts:'06,07,08,09'},
    'annual_leave.yearly_hours':{name:'年假年度额度',desc:'每年标准年假小时数',type:'number'},
    'annual_leave.cycle_type':{name:'年假周期规则',desc:'年假周期计算规则',type:'select',opts:'entry_anniversary,calendar_year'},
  }
  const list:ConfigItem[]=[]
  for(const k of Object.keys(info)){
    const i=info[k]
    list.push({key:k,name:i.name||k,desc:i.desc||'',value:raw[k]||'',value_type:i.type||'string',options:i.opts?i.opts.split(','):undefined})
  }
  configList.value=list
})

async function save(key:string,value:any){
  try{await request.put(`/system-configs/${key}`,{value:String(value)});ElMessage.success('保存成功')}catch{/* */}
}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
