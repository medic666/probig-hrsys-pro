<template>
  <div class="page-container">
    <div class="page-header">
      <h2>年假余额查询</h2>
      <el-radio-group v-model="viewMode" size="small" style="margin-left:16px">
        <el-radio-button value="cards">卡片</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
    </div>

    <template v-if="viewMode === 'list'">
      <ProTable ref="tableRef" :columns="columns" :fetch-api="fetchData" :search-fields="searchFields">
        <template #actions="{ row }">
          <el-button type="primary" link size="small" @click="openDetail(row.person_id, row.person_name)">明细</el-button>
        </template>
      </ProTable>
    </template>

    <template v-else>
      <CardGrid ref="cardGridRef" :fetch-fn="fetchCards">
        <template #default="{ item }">
          <PersonCard :person="item" @click="openDetail(item.person_id, item.name)">
            <template #extra>
              <div class="balance-line" :class="{ 'is-zero': !(balanceMap[item.person_id] ?? 0) }">
                年假 {{ hoursToDays(balanceMap[item.person_id] ?? 0).toFixed(2) }} 天
              </div>
            </template>
          </PersonCard>
        </template>
      </CardGrid>
    </template>

    <el-dialog v-model="detailVisible" :title="detailPersonName ? `${detailPersonName} 的假期明细` : '假期明细'" width="440px">
      <LeaveBalanceDetail :person-id="detailPersonId" :show-lil="false" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProTable from '@/components/ProTable.vue'
import CardGrid from '@/components/cards/CardGrid.vue'
import PersonCard from '@/components/cards/PersonCard.vue'
import LeaveBalanceDetail from '@/components/cards/LeaveBalanceDetail.vue'
import { getAllPersons, getPersonCards } from '@/api/person'
import request from '@/utils/request'
import { hoursToDays, formatDateTime } from '@/utils'

const tableRef = ref()
const viewMode = ref<'cards' | 'list'>('cards')
const cardGridRef = ref()
const balanceMap = ref<Record<number, number>>({})
const detailVisible = ref(false)
const detailPersonId = ref<number | null>(null)
const detailPersonName = ref('')

const columns = [
  { prop: 'person_name', label: '人员', width: '100' },
  { prop: 'balance_hours', label: '当前额度(天)', width: '120', formatter: (r: any) => hoursToDays(r.balance_hours).toFixed(2) },
  { prop: 'last_calc_at', label: '更新时间', width: '160', formatter: (r: any) => r.last_calc_at ? formatDateTime(r.last_calc_at) : '-' },
]
const searchFields = [
  { prop: 'person_id', label: '人员', type: 'person-select' as const, fetchApi: fetchPersonOpts },
]

async function fetchPersonOpts(k?: string) { const l = await getAllPersons() as any[]; return k ? l.filter(p => p.name.includes(k)) : l }
async function fetchData(p: any) {
  return (await request.get('/annual-leave-balances', { params: p })) as any
}

async function fetchCards() {
  const cards = (await getPersonCards()) as any[] || []
  const d = (await request.get('/annual-leave-balances', { params: { pageNum: 1, pageSize: 100 } })) as any
  const map: Record<number, number> = {}
  for (const row of d?.list || []) {
    map[row.person_id] = row.balance_hours ?? 0
  }
  balanceMap.value = map
  return { list: cards, total: cards.length }
}

function openDetail(personId: number, personName: string) {
  detailPersonId.value = personId
  detailPersonName.value = personName || ''
  detailVisible.value = true
}
</script>
<style scoped>
.page-container { padding: 0; background: transparent; }
.page-header { margin-bottom: 16px; display: flex; align-items: center; h2 { font-size: 18px; font-weight: 600; color: #303133; } }
.balance-line { font-size: 13px; color: #409eff; font-weight: 600; line-height: 20px; }
.balance-line.is-zero { color: #c0c4cc; font-weight: 400; }
</style>
