<template>
  <div class="attendance-details-editor">
    <el-table :data="rows" border size="small">
      <el-table-column label="类型" width="110">
        <template #default="{ row }">
          <el-select v-model="row.event_type" size="small" style="width:100%" :disabled="readonly" @change="onTypeChange(row)">
            <el-option v-for="t in eventTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="子类型" width="140">
        <template #default="{ row }">
          <el-select v-model="row.sub_type" size="small" style="width:100%" :disabled="readonly">
            <el-option v-for="s in subTypeMap[row.event_type] || []" :key="s" :label="s" :value="s" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="时长(小时)" width="120">
        <template #default="{ row }">
          <el-input-number v-if="row.event_type !== '违纪'" v-model="row.hours" :min="0" :precision="1" size="small" style="width:100%" :disabled="readonly" />
        </template>
      </el-table-column>
      <el-table-column label="分钟" width="100">
        <template #default="{ row }">
          <el-input-number v-if="row.sub_type === '迟到' || row.sub_type === '早退'" v-model="row.minutes" :min="0" size="small" style="width:100%" :disabled="readonly" />
        </template>
      </el-table-column>
      <el-table-column label="备注" min-width="140">
        <template #default="{ row }">
          <el-input v-model="row.remark" size="small" :disabled="readonly" />
        </template>
      </el-table-column>
      <el-table-column v-if="!readonly" label="操作" width="70">
        <template #default="{ $index }">
          <el-button type="danger" link size="small" @click="rows.splice($index, 1)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-button v-if="!readonly" size="small" style="margin-top:8px" @click="addEvent">+ 添加事件</el-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// 考勤明细统一编辑器：录入考勤/批量录入共用，支持只读模式（日记工时查看入口）。
// 所有事件类型统一渲染子类型下拉（含违纪的缺卡/迟到/早退）；违纪行隐藏时长、迟到/早退展示分钟输入。
const props = withDefaults(defineProps<{ modelValue: any[]; readonly?: boolean }>(), { readonly: false })
const emit = defineEmits<{ (e: 'update:modelValue', v: any[]): void }>()

const eventTypes = ['出勤', '休假', '加班', '违纪']
const subTypeMap: Record<string, string[]> = {
  '出勤': ['普通出勤', '补班出勤', '外勤出勤'],
  '休假': ['调休', '事假', '病假', '年假', '法定假', '福利假'],
  '加班': ['工作日加班', '节假日加班'],
  '违纪': ['缺卡', '迟到', '早退'],
}

const rows = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

function addEvent() {
  rows.value = [
    ...rows.value,
    { event_type: '出勤', sub_type: '普通出勤', hours: 8, minutes: 0, remark: '' },
  ]
}

function onTypeChange(row: any) {
  row.sub_type = ''
  row.hours = row.event_type === '违纪' ? 0 : 8
  row.minutes = 0
}
</script>
