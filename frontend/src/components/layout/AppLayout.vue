<template>
  <div
    class="ba-theme-shell"
    :class="shellClass"
  >
    <!-- Background Decoration -->
    <div class="ba-theme-backdrop pointer-events-none fixed inset-0"></div>

    <!-- 全局顶栏横跨侧栏和内容区，页面标题由内容区承载。 -->
    <AppHeader />

    <!-- Sidebar and Main Content Area -->
    <AppSidebar v-if="!hideSidebar" />

    <div
      class="relative z-10 flex min-w-0 flex-col pt-[var(--header-h)] transition-[margin-left] duration-layout"
      :class="[
        columnClass,
        hideSidebar
          ? ''
          : sidebarCollapsed
            ? 'lg:ml-[var(--sidebar-w-collapsed)]'
            : 'lg:ml-[var(--sidebar-w)]',
      ]"
    >
      <!-- Main Content：布局组件统一负责空间分配,子页面不再复制父级尺寸或抵消内边距。 -->
      <main v-content-reveal="route.path"
        class="app-main flex min-w-0 flex-1 flex-col"
        :class="mainClass"
      >
        <div v-if="pageTitle && !hidePageHeading" class="page-heading mb-4 flex flex-shrink-0 flex-wrap items-start justify-between gap-3">
          <div>
            <h1 class="page-title">{{ pageTitle }}</h1>
            <p v-if="pageDescription" class="page-description">{{ pageDescription }}</p>
          </div>
          <div v-if="$slots['page-heading-actions']" class="shrink-0">
            <slot name="page-heading-actions" />
          </div>
        </div>
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { vContentReveal } from '@/directives/contentReveal'

import '@/styles/onboarding.css'
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { usePageMeta } from '@/composables/usePageMeta'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

interface Props {
  // 全屏工作区使用动态视口锁定布局，并在组件存续期间禁止页面滚动。
  fullViewport?: boolean
  // 宽屏（lg 及以上）把内容区锁定为视口高度，保留页头与内边距，由页面内部区域自行滚动；窄屏仍按内容自然滚动。
  fitViewport?: boolean
  // 页面已有标题时，可隐藏布局提供的标题和说明。
  hidePageHeading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullViewport: false,
  fitViewport: false,
  hidePageHeading: false,
})

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
// 全屏工作区页面（如创作台）通过路由 meta 隐藏侧栏并取消内容区缩进。
const hideSidebar = computed(() => route.meta.hideSidebar === true)
const fullViewport = computed(() => props.fullViewport)

// 三种高度模式：全屏工作区始终锁定，宽屏锁定只在 lg 及以上生效，普通页面随内容增高。
const shellClass = computed(() => {
  if (fullViewport.value) return 'fixed inset-0 h-[100dvh] overflow-hidden'
  if (props.fitViewport) return 'min-h-screen lg:h-[100dvh] lg:min-h-0 lg:overflow-hidden'
  return 'min-h-screen'
})
const columnClass = computed(() => {
  if (fullViewport.value) return 'h-full min-h-0'
  if (props.fitViewport) return 'min-h-screen lg:h-full lg:min-h-0'
  return 'min-h-screen'
})
const mainClass = computed(() => {
  if (fullViewport.value) return 'min-h-0 p-0'
  const padding = 'px-4 pb-4 pt-4 md:px-6 md:pb-6 lg:px-8 lg:pb-8'
  return props.fitViewport ? `${padding} lg:min-h-0` : padding
})
const isAdmin = computed(() => authStore.user?.role === 'admin')

let previousHtmlOverflowY: string | null = null
let previousBodyOverflowY: string | null = null

// 动态视口变化可能触发浏览器恢复根页面滚动；全屏工作区必须只让内部控件滚动。
function lockDocumentScroll(): void {
  if (!fullViewport.value || typeof document === 'undefined') return
  previousHtmlOverflowY = document.documentElement.style.overflowY
  previousBodyOverflowY = document.body.style.overflowY
  document.documentElement.style.overflowY = 'hidden'
  document.body.style.overflowY = 'hidden'
}

// 路由离开时恢复进入创作台前的页面滚动策略，避免影响普通页面。
function restoreDocumentScroll(): void {
  if (typeof document === 'undefined') return
  if (previousHtmlOverflowY !== null) {
    document.documentElement.style.overflowY = previousHtmlOverflowY
    previousHtmlOverflowY = null
  }
  if (previousBodyOverflowY !== null) {
    document.body.style.overflowY = previousBodyOverflowY
    previousBodyOverflowY = null
  }
}

const { replayTour, startTeamTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()
const { pageTitle, pageDescription } = usePageMeta()

onMounted(() => {
  lockDocumentScroll()
  onboardingStore.setReplayCallback(replayTour)
  onboardingStore.setTeamGuideCallback(startTeamTour)
})

onBeforeUnmount(() => {
  restoreDocumentScroll()
})

defineExpose({ replayTour })
</script>

<!-- 空间分配全部经模板 flex 链完成:wrapper(flex-col, min-h-screen、宽屏锁定或全屏锁定)
     → app-main(flex-1) → page-heading(自然高度) + 页面内容(需要撑满时自取 flex-1)。
     不再维护 --main-pad-* / --page-heading-space 等与模板 padding 平行的镜像变量。 -->
