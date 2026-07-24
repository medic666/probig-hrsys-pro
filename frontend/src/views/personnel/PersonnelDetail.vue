<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/personnel')" content="人员详情" style="margin-bottom: 16px" />

    <el-tabs v-model="activeTab">
      <el-tab-pane label="当前信息" name="info">
        <el-descriptions v-if="snapshot" :column="2" border>
          <el-descriptions-item label="姓名">{{ snapshot.name }}</el-descriptions-item>
          <el-descriptions-item label="考勤组">{{ snapshot.attendance_group }}</el-descriptions-item>
          <el-descriptions-item label="入职日期">{{ snapshot.hire_date || '-' }}</el-descriptions-item>
          <el-descriptions-item label="生效日期">{{ snapshot.effective_date?.slice(0, 10) }}</el-descriptions-item>
          <el-descriptions-item label="基本工资">{{ snapshot.base_salary.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="绩效工资">{{ snapshot.performance_salary.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="计薪天数">{{ snapshot.pay_days }}</el-descriptions-item>
          <el-descriptions-item label="职位津贴">{{ snapshot.position_allowance.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="餐补">{{ snapshot.meal_subsidy.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="房补">{{ snapshot.housing_subsidy.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="交通补贴">{{ snapshot.transport_subsidy.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="高温补贴">{{ snapshot.heat_subsidy.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="保险补偿">{{ snapshot.insurance_compensation.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="公积金补偿">{{ snapshot.housing_fund_compensation.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="社保代扣">{{ snapshot.social_insurance_deduct.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="公积金代扣">{{ snapshot.housing_fund_deduct.toFixed(2) }}</el-descriptions-item>
        </el-descriptions>

        <h4 style="margin-top: 16px">扩展信息</h4>
        <el-descriptions v-if="snapshot" :column="2" border>
          <el-descriptions-item v-for="(v, k) in extEntries" :key="k" :label="extLabel(k as string)">{{ displayExtValue(v) }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <el-tab-pane label="事件记录" name="events">
        <PersonnelHistory :entity-id="entityId" />
      </el-tab-pane>

      <el-tab-pane label="历史快照" name="history">
        <el-timeline v-if="history.length">
          <el-timeline-item
            v-for="snap in history"
            :key="snap.id"
            :timestamp="snap.effective_date?.slice(0, 10)"
            placement="top"
          >
            <el-card>
              <p><strong>姓名:</strong> {{ snap.name }}</p>
              <p><strong>基本工资:</strong> {{ snap.base_salary.toFixed(2) }} | <strong>计薪天数:</strong> {{ snap.pay_days }}</p>
              <p><strong>考勤组:</strong> {{ snap.attendance_group || '-' }}</p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无历史快照" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPersonnel, getPersonnelHistory } from '../../api/personnel'
import type { PersonnelSnapshot } from '../../types'
import PersonnelHistory from './PersonnelHistory.vue'

const route = useRoute()
const entityId = Number(route.params.id)
const loading = ref(false)
const snapshot = ref<PersonnelSnapshot | null>(null)
const history = ref<PersonnelSnapshot[]>([])
const activeTab = ref('info')

const extLabelMap: Record<string, string> = {
  id_card: '身份证号',
  gender: '性别',
  birthday: '生日',
  phones: '电话',
  emails: '电子邮箱',
  ethnicity: '民族',
  native_place: '籍贯',
  address: '住址',
  political_status: '政治面貌',
  marital_status: '婚姻状态',
  bank_accounts: '银行卡号',
  alias: '别名',
}

const extEntries = computed(() => {
  const ext = snapshot.value?.extended_info
  if (!ext) return []
  return Object.entries(ext)
})

function extLabel(k: string) {
  return extLabelMap[k] || k
}

function displayExtValue(v: any) {
  if (v === null || v === undefined || v === '') return '-'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

onMounted(async () => {
  loading.value = true
  try {
    const [r1, r2] = await Promise.all([getPersonnel(entityId), getPersonnelHistory(entityId)])
    snapshot.value = r1.data
    history.value = r2.data
  } catch {} finally {
    loading.value = false
  }
})
</script>
