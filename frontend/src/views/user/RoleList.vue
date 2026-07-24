<template>
  <div>
    <el-card>
      <div style="display: flex; justify-content: flex-end; align-items: center; margin-bottom: 16px;">
        <el-button type="primary" @click="dialogVisible = true">新增角色</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="name" label="角色名称" min-width="120" />
        <el-table-column prop="remark" label="备注" min-width="200" />
        <el-table-column label="权限" min-width="300">
          <template #default="{ row }">
            <el-tag v-for="p in row.permissions" :key="p.id" size="small" style="margin: 2px;">{{ p.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="600px" @closed="resetForm">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
        </el-form-item>
        <el-form-item label="权限">
          <el-checkbox-group v-model="form.permission_ids">
            <div v-for="m in moduleList" :key="m.module" style="margin-bottom: 8px;">
              <span style="font-weight: bold; margin-right: 8px;">{{ m.label }}:</span>
              <el-checkbox v-for="p in m.perms" :key="p.id" :value="p.id" :label="p.id">{{ p.action }}</el-checkbox>
            </div>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '../../api/user'

const loading = ref(false)
const submitting = ref(false)
const list = ref<any[]>([])
const allPermissions = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)

const defaultForm = { id: 0, name: '', remark: '', permission_ids: [] as number[] }
const form = reactive({ ...defaultForm })

const moduleList = computed(() => {
  const modules: Record<string, { module: string; label: string; perms: any[] }> = {}
  for (const p of allPermissions.value) {
    if (!modules[p.module]) {
      const labels: Record<string, string> = {
        person: '人员', company: '公司', attendance: '假勤', salary: '工资',
        file: '文件', audit: '审计', user: '用户', system: '系统',
      }
      modules[p.module] = { module: p.module, label: labels[p.module] || p.module, perms: [] }
    }
    modules[p.module].perms.push(p)
  }
  return Object.values(modules)
})

function resetForm() { Object.assign(form, defaultForm) }

function openEdit(row: any) {
  isEdit.value = true
  form.id = row.id
  form.name = row.name
  form.remark = row.remark
  form.permission_ids = (row.permissions || []).map((p: any) => p.id)
  dialogVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const [roleRes, permRes] = await Promise.all([api.getRoleList(), api.getAllPermissions()])
    list.value = roleRes.data || []
    allPermissions.value = permRes.data || []
  } finally { loading.value = false }
}

async function handleSubmit() {
  submitting.value = true
  try {
    const data = {
      name: form.name,
      remark: form.remark,
      permissions: form.permission_ids.map((id: number) => ({ id })),
    }
    if (isEdit.value) {
      await api.updateRole(form.id, data)
      ElMessage.success('修改成功')
    } else {
      await api.createRole(data)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    fetchData()
  } finally { submitting.value = false }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除角色「${row.name}」？`, '确认删除', { type: 'warning' })
  await api.deleteRole(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>
