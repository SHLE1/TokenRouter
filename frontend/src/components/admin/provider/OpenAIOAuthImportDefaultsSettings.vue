<template>
  <SettingsCard
    id="openai-oauth-import-defaults"
    :title="t('admin.providers.openAIOAuthImportDefaultsTitle')"
    :description="t('admin.providers.openAIOAuthImportDefaultsDescription')"
  >
    <ContentSkeleton v-if="loading" variant="form" :rows="5" />

    <template v-else>
      <SettingsSection :title="t('admin.providers.openAIOAuthImportDefaultsProvider')">
        <div>
          <label :for="`${uid}-notes`" class="input-label">{{ t('admin.providers.notes') }}</label>
          <textarea :id="`${uid}-notes`" v-model="form.notes" rows="2" class="input"></textarea>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label :for="`${uid}-concurrency`" class="input-label">{{ t('admin.providers.concurrency') }}</label>
            <input :id="`${uid}-concurrency`" v-model="form.concurrency" type="number" min="0" step="1" class="input" />
          </div>
          <div>
            <label :for="`${uid}-priority`" class="input-label">{{ t('admin.providers.priority') }}</label>
            <input :id="`${uid}-priority`" v-model="form.priority" type="number" min="0" step="1" class="input" />
          </div>
          <div>
            <label :for="`${uid}-rate-multiplier`" class="input-label">{{ t('admin.providers.billingRateMultiplier') }}</label>
            <input :id="`${uid}-rate-multiplier`" v-model="form.rateMultiplier" type="number" min="0" step="0.01" class="input" />
          </div>
          <div>
            <label :for="`${uid}-expires-at`" class="input-label">{{ t('admin.providers.expiresAt') }}</label>
            <input :id="`${uid}-expires-at`" v-model="form.expiresAt" type="number" min="0" step="1" class="input" />
          </div>
        </div>
        <SettingRow
          :id="`${uid}-auto-pause-expired`"
          field
          :label-for="`${uid}-auto-pause-expired`"
          :label="t('admin.providers.autoPauseOnExpired')"
        >
          <Select
            :id="`${uid}-auto-pause-expired`"
            v-model="form.autoPauseOnExpired"
            :options="autoPauseOnExpiredOptions"
          />
        </SettingRow>
      </SettingsSection>

      <SettingsSection :title="t('admin.providers.sections.openaiCompatibility')">
        <SettingToggleRow
          :id="`${uid}-passthrough`"
          v-model="openaiPassthrough"
          :label="t('admin.providers.openai.oauthPassthrough')"
          :hint="t('admin.providers.openai.oauthPassthroughDesc')"
        />
        <CodexImageToolModeSelector
          v-model="codexImageToolMode"
          test-id-prefix="openai-oauth-default-codex-image-tool"
        />
        <SettingRow
          :id="`${uid}-ws-mode`"
          field
          :label-for="`${uid}-ws-mode`"
          :label="t('admin.providers.openai.wsMode')"
          :hint="t('admin.providers.openai.wsModeDesc')"
        >
          <Select :id="`${uid}-ws-mode`" v-model="wsMode" :options="wsModeOptions" />
        </SettingRow>
      </SettingsSection>

      <SettingsSection :title="t('admin.providers.sections.openaiClient')">
        <SettingRow
          :id="`${uid}-client-policy`"
          field
          :label-for="`${uid}-client-policy`"
          :label="t('admin.providers.openai.clientPolicy')"
          :hint="t('admin.providers.openai.clientPolicyDesc')"
        >
          <Select
            :id="`${uid}-client-policy`"
            v-model="openAIOAuthClientPolicy"
            :options="openAIOAuthClientPolicyOptions"
            data-testid="openai-oauth-default-client-policy"
          />
        </SettingRow>
        <Collapse :open="openAIOAuthClientPolicy === 'codex_only'" unmount-on-hide>
          <SettingsSubpanel>
            <SettingToggleRow
              :id="`${uid}-codex-allow-claude-code`"
              v-model="codexCLIOnlyAllowClaudeCode"
              :label="t('admin.providers.openai.codexCLIOnlyAllowClaudeCode')"
              :hint="t('admin.providers.openai.codexCLIOnlyAllowClaudeCodeDesc')"
              testid="openai-oauth-default-codex-allow-claude-code-toggle"
            />
          </SettingsSubpanel>
        </Collapse>
      </SettingsSection>

      <SettingsSection :title="t('admin.providers.sections.autoPause')">
        <SettingToggleRow
          :id="`${uid}-auto-pause-5h-disabled`"
          v-model="autoPause5hDisabled"
          :label="t('admin.providers.autoPause5hDisabled')"
          :hint="t('admin.providers.autoPauseDisabledHint')"
          testid="openai-oauth-default-auto-pause-5h-disabled"
        />
        <SettingToggleRow
          :id="`${uid}-auto-pause-7d-disabled`"
          v-model="autoPause7dDisabled"
          :label="t('admin.providers.autoPause7dDisabled')"
          :hint="t('admin.providers.autoPauseDisabledHint')"
          testid="openai-oauth-default-auto-pause-7d-disabled"
        />
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label :for="`${uid}-auto-pause-5h`" class="input-label">{{ t('admin.providers.autoPause5hThreshold') }}</label>
            <input
              :id="`${uid}-auto-pause-5h`"
              v-model="autoPause5hThreshold"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              :disabled="autoPause5hDisabled"
              data-testid="openai-oauth-default-auto-pause-5h-threshold"
            />
            <p class="input-hint">{{ t('admin.providers.autoPauseThresholdHint') }}</p>
          </div>
          <div>
            <label :for="`${uid}-auto-pause-7d`" class="input-label">{{ t('admin.providers.autoPause7dThreshold') }}</label>
            <input
              :id="`${uid}-auto-pause-7d`"
              v-model="autoPause7dThreshold"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              :disabled="autoPause7dDisabled"
              data-testid="openai-oauth-default-auto-pause-7d-threshold"
            />
            <p class="input-hint">{{ t('admin.providers.autoPauseThresholdHint') }}</p>
          </div>
        </div>
      </SettingsSection>

      <SettingsSection :title="t('admin.providers.sections.compaction')">
        <OpenAICompactionToggle
          v-model="nativeCompactV2Mode"
          test-id="openai-oauth-default-native-compaction-v2-mode"
          :label="t('admin.providers.openai.nativeCompactV2Mode')"
          :hint="t('admin.providers.openai.nativeCompactV2ModeDesc')"
        />
        <OpenAICompactionToggle
          v-model="compactMode"
          test-id="openai-oauth-default-compact-mode"
          :label="t('admin.providers.openai.compactMode')"
          :hint="t('admin.providers.openai.compactModeDesc')"
        />
      </SettingsSection>

      <TLSFingerprintFields
        v-model:enabled="tlsFingerprintEnabled"
        v-model:profile-id="tlsFingerprintProfileId"
        v-model:router-id="tlsFingerprintRouterId"
        :profile-options="tlsFingerprintProfileOptions"
        :router-options="tlsFingerprintRouterOptions"
        test-id-prefix="openai-oauth-default-tls-fingerprint"
      />

      <SettingsSection :title="t('admin.providers.modelWhitelist')">
        <ModelWhitelistSelector v-model="defaultAllowedModels" platform="openai" />
      </SettingsSection>

      <SettingsSection>
        <ProviderModelMappingEditor
          v-model="defaultModelMappings"
          :title="t('admin.providers.modelMapping')"
          :presets="presetMappings"
          @preset="addDefaultPresetMapping"
        />
      </SettingsSection>

      <SettingsSection>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label :for="`${uid}-credentials-json`" class="input-label">{{ t('admin.providers.openAIOAuthImportDefaultsCredentialsJson') }}</label>
            <textarea
              :id="`${uid}-credentials-json`"
              v-model="credentialsJson"
              rows="8"
              class="input font-mono text-xs"
              spellcheck="false"
            ></textarea>
          </div>
          <div>
            <label :for="`${uid}-extra-json`" class="input-label">{{ t('admin.providers.openAIOAuthImportDefaultsExtraJson') }}</label>
            <textarea
              :id="`${uid}-extra-json`"
              v-model="extraJson"
              rows="8"
              class="input font-mono text-xs"
              spellcheck="false"
            ></textarea>
          </div>
        </div>
      </SettingsSection>
    </template>
  </SettingsCard>
