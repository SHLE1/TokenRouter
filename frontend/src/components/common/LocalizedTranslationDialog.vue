<template>
  <BaseDialog
    :show="show"
    :title="t('localization.dialogTitle', { field: title })"
    :subtitle="t('localization.dialogHint')"
    :width="width"
    :z-index="Z_INDEX.MODAL_NESTED"
    @close="emit('close')"
  >
    <div v-if="draft" class="space-y-4">
      <section class="space-y-3 rounded-surface border border-primary-900/10 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-800/60">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-sm font-medium text-primary-900 dark:text-dark-50">{{ t('localization.original') }}</span>
          <Select
            class="w-40"
            :model-value="draft.source_locale || ''"
            :options="languageOptions"
            :placeholder="t('localization.originalLanguage')"
            :aria-label="t('localization.originalLanguage')"
            @update:model-value="changeSourceLocale(String($event))"
          />
        </div>
        <SettingsNotice v-if="languageConflict" tone="warning">{{ t('localization.languageConflict') }}</SettingsNotice>
        <SettingsNotice v-else-if="!draft.source_locale" tone="warning">{{ t('localization.unknownOriginal') }}</SettingsNotice>
        <slot :value="draft.source" :update="updateDraftSource" :locale="draft.source_locale" :id="`${uid}-source`" />
      </section>

      <section
        v-for="item in targets"
        :key="item.code"
        :ref="element => setBlock(item.code, element)"
        class="space-y-3 rounded-surface border border-primary-900/10 p-4 dark:border-dark-600"
      >
        <div class="flex min-h-8 flex-wrap items-center justify-between gap-2">
          <div class="flex min-w-0 items-center gap-2">
            <span class="text-sm font-medium text-primary-900 dark:text-dark-50">{{ item.name }}</span>
            <span class="inline-flex items-center gap-1 text-xs" :class="STATUS_CLASSES[item.status]">
              <Icon :name="STATUS_ICONS[item.status]" size="xs" :animate-on-hover="false" />
              {{ t(`localization.status.${item.status}`) }}
            </span>
          </div>
          <div v-if="item.status !== 'missing'" class="flex items-center gap-1">
            <button
              type="button"
              class="btn-icon-sm text-gray-500 hover:bg-gray-100 hover:text-primary-600 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-primary-500"
              :title="t('localization.useAsOriginal')"
              :aria-label="t('localization.useAsOriginal')"
              @click="apply(promoteToOriginal(draft, item.code))"
            >
              <Icon name="arrowUp" size="sm" />
            </button>
            <button
              type="button"
              class="btn-icon-sm text-gray-500 hover:bg-gray-100 hover:text-red-600 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-red-400"
              :title="t('localization.removeTranslation')"
              :aria-label="t('localization.removeTranslation')"
              @click="apply(removeTranslation(draft, item.code))"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
          <button v-else type="button" class="btn btn-secondary btn-sm gap-1" @click="add(item.code)">
            <Icon name="plus" size="sm" />
            {{ t('localization.addLocale') }}
          </button>
        </div>
        <SettingsNotice v-if="item.status === 'stale'" tone="warning">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <span>{{ t('localization.staleNotice') }}</span>
            <button type="button" class="shrink-0 font-medium underline underline-offset-2" @click="apply(markStillValid(draft, item.code))">
              {{ t('localization.stillValid') }}
            </button>
          </div>
        </SettingsNotice>
        <slot
          v-if="item.status !== 'missing'"
          :value="draft.translations[item.code].value"
          :update="(value: T) => apply(updateTranslation(draft!, item.code, value))"
          :locale="item.code"
          :id="`${uid}-${item.code}`"
        />
      </section>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" @click="save">{{ t('localization.done') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts" generic="T">
import { computed, nextTick, ref, watch, type ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import SettingsNotice from '@/components/common/settings/SettingsNotice.vue'
import Icon from '@/components/icons/Icon.vue'
import { Z_INDEX } from '@/constants/overlay'
import { availableLocales } from '@/i18n/catalog'
import type { LocalizedUpdate } from '@/i18n/content'
import {
  addTranslation,
  markStillValid,
  promoteToOriginal,
  removeTranslation,
  setSourceLocale,
  translationStatus,
  updateSource,
  updateTranslation,
  withDefaultSource,
  type TranslationStatus,
} from '@/i18n/contentEdit'

let dialogSequence = 0

const props = withDefaults(defineProps<{
  show: boolean
  title: string
  modelValue: LocalizedUpdate<T>
  /** 新内容的原文语言，以及原文语言未知时修改原文所用的语言。 */
  fallbackLocale?: string
  width?: 'normal' | 'wide'
}>(), { width: 'normal' })
const emit = defineEmits<{ close: []; save: [value: LocalizedUpdate<T>] }>()
defineSlots<{ default(props: { value: T; update: (value: T) => void; locale: string | null; id: string }): unknown }>()

const { t } = useI18n()
const uid = `localized-${++dialogSequence}`
const draft = ref<LocalizedUpdate<T>>()
const languageConflict = ref(false)
const blocks = new Map<string, Element>()
const languageOptions = availableLocales.map(item => ({ value: item.code, label: item.name }))

const STATUS_CLASSES: Record<TranslationStatus, string> = {
  translated: 'text-emerald-600 dark:text-emerald-400',
  stale: 'text-amber-600 dark:text-amber-400',
  missing: 'text-gray-500 dark:text-dark-400',
}
const STATUS_ICONS = { translated: 'checkCircle', stale: 'exclamationTriangle', missing: 'circle' } as const

// 原文语言以外的每种语言各占一块，未翻译的语言也列出来。
const targets = computed(() => {
  const content = draft.value
  if (!content) return []
  return availableLocales
    .filter(item => item.code !== content.source_locale)
    .map(item => ({ code: item.code, name: item.name, status: translationStatus(content, item.code) }))
})

// 每次打开都从外层表单的当前值复制草稿，取消时直接丢弃。
watch(() => props.show, open => {
  if (!open) return
  draft.value = withDefaultSource(JSON.parse(JSON.stringify(props.modelValue)) as LocalizedUpdate<T>, props.fallbackLocale)
  languageConflict.value = false
}, { immediate: true })

function apply(next: LocalizedUpdate<T>): void {
  draft.value = next
}
function updateDraftSource(value: T): void {
  apply(updateSource(draft.value!, value, props.fallbackLocale))
}
function changeSourceLocale(code: string): void {
  if (!code) return
  const result = setSourceLocale(draft.value!, code)
  languageConflict.value = result.conflict
  apply(result.content)
}
function setBlock(code: string, element: Element | ComponentPublicInstance | null): void {
  if (element instanceof Element) blocks.set(code, element)
  else blocks.delete(code)
}
// 添加译文后聚焦这一块的第一个输入控件，管理员可以直接改写预填的原文。
async function add(code: string): Promise<void> {
  apply(addTranslation(draft.value!, code))
  await nextTick()
  blocks.get(code)?.querySelector<HTMLElement>('input:not([type="file"]), textarea')?.focus()
}
function save(): void {
  if (draft.value) emit('save', draft.value)
}
</script>
