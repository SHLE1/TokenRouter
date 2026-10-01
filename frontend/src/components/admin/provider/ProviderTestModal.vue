<template>
  <BaseDialog
    :show="show"
    :title="t('admin.providers.testDialog.title', { name: provider?.name ?? '' })"
    :subtitle="t('admin.providers.testDialog.subtitle')"
    width="extra-wide"
    :body-scroll="false"
    @close="handleClose"
  >
    <template #header-icon>
      <span
        v-if="provider"
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-950"
      >
        <PlatformIcon :platform="provider.platform" size="lg" :class="platformIconClass(provider.platform)" />
      </span>
    </template>

    <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto md:h-[36rem] md:flex-row md:overflow-hidden">
      <!-- 左侧：本次测试的参数，不写回提供商配置 -->
      <aside
        :aria-label="t('admin.providers.testDialog.settings')"
        class="flex shrink-0 flex-col gap-4 rounded-surface border border-gray-200 bg-gray-50/60 p-4 dark:border-dark-600 dark:bg-dark-950 md:w-72 md:overflow-y-auto"
      >
        <div class="text-xs font-semibold text-gray-900 dark:text-dark-50">
          {{ t('admin.providers.testDialog.settings') }}
        </div>

        <div v-if="imageTestAvailable" class="space-y-1.5">
          <div class="input-label">{{ t('admin.providers.testDialog.type') }}</div>
          <SettingsSegmented
            v-model="testType"
            :options="testTypeOptions"
            :aria-label="t('admin.providers.testDialog.type')"
            :disabled="running"
            block
          />
        </div>

        <div class="space-y-1.5">
          <label class="input-label" :for="modelFieldId">{{ t('admin.providers.testDialog.model') }}</label>
          <Select
            :id="modelFieldId"
            v-model="selectedModelId"
            :options="availableModels"
            :disabled="loadingModels || running"
            value-key="id"
            label-key="display_name"
            creatable
            :placeholder="loadingModels ? t('common.loading') : t('admin.providers.testDialog.modelPlaceholder')"
          />
        </div>

        <div v-if="testType === 'text'" class="space-y-1.5">
          <label class="input-label" :for="protocolFieldId">{{ t('admin.providers.testDialog.protocol') }}</label>
          <Select
            :id="protocolFieldId"
            v-model="testProtocol"
            :options="protocolOptions"
            :disabled="running || !protocolPlan.selectable || isCompactTestMode"
            :placeholder="t('admin.providers.testDialog.protocolNone')"
            data-testid="provider-test-protocol"
          />
          <p v-if="protocolHint" class="input-hint">{{ protocolHint }}</p>
        </div>

        <div v-if="isOpenAIProvider && testType === 'text'" class="space-y-1.5">
          <label class="input-label" :for="modeFieldId">{{ t('admin.providers.openai.testMode') }}</label>
          <Select
            :id="modeFieldId"
            v-model="testMode"
            :options="openAITestModeOptions"
            :disabled="running"
          />
        </div>

        <TextArea
          v-if="!isCompactTestMode"
          v-model="testPrompt"
          :label="promptInputLabel"
          :placeholder="promptInputPlaceholder"
          :disabled="running"
          data-testid="provider-test-prompt"
          rows="4"
        />
      </aside>

      <!-- 右侧：状态、耗时和上游回复 -->
      <section
        :aria-label="t('admin.providers.testDialog.results')"
        class="flex min-h-0 min-w-0 flex-1 flex-col gap-3"
      >
        <div class="flex shrink-0 items-center justify-between gap-2">
          <span
            role="status"
            aria-live="polite"
            data-testid="provider-test-status"
            :class="['inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium', statusToneClass]"
          >
            <span :class="['h-1.5 w-1.5 rounded-full bg-current', { 'animate-pulse': running }]"></span>
            {{ statusLabel }}
          </span>
          <HelpTooltip :content="t('admin.providers.testDialog.timingHint')" width-class="w-64" />
        </div>

        <div class="grid shrink-0 grid-cols-3 divide-x divide-gray-200 rounded-surface border border-gray-200 dark:divide-dark-600 dark:border-dark-600">
          <div v-for="metric in metrics" :key="metric.key" class="min-w-0 px-3 py-2.5">
            <div class="truncate text-xs text-gray-500 dark:text-dark-400">{{ metric.label }}</div>
            <div
              class="mt-1 truncate font-mono text-sm font-semibold tabular-nums text-gray-900 dark:text-dark-50"
              :title="metric.value"
            >
              {{ metric.value }}
            </div>
          </div>
        </div>

        <div class="flex min-h-64 flex-1 flex-col overflow-hidden rounded-surface border border-gray-200 dark:border-dark-600">
          <div class="flex h-11 shrink-0 items-center justify-between gap-2 border-b border-gray-200 px-3 dark:border-dark-600">
            <span class="truncate text-xs font-medium text-gray-600 dark:text-dark-300">
              {{ t('admin.providers.testDialog.output') }}
            </span>
            <div class="flex items-center gap-1">
              <button
                type="button"
                class="btn-icon-sm text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:cursor-not-allowed disabled:opacity-40 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-dark-100"
                :disabled="!copyText"
                :title="t('admin.providers.testDialog.copy')"
                :aria-label="t('admin.providers.testDialog.copy')"
                @click="copyOutput"
              >
                <Icon name="copy" size="sm" />
              </button>
              <div v-segmented class="segmented" role="radiogroup" :aria-label="t('admin.providers.testDialog.output')">
                <button
                  v-for="item in viewOptions"
                  :key="item.value"
                  type="button"
                  role="radio"
                  :aria-checked="outputView === item.value"
                  :class="['segmented-item px-2.5 py-1 text-xs', { 'segmented-item-active': outputView === item.value }]"
                  @click="outputView = item.value"
                >
                  {{ item.label }}
                </button>
              </div>
            </div>
          </div>

          <div ref="outputRef" class="min-h-0 flex-1 overflow-auto overscroll-contain p-4" data-testid="provider-test-output">
            <template v-if="outputView === 'reply'">
              <SettingsNotice v-if="status === 'error'" tone="error" class="mb-3 break-words">
                {{ errorMessage }}
              </SettingsNotice>

              <div
                v-if="replyText"
                class="whitespace-pre-wrap break-words text-sm leading-relaxed text-gray-800 dark:text-dark-100"
              >{{ replyText }}<span v-if="running" class="ml-0.5 inline-block h-4 w-1.5 animate-pulse bg-primary-500 align-text-bottom"></span></div>

              <div v-if="generatedImages.length > 0" class="mt-3 flex flex-wrap gap-3">
                <button
                  v-for="(image, index) in generatedImages"
                  :key="`${image.url}-${index}`"
                  type="button"
                  class="group/img relative overflow-hidden rounded-surface border border-gray-200 bg-white transition hover:border-black/20 dark:border-dark-600 dark:bg-dark-900 dark:hover:border-dark-500"
                  @click="previewImageUrl = image.url"
                >
                  <img
                    :src="image.url"
                    :alt="t('admin.providers.imagePreviewAlt', { index: index + 1 })"
                    class="max-h-64 w-full object-contain"
                  />
                  <span class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover/img:bg-black/20">
                    <Icon
                      name="eye"
                      size="lg"
                      class="text-white opacity-0 drop-shadow-lg transition-opacity group-hover/img:opacity-100"
                    />
                  </span>
                </button>
              </div>

              <div
                v-if="!replyText && generatedImages.length === 0 && status !== 'error'"
                class="flex h-full min-h-40 flex-col items-center justify-center gap-2 px-4 text-center"
              >
                <span class="mb-1 flex h-12 w-12 items-center justify-center rounded-surface bg-gray-100 dark:bg-dark-800">
                  <Icon
                    v-if="running"
                    name="loader"
                    size="lg"
                    class="animate-spin text-primary-500"
                    :animate-on-hover="false"
                  />
                  <Icon v-else name="beaker" size="lg" class="text-gray-400 dark:text-dark-500" />
                </span>
                <div class="text-sm font-medium text-gray-900 dark:text-dark-50">{{ emptyTitle }}</div>
                <p class="max-w-sm text-xs leading-relaxed text-gray-500 dark:text-dark-400">{{ emptyDescription }}</p>
              </div>
            </template>

            <template v-else>
              <div v-if="logLines.length > 0" class="space-y-1 font-mono text-xs leading-relaxed">
                <div
                  v-for="(line, index) in logLines"
                  :key="index"
                  :class="['break-words', LOG_TONE_CLASSES[line.tone]]"
                >
                  {{ line.text }}
                </div>
              </div>
              <p v-else class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.providers.testDialog.logEmpty') }}
              </p>
            </template>
          </div>
        </div>
      </section>
    </div>

    <!-- 图片灯箱 -->
    <Teleport to="body">
      <MotionTransition name="fade">
        <div
          v-if="previewImageUrl"
          class="fixed inset-0 z-tooltip flex items-center justify-center bg-[var(--overlay-bg-strong)] p-4"
          @click.self="previewImageUrl = ''"
        >
          <button
            type="button"
            class="absolute right-4 top-4 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
            :aria-label="t('common.close')"
            @click="previewImageUrl = ''"
          >
            <Icon name="x" size="lg" />
          </button>
          <img
            :src="previewImageUrl"
            :alt="t('admin.providers.imageLightboxAlt')"
            class="max-h-[90vh] max-w-[90vw] rounded-control object-contain shadow-2xl"
          />
        </div>
      </MotionTransition>
    </Teleport>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.close') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          data-testid="provider-test-start"
          :disabled="!canTest"
          @click="startTest"
        >
          <Icon v-if="running" name="loader" size="sm" class="animate-spin" :animate-on-hover="false" />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" />
          <Icon v-else name="refresh" size="sm" />
          {{ startButtonLabel }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import MotionTransition from '@/components/common/MotionTransition.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import SettingsNotice from '@/components/common/settings/SettingsNotice.vue'
import SettingsSegmented from '@/components/common/settings/SettingsSegmented.vue'
import { Icon } from '@/components/icons'
import { vSegmented } from '@/directives/segmented'
import { useClipboard } from '@/composables/useClipboard'
import { buildApiUrl } from '@/api/client'
import { ADMIN_UI_REQUEST_HEADER } from '@/api/adminUIRequest'
import { adminAPI } from '@/api/admin'
import { platformIconClass } from '@/utils/platformColors'
import type { Provider, ClaudeModel } from '@/types'
import {
  defaultProviderTestProtocol,
  providerTestProtocolPlan,
  type ProviderTestProtocol
} from './providerTestProtocols'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

type LogTone = 'info' | 'muted' | 'success' | 'error'

interface LogLine {
  text: string
  tone: LogTone
}

interface PreviewImage {
  url: string
  mimeType?: string
}

const LOG_TONE_CLASSES: Record<LogTone, string> = {
  info: 'text-primary-600 dark:text-primary-400',
  muted: 'text-gray-500 dark:text-dark-400',
  success: 'text-green-600 dark:text-green-400',
  error: 'text-red-600 dark:text-red-400'
}

const props = defineProps<{
  show: boolean
  provider: Provider | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const modelFieldId = 'provider-test-model'
const protocolFieldId = 'provider-test-protocol'
const modeFieldId = 'provider-test-mode'

const outputRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const running = computed(() => status.value === 'connecting')
const logLines = ref<LogLine[]>([])
const replyText = ref('')
const errorMessage = ref('')
const resolvedModel = ref('')
const firstTokenMs = ref<number | null>(null)
const totalMs = ref<number | null>(null)
let startedAt = 0
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
let lastDefaultPrompt = ''
const loadingModels = ref(false)
let abortController: AbortController | null = null
const generatedImages = ref<PreviewImage[]>([])
const previewImageUrl = ref('')
const outputView = ref<'reply' | 'log'>('reply')

const isOpenAIProvider = computed(() => props.provider?.platform === 'openai')
const isCNProvider = computed(() => ['kimi', 'zhipu', 'deepseek'].includes(props.provider?.platform ?? ''))

// 测试协议只作用于本次请求，不改写提供商配置。
const protocolPlan = computed(() => providerTestProtocolPlan(props.provider))
const testProtocol = ref<ProviderTestProtocol | 'native' | 'all'>('native')
const protocolOptions = computed(() => {
  const options: Array<{ value: typeof testProtocol.value; label: string }> = protocolPlan.value.options.map((item) => ({
    value: item.value,
    label: `${item.label} · ${item.path}`
  }))
  // 国产平台启用了多个协议时，保留按顺序全部验证的选项。
  if (isCNProvider.value && protocolPlan.value.selectable) {
    options.push({ value: 'all', label: t('admin.providers.testDialog.protocolAll') })
  }
  return options
})
// 弹窗打开期间切换提供商时，原协议不在新列表里就回到默认值。
watch(protocolPlan, (plan) => {
  if (!protocolOptions.value.some((item) => item.value === testProtocol.value)) {
    testProtocol.value = defaultProviderTestProtocol(props.provider, plan)
  }
})
const protocolHint = computed(() => {
  if (isCompactTestMode.value) return t('admin.providers.testDialog.protocolCompact')
  if (protocolPlan.value.options.length === 0) return ''
  if (!protocolPlan.value.selectable) return t('admin.providers.testDialog.protocolFixed')
  return ''
})

const testMode = ref<'default' | 'compact' | 'legacy_compact'>('default')
const testType = ref<'text' | 'image'>('text')
// Compact 连接测试使用固定载荷和 Responses 端点，不显示可编辑提示词。
const isCompactTestMode = computed(() => isOpenAIProvider.value && testMode.value !== 'default')
watch(testMode, (mode) => {
  if (mode !== 'default' && protocolPlan.value.selectable) testProtocol.value = 'responses'
})
const openAITestModeOptions = computed(() => [
  { value: 'default', label: t('admin.providers.openai.testModeDefault') },
  { value: 'compact', label: t('admin.providers.openai.testModeCompact') },
  { value: 'legacy_compact', label: t('admin.providers.openai.testModeLegacyCompact') }
])

// 实际发给后端的协议：固定端点、全部协议和图片测试都不携带该字段。
const requestProtocol = computed<ProviderTestProtocol | undefined>(() => {
  if (testType.value !== 'text' || !protocolPlan.value.selectable) return undefined
  if (isCompactTestMode.value) return 'responses'
  if (testProtocol.value === 'native' || testProtocol.value === 'all') return undefined
  return testProtocol.value
})

const prioritizedGeminiModels = ['gemini-3.1-flash-image', 'gemini-2.5-flash-image', 'gemini-3.5-flash', 'gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-3-flash-preview', 'gemini-3-pro-preview', 'gemini-2.0-flash']

// 图片/文字请求类型完全由管理员选择，不再从模型名称推断。
const imageTestAvailable = computed(() => {
  const platform = props.provider?.platform
  return platform === 'openai' || platform === 'gemini' || platform === 'grok' ||
    (platform === 'antigravity' && props.provider?.type === 'apikey')
})

const testTypeOptions = computed(() => [
  { value: 'text' as const, label: t('admin.providers.testDialog.typeText'), icon: 'modalityText' as const },
  { value: 'image' as const, label: t('admin.providers.testDialog.typeImage'), icon: 'modalityImage' as const }
])

const viewOptions = computed(() => [
  { value: 'reply' as const, label: t('admin.providers.testDialog.viewReply') },
  { value: 'log' as const, label: t('admin.providers.testDialog.viewLog') }
])

const promptInputLabel = computed(() =>
  testType.value === 'image'
    ? t('admin.providers.imagePromptLabel')
    : t('admin.providers.textPromptLabel')
)
const promptInputPlaceholder = computed(() =>
  testType.value === 'image'
    ? t('admin.providers.imagePromptPlaceholder')
    : t('admin.providers.textPromptPlaceholder')
)

const canTest = computed(() => !running.value && !!selectedModelId.value)

const statusLabel = computed(() => {
  switch (status.value) {
    case 'connecting':
      return t('admin.providers.testDialog.statusRunning')
    case 'success':
      return t('admin.providers.testDialog.statusSuccess')
    case 'error':
      return t('admin.providers.testDialog.statusFailed')
    default:
      return t('admin.providers.testDialog.statusReady')
  }
})

const statusToneClass = computed(() => {
  switch (status.value) {
    case 'connecting':
      return 'bg-primary-500/10 text-primary-600 dark:text-primary-400'
    case 'success':
      return 'bg-green-500/10 text-green-600 dark:text-green-400'
    case 'error':
      return 'bg-red-500/10 text-red-600 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300'
  }
})

const formatSeconds = (value: number | null) => (value == null ? '—' : `${(value / 1000).toFixed(2)} s`)

const metrics = computed(() => [
  { key: 'model', label: t('admin.providers.testDialog.metricModel'), value: resolvedModel.value || '—' },
  { key: 'first', label: t('admin.providers.testDialog.metricFirstToken'), value: formatSeconds(firstTokenMs.value) },
  { key: 'total', label: t('admin.providers.testDialog.metricTotal'), value: formatSeconds(totalMs.value) }
])

const emptyTitle = computed(() => {
  if (running.value) return t('admin.providers.testDialog.waitingTitle')
  if (status.value === 'success') return t('admin.providers.testDialog.noContentTitle')
  return t('admin.providers.testDialog.emptyTitle')
})
const emptyDescription = computed(() => {
  if (running.value) return t('admin.providers.testDialog.waitingDescription')
  if (status.value === 'success') return t('admin.providers.testDialog.noContentDescription')
  return t('admin.providers.testDialog.emptyDescription')
})

const startButtonLabel = computed(() => {
  if (running.value) return t('admin.providers.testDialog.running')
  if (status.value === 'idle') return t('admin.providers.startTest')
  return t('admin.providers.testDialog.rerun')
})

const copyText = computed(() =>
  outputView.value === 'log'
    ? logLines.value.map((line) => line.text).join('\n')
    : replyText.value
)

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((a, b) => {
    const aPriority = priorityMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
    const bPriority = priorityMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
    return aPriority - bPriority
  })
}

// 打开弹窗时重置参数并加载可测试模型。
watch(
  () => props.show,
  async (open) => {
    if (open && props.provider) {
      testPrompt.value = ''
      lastDefaultPrompt = ''
      testMode.value = 'default'
      testType.value = 'text'
      testProtocol.value = defaultProviderTestProtocol(props.provider, protocolPlan.value)
      outputView.value = 'reply'
      resetState()
      await loadAvailableModels()
    } else {
      abortStream()
    }
  }
)

// 提示词未被手动修改时，跟随测试类型切换默认值。
watch([selectedModelId, testType], () => {
  const nextDefaultPrompt = testType.value === 'image'
    ? t('admin.providers.imagePromptDefault')
    : t('admin.providers.textPromptDefault')
  if (!testPrompt.value.trim() || testPrompt.value === lastDefaultPrompt) {
    testPrompt.value = nextDefaultPrompt
    lastDefaultPrompt = nextDefaultPrompt
  }
})

watch(testType, (nextType) => {
  if (nextType === 'image') {
    testMode.value = 'default'
  }
})

const loadAvailableModels = async () => {
  if (!props.provider) return

  loadingModels.value = true
  selectedModelId.value = ''
  try {
    const models = await adminAPI.providers.getAvailableModels(props.provider.id)
    availableModels.value = props.provider.platform === 'gemini' || props.provider.platform === 'antigravity'
      ? sortTestModels(models)
      : models
    if (availableModels.value.length > 0) {
      if (props.provider.platform === 'gemini') {
        selectedModelId.value = availableModels.value[0].id
      } else {
        // 优先选中 Sonnet，没有时取第一个模型。
        const sonnetModel = availableModels.value.find((m) => m.id.includes('sonnet'))
        selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
      }
    }
  } catch (error) {
    console.error('Failed to load available models:', error)
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  logLines.value = []
  replyText.value = ''
  errorMessage.value = ''
  resolvedModel.value = ''
  firstTokenMs.value = null
  totalMs.value = null
  generatedImages.value = []
  previewImageUrl.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

const addLine = (text: string, tone: LogTone = 'muted') => {
  logLines.value.push({ text, tone })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (outputRef.value) {
    outputRef.value.scrollTop = outputRef.value.scrollHeight
  }
}

const elapsed = () => Math.round(performance.now() - startedAt)

// markFirstToken 只记录首段文字或首张图片到达的时间。
const markFirstToken = () => {
  if (firstTokenMs.value == null) firstTokenMs.value = elapsed()
}

// finish 只处理第一次结束事件，后端在错误后补发的完成事件不再改写结果。
const finish = (nextStatus: 'success' | 'error', message = '') => {
  if (!running.value) return
  status.value = nextStatus
  totalMs.value = elapsed()
  if (nextStatus === 'error') {
    errorMessage.value = message
    addLine(t('admin.providers.errorPrefix', { message }), 'error')
  } else {
    addLine(t('admin.providers.testCompleted'), 'success')
  }
}

const startTest = async () => {
  if (!props.provider || !selectedModelId.value || running.value) return

  resetState()
  status.value = 'connecting'
  startedAt = performance.now()
  addLine(t('admin.providers.startingTestForProvider', { name: props.provider.name }), 'info')

  abortStream()
  abortController = new AbortController()

  try {
    const requestBody: {
      model_id: string
      prompt: string
      test_type: 'text' | 'image'
      mode?: 'default' | 'compact' | 'legacy_compact'
      protocol?: ProviderTestProtocol
    } = {
      model_id: selectedModelId.value,
      prompt: isCompactTestMode.value ? '' : testPrompt.value.trim(),
      test_type: testType.value
    }
    if (isOpenAIProvider.value) {
      requestBody.mode = testMode.value
    }
    if (requestProtocol.value) {
      requestBody.protocol = requestProtocol.value
    }

    // SSE 测试接口用 POST，只能走 fetch，必须显式套用配置的 API base。
    const url = buildApiUrl(`/admin/providers/${props.provider.id}/test`)
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json',
        [ADMIN_UI_REQUEST_HEADER]: '1'
      },
      body: JSON.stringify(requestBody),
      signal: abortController.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const jsonStr = line.slice(6).trim()
        if (!jsonStr) continue
        try {
          handleEvent(JSON.parse(jsonStr))
        } catch (e) {
          console.error('Failed to parse SSE event:', e)
        }
      }
    }

    // 连接提前结束且没有完成事件时，按失败处理，避免状态停在测试中。
    if (running.value) {
      finish('error', t('admin.providers.testDialog.streamEnded'))
    }
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      status.value = 'idle'
      return
    }
    finish('error', error instanceof Error ? error.message : 'Unknown error')
  }
}

const handleEvent = (event: {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.providers.connectedToApi'), 'success')
      if (event.model) {
        resolvedModel.value = event.model
        addLine(t('admin.providers.usingModel', { model: event.model }))
      }
      addLine(
        testType.value === 'image'
          ? t('admin.providers.sendingImageRequest')
          : t('admin.providers.sendingTestMessage')
      )
      break

    case 'content':
      if (event.text) {
        markFirstToken()
        replyText.value += event.text
        if (outputView.value === 'reply') scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        markFirstToken()
        generatedImages.value.push({ url: event.image_url, mimeType: event.mime_type })
        addLine(t('admin.providers.imageReceived', { count: generatedImages.value.length }), 'info')
      }
      break

    case 'status':
      if (event.text) {
        addLine(event.text, 'info')
      }
      break

    case 'test_complete':
      if (event.success) {
        finish('success')
      } else {
        finish('error', event.error || 'Test failed')
      }
      break

    case 'error':
      finish('error', event.error || 'Unknown error')
      break
  }
}

const copyOutput = () => {
  if (!copyText.value) return
  copyToClipboard(copyText.value, t('admin.providers.outputCopied'))
}
</script>