</template>

<script setup lang="ts">
import ContentSkeleton from '@/components/common/ContentSkeleton.vue'
import ProviderModelMappingEditor from '@/components/provider/ProviderModelMappingEditor.vue'
import type { ModelMappingRow } from '@/utils/modelMappingRules'
import { normalizeLegacyOpenAIExtra, normalizeOpenAICompactMode } from '@/utils/openaiLegacyConfiguration'
import OpenAICompactionToggle from '@/components/provider/OpenAICompactionToggle.vue'
import { computed, nextTick, onMounted, reactive, ref, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import type { OpenAIOAuthImportDefaults } from '@/api/admin/settings'
import {
  buildModelMappingObject,
  getPresetMappingsByPlatform,
  splitPersistedModelRestriction
} from '@/composables/useModelWhitelist'
import ModelWhitelistSelector from '@/components/provider/ModelWhitelistSelector.vue'
import CodexImageToolModeSelector from '@/components/provider/CodexImageToolModeSelector.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Collapse from '@/components/common/Collapse.vue'
import SettingRow from '@/components/common/settings/SettingRow.vue'
import SettingToggleRow from '@/components/common/settings/SettingToggleRow.vue'
import SettingsCard from '@/components/common/settings/SettingsCard.vue'
import SettingsSection from '@/components/common/settings/SettingsSection.vue'
import SettingsSubpanel from '@/components/common/settings/SettingsSubpanel.vue'
import TLSFingerprintFields from '@/components/provider/form/TLSFingerprintFields.vue'
import { useDirtyTracker } from '@/composables/useDirtyTracker'
import { useSettingsSaveTarget } from '@/composables/useSettingsSaveRegistry'
import { useAppStore } from '@/stores'
import {
  applyCodexImageToolMode,
  CODEX_IMAGE_GENERATION_BRIDGE_KEY,
  CODEX_IMAGE_GENERATION_POLICY_KEY,
  LEGACY_CODEX_IMAGE_GENERATION_BRIDGE_KEY,
  readCodexImageToolMode,
  type CodexImageToolMode
} from '@/utils/codexImageToolMode'
import {
  OPENAI_WS_MODE_OFF,
  isOpenAIWSModeEnabled,
  resolveOpenAIWSModeFromExtra,
  type OpenAIWSMode
} from '@/utils/openaiWsMode'
import type { OpenAICompactMode, OpenAIOAuthClientPolicy } from '@/types'

type AutoPauseDefault = 'unset' | 'true' | 'false'
type NumberInputValue = string | number
const { t } = useI18n()
const appStore = useAppStore()
const uid = useId()

const loading = ref(true)
const defaultAllowedModels = ref<string[]>([])
const defaultModelMappings = ref<ModelMappingRow[]>([])
const credentialsJson = ref('{}')
const extraJson = ref('{}')
const openaiPassthrough = ref(false)
const codexImageToolMode = ref<CodexImageToolMode>('inherit')
const openAIOAuthClientPolicy = ref<OpenAIOAuthClientPolicy>('any')
const codexCLIOnlyAllowClaudeCode = ref(false)
const wsMode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const compactMode = ref<OpenAICompactMode>('force_on')
const nativeCompactV2Mode = ref<OpenAICompactMode>('force_on')
const tlsFingerprintEnabled = ref(false)
const tlsFingerprintProfileId = ref<number | null>(null)
const tlsFingerprintProfiles = ref<{ id: number; name: string }[]>([])
const tlsFingerprintRouterId = ref<number | null>(null)
const tlsFingerprintRouters = ref<{ id: number; name: string }[]>([])
const autoPause5hThreshold = ref<NumberInputValue>('')
const autoPause7dThreshold = ref<NumberInputValue>('')
const autoPause5hDisabled = ref(false)
const autoPause7dDisabled = ref(false)
const form = reactive({
  notes: '',
  concurrency: '' as NumberInputValue,
  priority: '' as NumberInputValue,
  rateMultiplier: '' as NumberInputValue,
  expiresAt: '' as NumberInputValue,
  autoPauseOnExpired: 'unset' as AutoPauseDefault
})

const forbiddenCredentialFields = new Set([
  'access_token',
  'refresh_token',
  'id_token',
  'expires_at',
  'email',
  'client_id',
  'chatgpt_account_id',
  'chatgpt_user_id',
  'organization_id',
  'plan_type',
  'subscription_expires_at'
])

const forbiddenExtraFields = new Set(['email', 'name'])
const presetMappings = computed(() => getPresetMappingsByPlatform('openai'))
const structuredExtraKeys = [
  'openai_passthrough',
  'openai_oauth_passthrough',
  'openai_oauth_responses_websockets_v2_mode',
  'openai_oauth_responses_websockets_v2_enabled',
  'responses_websockets_v2_enabled',
  'openai_ws_enabled',
  'openai_oauth_client_policy',
  'codex_cli_only',
  'codex_cli_only_allowed_clients',
  CODEX_IMAGE_GENERATION_BRIDGE_KEY,
  LEGACY_CODEX_IMAGE_GENERATION_BRIDGE_KEY,
  CODEX_IMAGE_GENERATION_POLICY_KEY,
  'auto_pause_5h_threshold',
  'auto_pause_7d_threshold',
  'auto_pause_5h_disabled',
  'auto_pause_7d_disabled',
  'openai_native_compaction_v2_mode',
  'openai_compact_mode',
  'enable_tls_fingerprint',
  'tls_fingerprint_profile_id',
  'tls_fingerprint_router_id'
]

const autoPauseOnExpiredOptions = computed<SelectOption[]>(() => [
  { value: 'unset', label: t('admin.providers.openAIOAuthImportDefaultsUnset') },
  { value: 'true', label: t('common.yes') },
  { value: 'false', label: t('common.no') }
])

const wsModeOptions = computed<SelectOption[]>(() => [
  { value: 'off', label: t('admin.providers.openai.wsModeOff') },
  { value: 'ctx_pool', label: t('admin.providers.openai.wsModeCtxPool') },
  { value: 'passthrough', label: t('admin.providers.openai.wsModePassthrough') }
])

const openAIOAuthClientPolicyOptions = computed<SelectOption[]>(() => [
  { value: 'any', label: t('admin.providers.openai.clientPolicyAny') },
  { value: 'codex_only', label: t('admin.providers.openai.clientPolicyCodexOnly') },
  {
    value: 'tls_router_matched_only',
    label: t('admin.providers.openai.clientPolicyTLSRouterMatchedOnly')
  }
])

const tlsFingerprintProfileOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.providers.quotaControl.tlsFingerprint.defaultProfile') },
  ...(tlsFingerprintProfiles.value.length > 0
    ? [{ value: -1, label: t('admin.providers.quotaControl.tlsFingerprint.randomProfile') }]
    : []),
  ...tlsFingerprintProfiles.value.map((profile) => ({ value: profile.id, label: profile.name }))
])

const tlsFingerprintRouterOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.providers.quotaControl.tlsFingerprint.noRouter') },
  ...tlsFingerprintRouters.value.map((router) => ({ value: router.id, label: router.name }))
])

const numberToInput = (value: unknown): string => {
  return typeof value === 'number' && Number.isFinite(value) ? String(value) : ''
}

const stringifyJsonObject = (value: Record<string, unknown>): string => {
  return JSON.stringify(value, null, 2)
}

const parseJsonObject = (text: string, label: string): Record<string, unknown> => {
  const trimmed = text.trim()
  if (!trimmed) {
    return {}
  }

  const parsed = JSON.parse(trimmed)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(t('admin.providers.openAIOAuthImportDefaultsJsonObjectRequired', { label }))
  }
  return parsed as Record<string, unknown>
}

const parseOptionalNumber = (value: NumberInputValue, label: string, integer: boolean): number | undefined => {
  // number 输入框在运行时可能回传 number，先转成文本再校验。
  const trimmed = String(value).trim()
  if (!trimmed) {
    return undefined
  }

  const parsed = Number(trimmed)
  if (!Number.isFinite(parsed) || parsed < 0 || (integer && !Number.isInteger(parsed))) {
    throw new Error(t('admin.providers.openAIOAuthImportDefaultsInvalidNumber', { label }))
  }
  return parsed
}

