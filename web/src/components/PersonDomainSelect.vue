<template>
  <div class="person-domain-select">
    <el-select v-model="domainType" size="small" style="width:110px" @change="onDomainTypeChange">
      <el-option v-for="d in domainOptions" :key="d.value" :label="d.label" :value="d.value" />
    </el-select>
    <el-select
      v-if="domainType !== 'all'"
      v-model="domainValue"
      size="small"
      :placeholder="domainPlaceholder"
      clearable
      style="width:130px"
      @change="clearSelectionIfOutOfDomain"
    >
      <el-option v-for="v in domainValues" :key="v.value" :label="v.label" :value="v.value" />
    </el-select>
    <el-checkbox
      v-if="domainPersonCount > 0"
      :model-value="allDomainSelected"
      size="small"
      @change="toggleDomainSelectAll"
    >
      全选该域人员
    </el-checkbox>
    <el-select
      v-model="selectedIds"
      multiple
      filterable
      clearable
      collapse-tags
      collapse-tags-tooltip
      :max-collapse-tags="2"
      size="small"
      :placeholder="placeholder || '选择人员'"
      class="person-select"
      @update:model-value="handleChange"
    >
      <el-option v-for="p in visibleOptions" :key="p.id" :label="p.name" :value="p.id" />
    </el-select>
    <span class="sel-count">已选 {{ selectedIds.length }} 人</span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getAllPersons, type PersonOption } from '@/api/person'

// 多域人员多选组件：在全体人员/公司/考勤组/在职状态域内筛选后多选人员。
// - 域值选定后可一键「全选该域人员」（并集合并，已全选时再点取消该域全部）；
// - 三框统一 small 尺寸保持对齐；人员框弹性宽度 + 折叠标签（collapse-tags-tooltip），
//   选多人恒单行，悬停 tooltip 查看完整名单。
// 人员与域值均来自同一数据源（默认 /persons/all），域维度由选项字段推导，
// 与 NameSelect（单选）互补，供批量录入、批量核算等多人选择场景统一复用。
const props = withDefaults(
  defineProps<{
    modelValue: number[]
    fetchApi?: (keyword?: string) => Promise<PersonOption[]>
    placeholder?: string
  }>(),
  { fetchApi: getAllPersons, placeholder: '' },
)

const emit = defineEmits<{ (e: 'update:modelValue', v: number[]): void }>()

const domainType = ref<'all' | 'company' | 'group' | 'status'>('all')
const domainValue = ref<string | number | null>(null)
const selectedIds = ref<number[]>(props.modelValue || [])
const options = ref<PersonOption[]>([])

const domainOptions = [
  { label: '全体人员', value: 'all' },
  { label: '按公司', value: 'company' },
  { label: '按考勤组', value: 'group' },
  { label: '按在职状态', value: 'status' },
]

const domainPlaceholder = computed(() =>
  ({ company: '选择公司', group: '选择考勤组', status: '选择状态' })[domainType.value] || '',
)

const domainValues = computed<{ label: string; value: string | number }[]>(() => {
  switch (domainType.value) {
    case 'company': {
      const seen = new Set<number>()
      const list: { label: string; value: number }[] = []
      for (const p of options.value) {
        if (p.company_id && !seen.has(p.company_id)) {
          seen.add(p.company_id)
          list.push({ label: p.company_name || `公司#${p.company_id}`, value: p.company_id })
        }
      }
      return list
    }
    case 'group': {
      const seen = new Set<string>()
      const list: { label: string; value: string }[] = []
      for (const p of options.value) {
        if (p.attendance_group && !seen.has(p.attendance_group)) {
          seen.add(p.attendance_group)
          list.push({ label: p.attendance_group, value: p.attendance_group })
        }
      }
      return list
    }
    case 'status': {
      return [
        { label: '在职', value: 'active' },
        { label: '已离职', value: 'left' },
        { label: '未入职', value: 'not_entered' },
      ]
    }
  }
  return []
})

function personStatus(p: PersonOption): string {
  if (p.entry_date == null && !p.is_active) return 'not_entered'
  return p.is_active ? 'active' : 'left'
}

const visibleOptions = computed(() => {
  switch (domainType.value) {
    case 'all':
      return options.value
    case 'company':
      return options.value.filter((p) => p.company_id === domainValue.value)
    case 'group':
      return options.value.filter((p) => p.attendance_group === domainValue.value)
    case 'status':
      return options.value.filter((p) => personStatus(p) === domainValue.value)
  }
  return options.value
})

// 当前域内人员 id 集合与数量（全选勾选可见性/状态判定）
const visibleIds = computed(() => new Set(visibleOptions.value.map((p) => p.id)))
const domainPersonCount = computed(() => visibleOptions.value.length)
const allDomainSelected = computed(() =>
  domainPersonCount.value > 0 && visibleOptions.value.every((p) => selectedIds.value.includes(p.id)),
)

// 一键全选该域人员：勾选 = 并集合并（不清空已有选择）；已全选时再点 = 移除该域全部人员
function toggleDomainSelectAll() {
  if (allDomainSelected.value) {
    handleChange(selectedIds.value.filter((id) => !visibleIds.value.has(id)))
  } else {
    handleChange(Array.from(new Set([...selectedIds.value, ...visibleIds.value])))
  }
}

function handleChange(v: number[]) {
  selectedIds.value = v
  emit('update:modelValue', v)
}

function onDomainTypeChange() {
  domainValue.value = null
}

// 域变更后清理不在当前域内的已选项，避免"已选但不可见"的困惑
function clearSelectionIfOutOfDomain() {
  if (domainValue.value == null) return
  const visible = new Set(visibleOptions.value.map((p) => p.id))
  const kept = selectedIds.value.filter((id) => visible.has(id))
  if (kept.length !== selectedIds.value.length) {
    handleChange(kept)
  }
}

watch(
  () => props.modelValue,
  (v) => {
    selectedIds.value = v || []
  },
)

watch(domainType, () => {
  if (domainValue.value != null) clearSelectionIfOutOfDomain()
})

async function loadOptions() {
  try {
    options.value = (await props.fetchApi()) || []
  } catch {
    options.value = []
  }
}

loadOptions()
</script>

<style lang="scss" scoped>
.person-domain-select {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;

  // 人员多选框：弹性占满剩余宽度，标签折叠保持单行
  .person-select {
    flex: 1;
    min-width: 200px;
  }

  .sel-count {
    font-size: 12px;
    color: #909399;
    white-space: nowrap;
  }
}
</style>
