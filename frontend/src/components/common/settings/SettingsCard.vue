<template>
  <section class="card">
    <header class="card-header flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0 flex-1">
        <h2 class="text-lg font-semibold text-primary-900 dark:text-dark-50">
          {{ title }}
        </h2>
        <p
          v-if="description"
          class="mt-1 text-sm text-primary-900/80 dark:text-dark-300"
        >
          {{ description }}
        </p>
        <slot name="description" />
      </div>
      <slot name="actions" />
    </header>
    <div class="space-y-6 p-6">
      <slot />
    </div>
    <footer
      v-if="$slots.footer"
      class="flex flex-wrap items-center justify-end gap-2 border-t border-primary-900/10 px-6 py-4 dark:border-dark-600"
    >
      <slot name="footer" />
    </footer>
  </section>
</template>

<script setup lang="ts">
// 整页设置的一组设置项放在一张卡片里：头部是标题和说明，内容区按 SettingsSection 分区，
// 单独调用接口保存的卡片把保存按钮放进 footer 插槽，显示在卡片底部右侧。
defineProps<{
  title: string
  description?: string
}>()
</script>
