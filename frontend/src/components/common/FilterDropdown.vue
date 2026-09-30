<template>
  <div ref="dropdownRef" class="relative shrink-0" @keydown.esc.stop.prevent="closeWithFocus">
    <button
      ref="triggerRef"
      type="button"
      class="btn btn-secondary relative btn-icon"
      :aria-label="t('common.filter')"
      :title="t('common.filter')"
      :aria-expanded="open"
      @click="open = !open"
    >
      <Icon name="filter" size="sm" />
      <span v-if="activeCount" class="absolute -right-1 -top-1 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary-100 px-1.5 text-xs font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
        {{ activeCount }}
      </span>
    </button>

    <MotionTransition name="dropdown-fade">
      <div
        v-if="open"
        :inert="!open || undefined"
        class="dropdown right-0 top-full z-modal-nested mt-2 w-72 p-4 sm:left-0 sm:right-auto"
        @click.stop
      >
        <div class="mb-3 flex items-center justify-between">
          <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('common.filter') }}</div>
          <button v-if="activeCount" type="button" class="text-xs font-medium text-primary-600 dark:text-primary-400" @click="emit('reset')">
            {{ t('common.reset') }}
          </button>
        </div>
        <div class="space-y-3">
          <slot />
        </div>
      </div>
    </MotionTransition>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import MotionTransition from '@/components/common/MotionTransition.vue'

defineProps<{ activeCount: number }>()
const emit = defineEmits<{ (event: 'reset'): void }>()
const { t } = useI18n()
const open = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)

// 键盘关闭后回到筛选按钮，便于继续操作工具栏。
function closeWithFocus() {
  open.value = false
  triggerRef.value?.focus()
}

// Select 的 Teleport 面板会阻止点击冒泡，选择条件时不会误关外层筛选面板。
function closeOutside(event: MouseEvent) {
  if (event.target instanceof Node && !dropdownRef.value?.contains(event.target)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', closeOutside))
onBeforeUnmount(() => document.removeEventListener('click', closeOutside))
</script>
