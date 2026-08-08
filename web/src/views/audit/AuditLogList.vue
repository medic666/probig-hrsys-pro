<template>
  <div class="page-container"><div class="page-header"><h2>审计日志</h2></div>
    <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchLogs" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
      </template>
    </ProTable>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable.vue'
import { getAuditLogs, exportAuditLogs } from '@/api/audit'
import { useExport } from '@/composables/useExport'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref()

const targetTypeNames: Record<string,string> = {
  persons:'人员', companies:'公司', position_events:'职务事件',
  attendance_daily:'考勤记录', attendance_event_details:'考勤事件',
  annual_leave_account_events:'年假权益事件', salary_events:'工资事件',
  users:'用户', roles:'角色', sys_config:'系统配置',
  permissions:'权限', user_roles:'用户角色', role_permissions:'角色权限',
  files:'文件', file_relations:'文件关联',
  attendance_calculation_monthly:'月度考勤核算', salary_summaries:'工资汇总',
  annual_leave_carryover:'年假结转批次',
}
function targetTypeName(t:string){ return targetTypeNames[t] || t }

const columns=[
  {prop:'operator_name',label:'操作人',width:'100'},{prop:'action',label:'操作类型',width:'90'},
  {prop:'target_type',label:'对象类型',width:'120'},{prop:'target_name',label:'对象名称'},
  {prop:'created_at',label:'操作时间',width:'160'},
]
const searchFields=[
  {prop:'operator_name',label:'操作人',type:'input' as const,placeholder:'模糊搜索'},
  {prop:'action',label:'操作类型',type:'select' as const,options:['新增','修改','删除','恢复','核算','确认','结转','反结账','配置修改'].map(t=>({label:t,value:t}))},
  {prop:'target_type',label:'对象类型',type:'select' as const,options:Object.entries(targetTypeNames).map(([v,l])=>({label:l,value:v}))},
  {prop:'date',label:'时间范围',type:'date-range' as const,startKey:'date_start',endKey:'date_end'},
]
const actions=[{key:'export',label:'导出',type:'default' as const, permission: PERM.auditExport}]

async function fetchLogs(p:any){
  const d=(await getAuditLogs(p)) as any
  return { list:(d.list||[]).map((r:any)=>({...r,target_type:targetTypeName(r.target_type)})), total:d.total||0 }
}

const { run: handleExport } = useExport(exportAuditLogs, () => tableRef.value?.getSearchParams() || {})
async function handleAction(k:string){ if(k==='export'){ handleExport() } }
function showDetail(r:any){ router.push(`/audit-logs/${r.id}`) }
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
