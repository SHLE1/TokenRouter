<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ value?: string | null }>()
const route = useRoute()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copied = ref(false)
let timeout: ReturnType<typeof setTimeout> | undefined
async function copy() {
  const value = props.value
  if (!value || !await copyToClipboard(value, t('admin.usage.requestIdCopied')) || props.value !== value) return
  copied.value = true
  clearTimeout(timeout)
  timeout = setTimeout(() => { copied.value = false }, 2000)
}
watch(() => props.value, () => {
  copied.value = false
  clearTimeout(timeout)
})
onBeforeUnmount(() => clearTimeout(timeout))
const destination = computed(() => ({ path: route?.path.startsWith('/admin') ? '/admin/requests' : '/requests', query: { request_id: props.value || '' } }))
</script>

<template>
  <div v-if="value" class="flex min-w-0 items-center gap-2">
    <RouterLink :to="destination" :title="value" class="min-w-0 truncate font-mono text-xs text-primary-600 hover:underline dark:text-primary-400" @click.stop>{{ value }}</RouterLink>
    <button type="button" class="btn-icon-sm shrink-0 text-gray-500" :aria-label="t('keys.copyToClipboard')" :title="copied ? t('keys.copied') : t('keys.copyToClipboard')" @click.stop="copy">
      <Icon :name="copied ? 'check' : 'copy'" size="sm" />
    </button>
  </div>
  <span v-else class="text-gray-400">-</span>
</template>