const parseOptionalPercent = (value: NumberInputValue, label: string): number | undefined => {
  // 百分比阈值在界面按 0-100 展示，保存时再换算成 0-1 的比例。
  const parsed = parseOptionalNumber(value, label, false)
  if (parsed !== undefined && parsed > 100) {
    throw new Error(t('admin.providers.openAIOAuthImportDefaultsInvalidPercent', { label }))
  }
  return parsed
}

const rejectForbiddenFields = (
  fields: Record<string, unknown>,
  section: string,
  forbidden: Set<string>
): boolean => {
  for (const key of Object.keys(fields)) {
    if (forbidden.has(key.trim().toLowerCase())) {
      appStore.showError(t('admin.providers.openAIOAuthImportDefaultsForbiddenField', { section, field: key }))
      return false
    }
  }
  return true
}

const normalizeModelMappingObject = (value: unknown): Record<string, string> | undefined => {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, string>
    : undefined
}

const addDefaultPresetMapping = (from: string, to: string) => {
  if (defaultModelMappings.value.some((mapping) => mapping.from === from)) {
    appStore.showInfo(t('admin.providers.mappingExists', { model: from }))
    return
  }
  defaultModelMappings.value.push({ from, to })
}

const normalizeTLSFingerprintProfileId = (value: unknown): number | null => {
  // 导入模板保存为 JSON，profile_id 可能来自数字或数字字符串，这里统一归一化。
  if (typeof value === 'number' && Number.isInteger(value)) {
    return value === 0 ? null : value
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isInteger(parsed) && parsed !== 0 ? parsed : null
  }
  return null
}

