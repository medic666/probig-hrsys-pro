<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder || '请选择'"
    :disabled="disabled"
    :clearable="clearable"
    :multiple="multiple"
    :filterable="!isRemoteMode"
    :remote="isRemoteMode"
    :remote-method="handleRemoteSearch"
    :loading="loading"
    @update:model-value="handleChange"
    @visible-change="handleVisibleChange"
  >
    <el-option
      v-for="item in displayOptions"
      :key="item.id"
      :label="item.name"
      :value="item.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: number | number[] | null
    fetchApi: (keyword?: string) => Promise<{ id: number; name: string }[]>
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
    multiple?: boolean
    remoteThreshold?: number
  }>(),
  {
    modelValue: null,
    placeholder: '请选择',
    disabled: false,
    clearable: true,
    multiple: false,
    remoteThreshold: 500,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: number | number[] | null): void
}>()

const loading = ref(false)
const allOptions = ref<{ id: number; name: string }[]>([])
const searchKeyword = ref('')
const isRemoteMode = ref(false)

const displayOptions = computed(() => {
  if (isRemoteMode.value || allOptions.value.length > props.remoteThreshold) {
    return allOptions.value
  }
  if (!searchKeyword.value) return allOptions.value
  const kw = searchKeyword.value.toLowerCase()
  return allOptions.value.filter((item) => item.name.toLowerCase().includes(kw))
})

async function loadOptions(keyword?: string) {
  loading.value = true
  try {
    const data = await props.fetchApi(keyword)
    allOptions.value = data || []
    if (data && data.length >= props.remoteThreshold) {
      isRemoteMode.value = true
    }
  } catch {
    allOptions.value = []
  } finally {
    loading.value = false
  }
}

function handleChange(val: number | number[] | null) {
  emit('update:modelValue', val)
}

function handleRemoteSearch(keyword: string) {
  searchKeyword.value = keyword
  loadOptions(keyword)
}

function handleVisibleChange(visible: boolean) {
  if (visible && !isRemoteMode.value) {
    loadOptions()
  }
}

watch(
  () => props.modelValue,
  () => {},
  { immediate: true },
)

onMounted(() => {
  loadOptions()
})
</script>
