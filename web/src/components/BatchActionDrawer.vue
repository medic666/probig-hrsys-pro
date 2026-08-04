<template>
  <el-drawer v-model="visible" :title="title" direction="btt" size="50%" destroy-on-close>
    <div class="drawer-body">
      <BatchCalcForm :submit-fn="submitFn" :show-person="showPerson" @saved="onSaved" @cancel="visible = false" />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import BatchCalcForm from '@/components/BatchCalcForm.vue'

// 批量操作通用底部抽屉（移动端友好）：批量结转/批量核算共用。
// 参数复杂度分层：单参数/简参数操作用抽屉（与复杂多参数的页面化操作区分），
// 选月份→执行→关闭→通知调用方刷新，闭环在一个面板内完成。
withDefaults(
  defineProps<{
    title: string
    submitFn: (data: any) => Promise<any>
    showPerson?: boolean
  }>(),
  { showPerson: true },
)

const emit = defineEmits<{ (e: 'update:visible', v: boolean): void; (e: 'done'): void }>()

const visible = defineModel<boolean>('visible', { default: false })

function onSaved() {
  visible.value = false
  emit('done')
}
</script>

<style scoped>
.drawer-body {
  padding: 0 8px;
}
</style>
