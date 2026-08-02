<template>
  <div>
    <BalanceDetailList v-if="showAL" title="年假明细" empty-text="暂无年假余额记录" :rows="alRows" />
    <BalanceDetailList v-if="showLIL" title="调休明细" empty-text="暂无调休余额记录" :rows="lilRows" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import request from '@/utils/request'
import { hoursToDays } from '@/utils'
import BalanceDetailList from '@/components/cards/BalanceDetailList.vue'

const props = withDefaults(
  defineProps<{
    personId: number | null
    showAL?: boolean
    showLIL?: boolean
  }>(),
  {
    showAL: true,
    showLIL: true,
  },
)

const alDetail = ref<any>(null)
const lilDetail = ref<any>(null)

const alRows = computed(() =>
  alDetail.value
    ? [
        { label: '累计配发(天)', value: hoursToDays(alDetail.value.grant).toFixed(2) },
        { label: '累计已休(天)', value: hoursToDays(alDetail.value.consumed).toFixed(2) },
        { label: '累计人工调整(天)', value: hoursToDays(alDetail.value.adjust).toFixed(2) },
        { label: '累计结转扣除(天)', value: hoursToDays(alDetail.value.carryover).toFixed(2) },
        { label: '当前可用(天)', value: hoursToDays(alDetail.value.balance).toFixed(2) },
      ]
    : [],
)

const lilRows = computed(() =>
  lilDetail.value
    ? [
        { label: '累计补班(天)', value: hoursToDays(lilDetail.value.makeup).toFixed(2) },
        { label: '累计调休(天)', value: hoursToDays(lilDetail.value.consumed).toFixed(2) },
        { label: '当前可用(天)', value: hoursToDays(lilDetail.value.balance).toFixed(2) },
      ]
    : [],
)

watch(
  () => props.personId,
  (v) => {
    alDetail.value = null
    lilDetail.value = null
    if (!v) return
    if (props.showAL) {
      request.get(`/persons/${v}/annual-leave-balance-detail`)
        .then((d: any) => { alDetail.value = d })
        .catch(() => { alDetail.value = null })
    }
    if (props.showLIL) {
      request.get(`/persons/${v}/lil-balance-detail`)
        .then((d: any) => { lilDetail.value = d })
        .catch(() => { lilDetail.value = null })
    }
  },
  { immediate: true },
)
</script>
