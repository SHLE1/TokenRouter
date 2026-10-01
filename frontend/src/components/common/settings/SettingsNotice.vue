<template>
  <div
    class="flex items-start gap-2 rounded-control border px-3 py-2 text-xs"
    :class="TONE_CLASSES[tone]"
    :role="tone === 'warning' ? 'alert' : undefined"
  >
    <Icon
      :name="tone === 'warning' ? 'exclamationTriangle' : 'infoCircle'"
      size="sm"
      class="shrink-0"
    />
    <div class="min-w-0 flex-1 space-y-1">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

// 设置表单里的说明和风险提示只用两种语气，配色集中在这里维护。
withDefaults(defineProps<{ tone?: 'info' | 'warning' }>(), { tone: 'info' })

const TONE_CLASSES = {
  info: 'border-sky-200 bg-sky-50 text-sky-800 dark:border-sky-800/60 dark:bg-sky-900/20 dark:text-sky-200',
  warning: 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-800/60 dark:bg-amber-900/20 dark:text-amber-200',
} as const
</script>
