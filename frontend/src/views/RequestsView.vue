<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import RequestIdSearch from '@/components/common/RequestIdSearch.vue'
import RequestIdLink from '@/components/common/RequestIdLink.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import OpsErrorDetailModal from '@/views/admin/ops/components/OpsErrorDetailModal.vue'
import ContentSkeleton from '@/components/common/ContentSkeleton.vue'
import { findRequests, type RequestDetail } from '@/api/requests'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const requestId = ref('')
const selectedError = ref<number | null>(null)
const showError = ref(false)
const loading = ref(false)
const searched = ref(false)
const error = ref('')
const items = ref<RequestDetail[]>([])
const hasMore = ref(false)
const admin = computed(() => route.path.startsWith('/admin'))
let controller: AbortController | undefined

async function load() {
  controller?.abort()
  const current = new AbortController()
  controller = current
  const id = requestId.value.trim()
  if (!id) {
    items.value = []
    searched.value = false
    loading.value = false
    error.value = ''
    hasMore.value = false
    return
  }
  loading.value = true
  items.value = []
  hasMore.value = false
  error.value = ''
  try {
    const response = await findRequests(id, admin.value, current.signal)
    if (current.signal.aborted) return
    items.value = response.items || []
    hasMore.value = response.has_more
    searched.value = true
  } catch (cause: unknown) {
    if (current.signal.aborted) return
    items.value = []
    const message = (cause as { message?: unknown })?.message
    error.value = typeof message === 'string' ? message : t('requests.loadFailed')
  } finally {
    if (controller === current) loading.value = false
  }
}

async function search() {
  const id = requestId.value.trim()
  if (id === route.query.request_id) {
    await load()
    return
  }
  await router.replace({ query: { request_id: id || undefined } })
}

function openError(id: number) {
  selectedError.value = id
  showError.value = true
}

watch(() => [route.path, route.query.request_id], () => {
  requestId.value = typeof route.query.request_id === 'string' ? route.query.request_id : ''
  void load()
}, { immediate: true })
onBeforeUnmount(() => controller?.abort())
</script>

<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="card p-4"><RequestIdSearch v-model="requestId" :show-details="false" @search="search" /></div>
      <ContentSkeleton v-if="loading" />
      <p v-else-if="error" role="alert" class="text-red-600 dark:text-red-400">{{ error }}</p>
      <p v-else-if="searched && !items.length" class="card p-6 text-sm text-gray-500 dark:text-dark-300">{{ t('requests.notFound') }}</p>
      <p v-if="hasMore" class="text-sm text-amber-600 dark:text-amber-400">{{ t('requests.more') }}</p>
      <article v-for="item in items" :key="`${item.request_id}:${item.usage?.[0]?.id || ''}:${item.errors?.[0]?.id || ''}:${item.audit_ids?.[0] || ''}`" class="card space-y-4 p-6">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <RequestIdLink :value="item.request_id" />
          <span class="text-sm">{{ t(`requests.states.${item.state}`, item.state) }} <span v-if="item.status_code">({{ item.status_code }})</span></span>
        </div>
        <p v-if="item.error_code" class="text-sm text-red-600 dark:text-red-400">{{ t(`requests.errorCodes.${item.error_code}`, item.error_code) }}</p>
        <p v-if="item.legacy" class="text-sm text-amber-600 dark:text-amber-400">{{ t('requests.legacy') }}</p>
        <p v-if="item.pending" class="text-sm text-amber-600 dark:text-amber-400">{{ t('requests.pending') }}</p>
        <dl class="grid grid-cols-1 gap-4 text-sm sm:grid-cols-2 lg:grid-cols-3">
          <div><dt class="text-gray-500 dark:text-dark-300">{{ t('requests.endpoint') }}</dt><dd class="break-all">{{ item.method }} {{ item.path || '-' }}</dd></div>
          <div><dt class="text-gray-500 dark:text-dark-300">{{ t('requests.startedAt') }}</dt><dd>{{ new Date(item.started_at).toLocaleString() }}</dd></div>
          <div><dt class="text-gray-500 dark:text-dark-300">{{ t('requests.duration') }}</dt><dd>{{ item.duration_ms ?? '-' }} ms</dd></div>
          <div v-if="item.model"><dt class="text-gray-500 dark:text-dark-300">{{ t('usage.model') }}</dt><dd>{{ item.model }}</dd></div>
          <div v-if="item.parent_request_id"><dt class="text-gray-500 dark:text-dark-300">{{ t('requests.parent') }}</dt><dd><RequestIdLink :value="item.parent_request_id" /></dd></div>
        </dl>
        <div v-if="item.attempts?.length" class="space-y-2">
          <h2 class="text-sm font-medium">{{ t('requests.attempts') }}</h2>
          <div v-for="attempt in item.attempts" :key="attempt.number" class="flex flex-wrap items-center gap-2 rounded-surface bg-gray-50 p-4 text-sm dark:bg-dark-800">
            <span>#{{ attempt.number }}</span><span v-if="attempt.provider_id">{{ t('requests.provider') }} #{{ attempt.provider_id }}</span>
            <span>{{ attempt.status_code ? `HTTP ${attempt.status_code}` : t(`requests.states.${attempt.outcome}`, attempt.outcome || '-') }}</span><span v-if="attempt.duration_ms != null">{{ attempt.duration_ms }} ms</span>
            <span v-if="attempt.upstream_request_id" class="break-all font-mono text-xs">{{ attempt.upstream_request_id }}</span>
          </div>
        </div>
        <div v-if="item.timings && Object.keys(item.timings).length" class="space-y-2">
          <h2 class="text-sm font-medium">{{ t('requests.timings') }}</h2>
          <dl class="grid grid-cols-1 gap-2 text-sm sm:grid-cols-2"><div v-for="(value, name) in item.timings" :key="name"><dt class="text-gray-500 dark:text-dark-300">{{ t(`requests.timingNames.${name}`, String(name)) }}</dt><dd>{{ value }} ms</dd></div></dl>
        </div>
        <div v-if="item.usage?.length" class="space-y-2">
          <h2 class="text-sm font-medium">{{ t('requests.usage') }}</h2>
          <p v-for="usage in item.usage" :key="usage.id" class="text-sm">{{ usage.model }} · {{ t('requests.tokens', { input: usage.input_tokens, output: usage.output_tokens }) }} · ${{ usage.actual_cost.toFixed(8) }}</p>
        </div>
        <div v-if="item.errors?.length" class="space-y-2">
          <h2 class="text-sm font-medium">{{ t('requests.errors') }}</h2>
          <button v-for="failure in item.errors" :key="failure.id" type="button" class="btn btn-secondary" @click="openError(failure.id)">#{{ failure.id }} · HTTP {{ failure.status_code }} · {{ failure.phase }}</button>
        </div>
        <details v-if="admin && item.aliases?.length" class="text-sm"><summary class="cursor-pointer text-primary-600 dark:text-primary-400">{{ t('requests.aliases') }}</summary><ul class="mt-2 space-y-2"><li v-for="alias in item.aliases" :key="`${alias.kind}:${alias.value}`" class="break-all">{{ t(`requests.aliasKinds.${alias.kind}`, alias.kind) }}: <span class="font-mono text-xs">{{ alias.value }}</span></li></ul></details>
      </article>
    </div>
    <OpsErrorDetailModal v-if="admin" v-model:show="showError" :error-id="selectedError" error-type="request" />
    <UserErrorDetailModal v-else v-model:show="showError" :error-id="selectedError" />
  </AppLayout>
</template>