const normalizeOpenAIOAuthClientPolicy = (
  policy: unknown,
  legacyCodexCLIOnly: unknown
): OpenAIOAuthClientPolicy => {
  // 新字段优先；没有新字段时继续读取旧的 codex_cli_only 开关。
  if (policy === 'codex_only' || policy === 'tls_router_matched_only' || policy === 'any') {
    return policy
  }
  return legacyCodexCLIOnly === true ? 'codex_only' : 'any'
}

const hydrate = (defaults: OpenAIOAuthImportDefaults) => {
  const provider = defaults.provider || {}
  form.notes = typeof provider.notes === 'string' ? provider.notes : ''
  form.concurrency = numberToInput(provider.concurrency)
  form.priority = numberToInput(provider.priority)
  form.rateMultiplier = numberToInput(provider.rate_multiplier)
  form.expiresAt = numberToInput(provider.expires_at)
  form.autoPauseOnExpired =
    typeof provider.auto_pause_on_expired === 'boolean'
      ? provider.auto_pause_on_expired ? 'true' : 'false'
      : 'unset'

  const credentials = { ...(defaults.credentials || {}) }
  const modelRestriction = splitPersistedModelRestriction(
    normalizeModelMappingObject(credentials.model_mapping),
    credentials.model_whitelist
  )
  defaultAllowedModels.value = modelRestriction.allowedModels
  defaultModelMappings.value = modelRestriction.modelMappings
  delete credentials.model_whitelist
  delete credentials.model_mapping
  credentialsJson.value = stringifyJsonObject(credentials)

  const extra = normalizeLegacyOpenAIExtra(defaults.extra || {})
  openaiPassthrough.value = extra.openai_passthrough === true || extra.openai_oauth_passthrough === true
  codexImageToolMode.value = readCodexImageToolMode(extra)
  openAIOAuthClientPolicy.value = normalizeOpenAIOAuthClientPolicy(
    extra.openai_oauth_client_policy,
    extra.codex_cli_only
  )
  codexCLIOnlyAllowClaudeCode.value =
    Array.isArray(extra.codex_cli_only_allowed_clients) &&
    extra.codex_cli_only_allowed_clients.includes('claude_code')
  autoPause5hThreshold.value =
    typeof extra.auto_pause_5h_threshold === 'number' && Number.isFinite(extra.auto_pause_5h_threshold)
      ? String(extra.auto_pause_5h_threshold * 100)
      : ''
  autoPause7dThreshold.value =
    typeof extra.auto_pause_7d_threshold === 'number' && Number.isFinite(extra.auto_pause_7d_threshold)
      ? String(extra.auto_pause_7d_threshold * 100)
      : ''
  autoPause5hDisabled.value = extra.auto_pause_5h_disabled === true
  autoPause7dDisabled.value = extra.auto_pause_7d_disabled === true
  wsMode.value = resolveOpenAIWSModeFromExtra(extra, {
    modeKey: 'openai_oauth_responses_websockets_v2_mode',
    enabledKey: 'openai_oauth_responses_websockets_v2_enabled',
    fallbackEnabledKeys: ['responses_websockets_v2_enabled', 'openai_ws_enabled'],
    defaultMode: OPENAI_WS_MODE_OFF
  })
  compactMode.value = normalizeOpenAICompactMode(extra.openai_compact_mode)
  nativeCompactV2Mode.value = normalizeOpenAICompactMode(extra.openai_native_compaction_v2_mode)
  tlsFingerprintEnabled.value = extra.enable_tls_fingerprint === true
  tlsFingerprintProfileId.value = tlsFingerprintEnabled.value
    ? normalizeTLSFingerprintProfileId(extra.tls_fingerprint_profile_id)
    : null
  tlsFingerprintRouterId.value = tlsFingerprintEnabled.value
    ? normalizeTLSFingerprintProfileId(extra.tls_fingerprint_router_id)
    : null
  for (const key of structuredExtraKeys) {
    delete extra[key]
  }
  extraJson.value = stringifyJsonObject(extra)
}

