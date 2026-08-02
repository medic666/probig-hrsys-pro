<template>
  <div class="time-card-panel">
    <div class="panel-nav">
      <template v-if="level === 'months'">
        <el-button link size="small" @click="backToCards">← 全部人员</el-button>
        <span class="nav-title">{{ personName }} 的最近 12 个月</span>
      </template>
      <template v-else-if="level === 'days'">
        <el-button link size="small" @click="backToMonths">← {{ personName }} 的月度</el-button>
        <span class="nav-title">{{ personName }} / {{ selectedMonth }}</span>
      </template>
      <template v-else>
        <span class="nav-title">人员卡片</span>
      </template>
    </div>

    <div v-loading="loading" style="min-height:120px">
      <template v-if="level === 'cards'">
        <div class="panel-grid">
          <PersonCard v-for="p in personCards" :key="p.id" :person="p" @click="openPerson" />
        </div>
        <el-empty v-if="!loading && personCards.length === 0" description="暂无数据" :image-size="60" />
      </template>

      <template v-else-if="level === 'months'">
        <MonthCard :months="monthStats" title="月度概览" @select="openMonth" />
        <div v-loading="loadingMonths" style="min-height:60px" />
      </template>

      <template v-else-if="level === 'days' && hasDayLevel">
        <div class="panel-grid">
          <slot v-for="g in dayGroups" :key="g.date" name="day" :date="g.date" :items="g.items" />
        </div>
        <el-empty v-if="!loading && dayGroups.length === 0" description="该月暂无数据" :image-size="60" />
      </template>

      <template v-else-if="level === 'days' && !hasDayLevel">
        <slot name="month-list" :items="monthItems" :month="selectedMonth" />
        <el-empty v-if="!loading && monthItems.length === 0" description="该月暂无数据" :image-size="60" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PersonCard from '@/components/cards/PersonCard.vue'
import MonthCard from '@/components/cards/MonthCard.vue'
import { getPersonCards } from '@/api/person'
import { useMonthStats } from '@/composables/useMonthStats'

const props = withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    dateField?: string
    monthField?: string
    hasDayLevel?: boolean
    statusField?: string
    pendingValues?: string[]
  }>(),
  {
    dateField: '',
    monthField: '',
    hasDayLevel: true,
    statusField: 'status',
    pendingValues: () => [],
  },
)

const level = ref<'cards' | 'months' | 'days'>('cards')
const personCards = ref<any[]>([])
const personId = ref<number | null>(null)
const personName = ref('')
const selectedMonth = ref('')
const dayGroups = ref<{ date: string; items: any[] }[]>([])
const monthItems = ref<any[]>([])
const loading = ref(false)
const loadingMonths = ref(false)

const monthStatsLoader = useMonthStats({
  personId: 0,
  fetchFn: props.fetchFn,
  dateField: props.dateField,
  monthField: props.monthField,
  statusField: props.statusField,
  pendingValues: props.pendingValues,
})
const monthStats = monthStatsLoader.months

async function loadPersonCards() {
  loading.value = true
  try {
    personCards.value = (await getPersonCards()) as any[] || []
  } catch {
    personCards.value = []
  } finally {
    loading.value = false
  }
}

function openPerson(person: any) {
  personId.value = person.id
  personName.value = person.name
  level.value = 'months'
  loadMonthStats()
}

function backToCards() {
  level.value = 'cards'
  loadPersonCards()
}

function backToMonths() {
  level.value = 'months'
  loadMonthStats()
}

async function loadMonthStats() {
  if (!personId.value) return
  loadingMonths.value = true
  await monthStatsLoader.load(personId.value)
  loadingMonths.value = false
}

function openMonth(month: string) {
  selectedMonth.value = month
  level.value = 'days'
  if (props.hasDayLevel) {
    loadDays()
  } else {
    loadMonthItems()
  }
}

async function fetchAll(params: any): Promise<any[]> {
  const list: any[] = []
  let page = 1
  let total = Infinity
  while ((page - 1) * 100 < total && page <= 50) {
    const d = await props.fetchFn({ ...params, pageNum: page, pageSize: 100 })
    list.push(...(d.list || []))
    total = d.total || 0
    page++
  }
  return list
}

async function loadDays() {
  if (!personId.value || !selectedMonth.value) return
  loading.value = true
  try {
    const params: any = {
      person_id: personId.value,
      date_start: `${selectedMonth.value}-01`,
      date_end: `${selectedMonth.value}-31`,
    }
    if (!props.dateField) {
      params.belong_month = selectedMonth.value
    }
    const rows = await fetchAll(params)
    const map = new Map<string, any[]>()
    for (const row of rows) {
      const raw = row[props.dateField] || ''
      const date = String(raw).slice(0, 10)
      if (!date) continue
      if (!map.has(date)) map.set(date, [])
      map.get(date)!.push(row)
    }
    dayGroups.value = Array.from(map.entries())
      .map(([date, items]) => ({ date, items }))
      .sort((a, b) => a.date.localeCompare(b.date))
  } catch {
    dayGroups.value = []
  } finally {
    loading.value = false
  }
}

async function loadMonthItems() {
  if (!personId.value || !selectedMonth.value) return
  loading.value = true
  try {
    const params: any = { person_id: personId.value }
    if (props.monthField) {
      params[props.monthField] = selectedMonth.value
    }
    monthItems.value = await fetchAll(params)
  } catch {
    monthItems.value = []
  } finally {
    loading.value = false
  }
}

function reload() {
  if (level.value === 'cards') {
    loadPersonCards()
  } else if (level.value === 'months') {
    loadMonthStats()
  } else if (level.value === 'days') {
    if (props.hasDayLevel) loadDays()
    else loadMonthItems()
  }
}

onMounted(() => {
  loadPersonCards()
})

defineExpose({ reload })
</script>

<style lang="scss" scoped>
.time-card-panel {
  .panel-nav {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;

    .nav-title {
      font-weight: 600;
      color: #303133;
    }
  }

  .panel-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
}
</style>
