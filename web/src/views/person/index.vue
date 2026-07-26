<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, View, RefreshLeft } from '@element-plus/icons-vue'
import ProTable from '@/components/ProTable.vue'
import ProFormDialog from '@/components/ProFormDialog.vue'
import NameSelect from '@/components/NameSelect.vue'
import {
  listPersons, createPerson, updatePerson, deletePerson, restorePerson,
  listTrash
} from '@/api/person'
import type { Person, PersonListParams } from '@/api/person'
import { useRouter } from 'vue-router'

const router = useRouter()

const tableRef = ref()
const formVisible = ref(false)
const formTitle = ref('新增人员')
const editingRow = ref<Person | null>(null)
const trashVisible = ref(false)

const searchFields = [
  { prop: 'name', label: '姓名', type: 'input' as const, placeholder: '请输入姓名' },
  { prop: 'id_card', label: '身份证号', type: 'input' as const, placeholder: '请输入身份证号' },
  { prop: 'attendance_group', label: '考勤组', type: 'input' as const, placeholder: '请输入考勤组' },
  { prop: 'status', label: '在职状态', type: 'select' as const, options: [{ label: '在职', value: 'active' }, { label: '离职', value: 'inactive' }] }
]

const columns = [
  { prop: 'name', label: '姓名' },
  { prop: 'id_card', label: '身份证号' },
  { prop: 'attendance_group', label: '考勤组' },
  { prop: 'status', label: '在职状态' },
  { slot: 'actions', label: '操作', width: 280, fixed: 'right' as const }
]

const formFields = [
  { prop: 'name', label: '姓名', type: 'input' as const, required: true },
  { prop: 'id_card', label: '身份证号', type: 'input' as const, required: true },
  { prop: 'gender', label: '性别', type: 'select' as const, options: [{ label: '男', value: 1 }, { label: '女', value: 2 }] },
  { prop: 'birthday', label: '生日', type: 'date' as const },
  { prop: 'nation', label: '民族', type: 'input' as const },
  { prop: 'native_place', label: '籍贯', type: 'input' as const },
  { prop: 'address', label: '住址', type: 'textarea' as const },
  { prop: 'political_status', label: '政治面貌', type: 'input' as const },
  { prop: 'marital_status', label: '婚姻状态', type: 'select' as const, options: [{ label: '未婚', value: 1 }, { label: '已婚', value: 2 }, { label: '离异', value: 3 }, { label: '丧偶', value: 4 }] },
  { prop: 'alias', label: '别名', type: 'input' as const }
]

async function fetchPersonList(params: Record<string, unknown>) {
  return listPersons(params as unknown as PersonListParams)
}

async function fetchTrashList(params: Record<string, unknown>) {
  return listTrash({ pageNum: params.pageNum as number, pageSize: params.pageSize as number, name: params.name as string })
}

function handleAdd() {
  editingRow.value = null
  formTitle.value = '新增人员'
  formVisible.value = true
}

function handleEdit(row: Person) {
  editingRow.value = row
  formTitle.value = '编辑人员'
  formVisible.value = true
}

function getInitialData() {
  if (!editingRow.value) return {}
  return { ...editingRow.value }
}

async function handleSubmit(data: Record<string, unknown>) {
  if (editingRow.value) {
    await updatePerson(data as any)
  } else {
    await createPerson(data as any)
  }
}

async function handleDelete(row: Person) {
  try {
    await ElMessageBox.confirm(`确定要删除人员「${row.name}」吗？`, '确认删除', { type: 'warning' })
  } catch {
    return
  }
  await deletePerson(row.id)
  ElMessage.success('删除成功')
  tableRef.value?.refresh()
}

async function handleRestore(row: Person) {
  try {
    await ElMessageBox.confirm(`确定要恢复人员「${row.name}」吗？`, '确认恢复', { type: 'warning' })
  } catch {
    return
  }
  await restorePerson(row.id)
  ElMessage.success('恢复成功')
  tableRef.value?.refresh()
}

function handleViewDetail(row: Person) {
  ElMessage.info('详情页面开发中')
}

function handleJumpToPosition(row: Person) {
  router.push({ name: 'Position', query: { person_id: row.id } })
}

function handleJumpToAttendance(row: Person) {
  router.push({ name: 'AttendanceEvent', query: { person_id: row.id } })
}

function handleJumpToSalary(row: Person) {
  router.push({ name: 'SalaryEvent', query: { person_id: row.id } })
}

function handleFormSuccess() {
  tableRef.value?.refresh()
}

const trashColumns = [
  { prop: 'name', label: '姓名' },
  { prop: 'id_card', label: '身份证号' },
  { slot: 'trashActions', label: '操作', width: 120, fixed: 'right' as const }
]
</script>

<template>
  <div class="page-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增人员</el-button>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Delete" @click="trashVisible = true">回收站</el-button>
      </div>
    </div>

    <ProTable
      ref="tableRef"
      :columns="columns"
      :search-fields="searchFields"
      :api="fetchPersonList"
    >
      <template #actions="{ row }">
        <el-button type="primary" link :icon="View" @click="handleViewDetail(row)">详情</el-button>
        <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
        <el-button type="primary" link @click="handleJumpToPosition(row)">职务</el-button>
        <el-button type="primary" link @click="handleJumpToAttendance(row)">考勤</el-button>
        <el-button type="primary" link @click="handleJumpToSalary(row)">工资</el-button>
        <el-button type="danger" link :icon="Delete" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <ProFormDialog
      v-model:visible="formVisible"
      :title="formTitle"
      :form-fields="formFields"
      :initial-data="getInitialData()"
      :submit-api="handleSubmit"
      @success="handleFormSuccess"
    />

    <el-drawer v-model="trashVisible" title="回收站" size="800px">
      <ProTable
        :columns="trashColumns"
        :search-fields="searchFields.slice(0, 1)"
        :api="fetchTrashList"
      >
        <template #trashActions="{ row }">
          <el-button type="primary" link :icon="RefreshLeft" @click="handleRestore(row)">恢复</el-button>
        </template>
      </ProTable>
    </el-drawer>
  </div>
</template>
