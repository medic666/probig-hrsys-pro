<template>
  <BusinessPage>
    <BatchCalcForm :submit-fn="submitFn" @saved="goBack" @cancel="goBack" />
  </BusinessPage>
</template>

<script setup lang="ts">
import BusinessPage from '@/components/BusinessPage.vue'
import BatchCalcForm from '@/components/BatchCalcForm.vue'
import { calculateSalaries } from '@/api/salary'
import { useBusinessPage } from '@/composables/useBusinessPage'

// 月度工资批量核算（页面化）：月份 + 人员域多选，核算完成返回列表页
const { goBack } = useBusinessPage()
const submitFn = (data: any) => calculateSalaries(data)
</script>
