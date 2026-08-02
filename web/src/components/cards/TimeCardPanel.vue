<template>
  <div class="time-card-panel">
    <div v-if="level !== 'cards'" class="panel-nav">
      <PageBackButton v-if="level === 'months'" @click="backToCards" />
      <PageBackButton v-else-if="level === 'days'" @click="backToMonths" />
      <span class="nav-title">{{ level === 'months' ? `${personName} 的最近 12 个月` : `${personName} / ${selectedMonth}` }}</span>
    </div>

    <PersonScopeSwitch v-if="level === 'cards'" v-model="scope" />

    <div v-loading="loading" style="min-height:120px">
      <template v-if="level === 'cards'">
        <div class="panel-grid">
          <PersonCard v-for="p in visiblePersonCards" :key="p.id" :person="p" @click="openPerson" />
        </div>
        <el-empty v-if="!loading && visiblePersonCards.length === 0" description="暂无数据" :image-size="60" />
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
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PersonCard from '@/components/cards/PersonCard.vue'
import MonthCard from '@/components/cards/MonthCard.vue'
import PersonScopeSwitch from '@/components/cards/PersonScopeSwitch.vue'
import PageBackButton from '@/components/PageBackButton.vue'
import { getPersonCards } from '@/api/person'
import { useMonthStats } from '@/composables/useMonthStats'
import { filterPersons, type PersonScope } from '@/utils/personScope'

const props = withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    dateField?: string
    monthField?: string
    hasDayLevel?: boolean
    statusField?: string
    pendingValues?: string[]
    // URL 驱动模式：卡片→月份→日期层级状态同步到路由 query（?person=&name=&month=），
    // 返回/刷新后层级可恢复；后续模块逐个启用即完成推广
    urlDriven?: boolean
    // 月份点击跳转的业务逻辑页路由生成器（优先级高于内部日期层），如 `/module/:personId/:month`
    detailRoute?: (person: { id: number; name: string }, month: string) => string
  }>(),
  {
    dateField: '',
    monthField: '',
    hasDayLevel: true,
    statusField: 'status',
    pendingValues: () => [],
    urlDriven: false,
    detailRoute: undefined,
  },
)

const level = ref<'cards' | 'months' | 'days'>('cards')
const personCards = ref<any[]>([])
const scope = ref<PersonScope>('active')
const visiblePersonCards = computed(() => filterPersons(personCards.value, scope.value))
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

const route = useRoute()
const router = useRouter()

// syncUrl 将当前层级状态合并写入路由 query（urlDriven 模式，保留其它视图的休眠状态）
function syncUrl() {
  if (!props.urlDriven) return
  const query: Record<string, any> = { ...route.query }
  if (personId.value) {
    query.person = String(personId.value)
    query.name = personName.value
  } else {
    delete query.person
    delete query.name
  }
  if (selectedMonth.value && !props.detailRoute) {
    query.month = selectedMonth.value
  } else {
    delete query.month
  }
  query.scope = scope.value
  router.replace({ query })
}

// 卡片范围（活跃/全部）变化同步 URL（卡片视图状态完整化）
watch(scope, () => {
  if (props.urlDriven && level.value === 'cards') {
    syncUrl()
  }
})

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
  syncUrl()
  loadMonthStats()
}

function backToCards() {
  level.value = 'cards'
  personId.value = null
  personName.value = ''
  selectedMonth.value = ''
  syncUrl()
  loadPersonCards()
}

function backToMonths() {
  level.value = 'months'
  selectedMonth.value = ''
  syncUrl()
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
  if (props.detailRoute) {
    // 月份点击直接进入业务逻辑页（URL 携带人员+月份）
    syncUrl()
    router.push(props.detailRoute({ id: personId.value!, name: personName.value }, month))
    return
  }
  level.value = 'days'
  if (props.hasDayLevel) {
    loadDays()
  } else {
    loadMonthItems()
  }
  syncUrl()
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
  if (props.urlDriven) {
    // URL 恢复层级：?person=5&month=2026-06 → 直接进入对应层级
    const qp = route.query
    if (qp.scope === 'all' || qp.scope === 'active') {
      scope.value = qp.scope
    }
    if (qp.person) {
      personId.value = Number(qp.person)
      personName.value = String(qp.name || '')
      if (qp.month && !props.detailRoute) {
        selectedMonth.value = String(qp.month)
        level.value = 'days'
        if (props.hasDayLevel) {
          loadDays()
        } else {
          loadMonthItems()
        }
      } else {
        level.value = 'months'
        loadMonthStats()
      }
    }
  }
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
