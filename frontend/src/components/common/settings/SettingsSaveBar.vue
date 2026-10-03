<template>
  <MotionTransition name="pop-float">
    <!-- 外层占满宽度只负责居中，点击穿透到下面的内容。 -->
    <div
      v-if="visible"
      class="pointer-events-none sticky bottom-6 flex justify-center"
      data-testid="settings-save-bar"
    >
      <div
        class="pointer-events-auto flex max-w-full flex-wrap items-center gap-x-4 gap-y-2 rounded-full bg-gray-900 py-2 pl-5 pr-2 text-white shadow-xl ring-1 ring-black/5 dark:bg-dark-700 dark:text-dark-50 dark:ring-dark-500"
        role="status"
      >
        <p class="flex items-center gap-2 text-sm font-medium">
          <span class="h-2 w-2 shrink-0 rounded-full bg-amber-400" aria-hidden="true" />
          {{ message }}
        </p>
        <div class="flex items-center gap-1">
          <slot />
        </div>
      </div>
    </div>
  </MotionTransition>
</template>

<script setup lang="ts">
import MotionTransition from '@/components/common/MotionTransition.vue'

// 整页设置有未保存的修改时，在视口底部居中浮出深色操作条，放弃和保存按钮由调用方放进默认插槽。
defineProps<{
  visible: boolean
  message: string
}>()
</script>
