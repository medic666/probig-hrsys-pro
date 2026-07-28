<template>
  <div class="page-container"><div class="page-header"><h2>审计日志</h2></div>
    <ProTable :columns="columns" :fetch-api="fetchLogs" :search-fields="searchFields">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dv" title="操作详情" width="800px">
      <el-row v-if="detail" :gutter="16">
        <el-col :span="12"><h4>操作前</h4><pre class="json-block">{{ detail.before_snapshot || '(无)' }}</pre></el-col>
        <el-col :span="12"><h4>操作后</h4><pre class="json-block">{{ detail.after_snapshot || '(无)' }}</pre></el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import { getAuditLogs, getAuditLogDetail } from '@/api/audit'

const dv=ref(false), detail=ref<any>(null)

const columns=[
  {prop:'operator_name',label:'操作人',width:'100'},{prop:'action',label:'操作类型',width:'80'},
  {prop:'target_type',label:'对象类型',width:'120'},{prop:'target_name',label:'对象名称'},
  {prop:'created_at',label:'操作时间',width:'160'},
]
const searchFields=[
  {prop:'operator_name',label:'操作人',type:'input' as const,placeholder:'模糊搜索'},
  {prop:'action',label:'操作类型',type:'select' as const,options:['新增','修改','删除','恢复'].map(t=>({label:t,value:t}))},
]

async function fetchLogs(p:any){return (await getAuditLogs(p)) as any}
async function showDetail(r:any){detail.value=(await getAuditLogDetail(r.id)) as any;dv.value=true}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}
.json-block{background:#f5f7fa;padding:12px;border-radius:4px;max-height:400px;overflow:auto;font-size:12px;white-space:pre-wrap;word-break:break-all}</style>
