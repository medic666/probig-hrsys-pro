<template>
  <el-select
    v-model="selectedValue"
    :placeholder="placeholder"
    filterable
    :clearable="clearable"
    :loading="loading"
    :remote="remote"
    :remote-method="remote ? handleRemoteSearch : undefined"
    v-bind="$attrs"
  >
    <el-option
      v-for="item in options"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: any
  api: () => Promise<any>
  placeholder?: string
  clearable?: boolean
  valueKey?: string
  labelKey?: string
}>(), {
  placeholder: '请选择',
  clearable: true,
  valueKey: 'id',
  labelKey: 'name',
})

const emit = defineEmits<{ 'update:modelValue': [value: any] }>()

const selectedValue = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const options = ref<any[]>([])
const loading = ref(false)
const remote = ref(false)

async function fetchOptions() {
  loading.value = true
  try {
    const data = await props.api()
    if (data && data.list) {
      options.value = data.list.map((item: any) => ({
        value: item[props.valueKey],
        label: item[props.labelKey],
      }))
    } else if (Array.isArray(data)) {
      options.value = data.map((item: any) => ({
        value: item[props.valueKey],
        label: item[props.labelKey],
      }))
    }
    if (options.value.length > 500) {
      remote.value = true
    }
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
}

function handleRemoteSearch(query: string) {
  // handle remote search if needed
}

onMounted(() => {
  fetchOptions()
})
</script>
