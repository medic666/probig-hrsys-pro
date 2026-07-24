<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="showDialog()">新增角色</el-button>
        </div>
      </template>

      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column prop="name" label="角色名称" width="150" />
        <el-table-column prop="remark" label="备注" min-width="200" />
        <el-table-column label="权限" min-width="300">
          <template #default="{ row }">{{ row.permissions?.map((p: any) => p.name).join('、') }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showDialog(row)">编辑</el-button>
            <el-button size="small" @click="showPermDialog(row)">分配权限</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
              <template #reference><el-button size="small" type="danger">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="400px">
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules">
        <el-form-item label="名称" prop="name"><el-input v-model="roleForm.name" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="roleForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permDialog" title="分配权限" width="600px">
      <el-checkbox-group v-model="selectedPerms">
        <div v-for="group in permGroups" :key="group.module" style="margin-bottom:12px">
          <div style="font-weight:bold;margin-bottom:4px">{{ group.name }}</div>
          <el-checkbox v-for="p in group.perms" :key="p.id" :label="p.id" :value="p.id">{{ p.action_name }}</el-checkbox>
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="permDialog = false">取消</el-button>
        <el-button type="primary" @click="submitPerms">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const list = ref([]); const loading = ref(false)

const dialogVisible = ref(false); const isEdit = ref(false); const editId = ref(0); const roleFormRef = ref()
const roleForm = reactive({ name: '', remark: '' })
const roleRules = { name: [{ required: true }] }

const permDialog = ref(false); const currentRoleId = ref(0)
const selectedPerms = ref<number[]>([]); const allPerms = ref<any[]>([])

const actionLabels: Record<string, string> = { read: '查看', write: '编辑', delete: '删除', export: '导出' }
const moduleNames: Record<string, string> = { person: '人员管理', company: '公司管理', attendance: '假勤管理', salary: '工资管理', file: '文件管理', audit: '审计日志', system: '系统配置', user: '用户管理' }

const permGroups = computed(() => {
  const groups: Record<string, any> = {}
  for (const p of allPerms.value) {
    if (!groups[p.module]) groups[p.module] = { module: p.module, name: moduleNames[p.module] || p.module, perms: [] }
    groups[p.module].perms.push({ ...p, action_name: actionLabels[p.action] || p.action })
  }
  return Object.values(groups)
})

async function fetchList() {
  loading.value = true
  const res = await request.get('/roles', { params: { page_size: 100 } })
  list.value = res.data.list; loading.value = false
}

function showDialog(row?: any) {
  if (row) { isEdit.value = true; editId.value = row.id; roleForm.name = row.name; roleForm.remark = row.remark }
  else { isEdit.value = false; editId.value = 0; roleForm.name = ''; roleForm.remark = '' }
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await roleFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (isEdit.value) { await request.put(`/roles/${editId.value}`, roleForm) }
  else { await request.post('/roles', roleForm) }
  dialogVisible.value = false; fetchList()
}

async function handleDelete(id: number) { await request.delete(`/roles/${id}`); fetchList() }

async function showPermDialog(row: any) {
  currentRoleId.value = row.id
  const [roleRes, permRes] = await Promise.all([request.get(`/roles/${row.id}`), request.get('/permissions')])
  allPerms.value = permRes.data
  selectedPerms.value = (roleRes.data.permissions || []).map((p: any) => p.id)
  permDialog.value = true
}

async function submitPerms() {
  await request.put(`/roles/${currentRoleId.value}/permissions`, { permission_ids: selectedPerms.value })
  permDialog.value = false; fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
