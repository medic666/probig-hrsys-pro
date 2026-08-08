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
import { searchPersonOptions } from '@/api/person'

const props = withDefaults(
  defineProps<{
    modelValue: number | number[] | null
    fetchApi?: (keyword?: string) => Promise<{ id: number; name: string }[]>
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
    multiple?: boolean
    remoteThreshold?: number
    // 预知名：当前选中值在选项未加载完成时的显示名（如后端已返回的 matched_name），
    // 避免 el-select 找不到对应 option 时回退显示原始数字 ID
    valueLabel?: string
  }>(),
  {
    modelValue: null,
    fetchApi: searchPersonOptions,
    placeholder: '请选择',
    disabled: false,
    clearable: true,
    multiple: false,
    remoteThreshold: 500,
    valueLabel: '',
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
  let base: { id: number; name: string }[]
  if (isRemoteMode.value || allOptions.value.length > props.remoteThreshold) {
    base = allOptions.value
  } else if (!searchKeyword.value) {
    base = allOptions.value
  } else {
    const kw = searchKeyword.value.toLowerCase()
    base = allOptions.value.filter((item) => item.name.toLowerCase().includes(kw))
  }
  // 预知名兜底：选项未加载完成时用 valueLabel 渲染当前选中值，
  // 选项到达后自动被同名同 id 的真实选项替换，无闪烁
  if (props.valueLabel && typeof props.modelValue === 'number' && !base.some((o) => o.id === props.modelValue)) {
    return [{ id: props.modelValue, name: props.valueLabel }, ...base]
  }
  return base
})

// 初始全量加载共享 in-flight（按 fetchApi 去重）：同数据源并发实例只发一个请求
// （如钉钉导入预览 N 行 = 1 个全量请求），与 useBadges 的请求合并模式同构。
// 带关键字（远程搜索）各自独立，不共享。
const inflight = new Map<(...args: any[]) => Promise<any>, Promise<any>>()

function fetchShared(fetchApi: (...args: any[]) => Promise<any>) {
  let p = inflight.get(fetchApi)
  if (!p) {
    p = fetchApi().finally(() => {
      inflight.delete(fetchApi)
    })
    inflight.set(fetchApi, p)
  }
  return p
}

async function loadOptions(keyword?: string) {
  loading.value = true
  try {
    const data = keyword ? await props.fetchApi(keyword) : await fetchShared(props.fetchApi)
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
