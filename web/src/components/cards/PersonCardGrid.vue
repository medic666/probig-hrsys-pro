<template>
  <div class="person-card-grid">
    <CardGrid ref="gridRef" :fetch-fn="fetchFn" :filter-fn="filterByScope" :empty-text="emptyText" :page-size="pageSize">
      <template #default="{ item }">
        <PersonCard :person="item" :dot-color="dotColorOf ? dotColorOf(item) : ''" :badge-position="badgePosition" @click="$emit('select', item)">
          <template #badge>
            <slot name="badge" :person="item" />
          </template>
        </PersonCard>
      </template>
    </CardGrid>
    <PersonScopeSwitch v-model="scope" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import CardGrid from '@/components/cards/CardGrid.vue'
import PersonCard from '@/components/cards/PersonCard.vue'
import PersonScopeSwitch from '@/components/cards/PersonScopeSwitch.vue'
import { filterPersons, type PersonScope } from '@/utils/personScope'

// 人员卡片网格领域组件：人员卡片 + 活跃/全部范围开关 + 活跃过滤统一内置。
// 页面只需提供 fetchFn（拉全量人员卡片数据）、dotColorOf（颜色点派生函数）与
// #badge scoped slot（模块特有小组件）。佩戴规则由 PersonCard 内部统一判定。
withDefaults(
  defineProps<{
    fetchFn: (params: any) => Promise<{ list: any[]; total: number }>
    dotColorOf?: (person: any) => string
    badgePosition?: 'name' | 'meta'
    emptyText?: string
    pageSize?: number
  }>(),
  {
    emptyText: '暂无数据',
    pageSize: 100,
    badgePosition: 'name',
    dotColorOf: undefined,
  },
)

defineEmits<{ (e: 'select', person: any): void }>()

const scope = ref<PersonScope>('active')
const gridRef = ref()

function filterByScope(items: any[]) {
  return filterPersons(items, scope.value)
}

function reload() {
  gridRef.value?.reload()
}

defineExpose({ reload, scope })
</script>
