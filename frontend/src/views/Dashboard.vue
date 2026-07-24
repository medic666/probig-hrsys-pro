<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="6" v-for="card in cards" :key="card.title">
        <el-card shadow="hover">
          <div style="display: flex; align-items: center; justify-content: space-between">
            <div>
              <div style="font-size: 14px; color: #999">{{ card.title }}</div>
              <div style="font-size: 28px; font-weight: bold; margin-top: 8px">{{ card.value }}</div>
            </div>
            <el-icon :size="40" :color="card.color"><component :is="card.icon" /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { User, OfficeBuilding, Calendar, Money } from '@element-plus/icons-vue'
import { listPersonnel } from '../api/personnel'
import { listOrganizations } from '../api/organization'
import { listAttendanceSummaries } from '../api/attendance'
import { listSalarySummaries } from '../api/salary'

const cards = ref([
  { title: '人员总数', value: 0, icon: User, color: '#409EFF' },
  { title: '组织总数', value: 0, icon: OfficeBuilding, color: '#67C23A' },
  { title: '考勤记录', value: 0, icon: Calendar, color: '#E6A23C' },
  { title: '工资记录', value: 0, icon: Money, color: '#F56C6C' },
])

onMounted(async () => {
  try {
    const [p, o, a, s] = await Promise.allSettled([
      listPersonnel({ page: 1, page_size: 1 }),
      listOrganizations({ page: 1, page_size: 1 }),
      listAttendanceSummaries({ page: 1, page_size: 1 }),
      listSalarySummaries({ page: 1, page_size: 1 }),
    ])
    cards.value[0].value = p.status === 'fulfilled' ? p.value.data.total : 0
    cards.value[1].value = o.status === 'fulfilled' ? o.value.data.total : 0
    cards.value[2].value = a.status === 'fulfilled' ? a.value.data.total : 0
    cards.value[3].value = s.status === 'fulfilled' ? s.value.data.total : 0
  } catch {}
})
</script>
