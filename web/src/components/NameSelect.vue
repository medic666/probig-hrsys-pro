<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { get as requestGet } from '@/api/request'

interface Props {
  type: 'person' | 'company' | 'attendance_group' | 'role'
  modelValue?: number | null
  placeholder?: string
  disabled?: boolean
  multiple?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  placeholder: '请选择',
  disabled: false,
  multiple: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: number | number[] | null): void
}>()

const options = ref<{ label: string; value: number }[]>([])
const loading = ref(false)
const selectedValue = ref<number | number[] | null>(props.modelValue)

const apiMap: Record<string, string> = {
  person: '/api/persons',
  company: '/api/companies',
  attendance_group: '/api/sys-config',
  role: '/api/roles'
}

async function fetchOptions() {
  loading.value = true
  try {
    if (props.type === 'attendance_group') {
      const res = await requestGet('/api/sys-config', { config_key: 'attendance.groups' })
      const data = res as unknown as { config_value: string }
      if (data?.config_value) {
        const groups = JSON.parse(data.config_value) as string[]
        options.value = groups.map((g) => ({ label: g, value: g as unknown as number }))
      }
    } else {
      const res = await requestGet(apiMap[props.type], { pageNum: 1, pageSize: 500 })
      const data = res as unknown as { list: { id: number; name: string }[] }
      if (data?.list) {
        options.value = data.list.map((item) => ({ label: item.name, value: item.id }))
      }
    }
  } catch {
    options.value = []
  } finally {
    loading.value = false
  }
}

function handleChange(val: number | number[] | null) {
  emit('update:modelValue', val)
}

watch(() => props.modelValue, (val) => {
  selectedValue.value = val
})

onMounted(() => {
  fetchOptions()
})
</script>

<template>
  <el-select
    v-model="selectedValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :multiple="multiple"
    :loading="loading"
    filterable
    @change="handleChange"
  >
    <el-option
      v-for="item in options"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    />
  </el-select>
</template>
