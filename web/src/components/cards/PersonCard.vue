<template>
  <div class="person-card" @click="$emit('click', person)">
    <div class="pc-name">
      <span>{{ person.name }}</span>
      <slot name="badge" />
    </div>
    <div class="pc-meta">
      <div v-if="person.company_name" class="pc-line">公司：{{ person.company_name }}</div>
      <div v-if="person.department" class="pc-line">部门：{{ person.department }}</div>
      <div v-if="person.position" class="pc-line">职位：{{ person.position }}</div>
      <div v-if="!person.company_name && !person.department && !person.position" class="pc-line pc-empty">暂无职务信息</div>
      <slot name="extra" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  person: {
    id: number
    name: string
    company_id?: number
    company_name?: string
    department?: string
    position?: string
    is_active?: boolean
    entry_date?: string | null
    leave_date?: string | null
  }
}>()
defineEmits<{ (e: 'click', person: any): void }>()
</script>

<style lang="scss" scoped>
.person-card {
  width: 220px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .pc-name {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pc-meta {
    .pc-line {
      font-size: 12px;
      color: #606266;
      line-height: 20px;
    }

    .pc-empty {
      color: #c0c4cc;
    }
  }
}
</style>
