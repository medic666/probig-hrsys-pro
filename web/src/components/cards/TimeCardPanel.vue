<template>
  <div class="time-card-panel">
    <div v-if="level !== 'cards'" class="panel-nav">
      <PageBackButton v-if="level === 'periods'" @click="backToCards" />
      <PageBackButton v-else-if="level === 'days'" @click="backToPeriods" />
      <span class="nav-title">{{ navTitle }}</span>
    </div>

    <PersonScopeSwitch v-if="level === 'cards'" v-model="scope" />

    <div v-loading="loading" style="min-height:120px">
      <template v-if="level === 'cards'">
        <div class="panel-grid">
          <PersonCard
            v-for="p in visiblePersonCards"
            :key="p.id"
            :person="p"
            :dot-color="personDotMap?.[p.id] || ''"
            :badge-position="badgePosition"
            @click="openPerson"
          >
            <template #badge>
              <slot name="person-badge" :person="p" />
            </template>
          </PersonCard>
        </div>
        <el-empty v-if="!loading && visiblePersonCards.length === 0" description="暂无数据" :image-size="60" />
      </template>

      <template v-else-if="level === 'periods'">
        <PeriodCard :periods="periodStats" :aggregate="aggregate" :title="periodTitle" @select="openPeriod" />
        <div v-loading="loadingPeriods" style="min-height:60px" />
      </template>

      <template v-else-if="level === 'days' && hasDayLevel">
        <div class="panel-grid">
          <slot v-for="g in dayGroups" :key="g.date" name="day" :date="g.date" :items="g.items" />
        </div>
        <el-empty v-if="!loading && dayGroups.length === 0" description="该时段暂无数据" :image-size="60" />
      </template>

      <template v-else-if="level === 'days' && !hasDayLevel">
        <slot name="period-list" :items="periodItems" :period="selectedPeriod" />
        <el-empty v-if="!loading && periodItems.length === 0" description="该时段暂无数据" :image-size="60" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PersonCard from '@/components/cards/PersonCard.vue'
import PeriodCard from '@/components/cards/PeriodCard.vue'
import PersonScopeSwitch from '@/components/cards/PersonScopeSwitch.vue'
import PageBackButton from '@/components/PageBackButton.vue'
import { getPersonCards } from '@/api/person'
import { usePeriodStats, type PeriodAggregate } from '@/composables/usePeriodStats'
import { filterPersons, type PersonScope } from '@/utils/personScope'

// 时间卡片面板：人员卡片 → 时段聚合（月度/年度同构）→ 原子卡片（业务页 slot 渲染）。
// 卡片层可佩戴模块徽章（personDotMap：personId → 颜色点，仅活跃人员显示）；
// 时段窗口 month=[入职月..明年+1] / year=[入职年..明年]，均由人员入职日期裁剪。
const props = withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    dateField?: string
    monthField?: string
    hasDayLevel?: boolean
    statusField?: string
    pendingValues?: string[]
    aggregate?: PeriodAggregate
    personDotMap?: Record<number, string>
    badgePosition?: 'name' | 'meta'
    // URL 驱动模式：卡片→时段→日期层级状态同步到路由 query，返回/刷新后层级可恢复
    urlDriven?: boolean
    // 时段点击跳转的业务逻辑页路由生成器（优先级高于内部日期层），如 `/module/:personId/:month`
    detailRoute?: (person: { id: number; name: string }, period: string) => string
  }>(),
  {
    dateField: '',
    monthField: '',
    hasDayLevel: true,
    statusField: 'status',
    pendingValues: () => [],
    aggregate: 'month',
    personDotMap: undefined,
    badgePosition: 'name',
    urlDriven: false,
    detailRoute: undefined,
  },
)

const level = ref<'cards' | 'periods' | 'days'>('cards')
const personCards = ref<any[]>([])
const scope = ref<PersonScope>('active')
const visiblePersonCards = computed(() => filterPersons(personCards.value, scope.value))
const personId = ref<number | null>(null)
const personName = ref('')
const selectedPeriod = ref('')
const dayGroups = ref<{ date: string; items: any[] }[]>([])
const periodItems = ref<any[]>([])
const loading = ref(false)
const loadingPeriods = ref(false)

const periodStatsLoader = usePeriodStats({
  personId: 0,
  fetchFn: props.fetchFn,
  dateField: props.dateField,
  monthField: props.monthField,
  statusField: props.statusField,
  pendingValues: props.pendingValues,
  aggregate: props.aggregate,
})
const periodStats = periodStatsLoader.periods

const periodTitle = computed(() =>
  props.aggregate === 'year' ? `${personName.value} 的年度概览` : `${personName.value} 的最近 12 个月`,
)

const navTitle = computed(() => {
  if (level.value === 'periods') return periodTitle.value
  return `${personName.value} / ${selectedPeriod.value}`
})

const route = useRoute()
const router = useRouter()

// syncUrl 将当前层级状态合并写入路由 query（urlDriven 模式，保留其它视图的休眠状态）。
// 默认 push：层级钻取成为独立历史条目，浏览器返回可逐级回退；
// detailRoute 场景紧随其后的业务页跳转会顶替本条目的历史位置，改用 replace 避免重复条目。
function syncUrl(usePush = true) {
  if (!props.urlDriven) return
  const query: Record<string, any> = { ...route.query }
  if (personId.value) {
    query.person = String(personId.value)
    query.name = personName.value
  } else {
    delete query.person
    delete query.name
  }
  const periodKey = props.aggregate === 'year' ? 'year' : 'month'
  if (selectedPeriod.value && !props.detailRoute) {
    query[periodKey] = selectedPeriod.value
  } else {
    delete query.month
    delete query.year
  }
  query.scope = scope.value
  if (usePush) {
    router.push({ query })
  } else {
    router.replace({ query })
  }
}

