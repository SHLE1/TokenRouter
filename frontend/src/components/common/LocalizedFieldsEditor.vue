<template>
  <LocalizedEditor :model-value="modelValue || originalContent(source, sourceLocale)" @update:model-value="emit('update:modelValue', $event)">
    <template #default="{ value, update }">
      <div v-for="field in fields" :key="field.key" class="space-y-1">
        <label class="input-label">{{ field.label }}</label>
        <textarea dir="auto" :aria-label="field.label" v-if="field.multiline" :value="value[field.key]" class="input" rows="3" @input="update({ ...value, [field.key]: ($event.target as HTMLTextAreaElement).value })"></textarea>
        <input dir="auto" :aria-label="field.label" v-else :value="value[field.key]" :type="field.url ? 'url' : 'text'" class="input" @input="update({ ...value, [field.key]: ($event.target as HTMLInputElement).value })" />
      </div>
    </template>
  </LocalizedEditor>
</template>

<script setup lang="ts" generic="T extends Record<string, string>">
import LocalizedEditor from '@/components/common/LocalizedEditor.vue'
import { originalContent, type LocalizedUpdate } from '@/i18n/content'

// 字段定义来自业务表单，译文的编辑、预览和核对共用一个实现。
defineProps<{ modelValue?: LocalizedUpdate<T>; source: T; sourceLocale?: string | null; fields: { key: keyof T & string; label: string; multiline?: boolean; url?: boolean }[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: LocalizedUpdate<T>] }>()
</script>
