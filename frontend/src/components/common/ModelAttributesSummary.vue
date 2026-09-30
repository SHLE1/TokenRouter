<template>
  <div class="space-y-3">
    <p v-if="attributes.route_differences" :class="isTooltip ? 'text-amber-300' : 'text-sm text-amber-700 dark:text-amber-300'">{{ t('admin.modelAttributes.routeDifferences') }}</p>
    <dl :class="isTooltip ? 'grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5' : 'grid grid-cols-2 gap-x-4 gap-y-3 text-sm'">
      <template v-for="field in fields" :key="field">
        <dt :class="isTooltip ? 'text-gray-400 dark:text-dark-400' : 'text-gray-500 dark:text-dark-400'">{{ t(`admin.modelAttributes.fields.${field}`) }}</dt>
        <dd :class="isTooltip ? 'break-words text-right text-white dark:text-dark-100' : 'break-words text-gray-900 dark:text-dark-50'">{{ format(attributes[field]) }}</dd>
      </template>
    </dl>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { attributeCapabilities, attributeLimits, type ModelAttributes } from '@/types/modelAttributes'

const props = withDefaults(defineProps<{
  attributes: ModelAttributes
  // tooltip 变体用于深色浮层：反色文字，并省略标题旁已展示的显示名。
  variant?: 'default' | 'tooltip'
}>(), {
  variant: 'default',
})
const { t } = useI18n()
const isTooltip = computed(() => props.variant === 'tooltip')
const fields = computed(() => [
  ...(isTooltip.value ? [] : ['display_name'] as const),
  ...attributeLimits,
  'input_modalities',
  'output_modalities',
  ...attributeCapabilities,
] as const)
function format(value: ModelAttributes[keyof ModelAttributes]) {
  if (value === undefined || value === null) return t('admin.modelAttributes.unknown')
  if (typeof value === 'boolean') return t(value ? 'admin.modelAttributes.supported' : 'admin.modelAttributes.unsupported')
  if (Array.isArray(value)) return value.length ? value.map(item => t(`admin.modelAttributes.modalities.${item}`)).join(', ') : t('admin.modelAttributes.none')
  return typeof value === 'number' ? value.toLocaleString() : value
}
</script>
