<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center gap-2">
      <button type="button" class="btn btn-sm" :class="selected === 'source' ? 'btn-primary' : 'btn-secondary'" :aria-pressed="selected === 'source'" @click="selected = 'source'">
        {{ t('localization.original') }}
      </button>
      <button v-for="code in Object.keys(modelValue.translations)" :key="code" type="button" class="btn btn-sm" :class="selected === code ? 'btn-primary' : 'btn-secondary'" :aria-pressed="selected === code" @click="selected = code">
        {{ languageName(code) }}
        <span v-if="isStale(code)">{{ t('localization.reviewNeeded') }}</span>
      </button>
      <Select v-if="remainingLanguages.length" :model-value="''" :options="remainingLanguages" :placeholder="t('localization.addTranslation')" @update:model-value="addTranslation(String($event))" />
    </div>
    <div v-if="selected === 'source'">
      <label class="input-label">{{ t('localization.originalLanguage') }}</label>
      <Select :model-value="originalLanguage" :options="languageOptions" :placeholder="t('localization.chooseLanguage')" @update:model-value="setOriginalLanguage(String($event))" />
    </div>
    <p v-if="languageConflict" class="input-hint">{{ t('localization.languageConflict') }}</p>
    <slot :value="currentValue" :update="updateValue">
      <input dir="auto" v-if="rows === 1" :aria-label="selected === 'source' ? t('localization.original') : languageName(selected)" :value="String(currentValue ?? '')" class="input" @input="updateValue(($event.target as HTMLInputElement).value as T)" />
      <textarea dir="auto" v-else :aria-label="selected === 'source' ? t('localization.original') : languageName(selected)" :value="String(currentValue ?? '')" :rows="rows" class="input resize-y" @input="updateValue(($event.target as HTMLTextAreaElement).value as T)" />
    </slot>
    <div v-if="selected !== 'source'" class="flex flex-wrap items-center gap-2">
      <button type="button" class="btn btn-secondary btn-sm" @click="reviewTranslation">{{ t('localization.confirmReviewed') }}</button>
      <button type="button" class="btn btn-secondary btn-sm" @click="useAsOriginal">{{ t('localization.useAsOriginal') }}</button>
      <button type="button" class="btn btn-secondary btn-sm" @click="removeTranslation">{{ t('localization.removeTranslation') }}</button>
    </div>
    <p class="input-hint">{{ t('localization.fallbackHint') }}</p>
    <details>
      <summary class="cursor-pointer text-sm text-gray-500 dark:text-dark-400">{{ t('localization.preview') }}</summary>
      <Select v-model="previewLanguage" :options="languageOptions" />
      <p v-if="preview.fallback" class="input-hint">{{ t('localization.showingOriginal') }}</p>
      <slot name="preview" :value="preview.value">
        <pre class="mt-2 whitespace-pre-wrap break-words text-sm">{{ typeof preview.value === 'string' ? preview.value : JSON.stringify(preview.value, null, 2) }}</pre>
      </slot>
    </details>
  </div>
</template>

<script setup lang="ts" generic="T">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import { availableLocales, defaultLocale, normalizeLocale } from '@/i18n/catalog'
import { resolveContent, type LocalizedUpdate } from '@/i18n/content'

const props = withDefaults(defineProps<{ modelValue: LocalizedUpdate<T>; rows?: number; defaultSourceLocale?: string }>(), { rows: 1 })
const emit = defineEmits<{ 'update:modelValue': [value: LocalizedUpdate<T>] }>()
const { t, locale } = useI18n()
const selected = ref('source')
const languageConflict = ref(false)
const previewLanguage = ref<string>(normalizeLocale(locale?.value) || defaultLocale)
const languageOptions = availableLocales.map(item => ({ value: item.code, label: item.name }))
const remainingLanguages = computed(() => languageOptions.filter(item => item.value !== originalLanguage.value && !(item.value in props.modelValue.translations)))
const originalLanguage = computed(() => props.modelValue.source_locale || (props.modelValue.revision === 0 ? props.defaultSourceLocale : '') || '')
const currentValue = computed(() => selected.value === 'source' ? props.modelValue.source : props.modelValue.translations[selected.value]?.value)
const preview = computed(() => resolveContent(props.modelValue, previewLanguage.value))

// 草稿按值复制，编辑一个语言时其他语言的表单状态保持完整。
function draft(): LocalizedUpdate<T> {
  const next = JSON.parse(JSON.stringify(props.modelValue)) as LocalizedUpdate<T>
  if (next.revision === 0 && !next.source_locale && props.defaultSourceLocale) next.source_locale = props.defaultSourceLocale
  return next
}
function languageName(code: string): string {
  return availableLocales.find(item => item.code === code)?.name || code
}
function isStale(code: string): boolean {
  return !props.modelValue.reviewed_locales?.includes(code) && props.modelValue.translations[code]?.source_revision !== props.modelValue.source_revision
}
function setOriginalLanguage(code: string): void {
  const next = draft()
  if (code in next.translations) { languageConflict.value = true; return }
  languageConflict.value = false
  next.source_locale = code
  next.reviewed_locales = []
  for (const translation of Object.values(next.translations)) translation.source_revision = -1
  emit('update:modelValue', next)
}
function addTranslation(code: string): void {
  if (!code) return
  const next = draft()
  next.deleted_locales = next.deleted_locales?.filter(item => item !== code)
  next.translations[code] = { value: JSON.parse(JSON.stringify(next.source)) as T, source_revision: -1 }
  selected.value = code
  emit('update:modelValue', next)
}
function updateValue(value: T): void {
  const next = draft()
  if (selected.value === 'source') {
    next.source = value
    next.reviewed_locales = []
    for (const translation of Object.values(next.translations)) translation.source_revision = -1
  } else {
    next.translations[selected.value] = { value, source_revision: -1 }
    next.reviewed_locales = next.reviewed_locales?.filter(code => code !== selected.value)
  }
  emit('update:modelValue', next)
}
function reviewTranslation(): void {
  const next = draft()
  next.reviewed_locales = [...new Set([...(next.reviewed_locales || []), selected.value])]
  next.translations[selected.value].source_revision = next.source_revision
  emit('update:modelValue', next)
}
// 设为原文会替换当前原文，已知语言的旧原文转入译文列表。
function useAsOriginal(): void {
  const next = draft()
  const code = selected.value
  const translation = next.translations[code]
  if (!translation) return
  if (next.source_locale) next.translations[next.source_locale] = { value: next.source, source_revision: -1 }
  next.source = translation.value
  next.source_locale = code
  delete next.translations[code]
  next.deleted_locales = [...new Set([...(next.deleted_locales || []), code])]
  for (const entry of Object.values(next.translations)) entry.source_revision = -1
  next.reviewed_locales = []
  selected.value = 'source'
  languageConflict.value = false
  emit('update:modelValue', next)
}
function removeTranslation(): void {
  const next = draft()
  delete next.translations[selected.value]
  next.deleted_locales = [...new Set([...(next.deleted_locales || []), selected.value])]
  next.reviewed_locales = next.reviewed_locales?.filter(code => code !== selected.value)
  selected.value = 'source'
  emit('update:modelValue', next)
}
</script>
