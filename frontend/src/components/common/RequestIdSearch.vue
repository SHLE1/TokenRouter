<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  modelValue?: string | null
  withinTimeRange?: boolean
  showTimeRange?: boolean
  showDetails?: boolean
}>()
const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
  (event: 'update:withinTimeRange', value: boolean): void
  (event: 'search'): void
}>()
const { t } = useI18n()
const route = useRoute()
const destination = computed(() => ({
  path: route?.path.startsWith('/admin') ? '/admin/requests' : '/requests',
  query: { request_id: props.modelValue?.trim() || undefined },
}))

// 时间范围变更与搜索使用同一组父组件状态。
function changeTimeRange(event: Event) {
  emit('update:withinTimeRange', (event.target as HTMLInputElement).checked)
  emit('search')
}
</script>

<template>
  <div class="flex min-w-0 flex-wrap items-center gap-2">
    <input
      :value="modelValue"
      :aria-label="t('requests.id')"
      :placeholder="t('requests.searchPlaceholder')"
      class="input min-w-0 flex-1 font-mono sm:w-96 sm:flex-none"
      type="search"
      maxlength="255"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @keydown.enter.prevent="emit('search')"
      @search="emit('search')"
    />
    <button type="button" class="btn btn-secondary" @click="emit('search')">{{ t('common.search') }}</button>
    <RouterLink v-if="showDetails !== false && modelValue?.trim()" :to="destination" class="btn btn-secondary">{{ t('requests.details') }}</RouterLink>
    <label v-if="showTimeRange && modelValue?.trim()" class="flex items-center gap-2 text-sm text-gray-500 dark:text-dark-300">
      <input type="checkbox" :checked="withinTimeRange" @change="changeTimeRange" />
      {{ t('requests.withinTimeRange') }}
    </label>
  </div>
</template>
