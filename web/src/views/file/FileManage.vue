<template>
  <div class="page-container"><div class="page-header"><h2>文件管理</h2></div>
    <ProTable ref="tableRef" :url-driven="true" :columns="columns" :fetch-api="fetchFiles" :search-fields="searchFields" :actions="actions" @action="handleAction">
      <template #actions="{ row }">
        <el-button type="primary" link size="small" @click="showAssociations(row)">关联({{ row.assoc_count||0 }})</el-button>
        <el-button type="success" link size="small" @click="downloadFile(row.id, row.original_name)">下载</el-button>
        <el-button v-permission="PERM.fileWrite" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="assocVisible" title="文件关联" width="650px">
      <el-table :data="associations" border size="small">
        <el-table-column prop="target_name" label="关联实体" />
        <el-table-column label="操作" width="80">
          <template #default="{row:r}">
            <el-button type="primary" link size="small" @click="jumpTo(r)">跳转</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <RecycleBinDrawer v-model:visible="tv" :fetch-api="fd" :restore-api="rst" :columns="tc" @restored="onR">
      <template #footer>
        <el-button v-permission="PERM.fileWrite" type="danger" @click="handleCleanOrphans">清理孤儿文件(30天+)</el-button>
      </template>
    </RecycleBinDrawer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable.vue'
import RecycleBinDrawer from '@/components/RecycleBinDrawer.vue'
import { getFiles, deleteFile, restoreFile, getDeletedFiles, downloadFile, uploadFile, getFileAssociations, cleanOrphanFiles } from '@/api/file'
import { PERM } from '@/constants/permission'

const router = useRouter()
const tableRef=ref(), tv=ref(false), assocVisible=ref(false), associations=ref<any[]>([])

const columns=[
  {prop:'id',label:'ID',width:'60'},{prop:'original_name',label:'文件名',minWidth:'180'},
  {prop:'size',label:'大小',width:'100',formatter:(r:any)=>r.size?((r.size/1024/1024).toFixed(2)+' MB'):'-'},
  {prop:'md5',label:'MD5',width:'120'},{prop:'assoc_count',label:'关联数',width:'70'},
  {prop:'created_at',label:'上传时间',width:'160'},
]
const searchFields=[
  {prop:'name',label:'文件名',type:'input' as const,placeholder:'模糊搜索'},
  {prop:'mime_type',label:'类型',type:'select' as const,options:[
    {label:'图片',value:'image/'},{label:'PDF',value:'application/pdf'},
    {label:'Excel',value:'application/vnd.openxmlformats-officedocument.spreadsheetml'},
    {label:'Word',value:'application/vnd.openxmlformats-officedocument.wordprocessingml'},
    {label:'纯文本',value:'text/'},
  ]},
  {prop:'date',label:'上传时间',type:'date-range' as const,startKey:'date_start',endKey:'date_end'},
]
const actions=[
  {key:'upload',label:'上传文件',type:'primary' as const, permission: PERM.fileWrite},
  {key:'trash',label:'回收站',type:'default' as const, permission: PERM.fileWrite},
]
const tc=[{prop:'id',label:'ID',width:'60'},{prop:'original_name',label:'文件名'},{prop:'md5',label:'MD5',width:'120'}]

async function fetchFiles(p:any){
  const d=(await getFiles(p)) as any
  const list=(d.list||[]).map((r:any)=>{r.assoc_count=r.assoc_count??0;return r})
  for(const r of list){try{r.assoc_count=(await getFileAssociations(r.id) as any[]||[]).length}catch{/* */}}
  return {list,total:d.total||0}
}
async function fd(p:any){return (await getDeletedFiles(p)) as any}
async function rst(id:number){return restoreFile(id)}

function handleAction(k:string){
  if(k==='upload'){
    const input=document.createElement('input');input.type='file'
    input.onchange=async()=>{if(input.files?.[0]){try{const r=(await uploadFile(input.files[0])) as any;ElMessage.success(r.duplicate?'文件已存在(复用已有记录)':'上传成功');tableRef.value?.refresh()}catch{/* */}}}
    input.click()
  }else if(k==='trash') tv.value=true
}
async function handleDelete(r:any){try{await ElMessageBox.confirm('确认删除该文件？','提示',{type:'warning'})}catch{return};try{await deleteFile(r.id);ElMessage.success('已删除');tableRef.value?.refresh()}catch{/* */}}
async function showAssociations(r:any){associations.value=(await getFileAssociations(r.id)) as any[]||[];assocVisible.value=true}

const targetRouterMap:Record<string,string>={'person':'/person','company':'/company','position_event':'/position-event','attendance_event':'/attendance','annual_leave_event':'/annual-leave-events','salary_event':'/salary-events'}
function jumpTo(r:any){const p=targetRouterMap[r.target_type];if(p)router.push(p)}
async function handleCleanOrphans(){
  try{await ElMessageBox.confirm('确认清理30天前软删除且无关联的孤儿文件？','清理孤儿文件',{type:'warning'})}catch{return}
  try{const r=(await cleanOrphanFiles()) as any;ElMessage.success(r.msg||'已清理');tableRef.value?.refresh()}catch{/* */}
}
function onR(){tableRef.value?.refresh()}
</script>
<style scoped>.page-container{padding:0;background:transparent}.page-header{margin-bottom:16px}h2{font-size:18px;font-weight:600;color:#303133}</style>
