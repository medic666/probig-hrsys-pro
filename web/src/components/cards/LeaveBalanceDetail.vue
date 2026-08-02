<template>
  <div>
    <template v-if="showAL">
      <h4 class="lb-title">年假明细</h4>
      <el-descriptions v-if="alDetail" :column="1" border size="small">
        <el-descriptions-item label="累计配发(天)">{{ hoursToDays(alDetail.grant).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="累计已休(天)">{{ hoursToDays(alDetail.consumed).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="累计人工调整(天)">{{ hoursToDays(alDetail.adjust).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="累计结转扣除(天)">{{ hoursToDays(alDetail.carryover).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="当前可用(天)">{{ hoursToDays(alDetail.balance).toFixed(2) }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else description="暂无年假余额记录" :image-size="50" />
    </template>
    <template v-if="showLIL">
      <h4 class="lb-title">调休明细</h4>
      <el-descriptions v-if="lilDetail" :column="1" border size="small">
        <el-descriptions-item label="累计补班(天)">{{ hoursToDays(lilDetail.makeup).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="累计调休(天)">{{ hoursToDays(lilDetail.consumed).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="当前可用(天)">{{ hoursToDays(lilDetail.balance).toFixed(2) }}</el-descriptions-item>
      </el-descriptions>
      <el-empty v-else description="暂无调休余额记录" :image-size="50" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import request from '@/utils/request'
import { hoursToDays } from '@/utils'

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

<style scoped>
.lb-title {
  font-size: 13px;
  color: #606266;
  margin: 0 0 8px;
}
.lb-title + .el-descriptions,
.lb-title + .el-empty {
  margin-bottom: 12px;
}
</style>
