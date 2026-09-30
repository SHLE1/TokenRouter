<template>
  <AppLayout>
    <template #page-heading-actions>
      <UserDashboardUsageToolbar :refreshing="refreshing" @refresh="refreshAll" />
    </template>

    <div class="space-y-4">
      <UserDashboardUsageChart />
      <UserDashboardHeatmap ref="heatmapRef" />
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div class="lg:col-span-2"><UserDashboardAnnouncements /></div>
        <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useAnnouncementStore } from '@/stores/announcements'
import AppLayout from '@/components/layout/AppLayout.vue'
import UserDashboardUsageChart from '@/components/user/dashboard/UserDashboardUsageChart.vue'
import UserDashboardUsageToolbar from '@/components/user/dashboard/UserDashboardUsageToolbar.vue'
import { provideUsageChartState } from '@/components/user/dashboard/usageChartState'
import UserDashboardHeatmap from '@/components/user/dashboard/UserDashboardHeatmap.vue'
import UserDashboardAnnouncements from '@/components/user/dashboard/UserDashboardAnnouncements.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'

const authStore = useAuthStore()
const announcementStore = useAnnouncementStore()

// 用量状态由页面提供，标题行的工具栏和正文的图表共用同一份。
const usageState = provideUsageChartState()
const heatmapRef = ref<InstanceType<typeof UserDashboardHeatmap> | null>(null)
const refreshing = ref(false)

// refreshUser 刷新账户信息，顶栏余额随之更新。
const refreshUser = async () => {
  try {
    await authStore.refreshUser()
  } catch (error) {
    console.error('Failed to refresh user:', error)
  }
}

// App 负责首次预加载；用户主动刷新时同时绕过公告节流获取最新内容。
const refreshAll = async () => {
  refreshing.value = true
  try {
    // 各区块自行处理错误，这里只等全部结束再恢复刷新按钮。
    await Promise.allSettled([
      refreshUser(),
      usageState.load(),
      heatmapRef.value?.reload(),
      announcementStore.fetchAnnouncements(true),
    ])
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  void refreshUser()
  void usageState.load()
  void usageState.loadFilterOptions()
})
</script>