const load = async () => {
  loading.value = true
  try {
    const defaults = await adminAPI.settings.getOpenAIOAuthImportDefaults()
    hydrate(defaults)
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.providers.openAIOAuthImportDefaultsLoadFailed'))
  } finally {
    loading.value = false
  }
  // 等子组件规整完初始值再记录快照。
  await nextTick()
  markClean()
}

const loadTLSFingerprintProfiles = async () => {
  try {
    const profiles = await adminAPI.tlsFingerprintProfiles.list()
    tlsFingerprintProfiles.value = profiles.map((profile) => ({ id: profile.id, name: profile.name }))
  } catch {
    // 模板保存不依赖 profile 列表，加载失败时保留内置默认选项。
    tlsFingerprintProfiles.value = []
  }
}

const loadTLSFingerprintRouters = async () => {
  try {
    const routers = await adminAPI.tlsFingerprintRouters.list()
    tlsFingerprintRouters.value = routers.map((router) => ({ id: router.id, name: router.name }))
  } catch {
    // 路由器列表加载失败时清空可选项，其它默认配置仍可保存。
    tlsFingerprintRouters.value = []
  }
}

const buildProviderDefaults = (): OpenAIOAuthImportDefaults['provider'] => {
  const provider: NonNullable<OpenAIOAuthImportDefaults['provider']> = {}
  if (form.notes.trim() !== '') {
    provider.notes = form.notes
  }

  const concurrency = parseOptionalNumber(form.concurrency, t('admin.providers.concurrency'), true)
  if (concurrency !== undefined) provider.concurrency = concurrency

  const priority = parseOptionalNumber(form.priority, t('admin.providers.priority'), true)
  if (priority !== undefined) provider.priority = priority

  const rateMultiplier = parseOptionalNumber(form.rateMultiplier, t('admin.providers.billingRateMultiplier'), false)
  if (rateMultiplier !== undefined) provider.rate_multiplier = rateMultiplier

  const expiresAt = parseOptionalNumber(form.expiresAt, t('admin.providers.expiresAt'), true)
  if (expiresAt !== undefined) provider.expires_at = expiresAt

  if (form.autoPauseOnExpired !== 'unset') {
    provider.auto_pause_on_expired = form.autoPauseOnExpired === 'true'
  }

  return Object.keys(provider).length > 0 ? provider : undefined
}

