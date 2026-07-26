<template>
  <el-select
    v-model="selected"
    filterable
    remote
    :remote-method="searchPersons"
    :loading="loading"
    :multiple="multiple"
    :placeholder="multiple ? '请选择人员(可多选)' : '请选择人员'"
    clearable
    @focus="searchPersons('')"
  >
    <el-option
      v-for="item in options"
      :key="item.id"
      :label="item.name"
      :value="item.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { getPersonList } from '@/api/person'

const props = withDefaults(defineProps<{
  modelValue?: number | number[] | undefined
  multiple?: boolean
}>(), {
  modelValue: undefined,
  multiple: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | number[] | undefined]
}>()

const selected = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const loading = ref(false)
const options = ref<{ id: number; name: string }[]>([])

async function searchPersons(query: string) {
  loading.value = true
  try {
    const data = await getPersonList({ pageNum: 1, pageSize: 500, name: query })
    options.value = data.list || []
  } catch (e) {
    options.value = []
  } finally {
    loading.value = false
  }
}
</script>
