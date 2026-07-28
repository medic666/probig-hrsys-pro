<template>
  <div class="page-container"><div class="page-header"><h2>文件管理</h2></div>
    <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchFiles" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="downloadFile(row)">下载</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR"/>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import { getFiles, deleteFile, restoreFile, getDeletedFiles, downloadFile, uploadFile } from '@/api/file'

const tableRef=ref(), tv=ref(false)

const columns=[
  {prop:'id',label:'ID',width:'60'},{prop:'original_name',label:'文件名',minWidth:'180'},
  {prop:'size',label:'大小',width:'100',formatter:(r:any)=>r.size?((r.size/1024/1024).toFixed(2)+' MB'):'-'},
  {prop:'mime_type',label:'类型',width:'120'},{prop:'created_at',label:'上传时间',width:'160'},
]
const searchFields=[{prop:'name',label:'文件名',type:'input' as const,placeholder:'模糊搜索'}]
const actions=[{key:'upload',label:'上传文件',type:'primary' as const},{key:'trash',label:'回收站',type:'default' as const}]
const tc=[{prop:'id',label:'ID',width:'60'},{prop:'original_name',label:'文件名'}]

async function fetchFiles(p:any){return (await getFiles(p)) as any}
async function fd(p:any){return (await getDeletedFiles(p)) as any}
async function rst(id:number){return restoreFile(id)}

function handleAction(k:string){
  if(k==='upload'){
    const input=document.createElement('input');input.type='file'
    input.onchange=async()=>{if(input.files?.[0]){try{await uploadFile(input.files[0]);ElMessage.success('上传成功');tableRef.value?.refresh()}catch{/* */}}}
    input.click()
  }else if(k==='trash') tv.value=true
}
async function handleDelete(r:any){try{await ElMessageBox.confirm('确认?','提示',{type:'warning'})}catch{return};try{await deleteFile(r.id);ElMessage.success('已删除');tableRef.value?.refresh()}catch{ /* */ }}
function onR(){tableRef.value?.refresh()}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