// 卡片范围（活跃/全部）变化同步 URL（卡片视图状态完整化）
watch(scope, () => {
  if (props.urlDriven && level.value === 'cards') {
    syncUrl()
  }
})

// URL 回退（浏览器返回/外部导航）时按 query 重置层级状态：
// 同路由 query 变化组件不会重新 mount，必须由本 watch 反向同步；
// 自身 syncUrl 写入或无关键（view/筛选）变化触发时，关键状态与 URL 一致 → 跳过。
watch(
  () => route.query,
  () => {
    if (!props.urlDriven) return
    const qp = route.query
    const qScope = qp.scope === 'all' ? 'all' : 'active'
    const rawPerson = qp.person
    const qPerson = typeof rawPerson === 'string' && /^\d+$/.test(rawPerson) ? Number(rawPerson) : null
    const periodKey = props.aggregate === 'year' ? 'year' : 'month'
    const rawPeriod = qp[periodKey]
    const qPeriod = typeof rawPeriod === 'string' ? rawPeriod : ''

    if (scope.value === qScope && personId.value === qPerson && selectedPeriod.value === qPeriod) {
      return
    }

    scope.value = qScope
    personId.value = qPerson
    personName.value = String(qp.name || '')
    selectedPeriod.value = qPeriod
    if (qPerson === null) {
      level.value = 'cards'
    } else if (!qPeriod) {
      level.value = 'periods'
      const person = personCards.value.find((p) => p.id === qPerson)
      loadPeriodStats(qPerson, entryStartOf(person))
    } else {
      level.value = 'days'
      if (props.hasDayLevel) loadDays()
      else loadPeriodItems()
    }
  },
)

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

// entryStartOf 时段窗口起点：月度取入职月（YYYY-MM），年度取入职年（YYYY）；
// 未入职（无入职日期）返回空 → 走默认窗口
function entryStartOf(person: any): string {
  if (!person?.entry_date) return ''
  const raw = String(person.entry_date)
  return props.aggregate === 'year' ? raw.slice(0, 4) : raw.slice(0, 7)
}

function openPerson(person: any) {
  personId.value = person.id
  personName.value = person.name
  level.value = 'periods'
  syncUrl()
  loadPeriodStats(person.id, entryStartOf(person))
}

function backToCards() {
  level.value = 'cards'
  personId.value = null
  personName.value = ''
  selectedPeriod.value = ''
  syncUrl()
  loadPersonCards()
}

function backToPeriods() {
  level.value = 'periods'
  selectedPeriod.value = ''
  syncUrl()
  const person = personCards.value.find((p) => p.id === personId.value)
  loadPeriodStats(personId.value || undefined, entryStartOf(person))
}

async function loadPeriodStats(personIdParam?: number, start?: string) {
  const pid = personIdParam ?? personId.value
  if (!pid) return
  loadingPeriods.value = true
  await periodStatsLoader.load(pid, start)
  loadingPeriods.value = false
}

function openPeriod(period: string) {
  selectedPeriod.value = period
  if (props.detailRoute) {
    // 时段点击直接进入业务逻辑页（URL 携带人员+时段）；
    // 本条层级记录随即被业务页顶替，用 replace 避免历史重复条目
    syncUrl(false)
    router.push(props.detailRoute({ id: personId.value!, name: personName.value }, period))
    return
  }
  level.value = 'days'
  if (props.hasDayLevel) {
    loadDays()
  } else {
    loadPeriodItems()
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

function periodRange(period: string): { start: string; end: string } {
  if (props.aggregate === 'year') {
    return { start: `${period}-01-01`, end: `${period}-12-31` }
  }
  return { start: `${period}-01`, end: `${period}-31` }
}

async function loadDays() {
  if (!personId.value || !selectedPeriod.value) return
  loading.value = true
  try {
    const range = periodRange(selectedPeriod.value)
    const params: any = {
      person_id: personId.value,
      date_start: range.start,
      date_end: range.end,
    }
    if (!props.dateField) {
      params.belong_month = selectedPeriod.value
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

async function loadPeriodItems() {
  if (!personId.value || !selectedPeriod.value) return
  loading.value = true
  try {
    const params: any = { person_id: personId.value }
    if (props.monthField) {
      params[props.monthField] = selectedPeriod.value
    }
    periodItems.value = await fetchAll(params)
  } catch {
    periodItems.value = []
  } finally {
    loading.value = false
  }
}

function reload() {
  if (level.value === 'cards') {
    loadPersonCards()
  } else if (level.value === 'periods') {
    loadPeriodStats()
  } else if (level.value === 'days') {
    if (props.hasDayLevel) loadDays()
    else loadPeriodItems()
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
      const person = personCards.value.find((p) => p.id === personId.value)
      const start = entryStartOf(person)
      const periodKey = props.aggregate === 'year' ? 'year' : 'month'
      const qPeriod = qp[periodKey]
      if (qPeriod && !props.detailRoute) {
        selectedPeriod.value = String(qPeriod)
        level.value = 'days'
        if (props.hasDayLevel) {
          loadDays()
        } else {
          loadPeriodItems()
        }
      } else {
        level.value = 'periods'
        loadPeriodStats(personId.value, start)
      }
    }
  }
})

defineExpose({ reload })
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

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
    @include card-grid;
  }
}
</style>
