<template>
  <MotionTransition name="pop-float">
    <div
      v-if="visible"
      class="sticky bottom-4"
      data-testid="settings-save-bar"
    >
      <div
        class="flex flex-wrap items-center justify-between gap-3 rounded-surface border border-primary-900/10 bg-white px-4 py-3 shadow-lg dark:border-dark-600 dark:bg-dark-900"
        role="status"
      >
        <p class="flex items-center gap-2 text-sm font-medium text-primary-900 dark:text-dark-50">
          <span class="h-2 w-2 shrink-0 rounded-full bg-amber-500" aria-hidden="true" />
          {{ message }}
        </p>
        <div class="flex items-center gap-2">
          <slot />
        </div>
      </div>
    </div>
  </MotionTransition>
</template>

<script setup lang="ts">
import MotionTransition from '@/components/common/MotionTransition.vue'

// 整页设置有未保存的修改时，在视口底部吸附一条操作栏，保存按钮由调用方放进默认插槽。
defineProps<{
  visible: boolean
  message: string
}>()
</script>