// save 提交默认值，返回是否保存成功；由吸底保存条统一调用。
const save = async (): Promise<boolean> => {
  try {
    const credentials = parseJsonObject(
      credentialsJson.value,
      t('admin.providers.openAIOAuthImportDefaultsCredentialsJson')
    )
    const extra = normalizeLegacyOpenAIExtra(parseJsonObject(extraJson.value, t('admin.providers.openAIOAuthImportDefaultsExtraJson')))

    delete credentials.model_whitelist
    delete credentials.model_mapping
    for (const key of structuredExtraKeys) {
      delete extra[key]
    }
    applyCodexImageToolMode(extra, codexImageToolMode.value)

    if (!rejectForbiddenFields(credentials, 'credentials', forbiddenCredentialFields)) return false
    if (!rejectForbiddenFields(extra, 'extra', forbiddenExtraFields)) return false

    if (openaiPassthrough.value) {
      extra.openai_passthrough = true
    }
    if (wsMode.value !== OPENAI_WS_MODE_OFF) {
      extra.openai_oauth_responses_websockets_v2_mode = wsMode.value
      extra.openai_oauth_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(wsMode.value)
    }
    extra.openai_oauth_client_policy = openAIOAuthClientPolicy.value
    // 继续写旧字段，方便旧版本服务端或旧提供商逻辑读取；非 Codex 模式显式清 false。
    extra.codex_cli_only = openAIOAuthClientPolicy.value === 'codex_only'
    if (openAIOAuthClientPolicy.value === 'codex_only' && codexCLIOnlyAllowClaudeCode.value) {
      extra.codex_cli_only_allowed_clients = ['claude_code']
    }
    const autoPause5hPercent = parseOptionalPercent(
      autoPause5hThreshold.value,
      t('admin.providers.autoPause5hThreshold')
    )
    const autoPause7dPercent = parseOptionalPercent(
      autoPause7dThreshold.value,
      t('admin.providers.autoPause7dThreshold')
    )
    if (autoPause5hPercent !== undefined && autoPause5hPercent > 0) {
      extra.auto_pause_5h_threshold = autoPause5hPercent / 100
    }
    if (autoPause7dPercent !== undefined && autoPause7dPercent > 0) {
      extra.auto_pause_7d_threshold = autoPause7dPercent / 100
    }
    if (autoPause5hDisabled.value) {
      extra.auto_pause_5h_disabled = true
    }
    if (autoPause7dDisabled.value) {
      extra.auto_pause_7d_disabled = true
    }
    extra.openai_compact_mode = compactMode.value
    extra.openai_native_compaction_v2_mode = nativeCompactV2Mode.value
    if (tlsFingerprintEnabled.value) {
      extra.enable_tls_fingerprint = true
      if (tlsFingerprintProfileId.value !== null) {
        extra.tls_fingerprint_profile_id = tlsFingerprintProfileId.value
      }
      if (tlsFingerprintRouterId.value !== null) {
        extra.tls_fingerprint_router_id = tlsFingerprintRouterId.value
      }
    }

    const modelMapping = buildModelMappingObject('mapping', [], defaultModelMappings.value)
    const updatedCredentials: Record<string, unknown> = {
      ...credentials,
      model_whitelist: [...defaultAllowedModels.value]
    }
    if (modelMapping) {
      updatedCredentials.model_mapping = modelMapping
    }

    const updated = await adminAPI.settings.updateOpenAIOAuthImportDefaults({
      provider: buildProviderDefaults(),
      credentials: updatedCredentials,
      extra: Object.keys(extra).length > 0 ? extra : undefined
    })
    hydrate(updated)
    await nextTick()
    markClean()
    return true
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.providers.openAIOAuthImportDefaultsSaveFailed'))
    return false
  }
}

// 吸底保存条比较这些可编辑状态和上次加载或保存时的快照。
const { dirty, markClean } = useDirtyTracker({
  defaults: () => ({
    form,
    defaultAllowedModels: defaultAllowedModels.value,
    defaultModelMappings: defaultModelMappings.value,
    credentialsJson: credentialsJson.value,
    extraJson: extraJson.value,
    openaiPassthrough: openaiPassthrough.value,
    codexImageToolMode: codexImageToolMode.value,
    openAIOAuthClientPolicy: openAIOAuthClientPolicy.value,
    codexCLIOnlyAllowClaudeCode: codexCLIOnlyAllowClaudeCode.value,
    wsMode: wsMode.value,
    compactMode: compactMode.value,
    nativeCompactV2Mode: nativeCompactV2Mode.value,
    tlsFingerprintEnabled: tlsFingerprintEnabled.value,
    tlsFingerprintProfileId: tlsFingerprintProfileId.value,
    tlsFingerprintRouterId: tlsFingerprintRouterId.value,
    autoPause5hThreshold: autoPause5hThreshold.value,
    autoPause7dThreshold: autoPause7dThreshold.value,
    autoPause5hDisabled: autoPause5hDisabled.value,
    autoPause7dDisabled: autoPause7dDisabled.value,
  }),
})
useSettingsSaveTarget('openaiOAuthImportDefaults', { dirty, save })

onMounted(() => {
  void load()
  void loadTLSFingerprintProfiles()
  void loadTLSFingerprintRouters()
})
</script>
